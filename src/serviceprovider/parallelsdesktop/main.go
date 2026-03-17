package parallelsdesktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"io/ioutil"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/processlauncher"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
	"github.com/Parallels/prl-devops-service/serviceprovider/git"
	"github.com/Parallels/prl-devops-service/serviceprovider/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider/packer"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/google/uuid"

	"github.com/cjlapao/common-go/helper"
)

var (
	globalParallelsService    *ParallelsService
	eventsChannel             = make(chan models.ParallelsServiceEvent, 1000)
	configChangeTimers        = make(map[string]*time.Timer)
	configChangeTimersMutex   = &sync.Mutex{}
	configChangeCooldown      = make(map[string]*time.Timer)
	configChangeCooldownMutex = &sync.Mutex{}
	toolsStateTimers          = make(map[string]*time.Timer)
	toolsStateTimersMutex     = &sync.Mutex{}
)

const cooldownDelay = 2 * time.Second
const eventWorkerTicker = 1 * time.Second

// maxConcurrentSSH is the maximum number of simultaneous prlctl exec SSH logon
// attempts. macOS enforces a hard cap on simultaneous logons, so we keep this low.
const maxConcurrentSSH = 2

type ParallelsService struct {
	ctx              basecontext.ApiContext
	eventsProcessing bool
	sync.RWMutex
	cachedLocalVms []models.ParallelsVM
	// syncMu protects our Debounce and Cooldown maps
	syncMu sync.Mutex
	// pending holds VMs waiting to be synced (inherent deduplication)
	pending map[string]struct{}
	// inFlight tracks VMs currently running prlctl (prevents overlapping commands)
	inFlight map[string]struct{}
	// cooldown prevents the hypervisor echo loop
	cooldown map[string]time.Time
	// ipLastFetched tracks the last time we ran prlctl exec to fetch a macOS VM IP.
	// This prevents concurrent SSH logon floods when many cache refreshes and events
	// fire simultaneously (macOS enforces a hard cap on simultaneous logon attempts).
	ipLastFetched map[string]time.Time
	// sshSem is a semaphore that caps the number of simultaneous prlctl exec SSH
	// operations (waitForVMSSHReady probes + IP fetches) across all VMs.
	sshSem chan struct{}
	// fastStateUpdates is used for the fast state updates that we do not need to the
	// prlctl command and can update the database
	fastStateUpdates map[string]time.Time
	macVMsRunning    []string
	macVMsRunningMu  sync.RWMutex
	executable       string
	serverExecutable string
	Info             *models.ParallelsDesktopInfo
	Users            []*models.ParallelsDesktopUser
	isLicensed       bool
	installed        bool
	version          string
	build            string
	dependencies     []interfaces.Service
	cancelFunc       context.CancelFunc
	listenerCtx      context.Context
	processLauncher  processlauncher.ProcessLauncher
	databaseService  *data.JsonDatabase
}

func Get(ctx basecontext.ApiContext) *ParallelsService {
	if globalParallelsService != nil {
		return globalParallelsService
	}
	return New(ctx)
}

func New(ctx basecontext.ApiContext) *ParallelsService {
	// Initialize the context BEFORE we put it in the struct to avoid potential racing conditions
	listenerCtx, cancelFunc := context.WithCancel(context.Background())

	globalParallelsService = &ParallelsService{
		eventsProcessing: false,
		ctx:              ctx,
		processLauncher:  &processlauncher.RealProcessLauncher{},
		databaseService:  nil, // Will be injected later

		// Initialize maps for the debounce and cooldown logic
		// this will allow us to deduplicate events and prevent the echo loop
		// also will prevent multiple prlctl commands from being executed at the same time
		pending:          make(map[string]struct{}),
		inFlight:         make(map[string]struct{}),
		cooldown:         make(map[string]time.Time),
		ipLastFetched:    make(map[string]time.Time),
		fastStateUpdates: make(map[string]time.Time),
		sshSem:           make(chan struct{}, maxConcurrentSSH),

		// registered to the event listener to allow us to cancel it when needed
		listenerCtx: listenerCtx,
		cancelFunc:  cancelFunc,
	}
	if globalParallelsService.FindPath() == "" {
		ctx.LogWarnf("[ParallelsDesktop] [main] Running without support for Parallels Desktop")
	} else {
		globalParallelsService.installed = true
	}

	globalParallelsService.SetDependencies([]interfaces.Service{})

	cfg := config.Get()
	if cfg.IsApi() || cfg.IsHost() {
		ctx.LogInfof("[ParallelsDesktop] [main] Starting Parallels Desktop service")
		globalParallelsService.refreshCache(ctx)
		ctx.LogInfof("[ParallelsDesktop] [main] Starting Parallels Desktop service debounce worker")
		go globalParallelsService.startDebounceWorker()
		ctx.LogInfof("[ParallelsDesktop] [main] Starting Parallels Desktop service event listener")
		globalParallelsService.listenToParallelsEvents(ctx)
	}
	if cfg.IsCacheRefreshEnabled() {
		ctx.LogInfof("[ParallelsDesktop] [Cache] Auto cache refresh is enabled, starting the auto cache refresh routine")
		globalParallelsService.startAutoCacheRefresh(ctx)
	} else {
		ctx.LogInfof("[ParallelsDesktop] [Cache] Auto cache refresh is disabled, not starting the auto cache refresh routine")
	}
	return globalParallelsService
}

func (s *ParallelsService) Name() string {
	return "parallels_desktop"
}

func (s *ParallelsService) SetDatabaseService(dbService *data.JsonDatabase) {
	s.databaseService = dbService
}

func (s *ParallelsService) FindPath() string {
	s.ctx.LogInfof("[ParallelsDesktop] [main] Getting prlctl executable")
	cmd := helpers.Command{
		Command: "which",
		Args:    []string{"prlctl"},
	}
	out, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		s.ctx.LogWarnf("[ParallelsDesktop] [main] Parallels Desktop CLI executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.executable = path
		s.serverExecutable = strings.ReplaceAll(path, "prlctl", "prlsrvctl")
		s.ctx.LogInfof("[ParallelsDesktop] [main] Parallels Desktop CLI found at: %s", s.executable)
	} else {
		if _, err := os.Stat("/usr/bin/prlctl"); err == nil {
			s.executable = "/usr/bin/prlctl"
			s.serverExecutable = "/usr/bin/prlsrvctl"
			if err := os.Setenv("PATH", os.Getenv("PATH")+":/usr/bin"); err != nil {
				s.ctx.LogWarnf("[ParallelsDesktop] [main] Error setting PATH environment variable: %v", err)
			}
		} else if _, err := os.Stat("/usr/local/bin/prlctl"); err == nil {
			s.executable = "/usr/local/bin/prlctl"
			s.serverExecutable = "/usr/local/bin/prlsrvctl"
			if err := os.Setenv("PATH", os.Getenv("PATH")+":/usr/local/bin"); err != nil {
				s.ctx.LogWarnf("[ParallelsDesktop] [main] Error setting PATH environment variable: %v", err)
			}
		} else {
			s.ctx.LogWarnf("[ParallelsDesktop] [main] Parallels Desktop CLI executable not found, trying to install it")
			return s.executable
		}

		s.ctx.LogInfof("[ParallelsDesktop] [main] Parallels Desktop CLI found at: %s", s.executable)
	}

	return s.executable
}

func (s *ParallelsService) Version() string {
	if s.version != "" {
		return s.version
	}

	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"--version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "unknown"
	}

	v := strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(stdout, "prlctl version ", "")), "\n", "")
	vParts := strings.Split(v, " ")
	if len(vParts) > 0 {
		s.version = vParts[0]
		s.build = strings.ReplaceAll(vParts[1], "(", "")
		s.build = strings.ReplaceAll(s.build, ")", "")
	} else {
		s.version = v
	}

	if s.build == "" {
		return s.version
	}

	return fmt.Sprintf("%s.%s", s.version, s.build)
}

