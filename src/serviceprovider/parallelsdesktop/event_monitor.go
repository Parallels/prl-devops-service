package parallelsdesktop

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	diskspaceservice "github.com/Parallels/prl-devops-service/serviceprovider/diskSpace"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
)

func (s *ParallelsService) listenToParallelsEvents(ctx basecontext.ApiContext) {
	// Lock this check so concurrent startup calls don't spawn duplicate listeners!
	s.Lock()
	if s.eventsProcessing {
		s.Unlock()
		return
	}
	s.eventsProcessing = true
	s.Unlock()

	ctx.LogInfof("[ParallelsDesktop] [Events] Setting up Parallels events listener for all relevant users")

	users, err := s.getFilteredUsers(ctx)
	if err != nil {
		ctx.LogErrorf("[ParallelsDesktop] [Events] Failed to get filtered users: %v", err)

		s.Lock()
		s.eventsProcessing = false
		s.Unlock()

		return
	}

	if len(users) == 0 {
		ctx.LogWarnf("[ParallelsDesktop] [Events] No users found for event listening")

		s.Lock()
		s.eventsProcessing = false
		s.Unlock()

		return
	}

	// Spawn a monitor-events process for EVERY relevant user
	for _, user := range users {
		go func(u models.SystemUser) {
			helpersCmd := helpers.Command{
				Command: s.executable,
				Args:    []string{"monitor-events", "--json"},
			}.AsUser(u.Username)

			// Convert to exec.Cmd for processLauncher
			cmd := exec.CommandContext(s.listenerCtx, helpersCmd.Command, helpersCmd.Args...)

			// Use a PTY to avoid buffering
			file, err := s.processLauncher.Start(cmd)
			if err != nil {
				ctx.LogErrorf("[ParallelsDesktop] [Events] Error starting monitor-events PTY for user %s: %v\n", u.Username, err)
				return
			}
			defer file.Close()

			reader := bufio.NewReader(file)
			for {
				select {
				case <-s.listenerCtx.Done():
					ctx.LogInfof("[ParallelsDesktop] [Events] Stopping Parallels events listener for user %s", u.Username)
					if cmd.Process != nil {
						_ = cmd.Process.Kill()
					}
					_ = cmd.Wait()
					return
				default:
					line, err := reader.ReadString('\n')
					if err != nil {
						if err != io.EOF {
							ctx.LogErrorf("[ParallelsDesktop] [Events] Error reading output for user %s: %v\n", u.Username, err)
						}
						// Break out of the loop if the PTY dies
						break
					}

					var event models.ParallelsServiceEvent
					if err := json.Unmarshal([]byte(line), &event); err != nil {
						ctx.LogDebugf("[ParallelsDesktop] [Events] Non-JSON output: %s", line) // Optional debug
						continue
					}

					// Push to our global processing channel
					eventsChannel <- event
				}
			}
		}(user) // Pass the loop variable safely
	}

	// Start the single background worker to pull from eventsChannel
	s.processEventsChannel(ctx)
}

func (s *ParallelsService) processEventsChannel(ctx basecontext.ApiContext) {
	go func() {
		for {
			select {
			case <-s.listenerCtx.Done():
				ctx.LogInfof("[ParallelsDesktop] [Events] Stopping Parallels events processor")
				return
			case event := <-eventsChannel:
				s.processEvent(ctx, event)
			}
		}
	}()
}

