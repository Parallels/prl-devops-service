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
	"syscall"
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
	cooldown         map[string]time.Time
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

		// Initialize maps for the debounce and cooldown logic
		// this will allow us to deduplicate events and prevent the echo loop
		// also will prevent multiple prlctl commands from being executed at the same time
		pending:  make(map[string]struct{}),
		inFlight: make(map[string]struct{}),
		cooldown: make(map[string]time.Time),

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

func (s *ParallelsService) updateVmInCache(ctx basecontext.ApiContext, newVm *models.ParallelsVM) {
	s.Lock()
	defer s.Unlock()
	found := false
	changeType := MeaningfulChange // Default to full broadcast for brand new VMs

	for i, cachedVm := range s.cachedLocalVms {
		if cachedVm.ID == newVm.ID {
			// RACE CONDITION PROTECTION:
			// If the UI just set the state to "starting" (Fast Path), but our slow prlctl list
			// finally returned and says it's "stopped", we KEEP the "starting" state because it's newer.
			isCachedTransitional := cachedVm.State == "starting" || cachedVm.State == "stopping" ||
				cachedVm.State == "resuming" || cachedVm.State == "suspending"

			isNewStatic := newVm.State == "stopped" || newVm.State == "running" ||
				newVm.State == "paused" || newVm.State == "suspended"

			if isCachedTransitional && isNewStatic {
				ctx.LogInfof("[ParallelsDesktop] [Event] Preserving newer transitional state '%s' over older static state '%s'", cachedVm.State, newVm.State)
				newVm.State = cachedVm.State // Preserve the intermediate state!
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

func (s *ParallelsService) updateVMIPInCache(ctx basecontext.ApiContext, vmID string) {
	status, err := s.VmStatus(ctx, vmID)
	if err != nil {
		ctx.LogErrorf("Failed to get VM status for IP update: %v", err)
		return
	}

	s.Lock()
	defer s.Unlock()
	for i, cachedVm := range s.cachedLocalVms {
		if cachedVm.ID == vmID {
			for j, ip := range s.cachedLocalVms[i].NetworkInformation.IPAddresses {
				if strings.ToLower(ip.Type) == "ipv4" {
					s.cachedLocalVms[i].NetworkInformation.IPAddresses[j].IP = status.IPConfigured
				}
				if (s.cachedLocalVms[i].State == "running" && s.cachedLocalVms[i].NetworkInformation.IPAddresses[j].IP == "-") &&
					(s.cachedLocalVms[i].OS == "macosx" || strings.Contains(s.cachedLocalVms[i].Name, "mac")) {
					cmd := helpers.Command{
						Command: s.executable,
						Args:    []string{"exec", s.cachedLocalVms[i].ID, "ipconfig", "getifaddr", "en0"},
					}.AsUser(s.cachedLocalVms[i].User)
					ctx.LogDebugf("Executing command to get internal ip address: %s", cmd.String())
					out, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
					if err == nil {
						ip := strings.TrimSpace(out)
						ctx.LogDebugf("Got internal IP address %s for VM %s using command execution", ip, vmID)
						if ip != "" {
							s.cachedLocalVms[i].InternalIpAddress = ip
							ctx.LogInfof("Updated VM internal IP in cache for VM %s to %s", vmID, ip)
						}
					} else {
						ctx.LogErrorf("Failed to get internal IP address for VM %s: %v", vmID, err)
					}
				}
				break
			}

			ctx.LogInfof("Updated VM IP in cache for VM %s to %s", vmID, status.IPConfigured)
			VmUpdatedEvent := models.VmUpdated{
				VmID:  vmID,
				NewVm: s.cachedLocalVms[i],
			}

			go func() {
				if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
					msg := models.NewEventMessage(constants.EventTypePDFM, "VM_UPDATED", VmUpdatedEvent)
					if err := ee.BroadcastMessage(msg); err != nil {
						ctx.LogErrorf("Error broadcasting VM updated event: %v", err)
					}
				}
			}()
			break
		}
	}
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
	}
	s.Unlock()
}

func (s *ParallelsService) getUserVm(ctx basecontext.ApiContext, username string, vmId string) ([]models.ParallelsVM, error) {
	// vmId can be empty to get all VMs for the user
	ctx.LogInfof("Getting VMs for user: %s", username)

	// TODO: workaround for parallels bug (PDFM-126209) where some fields are not returned when vm id is not specified
	vmIds := []string{}
	if vmId == "" {
		ctx.LogDebugf("Getting all VMs for user %s", username)
		var err error
		vmIds, err = s.getUserVmIds(ctx, username)
		if err != nil {
			return nil, err
		}
	} else {
		vmIds = append(vmIds, vmId)
	}

	externalIp, _ := system.Get().GetExternalIp(ctx)
	userMachines := []models.ParallelsVM{}
	for _, id := range vmIds {
		cmd := helpers.Command{
			Command: s.executable,
			Args:    []string{"list", id, "-a", "-i", "--json"},
		}.AsUser(username)
		ctx.LogDebugf("[ParallelsDesktop] Executing command: %s", cmd.String())
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		defer cancel()

		stdout, err := helpers.ExecuteWithNoOutput(timeoutCtx, cmd, helpers.ExecutionTimeout)
		if err != nil {
			return nil, err
		}
		vms := []models.ParallelsVM{}
		err = json.Unmarshal([]byte(stdout), &vms)
		if err != nil {
			return nil, err
		}
		vms[0].User = username
		userMachines = append(userMachines, vms...)
		s.setCooldown(id)
	}

	// updating the internal and external IP address
	for i := range userMachines {
		if externalIp != "" {
			userMachines[i].HostExternalIpAddress = externalIp
		}
		if len(userMachines[i].NetworkInformation.IPAddresses) > 0 {
			userMachines[i].InternalIpAddress = userMachines[i].NetworkInformation.IPAddresses[0].IP
		}

		if userMachines[i].InternalIpAddress == "" || userMachines[i].InternalIpAddress == "-" {
			// If the machine is running and it is a macos machine we will try to get the ip address by running a command
			// inside the machine
			if userMachines[i].State == "running" && (userMachines[i].OS == "macosx" || strings.Contains(userMachines[i].Name, "mac")) {
				cmd := helpers.Command{
					Command: s.executable,
					Args:    []string{"exec", userMachines[i].ID, "ipconfig", "getifaddr", "en0"},
				}.AsUser(username)
				ctx.LogDebugf("[ParallelsDesktop] Executing command to get internal ip address: %s", cmd.String())
				out, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
				if err == nil {
					ip := strings.TrimSpace(out)
					if ip != "" {
						userMachines[i].InternalIpAddress = ip
					}
				}
			}
		}
	}

	if vmId == "" {
		ctx.LogInfof("[ParallelsDesktop] User %s has %v VMs", username, len(userMachines))
	} else if vmId != "" && len(userMachines) > 0 {
		ctx.LogInfof("[ParallelsDesktop] User %s VM %s found", username, vmId)
	} else {
		ctx.LogInfof("[ParallelsDesktop] User %s VM %s not found", username, vmId)
	}
	return userMachines, nil
}

func (s *ParallelsService) getUserVmIds(ctx basecontext.ApiContext, username string) ([]string, error) {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"list", "-a", "-f", "--json"},
	}.AsUser(username)

	output, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	var status []models.VirtualMachineStatus
	err = json.Unmarshal([]byte(output), &status)
	if err != nil {
		return nil, err
	}
	listOfVms := make([]string, 0)
	for _, vm := range status {
		listOfVms = append(listOfVms, vm.UUID)
	}

	return listOfVms, nil
}