func (s *ParallelsService) Install(asUser, version string, flags map[string]string) error {
	if s.installed {
		s.ctx.LogInfof("[ParallelsDesktop] [main] %s already installed", s.Name())
	} else {

		// Installing service dependency
		if s.dependencies != nil {
			for _, dependency := range s.dependencies {
				if dependency == nil {
					return errors.New("Dependency is nil")
				}
				s.ctx.LogInfof("[ParallelsDesktop] [main] Installing dependency %s for %s", dependency.Name(), s.Name())
				if err := dependency.Install(asUser, "latest", flags); err != nil {
					return err
				}
			}
		}

		// TODO need to verify if brew is working fine
		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew"},
			}
		}

		if version == "" || version == "latest" {
			cmd.Args = append(cmd.Args, "install", "parallels")
		} else {
			cmd.Args = append(cmd.Args, "install", "parallels@"+version)
		}

		s.ctx.LogInfof("[ParallelsDesktop] [main] Installing %s with command: %v", s.Name(), cmd.String())
		_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
		if err != nil {
			return err
		}
		s.installed = true
	}

	license := ""
	username := ""
	password := ""

	for flag, value := range flags {
		switch flag {
		case "license":
			license = value
		case "my_account_username":
			username = value
		case "my_account_password":
			password = value
		}
	}

	if license != "" {
		s.ctx.LogInfof("[ParallelsDesktop] [main] Activating Parallels Desktop with license %s", license)
		if err := s.InstallLicense(license, username, password); err != nil {
			return err
		}

		if _, err := s.GetInfo(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ParallelsService) InstallFromDmg(asUser, version string, flags map[string]string) error {
	if s.installed {
		s.ctx.LogInfof("[ParallelsDesktop] [main] %s already installed", s.Name())
	} else {
		if version == "" || version == "latest" {
			// fallback to a known default if latest is requested and we can't fetch the xml easily
			// ideally we'd parse the livecheck xml, but for this iteration we'll use a hardcoded recent version or fail
			version = "20.1.0-55732"
			s.ctx.LogWarnf("[ParallelsDesktop] [main] Version not specified, defaulting to %s", version)
		}

		vParts := strings.Split(version, ".")
		if len(vParts) < 1 {
			return errors.New("Invalid version format")
		}
		majorVersion := vParts[0]

		dmgUrl := fmt.Sprintf("https://download.parallels.com/desktop/v%s/%s/ParallelsDesktop-%s.dmg", majorVersion, version, version)

		s.ctx.LogInfof("[ParallelsDesktop] [main] Downloading Parallels Desktop %s from %s", version, dmgUrl)

		// Create a temporary file for the dmg
		tmpFile, err := ioutil.TempFile("", "parallels-*.dmg")
		if err != nil {
			return errors.New("Failed to create temporary file for dmg: " + err.Error())
		}
		defer os.Remove(tmpFile.Name())

		resp, err := http.Get(dmgUrl)
		if err != nil {
			return errors.New("Failed to download Parallels Desktop: " + err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.New(fmt.Sprintf("Failed to download Parallels Desktop, status code: %d", resp.StatusCode))
		}

		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			return errors.New("Failed to save downloaded dmg: " + err.Error())
		}
		tmpFile.Close()

		s.ctx.LogInfof("[ParallelsDesktop] [main] Mounting the downloaded dmg: %s", tmpFile.Name())
		mountCmd := helpers.Command{
			Command: "hdiutil",
			Args:    []string{"attach", "-nobrowse", tmpFile.Name()},
		}
		if asUser != "" {
			mountCmd.Command = "sudo"
			mountCmd.Args = append([]string{"-u", asUser, "hdiutil", "attach", "-nobrowse", tmpFile.Name()})
		}
		out, _, _, err := helpers.ExecuteWithOutput(s.ctx.Context(), mountCmd, helpers.ExecutionTimeout)
		if err != nil {
			return errors.New("Failed to mount dmg: " + err.Error() + ": " + out)
		}

		// parse the output to find the mount point
		mountPoint := ""
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "/Volumes/Parallels Desktop") {
				parts := strings.Split(line, "\t")
				if len(parts) > 2 {
					mountPoint = strings.TrimSpace(parts[2])
				} else {
					// fallback heuristic
					mountPoint = "/Volumes/Parallels Desktop"
				}
				break
			}
		}

		if mountPoint == "" {
			// default fallback
			mountPoint = "/Volumes/Parallels Desktop"
		}

		defer func() {
			s.ctx.LogInfof("Unmounting dmg at %s", mountPoint)
			unmountCmd := helpers.Command{
				Command: "hdiutil",
				Args:    []string{"detach", mountPoint},
			}
			if asUser != "" {
				unmountCmd.Command = "sudo"
				unmountCmd.Args = append([]string{"-u", asUser, "hdiutil", "detach", mountPoint})
			}
			helpers.ExecuteWithNoOutput(s.ctx.Context(), unmountCmd, helpers.ExecutionTimeout)
		}()

		appSource := filepath.Join(mountPoint, "Parallels Desktop.app")
		appDest := "/Applications/Parallels Desktop.app"

		s.ctx.LogInfof("Copying %s to %s", appSource, appDest)
		copyCmd := helpers.Command{
			Command: "sudo",
			Args:    []string{"cp", "-R", appSource, "/Applications/"},
		}
		if asUser != "" {
			copyCmd.Args = append([]string{"-u", asUser, "cp", "-R", appSource, "/Applications/"})
		}

		_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), copyCmd, helpers.ExecutionTimeout)
		if err != nil {
			return errors.New("Failed to copy Parallels Desktop.app to /Applications: " + err.Error())
		}

		s.ctx.LogInfof("Adjusting permissions for Parallels Desktop")
		helpers.ExecuteWithNoOutput(s.ctx.Context(), helpers.Command{Command: "sudo", Args: []string{"chflags", "nohidden", appDest}}, helpers.ExecutionTimeout)
		helpers.ExecuteWithNoOutput(s.ctx.Context(), helpers.Command{Command: "sudo", Args: []string{"xattr", "-d", "com.apple.FinderInfo", appDest}}, helpers.ExecutionTimeout)

		initToolPath := filepath.Join(appDest, "Contents", "MacOS", "inittool")
		s.ctx.LogInfof("Running Parallels inittool")
		initCmd := helpers.Command{
			Command: "sudo",
			Args:    []string{initToolPath, "init"},
		}
		_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), initCmd, helpers.ExecutionTimeout)
		if err != nil {
			return errors.New("Failed to run Parallels inittool: " + err.Error())
		}

		s.installed = true
		s.executable = "/usr/local/bin/prlctl"
		s.serverExecutable = "/usr/local/bin/prlsrvctl"
	}

	license := ""
	username := ""
	password := ""

	for flag, value := range flags {
		switch flag {
		case "license":
			license = value
		case "my_account_username":
			username = value
		case "my_account_password":
			password = value
		}
	}

	if license != "" {
		s.ctx.LogInfof("Activating Parallels Desktop with license %s", license)
		if err := s.InstallLicense(license, username, password); err != nil {
			return err
		}

		if _, err := s.GetInfo(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ParallelsService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		s.ctx.LogInfof("Uninstalling %s", s.Name())

		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
				Args:    []string{"uninstall", "parallels"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew", "uninstall", "parallels"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
		if err != nil {
			return err
		}
	}

	if err := s.DeactivateLicense(); err != nil {
		return err
	}

	if uninstallDependencies {
		// Uninstall service dependency
		if s.dependencies != nil {
			for _, dependency := range s.dependencies {
				if dependency == nil {
					continue
				}
				s.ctx.LogInfof("Uninstalling dependency %s for %s", dependency.Name(), s.Name())
				if err := dependency.Uninstall(asUser, uninstallDependencies); err != nil {
					return err
				}
			}
		}
	}

	s.installed = false
	return nil
}

func (s *ParallelsService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *ParallelsService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *ParallelsService) Installed() bool {
	return s.installed && s.executable != "" && s.serverExecutable != ""
}

func (s *ParallelsService) IsLicensed() bool {
	return s.isLicensed
}

func (s *ParallelsService) updateVmInCache(ctx basecontext.ApiContext, newVm *models.ParallelsVM, syncStartTime time.Time) {
	s.Lock()
	defer s.Unlock()
	found := false
	changeType := MeaningfulChange // Default to full broadcast for brand new VMs

	for i, cachedVm := range s.cachedLocalVms {
		if cachedVm.ID == newVm.ID {
			// // RACE CONDITION PROTECTION:
			// // If the UI just set the state to "starting" (Fast Path), but our slow prlctl list
			// // finally returned and says it's "stopped", we KEEP the "starting" state because it's newer.
			// isCachedTransitional := cachedVm.State == "starting" || cachedVm.State == "stopping" ||
			// 	cachedVm.State == "resuming" || cachedVm.State == "suspending"

			// isNewStatic := newVm.State == "stopped" || newVm.State == "running" ||
			// 	newVm.State == "paused" || newVm.State == "suspended"

			// if isCachedTransitional && isNewStatic {
			// 	ctx.LogInfof("[ParallelsDesktop] [Event] Preserving newer transitional state '%s' over older static state '%s'", cachedVm.State, newVm.State)
			// 	newVm.State = cachedVm.State // Preserve the intermediate state!
			// }

			// the deterministic timestamp shield to avoid loosing some of the fast path updates
			if lastFastUpdate, exists := s.fastStateUpdates[newVm.ID]; exists {
				// If the Fast Path updated the state AFTER we started our slow prlctl command...
				// The Fast Path is newer and more accurate. Reject the slow path's state!
				if lastFastUpdate.After(syncStartTime) {
					ctx.LogInfof("[ParallelsDesktop] [Cache] Shielding newer fast-path state '%s' from older slow-path state '%s'", cachedVm.State, newVm.State)
					newVm.State = cachedVm.State // Preserve the fast path state!
				}
			}

			// DIFF ENGINE: Classify the change
			changeType = s.evaluateVmChanges(cachedVm, *newVm)

			s.cachedLocalVms[i] = *newVm
			found = true
			break
		}
	}

	if !found {
		s.cachedLocalVms = append(s.cachedLocalVms, *newVm)
	}

	go func() {
		ee := eventemitter.Get()
		if ee == nil || !ee.IsRunning() {
			return
		}

		switch changeType {
		case OnlyUptimeChanged:
			ctx.LogDebugf("[ParallelsDesktop] [Event] VM %s uptime ticked. Broadcasting lightweight event.", newVm.ID)
			_ = ee.BroadcastMessage(models.NewEventMessage(constants.EventTypePDFM, "VM_UPTIME_CHANGED", models.VmUptimeChanged{
				VmID:   newVm.ID,
				Uptime: newVm.Uptime,
			}))
		case MeaningfulChange:
			ctx.LogInfof("[ParallelsDesktop] [Event] VM %s had meaningful changes. Broadcasting full update.", newVm.ID)
			_ = ee.BroadcastMessage(models.NewEventMessage(constants.EventTypePDFM, "VM_UPDATED", models.VmUpdated{
				VmID:  newVm.ID,
				NewVm: *newVm,
			}))
		}
	}()
}

func (s *ParallelsService) InitSnapshotTreeInDB(ctx basecontext.ApiContext) {
	s.RLock()
	cachedVMs := s.cachedLocalVms
	s.RUnlock()

	for _, vm := range cachedVMs {
		snapshotResponse, err := s.listVMSnapshots(ctx, vm.ID)
		if err != nil {
			ctx.LogErrorf("[parallelsdesktop][snapshots] Failed to get snapshots for VM %s: %v", vm.ID, err)
			continue
		}
		if s.databaseService == nil {
			ctx.LogErrorf("[parallelsdesktop][snapshots] Database service not available")
			return
		}

		dtoSnapshots := mappers.VMSnapshotsApiToDto(snapshotResponse.Snapshots)

		s.databaseService.SetListVMSnapshotsByVMId(vm.ID, data_models.VMSnapshots{
			VMId:       vm.ID,
			VMSnapshot: dtoSnapshots,
		})
	}
}

func (s *ParallelsService) GetVMSnapshotsFromDB(ctx basecontext.ApiContext, vmID string) (*models.ListVMSnapshotResponse, error) {
	if s.databaseService == nil {
		ctx.LogErrorf("[parallelsdesktop][snapshots] Database service not available")
		return nil, nil
	}

	dbSnaps, err := s.databaseService.GetListVMSnapshotsByVMId(vmID)
	if err != nil {
		return nil, err
	}
	mappedSnaps := mappers.VMSnapshotsDtoToApi(dbSnaps)
	resp := &models.ListVMSnapshotResponse{
		Snapshots: mappedSnaps,
	}
	return resp, nil
}

// waitForVMSSHReady probes a macOS VM's SSH readiness by executing a trivial command
// via prlctl exec. It retries every 3 seconds for up to 30 seconds total.
// Returns true as soon as the command succeeds, false if the timeout is exceeded.
// This is safe to call from any goroutine and has no side-effects on the cache.
func (s *ParallelsService) waitForVMSSHReady(ctx basecontext.ApiContext, vmID, username string) bool {
	const (
		retryInterval = 5 * time.Second
		totalTimeout  = 30 * time.Second
	)
	ctx.LogInfof("[ParallelsDesktop] [SSH] Waiting for VM %s SSH to become ready (timeout: %v)", vmID, totalTimeout)
	deadline := time.Now().Add(totalTimeout)
	for time.Now().Before(deadline) {
		s.sshSem <- struct{}{}
		cmd := helpers.Command{
			Command: s.executable,
			Args:    []string{"exec", vmID, "echo", "hello"},
		}.AsUser(username)
		_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, retryInterval)
		<-s.sshSem
		if err == nil {
			ctx.LogInfof("[ParallelsDesktop] [SSH] VM %s is ready for SSH commands", vmID)
			return true
		}
		ctx.LogDebugf("[ParallelsDesktop] [SSH] VM %s not ready yet (%v), retrying in %v", vmID, err, retryInterval)
		time.Sleep(retryInterval)
	}
	ctx.LogWarnf("[ParallelsDesktop] [SSH] VM %s did not become ready within %v", vmID, totalTimeout)
	return false
}