func (s *ParallelsService) startAutoCacheRefresh(ctx basecontext.ApiContext) {
	if !s.eventsProcessing {
		s.ctx.LogInfof("[ParallelsDesktop] [Events] eventsProcessing is false, not starting auto cache refresh")
		return
	}
	cfg := config.Get()
	interval := cfg.ForceCacheRefreshInterval()
	ctx.LogInfof("[ParallelsDesktop] [Events] Starting auto cache refresh for Parallels VMs every %v", interval)
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-s.listenerCtx.Done():
				ctx.LogInfof("[ParallelsDesktop] [Events] Stopping auto cache refresh for Parallels VMs")
				ticker.Stop()
				return
			case <-ticker.C:
				s.RLock()
				// Make a safe copy of the cache to compare against
				cachedVMs := make([]models.ParallelsVM, len(s.cachedLocalVms))
				copy(cachedVMs, s.cachedLocalVms)
				s.RUnlock()

				currentVMs, err := s.GetVms(ctx, "")
				if err != nil {
					ctx.LogErrorf("[ParallelsDesktop] [Events] Error getting current VMs: %v", err)
					continue
				}

				needsRefresh := false

				// Check 1: Length mismatch (VM added or deleted)
				if len(cachedVMs) != len(currentVMs) {
					ctx.LogWarnf(
						"[ParallelsDesktop] [Events] This shouldn't happen: Cached VMs count %d does not match current VMs count %d, refreshing cache",
						len(cachedVMs),
						len(currentVMs),
					)
					needsRefresh = true
				} else {
					// Check 2: State mismatch
					for _, cachedVM := range cachedVMs {
						for _, currentVM := range currentVMs {
							if cachedVM.ID == currentVM.ID && cachedVM.State != currentVM.State {
								ctx.LogWarnf(
									"[ParallelsDesktop] [Events] This shouldn't happen: Cached VM %s state %s does not match current VM state %s, refreshing cache",
									cachedVM.ID,
									cachedVM.State,
									currentVM.State,
								)
								needsRefresh = true
								break // Break inner loop
							}
						}
						if needsRefresh {
							break // Break outer loop to avoid redundant logs
						}
					}
				}

				// If we found a mismatch in either check, execute the safe refresh
				if needsRefresh {
					// 1. SAFELY SWAP THE CACHE (No nested function calls!)
					s.Lock()
					s.cachedLocalVms = currentVMs
					s.Unlock()

					// 2. BROADCAST UPDATES (Without holding the lock)
					go func(vmsToBroadcast []models.ParallelsVM) {
						ee := eventemitter.Get()
						if ee != nil && ee.IsRunning() {
							for _, vm := range vmsToBroadcast {
								_ = ee.BroadcastMessage(models.NewEventMessage(constants.EventTypePDFM, "VM_UPDATED", models.VmUpdated{
									VmID:  vm.ID,
									NewVm: vm,
								}))
							}
						}
					}(currentVMs)
				}
			}
		}
	}()
}

func (s *ParallelsService) StopListeners() {
	if s.eventsProcessing {
		s.ctx.LogInfof("[ParallelsDesktop] [Event] Stopping all Parallels event listeners and workers")
		s.cancelFunc() // This safely kills both the event tailer AND the Debounce Worker
		s.eventsProcessing = false
	}
}

func (s *ParallelsService) processEvent(ctx basecontext.ApiContext, event models.ParallelsServiceEvent) {
	switch event.EventName {
	case "vm_state_changed":
		s.processVmStateChanged(ctx, event)
	case "vm_added":
		s.processVmAdded(ctx, event)
	case "vm_unregistered", "vm_deleted":
		s.processVmUnregistered(ctx, event)
	case "vm_config_changed":
		s.processVmConfigChanged(ctx, event)
	case "vm_tools_state_changed":
		s.processVmToolsStateChanged(ctx, event)
	case "vm_snapshots_tree_changed":
		s.processVmSnapshotsTreeChanged(ctx, event)
	default:
		ctx.LogInfof("[ParallelsDesktop] [Event] Unhandled event: %v", event)
	}
}

func (s *ParallelsService) processVmStateChanged(ctx basecontext.ApiContext, event models.ParallelsServiceEvent) {
	if event.AdditionalInfo == nil || event.AdditionalInfo.VmStateName == "" {
		return
	}

	newState := event.AdditionalInfo.VmStateName
	var prevState string
	found := false

	s.Lock() // Grab write lock for cache
	for i, vm := range s.cachedLocalVms {
		if vm.ID == event.VMID {
			prevState = vm.State
			if prevState == newState {
				s.Unlock()
				return // Nothing to do
			}

			ctx.LogInfof("[ParallelsDesktop] [Event] VM %s state changed: %s -> %s", vm.ID, prevState, newState)
			s.cachedLocalVms[i].State = newState

			// Add to fast state updates timestamp
			s.fastStateUpdates[vm.ID] = time.Now()

			found = true
			break
		}
	}
	s.Unlock()

	// If we missed the vm_added event, fallback to the slow debouncer
	if !found {
		ctx.LogWarnf("[ParallelsDesktop] [Event] VM %s not in cache during state change, falling back to full sync", event.VMID)
		s.queueVmForSync(event.VMID, event.EventName)
		return
	}

	// Instantly tell the UI the VM is "resuming", "stopping", etc!
	go func() {
		if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
			ctx.LogInfof("[ParallelsDesktop] [Event] Broadcasting VM state changed event for VM %s", event.VMID)
			_ = ee.BroadcastMessage(models.NewEventMessage(constants.EventTypePDFM, "VM_STATE_CHANGED", models.VmStateChange{
				PreviousState: prevState, CurrentState: newState, VmID: event.VMID,
			}))
		}
	}()

	// When a macOS VM finishes starting up, kick off async IP resolution.
	// waitForVMSSHReady will probe SSH readiness before fetching the IP,
	// so this is safe to call immediately without any extra delay.
	if newState == "running" {
		s.RLock()
		var isMacOS bool
		var vmName string
		for _, vm := range s.cachedLocalVms {
			if vm.ID == event.VMID {
				isMacOS = vm.OS == "macosx" || strings.Contains(strings.ToLower(vm.Name), "mac")
				vmName = vm.Name
				break
			}
		}
		s.RUnlock()
		if isMacOS {
			ctx.LogInfof("[ParallelsDesktop] [IP] VM %s (%s) is now running, scheduling IP resolution", event.VMID, vmName)
			go s.updateVMIPInCache(ctx, event.VMID)
		}
	}
}