func (s *ParallelsService) GetCachedVms(ctx basecontext.ApiContext, filter string) ([]models.ParallelsVM, error) {
	ctx.LogInfof("Getting all VMs for all users with cache")
	var systemMachines []models.ParallelsVM
	var err error

	cfg := config.Get()
	if cfg.IsApi() || cfg.IsOrchestrator() {
		s.RLock()
		systemMachines = s.cachedLocalVms
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

func (s *ParallelsService) DeleteVm(ctx basecontext.ApiContext, id string) error {
	vm, err := s.findVmSync(ctx, id)
	if err != nil {
		return err
	}

	if vm == nil {
		return errors.Newf("VM with id %s was not found", id)
	}

	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
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

// CreateSnapshot creates a new snapshot for the specified VM
func (s *ParallelsService) CreateSnapshot(ctx basecontext.ApiContext, vmID string, request *models.CreateSnapShotRequest) (*models.CreateSnapShotResponse, error) {
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
		ctx.LogInfof("Creating snapshot '%s' for VM %s", request.SnapshotName, vmID)
	}
	if request.SnapshotDescription != "" {
		args = append(args, "-d", request.SnapshotDescription)
		ctx.LogInfof("Creating snapshot with description '%s' for VM %s", request.SnapshotDescription, vmID)
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

	return &models.CreateSnapShotResponse{
		SnapshotName: request.SnapshotName,
		SnapshotId:   snapshotId,
	}, nil
}

// DeleteSnapshot deletes a snapshot from the specified VM
func (s *ParallelsService) DeleteSnapshot(ctx basecontext.ApiContext, vmId string, snapshotId string, request *models.DeleteSnapshotRequest) error {
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

	ctx.LogInfof("Deleting snapshot %s for VM %s", snapshotId, vmId)

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

func (s *ParallelsService) RevertSnapshot(ctx basecontext.ApiContext, vmId string, snapshotId string, request *models.RevertSnapshotRequest) error {
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

	ctx.LogInfof("Reverting snapshot %s for VM %s", snapshotId, vmId)

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
func (s *ParallelsService) ListSnapshots(ctx basecontext.ApiContext, vmId string) (*models.ListSnapshotResponse, error) {
	vm, err := s.findVmSync(ctx, vmId)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.Newf("VM with id %s was not found", vmId)
	}

	ctx.LogInfof("Listing snapshots for VM %s", vmId)
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"snapshot-list", vmId, "--json"},
	}.AsUser(vm.User)

	output, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}

	// Parse the JSON which has snapshot IDs as keys
	var snapshotMap map[string]models.Snapshot
	err = json.Unmarshal([]byte(output), &snapshotMap)
	if err != nil && output != "" {
		return nil, errors.Newf("failed to parse snapshot list output: %v", err)
	}

	// Convert the map to a slice and set the ID field
	var snapshotList []models.Snapshot
	for id, snapshot := range snapshotMap {
		snapshot.ID = id
		snapshotList = append(snapshotList, snapshot)
	}

	snapshots := models.ListSnapshotResponse{
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
	cfg := config.Get()
	locationPath := cfg.GetKey(constants.VIRTUAL_MACHINES_FOLDER_ENV_VAR)
	if locationPath != "" {
		return locationPath, nil
	}

	fmt.Printf("%s\n", locationPath)

	if s.Users != nil || len(s.Users) == 0 {
		_, err := s.GetUsers(ctx)
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
	// Get the file owner
	sys := fileInfo.Sys().(*syscall.Stat_t)
	uid := sys.Uid
	gid := int(sys.Gid)

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