// updateVMIPInCache resolves the internal IP of a macOS VM via `prlctl exec ipconfig getifaddr en0`
// and updates the entry in the in-memory cache. It must be called in a goroutine as it blocks
// until SSH is ready (up to 30s). Rate-limited to once per 30s per VM ID.
func (s *ParallelsService) updateVMIPInCache(ctx basecontext.ApiContext, vmID string) {
	// Snapshot the VM's username while holding the read lock, so we can release it
	// before any slow SSH operations.
	s.RLock()
	var username string
	var isMacOS bool
	for _, vm := range s.cachedLocalVms {
		if vm.ID == vmID {
			username = vm.User
			isMacOS = vm.OS == "macosx" || strings.Contains(strings.ToLower(vm.Name), "mac")
			break
		}
	}
	s.RUnlock()

	if username == "" {
		ctx.LogWarnf("[ParallelsDesktop] [IP] VM %s not found in cache, skipping IP update", vmID)
		return
	}

	if !isMacOS {
		// Only macOS VMs need the prlctl exec SSH path; other platforms expose their IP
		// via prlctl list network info directly.
		return
	}

	// Rate-limit: skip if we fetched this VM's IP less than 30 seconds ago.
	const macOSIPFetchCooldown = 30 * time.Second
	s.syncMu.Lock()
	if last, ok := s.ipLastFetched[vmID]; ok && time.Since(last) < macOSIPFetchCooldown {
		s.syncMu.Unlock()
		ctx.LogDebugf("[ParallelsDesktop] [IP] Skipping IP fetch for VM %s (in cooldown, last fetched %v ago)", vmID, time.Since(last).Round(time.Second))
		return
	}
	s.ipLastFetched[vmID] = time.Now()
	s.syncMu.Unlock()

	// Wait until the VM's SSH accepts commands before attempting to fetch the IP.
	if !s.waitForVMSSHReady(ctx, vmID, username) {
		ctx.LogWarnf("[ParallelsDesktop] [IP] Giving up on IP fetch for VM %s: SSH not ready within timeout", vmID)
		return
	}

	// Fetch the internal IP via prlctl exec, holding the semaphore to avoid
	// flooding macOS with simultaneous SSH logon attempts.
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"exec", vmID, "ipconfig", "getifaddr", "en0"},
	}.AsUser(username)
	ctx.LogDebugf("[ParallelsDesktop] [IP] Fetching internal IP for VM %s: %s", vmID, cmd.String())
	s.sshSem <- struct{}{}
	out, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	<-s.sshSem
	if err != nil {
		ctx.LogErrorf("[ParallelsDesktop] [IP] Failed to get internal IP for VM %s: %v", vmID, err)
		return
	}
	ip := strings.TrimSpace(out)
	if ip == "" {
		ctx.LogDebugf("[ParallelsDesktop] [IP] Empty IP result for VM %s", vmID)
		return
	}
	ctx.LogInfof("[ParallelsDesktop] [IP] Resolved internal IP for VM %s: %s", vmID, ip)

	// Update the cache and broadcast the change.
	s.Lock()
	var updatedVm *models.ParallelsVM
	for i, vm := range s.cachedLocalVms {
		if vm.ID == vmID {
			s.cachedLocalVms[i].InternalIpAddress = ip
			// Also update the matching IPv4 entry in NetworkInformation so it stays consistent.
			for j, addr := range s.cachedLocalVms[i].NetworkInformation.IPAddresses {
				if strings.ToLower(addr.Type) == "ipv4" {
					s.cachedLocalVms[i].NetworkInformation.IPAddresses[j].IP = ip
					break // found and updated the IPv4 entry — done
				}
			}
			copy := s.cachedLocalVms[i]
			updatedVm = &copy
			break
		}
	}
	s.Unlock()

	if updatedVm == nil {
		return
	}
	go func() {
		if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypePDFM, "VM_UPDATED", models.VmUpdated{
				VmID:  vmID,
				NewVm: *updatedVm,
			})
			if err := ee.BroadcastMessage(msg); err != nil {
				ctx.LogErrorf("[ParallelsDesktop] [IP] Error broadcasting VM_UPDATED: %v", err)
			}
		}
	}()
}

func (s *ParallelsService) getFilteredUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	users, err := system.Get().GetSystemUsers(ctx)
	if err != nil {
		return nil, err
	}

	currentUser := "root"
	if user, err := system.Get().GetCurrentUser(ctx); err == nil {
		currentUser = user
	}
	if currentUser != "root" {
		newAllUsers := make([]models.SystemUser, 0)
		for _, user := range users {
			if strings.EqualFold(user.Username, currentUser) {
				newAllUsers = append(newAllUsers, user)
				break
			}
		}

		users = newAllUsers
	}

	return users, nil
}

func (s *ParallelsService) getVmInMachine(ctx basecontext.ApiContext, vmId string) (*models.ParallelsVM, error) {
	// FAST PATH: Check the cache first to find out who owns this VM
	s.RLock()
	var knownOwner string
	for _, cachedVm := range s.cachedLocalVms {
		if cachedVm.ID == vmId {
			knownOwner = cachedVm.User
			break
		}
	}
	s.RUnlock()

	// If we know the owner, target them directly! (Saves doing a prlctl list for every user)
	if knownOwner != "" {
		userMachines, err := s.getUserVm(ctx, knownOwner, vmId)
		if err == nil && len(userMachines) == 1 {
			return &userMachines[0], nil
		}
	}

	// SLOW PATH: Fallback to checking all users on the host
	users, err := s.getFilteredUsers(ctx)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		userMachines, err := s.getUserVm(ctx, user.Username, vmId)
		if err == nil && len(userMachines) == 1 {
			return &userMachines[0], nil
		}
	}
	return nil, errors.New("VM not found")
}

func (s *ParallelsService) refreshCache(ctx basecontext.ApiContext) {
	ctx.LogInfof("Refreshing Parallels VMs cache")
	vms, err := s.getVmsInMachineForCurrentUser(ctx)
	s.Lock()
	if err != nil {
		ctx.LogErrorf("Error refreshing Parallels VMs cache: %v", err)
		s.cachedLocalVms = []models.ParallelsVM{} // Clear cache on error for consistency
	} else {
		s.cachedLocalVms = vms
		go s.syncMacVmRunningStatus(ctx)
	}
	s.Unlock()

	// For every running macOS VM that's already up when we first load the cache,
	// kick off an async IP resolution so InternalIpAddress is populated quickly.
	if err == nil {
		stagger := 0
		for _, vm := range vms {
			if vm.State == "running" &&
				(vm.OS == "macosx" || strings.Contains(strings.ToLower(vm.Name), "mac")) {
				ctx.LogInfof("[ParallelsDesktop] [IP] Scheduling IP resolution for running macOS VM %s (%s)", vm.ID, vm.Name)
				vmID := vm.ID
				delay := time.Duration(stagger) * 3 * time.Second
				stagger++
				go func() {
					if delay > 0 {
						time.Sleep(delay)
					}
					s.updateVMIPInCache(ctx, vmID)
				}()
			}
		}
	}
}

func (s *ParallelsService) getUserVm(ctx basecontext.ApiContext, username string, vmId string) ([]models.ParallelsVM, error) {
	ctx.LogInfof("[ParallelsDesktop] [main] Getting VMs for user: %s", username)
	var syncVms []models.ParallelsVM

	// fetch the base data
	if vmId == "" {
		ctx.LogDebugf("[ParallelsDesktop] [main] Getting all VMs fast for user %s", username)
		var err error
		syncVms, err = s.getAllUserVmSync(ctx, username)
		if err != nil {
			return nil, err
		}
	} else {
		vm, err := s.getUserVmSync(ctx, username, vmId)
		if err != nil {
			return nil, err
		}
		syncVms = append(syncVms, *vm)
	}

	// build a fast-lookup cache map
	s.RLock()
	cacheMap := make(map[string]models.ParallelsVM)
	for _, cvm := range s.cachedLocalVms {
		cacheMap[cvm.ID] = cvm
	}
	s.RUnlock()

	// cache patching engine
	for i := range syncVms {
		// Parallels bug: Home is missing on fast bulk fetches
		if syncVms[i].Home == "" {

			// FAST PATH: Try to patch it from our memory cache Map
			if cached, exists := cacheMap[syncVms[i].ID]; exists && cached.Home != "" {
				ctx.LogTracef("[ParallelsDesktop] [main] Patched Home path for %s from cache", syncVms[i].ID)
				syncVms[i].Home = cached.Home
				syncVms[i].HomePath = cached.HomePath
			} else {
				// SLOW PATH FALLBACK: Not in cache! We must run the slow command.
				ctx.LogDebugf("[ParallelsDesktop] [main] Cached vm is missing the Home path %s, executing fallback prlctl list...", syncVms[i].ID)
				fallbackVm, err := s.getUserVmSync(ctx, username, syncVms[i].ID)
				if err == nil {
					syncVms[i].Home = fallbackVm.Home
					syncVms[i].HomePath = fallbackVm.HomePath
				}
			}
		}

		// Ensure HomePath is set if Home exists but HomePath doesn't
		if syncVms[i].Home != "" && syncVms[i].HomePath == "" {
			syncVms[i].HomePath = fmt.Sprintf("%s/config.pvs", syncVms[i].Home)
		}
	}

	// ip resolution
	externalIp, _ := system.Get().GetExternalIp(ctx)
	for i := range syncVms {
		if externalIp != "" {
			syncVms[i].HostExternalIpAddress = externalIp
		}
		if len(syncVms[i].NetworkInformation.IPAddresses) > 0 {
			syncVms[i].InternalIpAddress = syncVms[i].NetworkInformation.IPAddresses[0].IP
		}
	}

	if vmId == "" {
		ctx.LogInfof("[ParallelsDesktop] [main] User %s has %v VMs", username, len(syncVms))
	} else if len(syncVms) > 0 {
		ctx.LogInfof("[ParallelsDesktop] [main] User %s VM %s found", username, vmId)
	}

	return syncVms, nil
}