func (s *ParallelsService) processVmAdded(ctx basecontext.ApiContext, event models.ParallelsServiceEvent) {
	machine, err := s.getVmInMachine(ctx, event.VMID)
	if err != nil {
		ctx.LogErrorf("[ParallelsDesktop] [Events] Failed to get VM in machine: %v", err)
		return
	}
	s.Lock()
	s.cachedLocalVms = append(s.cachedLocalVms, *machine)
	s.Unlock()
	ctx.LogInfof("[ParallelsDesktop] [Events] Added VM %s to cache", event.VMID)
	VmAddedEvent := models.VmAdded{
		VmID:  event.VMID,
		NewVm: *machine,
	}

	go func() {
		if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypePDFM, "VM_ADDED", VmAddedEvent)
			if err := ee.BroadcastMessage(msg); err != nil {
				ctx.LogErrorf("[ParallelsDesktop] [Events] Error broadcasting VM added event: %v", err)
			}
		}
		diskspaceservice.Get(ctx).CheckDiskSpaceAndBroadcast()
	}()

}

func (s *ParallelsService) processVmUnregistered(ctx basecontext.ApiContext, event models.ParallelsServiceEvent) {
	for i, vm := range s.cachedLocalVms {
		if vm.ID == event.VMID {
			s.Lock()
			s.cachedLocalVms = append(s.cachedLocalVms[:i], s.cachedLocalVms[i+1:]...)
			s.Unlock()
			ctx.LogInfof("[ParallelsDesktop] [Events] Removed VM %s from cache", event.VMID)

			VmRemoved := models.VmRemoved{
				VmID: event.VMID,
			}
			go func() {
				if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
					msg := models.NewEventMessage(constants.EventTypePDFM, "VM_REMOVED", VmRemoved)
					if err := ee.BroadcastMessage(msg); err != nil {
						ctx.LogErrorf("[ParallelsDesktop] [Events] Error broadcasting VM removed event: %v", err)
					}
				}
				diskspaceservice.Get(ctx).CheckDiskSpaceAndBroadcast()
			}()

			break
		}
	}
}

func (s *ParallelsService) processVmConfigChanged(ctx basecontext.ApiContext, event models.ParallelsServiceEvent) {
	ctx.LogInfof("[ParallelsDesktop] [Events] VM %s config changed, queuing for sync", event.VMID)
	s.queueVmForSync(event.VMID, event.EventName)
}

func (s *ParallelsService) processVmToolsStateChanged(ctx basecontext.ApiContext, event models.ParallelsServiceEvent) {
	ctx.LogInfof("[ParallelsDesktop] [Events] VM %s tools state changed, queuing for sync", event.VMID)
	s.queueVmForSync(event.VMID, event.EventName)
}

func (s *ParallelsService) processVmSnapshotsTreeChanged(ctx basecontext.ApiContext, event models.ParallelsServiceEvent) {
	VMSnapshots, err := s.listVMSnapshots(ctx, event.VMID)
	if err != nil {
		ctx.LogErrorf("[parallelsdesktop][snapshots] Failed to get snapshots for VM %s: %v", event.VMID, err)
		return
	}
	if s.databaseService == nil {
		ctx.LogErrorf("[parallelsdesktop][snapshots] Database service not available")
		return
	}
	var dtoVMSnaps []data_models.VMSnapshot
	if VMSnapshots != nil {
		for _, snap := range VMSnapshots.Snapshots {
			dtoVMSnaps = append(dtoVMSnaps, data_models.VMSnapshot{
				ID:      snap.ID,
				Name:    snap.Name,
				Date:    snap.Date,
				State:   snap.State,
				Current: snap.Current,
				Parent:  snap.Parent,
			})
		}
	}
	s.databaseService.SetListVMSnapshotsByVMId(event.VMID, data_models.VMSnapshots{
		VMId:       event.VMID,
		VMSnapshot: dtoVMSnaps,
	})

	VmSnapshotsUpdatedEvent := models.VmSnapshotsUpdated{
		VmID:        event.VMID,
		VMSnapshots: mappers.VMSnapshotsDtoToApi(dtoVMSnaps),
	}

	go func() {
		if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypePDFM, "VM_SNAPSHOTS_UPDATED", VmSnapshotsUpdatedEvent)
			if err := ee.BroadcastMessage(msg); err != nil {
				ctx.LogErrorf("[parallelsdesktop][snapshots] Error broadcasting VM snapshots updated event: %v", err)
			}
		}
	}()
}