// Fetches a single VM with full details, including the Home path.
func (s *ParallelsService) getUserVmSync(ctx basecontext.ApiContext, username string, vmId string) (*models.ParallelsVM, error) {
	ctx.LogDebugf("[ParallelsDesktop] [main] Getting VM %s for user %s", vmId, username)
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"list", vmId, "-a", "-i", "--json"},
	}.AsUser(username)

	output, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	var vms []models.ParallelsVM
	if err = json.Unmarshal([]byte(output), &vms); err != nil || len(vms) == 0 {
		return nil, errors.New("VM not found")
	}
	vms[0].User = username

	// ARM THE SHIELD! Prevent this single execution from causing an echo loop
	s.setCooldown(vmId)

	return &vms[0], nil
}

// Fetches all VMs for a user quickly, but suffers from the PD bug where 'Home' is empty
func (s *ParallelsService) getAllUserVmSync(ctx basecontext.ApiContext, username string) ([]models.ParallelsVM, error) {
	ctx.LogDebugf("[ParallelsDesktop] [main] Getting all VMs for user %s", username)
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"list", "-a", "-i", "--json"},
	}.AsUser(username)

	output, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	var vms []models.ParallelsVM
	if err = json.Unmarshal([]byte(output), &vms); err != nil {
		return nil, err
	}

	for i := range vms {
		vms[i].User = username
	}
	ctx.LogDebugf("[ParallelsDesktop] [main] Found %v VMs for user %s", len(vms), username)
	return vms, nil
}

func (s *ParallelsService) GetCachedVms(ctx basecontext.ApiContext, filter string) ([]models.ParallelsVM, error) {
	ctx.LogInfof("Getting all VMs for all users with cache")
	var systemMachines []models.ParallelsVM
	var err error

	cfg := config.Get()
	if cfg.IsApi() || cfg.IsOrchestrator() {
		s.RLock()
		// Making a copy of the slice so we can release the lock IMMEDIATELY
		// instead of holding it during the filter operations below.
		systemMachines = make([]models.ParallelsVM, len(s.cachedLocalVms))
		copy(systemMachines, s.cachedLocalVms)
		s.RUnlock()
	} else { // if not API or Orchestrator, we will not maintain the cache and get the VMs directly from the system
		systemMachines, err = s.getVmsInMachineForCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
	}

	dbFilter, err := data.ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := data.FilterByProperty(systemMachines, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (s *ParallelsService) GetVms(ctx basecontext.ApiContext, filter string) ([]models.ParallelsVM, error) {
	ctx.LogInfof("Getting all VMs for all users without cache")
	var systemMachines []models.ParallelsVM
	var err error

	systemMachines, err = s.getVmsInMachineForCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	dbFilter, err := data.ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := data.FilterByProperty(systemMachines, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

// Gets all VMs for current user if the user is not root, otherwise gets all VMs for all users
// This is non cached
func (s *ParallelsService) getVmsInMachineForCurrentUser(ctx basecontext.ApiContext) ([]models.ParallelsVM, error) {
	ctx.LogDebugf("Getting all VMs for all users without cache")
	var systemMachines []models.ParallelsVM

	users, err := s.getFilteredUsers(ctx)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.ErrNoSystemUserFound()
	}

	for _, user := range users {
		userMachines, err := s.getUserVm(ctx, user.Username, "")
		if err != nil {
			return nil, err
		}

		for _, machine := range userMachines {
			found := false
			for _, globalMachine := range systemMachines {
				if strings.EqualFold(machine.ID, globalMachine.ID) {
					found = true
					break
				}
			}
			if !found {
				machine.User = user.Username
				systemMachines = append(systemMachines, machine)
			}
		}
	}

	return systemMachines, nil
}

func (s *ParallelsService) GetVm(ctx basecontext.ApiContext, id string) (*models.ParallelsVM, error) {
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *ParallelsService) GetVmSync(ctx basecontext.ApiContext, id string) (*models.ParallelsVM, error) {
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *ParallelsService) SetVmState(ctx basecontext.ApiContext, id string, desiredState ParallelsVirtualMachineDesiredState,
	flags DesiredStateFlags) error {
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.ErrNoVirtualMachineFound(id)
	}

	if vm.User == "" {
		vm.User = "root"
	}
	isStopForceFlagSet := false
	for _, flag := range flags.flags {
		if flag == "--force" {
			isStopForceFlagSet = true
			break
		}
	}

	switch desiredState {
	case ParallelsVirtualMachineDesiredStateStart:
		if vm.State == ParallelsVirtualMachineStateRunning.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateStopped.String() {
			return errors.New("VM is not stopped")
		}
	case ParallelsVirtualMachineDesiredStateStop:
		if vm.State == ParallelsVirtualMachineStateStopped.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() && !isStopForceFlagSet {
			return errors.New("VM is not running")
		}
		if (vm.State == ParallelsVirtualMachineStateRunning.String() ||
			vm.State == ParallelsVirtualMachineStatePaused.String()) && isStopForceFlagSet {
			ctx.LogDebugf("Adding --kill flag to stop running VM %s", id)
			flags.flags = []string{"--kill"}
		} else if vm.State == ParallelsVirtualMachineStateSuspended.String() && isStopForceFlagSet {
			ctx.LogDebugf("Adding --drop-state flag to stop VM %s from suspended/paused state", id)
			flags.flags = []string{"--drop-state"}
		}
	case ParallelsVirtualMachineDesiredStatePause:
		if vm.State == ParallelsVirtualMachineStatePaused.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() {
			return errors.New("VM is not running")
		}
	case ParallelsVirtualMachineDesiredStateSuspend:
		if vm.State == ParallelsVirtualMachineStateSuspended.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() {
			return errors.New("VM is not running")
		}
	case ParallelsVirtualMachineDesiredStateResume:
		if vm.State != ParallelsVirtualMachineStatePaused.String() &&
			vm.State != ParallelsVirtualMachineStateSuspended.String() {
			return errors.New("VM is not paused or suspended")
		}
	case ParallelsVirtualMachineDesiredStateReset:
		if vm.State == ParallelsVirtualMachineStateStopped.String() {
			return nil
		}
	case ParallelsVirtualMachineDesiredStateRestart:
		if vm.State == ParallelsVirtualMachineStateStopped.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() {
			return errors.New("VM is not running")
		}
	default:
		return errors.New("Invalid desired state")
	}
	cmd := helpers.Command{
		Command: s.executable,
		Args:    append([]string{desiredState.String(), id}, flags.flags...),
	}.AsUser(vm.User)
	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}
	return nil
}

func (s *ParallelsService) CloneVm(ctx basecontext.ApiContext, id string, cloneName string, destinationPath string) error {
	configure := models.VirtualMachineConfigRequest{
		Operations: []*models.VirtualMachineConfigRequestOperation{
			{
				Group:     "machine",
				Operation: "clone",
				Options: []*models.VirtualMachineConfigRequestOperationOption{
					{
						Flag:  "name",
						Value: cloneName,
					},
				},
			},
		},
	}

	if destinationPath != "" {
		configure.Operations[0].Options = append(configure.Operations[0].Options, &models.VirtualMachineConfigRequestOperationOption{
			Flag:  "dst",
			Value: destinationPath,
		})
	}

	if err := s.ConfigureVm(ctx, id, &configure); err != nil {
		return err
	}

	if err := s.RegenerateMacAddress(ctx, id, "root"); err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) StartVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateStart, DesiredStateFlags{})
}

func (s *ParallelsService) StopVm(ctx basecontext.ApiContext, id string, flags DesiredStateFlags) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateStop, flags)
}

func (s *ParallelsService) RestartVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateRestart, DesiredStateFlags{})
}

func (s *ParallelsService) SuspendVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateSuspend, DesiredStateFlags{})
}

func (s *ParallelsService) ResumeVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateResume, DesiredStateFlags{})
}

func (s *ParallelsService) ResetVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateReset, DesiredStateFlags{})
}

func (s *ParallelsService) PauseVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStatePause, DesiredStateFlags{})
}

func (s *ParallelsService) DeleteVm(ctx basecontext.ApiContext, id string, force bool) error {
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		return err
	}

	if vm == nil {
		return errors.Newf("VM with id %s was not found", id)
	}

	if vm.State != "stopped" && vm.State != "invalid" && !force {
		return errors.New("VM is not stopped")
	}

	// we have a vm that is not stopped or invalid and force is true, we need to stop it first
	if force && vm.State != "invalid" && vm.State != "stopped" {
		forceFlag := DesiredStateFlags{flags: []string{"--force"}}
		if err := s.StopVm(ctx, id, forceFlag); err != nil {
			return err
		}
	}

	// If the VM is in an invalid state, we need to destroy it first
	if vm.State == "invalid" {
		cmd := helpers.Command{
			Command: s.executable,
			Args:    []string{"unregister", id},
		}.AsUser(vm.User)
		_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
		if err != nil {
			return err
		}
		// now lets check if the folder exists and remove it
		if vm.Home != "" {
			if err := os.RemoveAll(vm.Home); err != nil {
				return err
			}
		}
		return nil
	}

	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"delete", id},
	}.AsUser(vm.User)
	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) VmStatus(ctx basecontext.ApiContext, id string) (*models.VirtualMachineStatus, error) {
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.New("VM not found")
	}
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"list", id, "-a", "-f", "--json"},
	}.AsUser(vm.User)

	output, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	var status []models.VirtualMachineStatus
	err = json.Unmarshal([]byte(output), &status)
	if err != nil {
		return nil, err
	}

	if len(status) == 1 {
		return &status[0], nil
	}

	return nil, errors.New("VM not found")
}

// CreateVMSnapshot creates a new snapshot for the specified VM
func (s *ParallelsService) CreateVMSnapshot(ctx basecontext.ApiContext, vmID string, request *models.CreateVMSnapshotRequest) (*models.CreateVMSnapshotResponse, error) {
	if request == nil {
		return nil, errors.New("snapshot create request is required")
	}

	vm, err := s.findVmSync(ctx, vmID)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.Newf("VM with id %s was not found", vmID)
	}

	args := []string{"snapshot", vmID}
	if request.SnapshotName != "" {
		args = append(args, "-n", request.SnapshotName)
		ctx.LogInfof("[parallelsdesktop][snapshots] Creating snapshot '%s' for VM %s", request.SnapshotName, vmID)
	}
	if request.SnapshotDescription != "" {
		args = append(args, "-d", request.SnapshotDescription)
		ctx.LogInfof("[parallelsdesktop][snapshots] Creating snapshot with description '%s' for VM %s", request.SnapshotDescription, vmID)
	}
	cmd := helpers.Command{
		Command: s.executable,
		Args:    args,
	}.AsUser(vm.User)

	output, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}
	output = strings.TrimSpace(output)

	// Extract snapshot ID from output string in format: "The snapshot with id {snapshot-id} has been successfully created."
	snapshotId := extractSnapshotId(output)
	if snapshotId == "" {
		return nil, errors.New("failed to extract snapshot ID from command output")
	}

	return &models.CreateVMSnapshotResponse{
		SnapshotName: request.SnapshotName,
		SnapshotId:   snapshotId,
	}, nil
}

// DeleteVMSnapshot deletes a snapshot from the specified VM
func (s *ParallelsService) DeleteVMSnapshot(ctx basecontext.ApiContext, vmId string, snapshotId string, request *models.DeleteVMSnapshotRequest) error {
	if snapshotId == "" {
		return errors.New("snapshot ID is required")
	}

	vm, err := s.findVmSync(ctx, vmId)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.Newf("VM with id %s was not found", vmId)
	}

	ctx.LogInfof("[parallelsdesktop][snapshots] Deleting snapshot %s for VM %s", snapshotId, vmId)

	args := []string{"snapshot-delete", vmId, "--id", snapshotId}
	if request.DeleteChildren {
		args = append(args, "-c")
	}

	cmd := helpers.Command{
		Command: s.executable,
		Args:    args,
	}.AsUser(vm.User)

	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) RevertVMSnapshot(ctx basecontext.ApiContext, vmId string, snapshotId string, request *models.RevertVMSnapshotRequest) error {
	if snapshotId == "" {
		return errors.New("snapshot ID is required")
	}

	vm, err := s.findVmSync(ctx, vmId)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.Newf("VM with id %s was not found", vmId)
	}

	ctx.LogInfof("[parallelsdesktop][snapshots] Reverting snapshot %s for VM %s", snapshotId, vmId)

	args := []string{"snapshot-switch", vmId, "--id", snapshotId}
	if request.SkipResume {
		args = append(args, "--skip-resume")
	}

	cmd := helpers.Command{
		Command: s.executable,
		Args:    args,
	}.AsUser(vm.User)

	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

// ListSnapshots lists all snapshots for the specified VM
func (s *ParallelsService) listVMSnapshots(ctx basecontext.ApiContext, vmId string) (*models.ListVMSnapshotResponse, error) {
	vm, err := s.findVmSync(ctx, vmId)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.Newf("VM with id %s was not found", vmId)
	}

	ctx.LogInfof("[parallelsdesktop][snapshots] Listing snapshots for VM %s", vmId)
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"snapshot-list", vmId, "--json"},
	}.AsUser(vm.User)

	output, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	// Parse the JSON which has snapshot IDs as keys
	var snapshotMap map[string]models.VMSnapshot
	err = json.Unmarshal([]byte(output), &snapshotMap)
	if err != nil && output != "" {
		return nil, errors.Newf("failed to parse snapshot list output: %v", err)
	}

	// Convert the map to a slice and set the ID field
	var snapshotList []models.VMSnapshot
	for id, snapshot := range snapshotMap {
		snapshot.ID = strings.Trim(id, "{}")
		snapshotList = append(snapshotList, snapshot)
	}

	snapshots := models.ListVMSnapshotResponse{
		Snapshots: snapshotList,
	}

	return &snapshots, nil
}

func (s *ParallelsService) RegisterVm(ctx basecontext.ApiContext, r models.RegisterVirtualMachineRequest) error {
	if r.Uuid != "" {
		vm, err := s.findVmInCacheAndSystem(ctx, r.Uuid)
		if err != nil {
			return err
		}
		if vm != nil {
			return errors.Newf("VM with UUID %s already exists", r.Uuid)
		}
	} else {
		r.Uuid = uuid.New().String()
	}

	if r.MachineName != "" {
		vm, err := s.findVmInCacheAndSystem(ctx, r.MachineName)
		if err != nil && errors.GetSystemErrorCode(err) != 404 {
			return err
		}
		if vm != nil {
			return errors.Newf("VM with name %s already exists", r.MachineName)
		} else if err := s.ReplaceMachineNameInConfigPvs(r.Path, r.MachineName); err != nil {
			return err
		}
	}

	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"register", r.Path},
	}

	var cmd helpers.Command
	if r.Owner != "" && r.Owner != "root" {
		cmd = baseCmd.AsUser(r.Owner)
	} else {
		cmd = baseCmd
	}
	if r.Uuid != "" {
		cmd.Args = append(cmd.Args, "--uuid", r.Uuid)
	}

	if r.RegenerateSourceUuid {
		cmd.Args = append(cmd.Args, "--regenerate-src-uuid")
	}
	if r.Force {
		cmd.Args = append(cmd.Args, "--force")
	}
	if r.DelayApplyingRestrictions {
		cmd.Args = append(cmd.Args, "--delay-applying-restrictions")
	}

	ctx.LogDebugf("Executing command: %s", cmd.String())
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	if err := s.RegenerateMacAddress(ctx, r.Uuid, r.Owner); err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) UnregisterVm(ctx basecontext.ApiContext, r models.UnregisterVirtualMachineRequest) error {
	vm, err := s.findVmSync(ctx, r.ID)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.ErrNoVirtualMachineFound(r.ID)
	}
	r.Owner = vm.User

	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"unregister", r.ID},
	}

	var cmd helpers.Command
	if r.Owner != "" && r.Owner != "root" {
		cmd = baseCmd.AsUser(r.Owner)
	} else {
		cmd = baseCmd
	}
	if r.CleanSourceUuid {
		cmd.Args = append(cmd.Args, "--clean-src-uuid")
	}

	ctx.LogInfof(cmd.String())
	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return errors.NewFromErrorf(err, "Error unregistering VM %s", r.ID)
	}

	return nil
}

func (s *ParallelsService) RenameVm(ctx basecontext.ApiContext, r models.RenameVirtualMachineRequest) error {
	vm, err := s.findVmSync(ctx, r.GetId())
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}
	if vm.State != "stopped" {
		ctx.LogWarnf("VM %s is not stopped, we cannot rename it", vm.ID)
		return errors.New("VM is not stopped")
	}

	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"set", r.GetId(), "--name", r.NewName},
	}

	var cmd helpers.Command
	if vm.User != "" && vm.User != "root" {
		cmd = baseCmd.AsUser(vm.User)
	} else {
		cmd = baseCmd
	}
	if r.Description != "" {
		cmd.Args = append(cmd.Args, "--description", r.Description)
	}

	ctx.LogInfof(cmd.String())
	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) PackVm(ctx basecontext.ApiContext, idOrName string) error {
	vm, err := s.findVmSync(ctx, idOrName)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.Newf("VM with ID %s was not found", idOrName)
	}

	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"pack", vm.ID},
	}

	var cmd helpers.Command
	if vm.User != "" && vm.User != "root" {
		cmd = baseCmd.AsUser(vm.User)
	} else {
		cmd = baseCmd
	}
	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)

	return err
}

func (s *ParallelsService) UnpackVm(ctx basecontext.ApiContext, idOrName string) error {
	vm, err := s.findVmSync(ctx, idOrName)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.Newf("VM with ID %s was not found", idOrName)
	}

	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"unpack", vm.ID},
	}

	var cmd helpers.Command
	if vm.User != "" && vm.User != "root" {
		cmd = baseCmd.AsUser(vm.User)
	} else {
		cmd = baseCmd
	}
	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)

	return err
}

func (s *ParallelsService) GetInfo() (*models.ParallelsDesktopInfo, error) {
	if s.Info != nil {
		return s.Info, nil
	}

	stdout, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), helpers.Command{
		Command: s.serverExecutable,
		Args:    []string{"info", "--json"},
	}, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	var info models.ParallelsDesktopInfo
	err = json.Unmarshal([]byte(stdout), &info)
	if err != nil {
		return nil, err
	}

	s.Info = &info
	if info.License.State != "valid" {
		s.ctx.LogErrorf("Parallels license is not active")
	} else {
		s.isLicensed = true
	}

	return s.Info, nil
}

func (s *ParallelsService) GetUsers(ctx basecontext.ApiContext) ([]*models.ParallelsDesktopUser, error) {
	if s.Users != nil {
		return s.Users, nil
	}

	stdout, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), helpers.Command{
		Command: s.serverExecutable,
		Args:    []string{"user", "list", "--json"},
	}, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	var users []*models.ParallelsDesktopUser
	err = json.Unmarshal([]byte(stdout), &users)
	if err != nil {
		return nil, err
	}

	s.Users = users

	return s.Users, nil
}

func (s *ParallelsService) GetUser(ctx basecontext.ApiContext, user string) (*models.ParallelsDesktopUser, error) {
	if s.Users != nil || len(s.Users) == 0 {
		s.GetUsers(ctx)
	}

	for _, u := range s.Users {
		currentName := strings.Split(u.Name, "@")
		if strings.EqualFold(currentName[0], user) {
			return u, nil
		}
	}

	return nil, errors.Newf("User %s not found", user)
}

func (s *ParallelsService) GetUserHome(ctx basecontext.ApiContext, user string) (string, error) {

	if user == "" {
		var err error
		user, err = system.Get().GetCurrentUser(ctx)
		if err != nil {
			return "", err
		}
	}

	parallelsUser, err := s.GetUser(ctx, user)
	if err != nil {
		return "", err
	}

	return parallelsUser.DefVMHome, nil
}