// queueVmForSync safely adds a VM to the processing queue while protecting against echo loops.
func (s *ParallelsService) queueVmForSync(vmID string, eventName string) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	// The Two-Stage Shield: We only apply this to config_changed events,
	// as these are the ones that echo when we run prlctl list.
	if eventName == "vm_config_changed" {

		// Stage 1: The "During" Shield. If we are currently running a command, drop the echo.
		if _, active := s.inFlight[vmID]; active {
			s.ctx.LogDebugf("[ParallelsDesktop] [Debounce] Dropped config_changed event for VM %s (Currently in-flight)", vmID)
			return
		}

		// Stage 2: The "After" Shield. If we just finished a command < X ago, drop the echo.
		if lastSync, exists := s.cooldown[vmID]; exists {
			if time.Since(lastSync) < cooldownDelay {
				s.ctx.LogDebugf("[ParallelsDesktop] [Debounce] Dropped config_changed event for VM %s (In cooldown)", vmID)
				return
			}
		}
	}

	// Queue it up! (If it's already there, this just safely overwrites it - zero memory growth)
	s.ctx.LogDebugf("[ParallelsDesktop] [Debounce] Queuing config_changed event for VM %s", vmID)
	s.pending[vmID] = struct{}{}
}

// startDebounceWorker runs continuously in the background, checking the queue every second.
func (s *ParallelsService) startDebounceWorker() {
	ticker := time.NewTicker(eventWorkerTicker)
	defer ticker.Stop()

	for {
		select {
		case <-s.listenerCtx.Done():
			s.ctx.LogInfof("[ParallelsDesktop] [Debounce] Stopping Debounce Worker")
			return
		case <-ticker.C:
			s.processPendingSyncs()
		}
	}
}

// processPendingSyncs moves VMs from pending to inFlight and spawns the prlctl jobs.
func (s *ParallelsService) processPendingSyncs() {
	s.syncMu.Lock()
	toProcess := make([]string, 0)

	for vmID := range s.pending {
		// Only sync if this VM isn't ALREADY running a prlctl list
		if _, active := s.inFlight[vmID]; !active {
			toProcess = append(toProcess, vmID)
			s.inFlight[vmID] = struct{}{} // Mark as busy
			delete(s.pending, vmID)       // Remove from queue
		}
	}
	s.syncMu.Unlock()

	// Run all the pending syncs concurrently
	for _, vmID := range toProcess {
		go s.syncVmTask(vmID)
	}
}

// syncVmTask handles the slow CLI execution and cache updating.
func (s *ParallelsService) syncVmTask(vmID string) {
	// Record exactly when we started asking PD for data
	cmdStartTime := time.Now()

	// 1. Fetch the new state (The slow path)
	vm, err := s.getVmInMachine(s.ctx, vmID)

	// 2. Update the cache
	if err == nil {
		s.ctx.LogDebugf("[ParallelsDesktop] [Debounce] Updating VM %s in cache", vm.ID)
		s.updateVmInCache(s.ctx, vm, cmdStartTime)
	} else {
		s.ctx.LogErrorf("[ParallelsDesktop] [Debounce] Failed to get VM during debounce sync: %v", err)
	}

	// 3. Cleanup & Arm the Shield
	s.syncMu.Lock()
	delete(s.inFlight, vmID)      // Unblock the VM
	s.cooldown[vmID] = time.Now() // Start the x time echo shield
	s.syncMu.Unlock()
}

// setCooldown is a helper so our 2-minute refresh loop can also arm the shield
func (s *ParallelsService) setCooldown(vmID string) {
	s.syncMu.Lock()
	s.cooldown[vmID] = time.Now()
	s.syncMu.Unlock()
}