func (s *ParallelsService) ConfigureVm(ctx basecontext.ApiContext, id string, setOperations *models.VirtualMachineConfigRequest) error {
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.ErrNoVirtualMachineFound(id)
	}

	for _, op := range setOperations.Operations {
		op.Owner = vm.User
		switch op.Group {
		case "state":
			ctx.LogInfof("Setting machine state to %s", op.Operation)
			if err := s.SetVmState(ctx, vm.ID, ParallelsVirtualMachineDesiredStateFromString(op.Operation), DesiredStateFlags{}); err != nil {
				op.Error = err
			}
		case "machine":
			if err := s.SetVmMachineOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "cpu":
			ctx.LogInfof("Setting cpu property %s to %s", op.Operation, op.Value)
			if err := s.SetVmCpu(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "memory":
			ctx.LogInfof("Setting memory property %s to %s", op.Operation, op.Value)
			if err := s.SetVmMemory(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "boot-order":
			ctx.LogInfof("Setting boot order property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "efi-secure-boot":
			ctx.LogInfof("Setting boot efi secure boot property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "select-boot-device":
			ctx.LogInfof("Setting select boot device property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "external-boot-device":
			ctx.LogInfof("Setting external boot device property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "time":
			ctx.LogInfof("Setting time sync property %s to %s", op.Operation, op.Value)
			if err := s.SetTimeSyncOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "network":
			ctx.LogInfof("Setting network property %s to %s", op.Operation, op.Value)
		case "device":
			ctx.LogInfof("Setting device property %s to %s", op.Operation, op.Value)
			if err := s.SetVmDeviceOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "shared-folder":
			ctx.LogInfof("Setting shared_folder property %s to %s", op.Operation, op.Value)
			if err := s.SetVmSharedFolderOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "rosetta":
			ctx.LogInfof("Setting rosetta property %s to %s", op.Operation, op.Value)
			if err := s.SetVmRosettaEmulation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "cmd":
			ctx.LogInfof("Setting custom property %s to %s", op.Operation, op.Value)
			if err := s.RunCustomCommand(ctx, vm, op); err != nil {
				op.Error = err
			}
		default:
			return errors.Newf("Invalid group %s", op.Group)
		}
	}

	for _, op := range setOperations.Operations {
		if op.Error != nil {
			return op.Error
		}
	}

	return nil
}

func (s *ParallelsService) CreateVm(ctx basecontext.ApiContext, template data_models.PackerTemplate, desiredState string) (*models.ParallelsVM, error) {
	return s.CreatePackerTemplateVm(ctx, template, desiredState)
}

func (s *ParallelsService) CreatePackerTemplateVm(ctx basecontext.ApiContext, template data_models.PackerTemplate, desiredState string) (*models.ParallelsVM, error) {
	ctx.LogInfof("Creating Packer Virtual Machine %s", template.Name)
	existVm, err := s.findVmSync(ctx, template.Name)
	if existVm != nil || err != nil {
		if errors.GetSystemErrorCode(err) != 404 {
			return nil, errors.Newf("Machine %v with ID %v already exists and is %s", template.Name, existVm.ID, existVm.State)
		}
	}

	git := git.Get(ctx)
	repoPath, err := git.Clone(ctx, "https://github.com/Parallels/packer-examples", template.Owner, "packer-examples")
	if err != nil {
		ctx.LogErrorf("Error cloning packer-examples repository: %s", err.Error())
		return nil, err
	}

	ctx.LogInfof("Cloned packer-examples repository to %s", repoPath)

	packer := packer.Get(ctx)
	scriptPath := fmt.Sprintf("%s/%s", repoPath, template.PackerFolder)
	overrideFilePath := fmt.Sprintf("%s/%s/override.pkrvars.hcl", repoPath, template.PackerFolder)
	overrideFile := make(map[string]interface{})
	if template.Name != "" {
		overrideFile["machine_name"] = template.Name
	}
	if template.Hostname != "" {
		overrideFile["hostname"] = template.Hostname
	}
	overrideFile["create_vagrant_box"] = false
	overrideFile["machine_specs"] = map[string]interface{}{}
	if template.Specs["memory"] != "" {
		memory, err := strconv.Atoi(template.Specs["memory"])
		if err != nil {
			memory = 2048
		}
		overrideFile["machine_specs"].(map[string]interface{})["memory"] = memory
	}
	if template.Specs["cpu"] != "" {
		cpu, err := strconv.Atoi(template.Specs["cpu"])
		if err != nil {
			cpu = 2
		}
		overrideFile["machine_specs"].(map[string]interface{})["cpus"] = cpu
	}
	if template.Specs["disk"] != "" {
		disk, err := strconv.Atoi(template.Specs["disk"])
		if err != nil {
			disk = 40960
		}
		overrideFile["machine_specs"].(map[string]interface{})["disk_size"] = disk
	}

	template.Addons = append(template.Addons, "parallels-tools")
	if len(template.Addons) > 0 {
		overrideFile["addons"] = template.Addons
	}

	for key, value := range template.Variables {
		overrideFile[key] = value
	}

	overrideFileContent := helpers.ToHCL(overrideFile, 0)
	if err := helper.WriteToFile(overrideFileContent, overrideFilePath); err != nil {
		ctx.LogErrorf("Error writing override file %s: %s", overrideFilePath, err.Error())
		return nil, err
	}

	ctx.LogInfof("Created override file")

	ctx.LogInfof("Initializing packer repository")
	if err = packer.Init(ctx, template.Owner, scriptPath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}
	ctx.LogInfof("Initialized packer repository")

	ctx.LogInfof("Building packer machine")
	if err = packer.Build(ctx, template.Owner, scriptPath, overrideFilePath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	ctx.LogInfof("Built packer machine")

	users, err := system.Get().GetSystemUsers(ctx)
	if err != nil {
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	userExists := false
	if template.Owner == "root" {
		userExists = true
	} else {
		for _, user := range users {
			if user.Username == template.Owner {
				userExists = true
				break
			}
		}
	}

	if !userExists {
		ctx.LogErrorf("User %s does not exist", template.Owner)
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, errors.New("User does not exist")
	}

	userFolder := fmt.Sprintf("/Users/%s/Parallels", template.Owner)
	if template.Owner == "root" {
		userFolder = "/var/root"
	}

	err = helpers.CreateDirIfNotExist(userFolder)
	if err != nil {
		ctx.LogErrorf("Error creating user folder %s: %s", userFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	ctx.LogInfof("Created user folder %s", userFolder)

	destinationFolder := fmt.Sprintf("%s/%s.pvm", userFolder, template.Name)
	sourceFolder := fmt.Sprintf("%s/out/%s.pvm", scriptPath, template.Name)
	err = helpers.MoveFolder(sourceFolder, destinationFolder)
	if err != nil {
		ctx.LogErrorf("Error moving folder %s to %s: %s", sourceFolder, destinationFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		if helper.DirectoryExists(sourceFolder) {
			if cleanError := helpers.RemoveFolder(sourceFolder); cleanError != nil {
				ctx.LogErrorf("Error removing destination folder %s: %s", repoPath, cleanError.Error())
				return nil, cleanError
			}
		}
		return nil, err
	}

	if template.Owner != "root" {
		cmd := helpers.Command{
			Command: "sudo",
			Args:    make([]string, 0),
		}
		cmd.Args = append(cmd.Args, "chown", "-R", template.Owner, destinationFolder)

		_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
		if err != nil {
			ctx.LogErrorf("Error changing owner of folder %s to %s: %s", destinationFolder, template.Owner, err.Error())
			if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
				ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
				return nil, cleanError
			}
			return nil, err
		}
	}

	ctx.LogInfof("Moved folder %s to %s", sourceFolder, destinationFolder)
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"register", destinationFolder},
	}.AsUser(template.Owner)

	_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		ctx.LogErrorf("Error registering VM %s: %s", destinationFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	ctx.LogInfof("Registered VM %s", destinationFolder)

	existVm, err = s.findVmSync(ctx, template.Name)
	if existVm == nil || err != nil {
		ctx.LogErrorf("Error finding VM %s: %s", template.Name, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, errors.Newf("Something went wrong with creating machine %v, it does not exist, err: %v", template.Name, err.Error())
	}

	if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
		ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
		return nil, cleanError
	}

	switch desiredState {
	case "running":
		if err := s.StartVm(ctx, existVm.ID); err != nil {
			ctx.LogErrorf("Error starting VM %s: %s", existVm.ID, err.Error())
			return nil, err
		}
	default:
		ctx.LogInfof("Desired state is %s, not starting VM %s", desiredState, existVm.ID)
	}

	ctx.LogInfof("Created VM %s", existVm.ID)
	return existVm, nil
}

// Config Region
func (s *ParallelsService) SetVmMachineOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    make([]string, 0),
	}

	switch op.Operation {
	case "clone":
		baseCmd.Args = append(baseCmd.Args, "clone", vm.ID)
		if op.Value != "" {
			baseCmd.Args = append(baseCmd.Args, "--name", fmt.Sprintf("\"%s\"", op.Value))
		}
		baseCmd.Args = append(baseCmd.Args, op.GetCmdArgs()...)
	case "archive":
		baseCmd.Args = append(baseCmd.Args, "archive", vm.ID)
	case "unarchive":
		baseCmd.Args = append(baseCmd.Args, "unarchive", vm.ID)
	case "pack":
		baseCmd.Args = append(baseCmd.Args, "pack", vm.ID)
	case "unpack":
		baseCmd.Args = append(baseCmd.Args, "unpack", vm.ID)
	case "encrypt":
		baseCmd.Args = append(baseCmd.Args, "encrypt", vm.ID)
		baseCmd.Args = append(baseCmd.Args, op.GetCmdArgs()...)
	case "decrypt":
		baseCmd.Args = append(baseCmd.Args, "decrypt", vm.ID)
		baseCmd.Args = append(baseCmd.Args, op.GetCmdArgs()...)
	case "reset-uptime":
		baseCmd.Args = append(baseCmd.Args, "reset-uptime", vm.ID)
	case "install-tools":
		baseCmd.Args = append(baseCmd.Args, "install-tools", vm.ID)
	case "rename":
		baseCmd.Args = append(baseCmd.Args, "set", vm.ID, "--name", fmt.Sprintf("\"%s\"", op.Value))
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}
	cmd := baseCmd.AsUser(vm.User)
	ctx.LogDebugf(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmBootOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"set", vm.ID},
	}.AsUser(vm.User)

	switch op.Operation {
	case "boot-order":
		cmd.Args = append(cmd.Args, "--device-bootorder", op.Value)
	case "bios-type":
		if op.Value != "legacy" && op.Value != "efi32" && op.Value != "efi64" && op.Value != "efi-arm64" {
			return errors.ErrConfigInvalidBiosType(op.Value)
		}
		cmd.Args = append(cmd.Args, "--device-bootorder", op.Value)
	case "efi-secure-boot":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--efi-secure-boot", "on")
		} else {
			cmd.Args = append(cmd.Args, "--efi-secure-boot", "off")
		}
	case "select-boot-device":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--select-boot-device", "on")
		} else {
			cmd.Args = append(cmd.Args, "--select-boot-device", "off")
		}
	case "external-boot-device":
		cmd.Args = append(cmd.Args, "--external-boot-device", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmSharedFolderOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"set", vm.ID},
	}.AsUser(vm.User)

	switch op.Operation {
	case "add":
		if op.GetOption("path").Value == "" {
			return errors.ErrConfigMissingSharedFolderPath()
		}
		cmd.Args = append(cmd.Args, "--shf-host-add", op.Value)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "set":
		cmd.Args = append(cmd.Args, "--shf-host-set", op.Value)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "delete":
		cmd.Args = append(cmd.Args, "--shf-host-del", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmDeviceOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"set", vm.ID},
	}.AsUser(vm.User)

	switch op.Operation {
	case "add":
		switch op.Value {
		case "cdrom":
			cmd.Args = append(cmd.Args, "--device-add", "cdrom")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "fdd":
			cmd.Args = append(cmd.Args, "--device-add", "fdd")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "hdd":
			cmd.Args = append(cmd.Args, "--device-add", "hdd")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "net":
			cmd.Args = append(cmd.Args, "--device-add", "net")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "serial":
			cmd.Args = append(cmd.Args, "--device-add", "serial")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "parallel":
			cmd.Args = append(cmd.Args, "--device-add", "parallel")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "usb":
			cmd.Args = append(cmd.Args, "--device-add", "usb")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "sound":
			cmd.Args = append(cmd.Args, "--device-add", "sound")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		default:
			return errors.ErrConfigInvalidOperation(op.Value)
		}
	case "set":
		cmd.Args = append(cmd.Args, "--device-set", op.Value)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "connect":
		cmd.Args = append(cmd.Args, "--device-connect", op.Value)
	case "disconnect":
		cmd.Args = append(cmd.Args, "--device-disconnect", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmCpu(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    make([]string, 0),
	}
	var cmd helpers.Command

	switch op.Operation {
	case "set":
		if op.Value != "auto" {
			_, err := strconv.Atoi(op.Value)
			if err != nil {
				return err
			}
		}
		baseCmd.Args = append(baseCmd.Args, "set", vm.ID, "--cpus", op.Value)
	case "set_type":
		if op.Value != "x86" && op.Value != "arm" {
			return errors.Newf("Invalid CPU type %s", op.Value)
		}
		baseCmd.Args = append(baseCmd.Args, "set", vm.ID, "--cpu-type", op.Value)
	default:
		return errors.Newf("Invalid operation %s", op.Operation)
	}
	cmd = baseCmd.AsUser(op.Owner)
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmMemory(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    make([]string, 0),
	}
	var cmd helpers.Command

	switch op.Operation {
	case "set":
		if op.Value != "auto" {
			_, err := strconv.Atoi(op.Value)
			if err != nil {
				return err
			}
		}
		baseCmd.Args = append(baseCmd.Args, "set", vm.ID, "--memsize", op.Value)
	default:
		return errors.Newf("Invalid operation %s", op.Operation)
	}
	cmd = baseCmd.AsUser(op.Owner)
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmRosettaEmulation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    make([]string, 0),
	}

	switch op.Operation {
	case "set":
		if op.Value != "on" && op.Value != "off" && op.Value != "true" && op.Value != "false" {
			return errors.Newf("Invalid value %s", op.Value)
		}

		if op.Value == "on" || op.Value == "true" {
			baseCmd.Args = append(baseCmd.Args, "set", vm.ID, "--rosetta-linux", "on")
		}
		if op.Value == "off" || op.Value == "false" {
			baseCmd.Args = append(baseCmd.Args, "set", vm.ID, "--rosetta-linux", "off")
		}
	default:
		return errors.Newf("Invalid operation %s", op.Operation)
	}
	cmd := baseCmd.AsUser(op.Owner)
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetTimeSyncOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"set", vm.ID},
	}.AsUser(vm.User)

	switch op.Operation {
	case "time-sync":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--time-sync", "on")
		} else {
			cmd.Args = append(cmd.Args, "--time-sync", "off")
		}
	case "time-sync-smart-mode":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--time-sync-smart-mode", "on")
		} else {
			cmd.Args = append(cmd.Args, "--time-sync-smart-mode", "off")
		}
	case "disable-timezone-synct":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--disable-timezone-sync", "on")
		} else {
			cmd.Args = append(cmd.Args, "--disable-timezone-sync", "off")
		}
	case "time-sync-interval":
		cmd.Args = append(cmd.Args, "--time-sync-interval", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) LocalUploadToVm(ctx basecontext.ApiContext, id string, r *models.VirtualMachineUploadRequest) (*models.VirtualMachineUploadResponse, error) {
	response := models.VirtualMachineUploadResponse{}

	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		response.Error = err.Error()
		return &response, err
	}
	if vm == nil {
		err := errors.Newf("VM with ID %s was not found", id)
		response.Error = err.Error()
		return &response, err
	}

	if vm.State != "running" {
		err := errors.New("VM is not running")
		response.Error = err.Error()
		return &response, err
	}

	if r.LocalPath == "" {
		err := errors.New("Local path is required")
		response.Error = err.Error()
		return &response, err
	}

	if _, err := os.Stat(r.LocalPath); os.IsNotExist(err) {
		err := errors.Newf("file %s does not exist", r.LocalPath)
		response.Error = err.Error()
		return &response, err
	}

	prlcopySupportedVersion := helpers.NewVersion("20.2.0")
	currentVersion := helpers.NewVersion(s.Version())

	if currentVersion.LessThan(prlcopySupportedVersion) {
		if r.RemotePath == "" {
			r.RemotePath = "/tmp"
		}

		response.LocalPath = r.RemotePath

		// First command to compress the file / folder
		cmd := helpers.Command{
			Command: "tar",
			Args:    make([]string, 0),
		}

		cmd.Args = append(cmd.Args, "czf", "-", "--no-mac-metadata", "--no-xattrs", "--no-fflags")
		cmd.Args = append(cmd.Args, "-C", filepath.Dir(r.LocalPath), filepath.Base(r.LocalPath))

		ctx.LogInfof("Executing command %s %s", cmd.Command, strings.Join(cmd.Args, " "))
		cmd1 := exec.Command(cmd.Command, cmd.Args...)
		outPipe, _ := cmd1.StdoutPipe()

		// stderr1 is to capture the error message from the command
		stderr1 := &bytes.Buffer{}
		cmd1.Stderr = stderr1

		// Constructing second command, to copy the file / folder to the VM
		if vm.User != "root" {
			cmd.Args = append(cmd.Args, "-u", vm.User)
		}

		cmd = helpers.Command{
			Command: s.executable,
			Args:    make([]string, 0),
		}

		cmd.Args = append(cmd.Args, "exec", vm.ID, "--current-user", "tar", "xzf", "-", "-C", r.RemotePath)

		ctx.LogInfof("Executing command %s %s", cmd.Command, strings.Join(cmd.Args, " "))
		cmd2 := exec.Command(cmd.Command, cmd.Args...)
		inPipe, _ := cmd2.StdinPipe()
		cmd2.Stdout = os.Stdout // Output to terminal

		stderr2 := &bytes.Buffer{}
		cmd2.Stderr = stderr2

		// Start first command
		if err := cmd1.Start(); err != nil {
			response.Error = err.Error()
			return &response, err
		}

		// Pipe data from cmd1 to cmd2
		go func() {
			io.Copy(inPipe, outPipe)
			inPipe.Close()
		}()

		// Start second command
		if err := cmd2.Start(); err != nil {
			response.Error = "q" + err.Error()
			return &response, err
		}

		// Wait for first command to finish
		if err := cmd1.Wait(); err != nil {
			ctx.LogInfof("Compressing the file/dir failed: %s Error : %s", err.Error(), stderr1.String())
			response.Error = err.Error()
			return &response, err
		}

		// Wait for second command to finish
		if err := cmd2.Wait(); err != nil {
			ctx.LogInfof("Copy file to VM failed: %s Error : %s", err.Error(), stderr2.String())
			response.Error = err.Error()
			return &response, err
		}

	} else {
		cmd := helpers.Command{
			Command: "/usr/local/bin/prlcopy",
			Args:    make([]string, 0),
		}

		cmd.Args = append(cmd.Args, "--vm", vm.ID, r.LocalPath)

		if r.RemotePath != "" {
			cmd.Args = append(cmd.Args, r.RemotePath)
		}

		ctx.LogInfof("Executing command %s %s", cmd.Command, strings.Join(cmd.Args, " "))
		_, err = helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
		if err != nil {
			response.Error = err.Error()
			return &response, err
		}
	}

	return &response, nil
}

func (s *ParallelsService) ExecuteCommandOnVm(ctx basecontext.ApiContext, id string, r *models.VirtualMachineExecuteCommandRequest) (*models.VirtualMachineExecuteCommandResponse, error) {
	response := &models.VirtualMachineExecuteCommandResponse{}
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		ctx.LogErrorf("Error finding VM %s: %s", id, err.Error())
		return nil, err
	}
	if vm == nil {
		ctx.LogErrorf("Error finding VM %s: VM not found", id)
		return nil, errors.New("VM not found")
	}

	if vm.State != "running" {
		return nil, errors.New("VM is not running")
	}
	ctx.LogInfof("Preparing to execute command %s on VM %s", r.Command, vm.ID)
	envVars := ""
	bashCommand := ""
	commandParts := strings.Split(r.Command, " ")
	command := ""
	if r.UseSudo {
		command = fmt.Sprintf("sudo %s", command)
	}
	if len(commandParts) > 1 {
		command = strings.Join(commandParts, " ")
	} else {
		command = commandParts[0]
	}

	for key, value := range r.EnvironmentVariables {
		envVars += fmt.Sprintf(`%s='%s' `, key, value)
	}

	envVars = strings.TrimSpace(envVars)
	if envVars != "" {
		bashCommand = fmt.Sprintf("env %s bash -c '%s'", envVars, command)
	} else {
		bashCommand = command
	}
	cmd := helpers.Command{
		Command: s.executable,
	}
	cmd.Args = make([]string, 0)

	cmd.Args = append(cmd.Args, "exec", vm.ID, bashCommand)
	cmd = cmd.AsUser(vm.User)
	ctx.LogInfof("Executing command %s %s", cmd.Command, strings.Join(cmd.Args, " "))
	stdout, stderr, exitCode, cmdError := helpers.ExecuteWithOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	response.Stdout = stdout
	response.Stderr = stderr
	response.ExitCode = exitCode
	if cmdError != nil {
		response.Error = cmdError.Error()
	}

	ctx.LogInfof("Command %s %s executed with exit code %v", cmd.Command, strings.Join(cmd.Args, " "), exitCode)
	return response, nil
}

func (s *ParallelsService) ReplaceMachineNameInConfigPvs(path string, newName string) error {
	configPath := filepath.Join(path, "config.pvs")
	if !helper.FileExists(configPath) {
		return errors.Newf("Config file %s not found", configPath)
	}

	// fileInfo, err := os.Stat("filename")
	// if err != nil {
	// 	return err
	// }

	// fileMode := fileInfo.Mode()
	// // Get the file owner
	// sys := fileInfo.Sys().(*syscall.Stat_t)
	// uid := sys.Uid
	// gid := int(sys.Gid)

	file, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := helper.ReadFromFile(configPath)
	if err != nil {
		return err
	}

	pattern := regexp.MustCompile(`<[Vv]m[Nn]ame>[^<]*</[Vv]m[Nn]ame>`)

	newContent := pattern.ReplaceAllString(string(content), fmt.Sprintf("<VmName>%s</VmName>", newName))

	err = helper.WriteToFile(newContent, configPath)
	if err != nil {
		return err
	}

	// // Set the mode and owner of another file to be the same as configPath
	// err = os.Chmod(configPath, fileMode)
	// if err != nil {
	// 	return err
	// }
	// err = os.Chown(configPath, uid, gid)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (s *ParallelsService) ReplaceMacAddressInConfigPvs(path string) error {
	configPath := filepath.Join(path, "config.pvs")
	if !helper.FileExists(configPath) {
		return errors.Newf("Config file %s not found", configPath)
	}
	// lets get the config.pvs current owner
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return err
	}

	fileMode := fileInfo.Mode()
	// Get the file owner (platform-specific)
	uid, gid := getFileOwner(fileInfo)

	file, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := helper.ReadFromFile(configPath)
	if err != nil {
		return err
	}
	macPrefix := "001C42"
	macRandom3Octets := fmt.Sprintf("%02X%02X%02X", rand.Intn(256), rand.Intn(256), rand.Intn(256))
	macAddress := macPrefix + macRandom3Octets

	pattern := regexp.MustCompile(`<[Mm]ac>[^<]*</[Mm]ac>`)

	newContent := pattern.ReplaceAllString(string(content), fmt.Sprintf("<Mac>%s</Mac>", macAddress))

	err = helper.WriteToFile(newContent, configPath)
	if err != nil {
		return err
	}

	// Set the mode and owner of another file to be the same as configPath
	err = os.Chmod(configPath, fileMode)
	if err != nil {
		return err
	}
	err = os.Chown(configPath, int(uid), int(gid))
	if err != nil {
		return err
	}
	return nil
}

func (s *ParallelsService) RunCustomCommand(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	baseCmd := helpers.Command{
		Command: s.executable,
		Args:    []string{op.Operation, vm.ID},
	}

	var cmd helpers.Command
	// Setting the owner in the command
	if op.Owner != "root" {
		cmd = baseCmd.AsUser(op.Owner)
	} else {
		cmd = baseCmd
	}
	cmd.Args = append(cmd.Args, op.GetCmdArgs()...)

	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) GetHardwareUsage(ctx basecontext.ApiContext) (*models.SystemUsageResponse, error) {
	result := &models.SystemUsageResponse{
		TotalInUse:     &models.SystemUsageItem{},
		TotalReserved:  &models.SystemUsageItem{},
		SystemReserved: &models.SystemUsageItem{},
		Total:          &models.SystemUsageItem{},
	}

	vms, err := s.GetCachedVms(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.State == "running" {
			if result.TotalInUse == nil {
				result.TotalInUse = &models.SystemUsageItem{}
			}
			result.TotalInUse.LogicalCpuCount += vm.Hardware.CPU.Cpus
			memorySizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Memory.Size)
			if err != nil {
				return nil, err
			}
			result.TotalInUse.MemorySize += helpers.ConvertByteToMegabyte(memorySizeInt)
			if vm.Hardware.Hdd0.Size != "" {
				hddSizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Hdd0.Size)
				if err != nil {
					return nil, err
				}
				result.TotalInUse.DiskSize += helpers.ConvertByteToMegabyte(hddSizeInt)
			}
		} else {
			if result.TotalReserved == nil {
				result.TotalReserved = &models.SystemUsageItem{}
			}
			result.TotalReserved.LogicalCpuCount += vm.Hardware.CPU.Cpus
			memorySizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Memory.Size)
			if err != nil {
				return nil, err
			}
			result.TotalReserved.MemorySize += helpers.ConvertByteToMegabyte(memorySizeInt)
			if vm.Hardware.Hdd0.Size != "" {
				hddSizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Hdd0.Size)
				if err != nil {
					return nil, err
				}
				result.TotalReserved.DiskSize += helpers.ConvertByteToMegabyte(hddSizeInt)
			}
		}
	}
	result.TotalInUse.MacVMsRunning = s.getMacVMsRunning()

	cfg := config.Get()
	systemSrv := system.Get()
	systemInfo, err := systemSrv.GetHardwareInfo(ctx)
	if err != nil {
		return nil, err
	}
	result.CpuType = systemInfo.CpuType
	result.CpuBrand = systemInfo.CpuBrand
	result.DevOpsVersion = system.VersionSvc.String()
	result.ParallelsDesktopVersion = s.Version()
	result.ParallelsDesktopLicensed = s.isLicensed

	result.SystemReserved = &models.SystemUsageItem{}
	result.SystemReserved.LogicalCpuCount = int64(cfg.SystemReservedCpu())
	result.SystemReserved.MemorySize = float64(cfg.SystemReservedMemory())
	result.SystemReserved.DiskSize = float64(cfg.SystemReservedDisk())

	result.Total = &models.SystemUsageItem{}
	result.Total.LogicalCpuCount = int64(systemInfo.LogicalCpuCount)
	result.Total.MemorySize = systemInfo.MemorySize
	result.Total.DiskSize = systemInfo.DiskSize - systemInfo.FreeDiskSize

	result.TotalAvailable = &models.SystemUsageItem{}
	result.TotalAvailable.DiskSize = systemInfo.FreeDiskSize
	result.TotalAvailable.MemorySize = result.Total.MemorySize - result.SystemReserved.MemorySize - result.TotalInUse.MemorySize
	result.TotalAvailable.LogicalCpuCount = result.Total.LogicalCpuCount - result.SystemReserved.LogicalCpuCount - result.TotalInUse.LogicalCpuCount
	result.TotalAvailable.PrlHomeFreeSize, _ = s.GetParallelsHomeDiskSpaceInfo(ctx, "")

	homeDir, err := s.GetUserHome(ctx, "")
	if err != nil {
		return nil, err
	}
	result.TotalAvailable.PrlHomeTotalSize, _ = helpers.GetTotalDiskSpace(homeDir)

	external_ip, err := systemSrv.GetExternalIp(ctx)
	if err == nil {
		result.ExternalIpAddress = external_ip
	}
	osVersion, err := systemSrv.GetOsVersion(ctx)
	if err == nil {
		result.OsVersion = osVersion
	}
	result.OsName = systemSrv.GetOSName()

	return result, nil
}
func (s *ParallelsService) RegenerateMacAddress(ctx basecontext.ApiContext, vmID string, owner string) error {
	// getting the VM to check state
	vm, err := s.findVmSync(ctx, vmID)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}
	if vm.State != "stopped" {
		if vm.Home == "" {
			ctx.LogWarnf("VM %s has no home path, skipping MAC address regeneration", vm.ID)
			return nil
		}
		err := s.ReplaceMacAddressInConfigPvs(vm.Home)
		if err != nil {
			return err
		}
		return nil
	}

	// lets regenerate the MAC address for the VM
	regenerateMacAddressCmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"set", vmID, "--device-set", "net0", "--mac", "auto"},
	}

	if owner != "" && owner != "root" {
		regenerateMacAddressCmd = regenerateMacAddressCmd.AsUser(owner)
	}

	ctx.LogDebugf("Executing command: %s", regenerateMacAddressCmd.String())
	_, err = helpers.ExecuteWithNoOutput(ctx.Context(), regenerateMacAddressCmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) GetParallelsHomeDiskSpaceInfo(ctx basecontext.ApiContext, username string) (int64, error) {

	if username == "" {
		if user, err := system.Get().GetCurrentUser(ctx); err == nil {
			ctx.LogInfof("No username provided, using current user %s for disk space info", user)
			username = user
		}
	}
	path, err := s.GetUserHome(ctx, username)
	if err != nil {
		return 0, err
	}
	return helpers.GetFreeDiskSpace(path)
}
func (s *ParallelsService) getMacVMsRunning() []string {
	s.macVMsRunningMu.RLock()
	defer s.macVMsRunningMu.RUnlock()
	return append([]string(nil), s.macVMsRunning...)
}

// resetMacVMsRunning replaces the entire macVMsRunning list atomically.
// Returns a snapshot and whether the list changed.
func (s *ParallelsService) resetMacVMsRunning(ids []string) ([]string, bool) {
	s.macVMsRunningMu.Lock()
	defer s.macVMsRunningMu.Unlock()

	changed := len(ids) != len(s.macVMsRunning)
	if !changed {
		existing := make(map[string]struct{}, len(s.macVMsRunning))
		for _, id := range s.macVMsRunning {
			existing[id] = struct{}{}
		}
		for _, id := range ids {
			if _, ok := existing[id]; !ok {
				changed = true
				break
			}
		}
	}

	s.macVMsRunning = ids
	return append([]string(nil), ids...), changed
}

func (s *ParallelsService) syncMacVmRunningStatus(ctx basecontext.ApiContext) {
	s.RLock()
	running := make([]string, 0)
	for _, vm := range s.cachedLocalVms {
		if (vm.OS == "macosx" || strings.Contains(strings.ToLower(vm.Name), "mac")) && vm.State == "running" {
			running = append(running, vm.ID)
		}
	}
	s.RUnlock()

	snapshot, changed := s.resetMacVMsRunning(running)
	ctx.LogInfof("[SyncMacVmRunningStatus] %d Mac VM(s) running: %v", len(snapshot), snapshot)
	if changed {
		if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
			_ = ee.BroadcastMessage(models.NewEventMessage(constants.EventTypePDFM, "MAC_VMS_RUNNING_NOW", models.MacVMsRunningNowEvent{
				MacVmsRunning: snapshot,
			}))
		}
	}
}

func escapeForBashC(command string) string {
	var escaped strings.Builder
	for i := 0; i < len(command); i++ {
		switch command[i] {
		case '"':
			if i == 0 || command[i-1] != '\\' {
				escaped.WriteString("\\\"")
			} else {
				escaped.WriteByte('"')
			}
		case '$':
			escaped.WriteString("\\$")
		default:
			escaped.WriteByte(command[i])
		}
	}
	result := escaped.String()

	return result
}

// extractSnapshotId extracts the snapshot ID from output string in format:
// "The snapshot with id {snapshot-id} has been successfully created."
func extractSnapshotId(output string) string {
	// Use regex to find content within curly braces
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
