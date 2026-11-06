package parallelsdesktop

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
)

// IP resolution status values reported on ParallelsVM.IpStatus.
const (
	IpStatusPending   = "pending"
	IpStatusResolving = "resolving"
	IpStatusResolved  = "resolved"
	IpStatusFailed    = "failed"
)

const (
	// ipProbeBaseDelay is the delay after the first attempt to resolve a VM IP.
	// Each subsequent round adds ipProbeBackoffStep, capped at ipProbeMaxDelay.
	ipProbeBaseDelay   = 10 * time.Second
	ipProbeBackoffStep = 10 * time.Second
	ipProbeMaxDelay    = 60 * time.Second
	// ipProbeTimeout bounds one worker's total lifetime: we stop poking (and mark
	// the VM ip_status=failed) once this much time has passed since the worker start.
	ipProbeTimeout = 280 * time.Second
	// ipProbeExecTimeout is the short per-attempt timeout for prlctl exec probes.
	// The poke loop itself plays the role of waiting for guest readiness.
	ipProbeExecTimeout = 5 * time.Second
	// ipProbeAttemptCooldown deduplicates attempts triggered by several sources at
	// once (startup sweep, state events, live reads).
	ipProbeAttemptCooldown = 5 * time.Second
	// guestIpInterfaces are tried in order via `ipconfig getifaddr` on macOS guests.
	// en0 is the primary NIC, but docks/multi-NIC VMs may carry the lease on en1/en2.
	macIpProbeInterfaces = "en0 en1 en2"
)

// vmIsMacOS reports whether the VM is a macOS guest.
func vmIsMacOS(vm models.ParallelsVM) bool {
	return vm.OS == "macosx" || strings.Contains(strings.ToLower(vm.Name), "mac")
}

// vmToolsProbablyInstalled treats an unknown/empty tools state as "probably installed"
// (bulk prlctl lists often omit it) and only skips probing when the state explicitly
// reports the tools as not installed.
func vmToolsProbablyInstalled(vm models.ParallelsVM) bool {
	state := strings.ToLower(vm.GuestTools.State)
	state = strings.NewReplacer(" ", "", "_", "", "-", "").Replace(state)
	if state == "" {
		return true
	}
	return !strings.Contains(state, "notinstalled")
}

// shouldProbeIp decides whether the IP poke worker is worth running for a VM:
//   - any VM: running, no IP yet, tools installed or probably installed.
//   - macOS VMs additionally: running with an IP that we have not confirmed via
//     `prlctl exec` (ip_status != resolved) — prlctl's own network info is often
//     missing or stale for macOS guests, so we always want the authoritative one.
func shouldProbeIp(vm models.ParallelsVM) bool {
	if !strings.EqualFold(vm.State, "running") || !vmToolsProbablyInstalled(vm) {
		return false
	}
	if vm.InternalIpAddress == "" {
		return true
	}
	return vmIsMacOS(vm) && vm.IpStatus != IpStatusResolved
}

// preserveResolvedIp merges a cached VM's IP and resolution status into fresh
// prlctl-sourced data: when the live entry has no IP we keep the cached one (and
// its status), instead of clobbering an exec-resolved IP with prlctl silence.
func preserveResolvedIp(cached, live models.ParallelsVM) models.ParallelsVM {
	if live.InternalIpAddress == "" {
		if cached.InternalIpAddress != "" &&
			strings.EqualFold(live.State, "running") &&
			strings.EqualFold(cached.State, "running") {
			live.InternalIpAddress = cached.InternalIpAddress
			// Keep any IPv4 entry in NetworkInformation consistent too.
			for j, addr := range live.NetworkInformation.IPAddresses {
				if strings.ToLower(addr.Type) == "ipv4" {
					live.NetworkInformation.IPAddresses[j].IP = cached.InternalIpAddress
					break
				}
			}
		}
		if live.IpStatus == "" {
			live.IpStatus = cached.IpStatus
			live.IpStatusDetail = cached.IpStatusDetail
		}
	}
	return live
}

// ensureIpProbeMapsLocked lazily initializes the probe bookkeeping so that services
// built directly (e.g. in tests) do not nil-panic. Caller holds syncMu.
func (s *ParallelsService) ensureIpProbeMapsLocked() {
	if s.ipLastFetched == nil {
		s.ipLastFetched = make(map[string]time.Time)
	}
	if s.ipPokeWorkers == nil {
		s.ipPokeWorkers = make(map[string]context.CancelFunc)
	}
}

// snapshotVm returns a copy of the cached VM entry for vmID.
func (s *ParallelsService) snapshotVm(vmID string) (models.ParallelsVM, bool) {
	s.RLock()
	defer s.RUnlock()
	for _, vm := range s.cachedLocalVms {
		if vm.ID == vmID {
			return vm, true
		}
	}
	return models.ParallelsVM{}, false
}

// startIpPokeWorker arms (at most one) poke worker per VM. It is safe to call from
// any goroutine and from any trigger (startup sweep, state events, cache refresh,
// live reads): repeats while a worker is active are no-ops.
func (s *ParallelsService) startIpPokeWorker(ctx basecontext.ApiContext, vmID string) {
	vm, ok := s.snapshotVm(vmID)
	if !ok || !shouldProbeIp(vm) {
		return
	}

	baseCtx := s.listenerCtx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	s.syncMu.Lock()
	s.ensureIpProbeMapsLocked()
	if _, active := s.ipPokeWorkers[vmID]; active {
		s.syncMu.Unlock()
		ctx.LogDebugf("[ParallelsDesktop] [IP] Poke worker already active for VM %s", vmID)
		return
	}
	workerCtx, cancel := context.WithCancel(baseCtx)
	s.ipPokeWorkers[vmID] = cancel
	s.syncMu.Unlock()

	ctx.LogInfof("[ParallelsDesktop] [IP] Arming IP poke worker for VM %s (%s)", vmID, vm.Name)
	go s.runIpPokeWorker(ctx, workerCtx, vmID)
}

// stopIpPokeWorker cancels and forgets the worker for a VM (no status change).
func (s *ParallelsService) stopIpPokeWorker(vmID string) {
	s.syncMu.Lock()
	if cancel, ok := s.ipPokeWorkers[vmID]; ok {
		delete(s.ipPokeWorkers, vmID)
		s.syncMu.Unlock()
		cancel()
		return
	}
	s.syncMu.Unlock()
}

// stopAllIpPokeWorkers tears down every worker (service shutdown).
func (s *ParallelsService) stopAllIpPokeWorkers() {
	s.syncMu.Lock()
	workers := s.ipPokeWorkers
	s.ipPokeWorkers = make(map[string]context.CancelFunc)
	s.syncMu.Unlock()
	for _, cancel := range workers {
		cancel()
	}
}

// NotifyMissingIps sweeps the given VMs and arms poke workers for eligible ones,
// staggering starts so we never flood macOS guests with simultaneous exec probes.
// Fire-and-forget.
func (s *ParallelsService) NotifyMissingIps(vms []models.ParallelsVM) {
	go func() {
		ctx := s.ctx
		if ctx == nil {
			ctx = basecontext.NewRootBaseContext()
		}
		stagger := 0
		for _, vm := range vms {
			if !shouldProbeIp(vm) {
				continue
			}
			delay := time.Duration(stagger) * 3 * time.Second
			stagger++
			go func(vmID string, d time.Duration) {
				if d > 0 {
					time.Sleep(d)
				}
				s.startIpPokeWorker(ctx, vmID)
			}(vm.ID, delay)
		}
		ctx.LogDebugf("[ParallelsDesktop] [IP] NotifyMissingIps scheduled %d probe workers", stagger)
	}()
}

// ipProbeWaitDefault blocks for d or until the worker context is cancelled.
// Returns false when cancelled.
func ipProbeWaitDefault(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// runIpPokeWorker is the per-VM poke loop: attempt immediately, then retry with a
// linearly growing delay (10s -> 60s cap) until the IP is resolved or the total
// 280s lifetime expires, at which point the VM is marked ip_status=failed.
func (s *ParallelsService) runIpPokeWorker(ctx basecontext.ApiContext, workerCtx context.Context, vmID string) {
	defer func() {
		s.syncMu.Lock()
		delete(s.ipPokeWorkers, vmID)
		s.syncMu.Unlock()
	}()

	now := time.Now
	if s.ipProbeNow != nil {
		now = s.ipProbeNow
	}
	wait := ipProbeWaitDefault
	if s.ipProbeWait != nil {
		wait = s.ipProbeWait
	}
	attempt := func(vmID string) (string, bool) {
		if s.ipProbeAttempt != nil {
			return s.ipProbeAttempt(ctx, vmID)
		}
		return s.attemptResolveVMIP(ctx, vmID)
	}

	start := now()
	delay := ipProbeBaseDelay
	attempts := 0
	first := true

	giveUp := func() {
		detail := fmt.Sprintf("no IP address obtained after %d attempts within %v", attempts, ipProbeTimeout)
		ctx.LogWarnf("[ParallelsDesktop] [IP] Giving up on IP poke for VM %s: %s", vmID, detail)
		s.markIpStatus(ctx, vmID, IpStatusFailed, detail)
	}

	for {
		select {
		case <-workerCtx.Done():
			return
		default:
		}

		if now().Sub(start) >= ipProbeTimeout {
			giveUp()
			return
		}

		vm, ok := s.snapshotVm(vmID)
		if !ok || vm.IpStatus == IpStatusResolved || !shouldProbeIp(vm) {
			// Resolved elsewhere, VM stopped/removed, or tools explicitly missing.
			ctx.LogDebugf("[ParallelsDesktop] [IP] Stopping poke worker for VM %s (no longer eligible)", vmID)
			return
		}

		if first {
			first = false
			s.markIpStatus(ctx, vmID, IpStatusResolving, "")
		}

		attempts++
		if ip, resolved := attempt(vmID); resolved {
			// No-op when the real attempt already applied it; required when a
			// test hook supplies the resolution.
			s.applyResolvedIp(ctx, vmID, ip)
			ctx.LogInfof("[ParallelsDesktop] [IP] Poke worker resolved internal IP for VM %s after %d attempt(s): %s", vmID, attempts, ip)
			return
		}

		// Sleep the current round's delay (base 10s for the first retry), capped
		// by the remaining budget; the next round's delay grows afterwards.
		remaining := ipProbeTimeout - now().Sub(start)
		if remaining <= 0 {
			giveUp()
			return
		}
		sleepFor := delay
		if sleepFor > remaining {
			sleepFor = remaining
		}
		if !wait(workerCtx, sleepFor) {
			return
		}
		delay += ipProbeBackoffStep
		if delay > ipProbeMaxDelay {
			delay = ipProbeMaxDelay
		}
	}
}

// attemptResolveVMIP performs ONE short IP resolution attempt and, on success,
// applies it to the cache and broadcasts VM_UPDATED. It never long-waits: guest
// readiness is handled by the poke cadence itself.
func (s *ParallelsService) attemptResolveVMIP(ctx basecontext.ApiContext, vmID string) (string, bool) {
	vm, ok := s.snapshotVm(vmID)
	if !ok {
		return "", false
	}

	// Short cooldown so multiple triggers cannot stack prlctl exec probes.
	s.syncMu.Lock()
	s.ensureIpProbeMapsLocked()
	if last, seen := s.ipLastFetched[vmID]; seen && time.Since(last) < ipProbeAttemptCooldown {
		s.syncMu.Unlock()
		return "", false
	}
	s.ipLastFetched[vmID] = time.Now()
	s.syncMu.Unlock()

	var ip string
	if vmIsMacOS(vm) {
		guessed, err := s.execGuestIp(ctx, vm)
		if err != nil || guessed == "" {
			if err != nil {
				ctx.LogDebugf("[ParallelsDesktop] [IP] exec probe failed for VM %s: %v", vmID, err)
			}
			return "", false
		}
		ip = guessed
	} else {
		// Non-macOS guests report their IP through prlctl directly once tools do.
		live, err := s.getVmInMachine(ctx, vmID)
		if err != nil || live == nil {
			return "", false
		}
		ip = live.InternalIpAddress
		if ip == "" && len(live.NetworkInformation.IPAddresses) > 0 {
			ip = live.NetworkInformation.IPAddresses[0].IP
		}
		if ip == "" {
			return "", false
		}
	}

	s.applyResolvedIp(ctx, vmID, ip)
	return ip, true
}

// execGuestIp runs `prlctl exec <vmID> ipconfig getifaddr <if>` for each candidate
// interface, returning the first non-empty lease. Each probe uses a short timeout.
func (s *ParallelsService) execGuestIp(ctx basecontext.ApiContext, vm models.ParallelsVM) (string, error) {
	if vm.User == "" {
		return "", fmt.Errorf("VM %s has no owner user, cannot exec", vm.ID)
	}
	var lastErr error
	for _, iface := range strings.Fields(macIpProbeInterfaces) {
		cmd := helpers.Command{
			Command: s.executable,
			Args:    []string{"exec", vm.ID, "ipconfig", "getifaddr", iface},
		}.AsUser(vm.User)
		ctx.LogDebugf("[ParallelsDesktop] [IP] Probe %s for VM %s: %s", iface, vm.ID, cmd.String())
		if s.sshSem != nil {
			s.sshSem <- struct{}{}
		}
		out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, ipProbeExecTimeout)
		if s.sshSem != nil {
			<-s.sshSem
		}
		if err != nil {
			lastErr = err
			continue
		}
		if ip := strings.TrimSpace(out); ip != "" {
			return ip, nil
		}
	}
	return "", lastErr
}

// resetIpStatusForStop clears a stale resolution status when the VM leaves the
// running state (the last known IP itself is kept for display purposes). No
// broadcast — the accompanying state change is broadcast separately.
func (s *ParallelsService) resetIpStatusForStop(ctx basecontext.ApiContext, vmID string) {
	s.mutateCachedVm(vmID, func(entry *models.ParallelsVM) {
		entry.IpStatus = ""
		entry.IpStatusDetail = ""
	})
}

// applyResolvedIp writes a freshly resolved IP into the cache (InternalIpAddress
// plus any IPv4 entry in NetworkInformation), marks ip_status=resolved and
// broadcasts VM_UPDATED with the changed VM.
func (s *ParallelsService) applyResolvedIp(ctx basecontext.ApiContext, vmID, ip string) {
	updated := s.mutateCachedVm(vmID, func(entry *models.ParallelsVM) {
		entry.InternalIpAddress = ip
		for j, addr := range entry.NetworkInformation.IPAddresses {
			if strings.ToLower(addr.Type) == "ipv4" {
				entry.NetworkInformation.IPAddresses[j].IP = ip
				break // found and updated the IPv4 entry — done
			}
		}
		entry.IpStatus = IpStatusResolved
		entry.IpStatusDetail = ""
	})
	if updated == nil {
		return
	}
	ctx.LogInfof("[ParallelsDesktop] [IP] Resolved internal IP for VM %s: %s", vmID, ip)
	s.broadcastVmUpdated(ctx, *updated)
}

// markIpStatus updates ip_status/ip_status_detail on the cached VM and broadcasts
// VM_UPDATED only when something actually changed.
func (s *ParallelsService) markIpStatus(ctx basecontext.ApiContext, vmID, status, detail string) *models.ParallelsVM {
	updated := s.mutateCachedVm(vmID, func(entry *models.ParallelsVM) {
		entry.IpStatus = status
		entry.IpStatusDetail = detail
	})
	if updated == nil {
		return nil
	}
	s.broadcastVmUpdated(ctx, *updated)
	return updated
}

// mutateCachedVm applies mutate to the cached entry under the write lock and
// returns a copy of the changed VM, or nil when the VM is unknown or unchanged.
func (s *ParallelsService) mutateCachedVm(vmID string, mutate func(*models.ParallelsVM)) *models.ParallelsVM {
	s.Lock()
	defer s.Unlock()
	for i, vm := range s.cachedLocalVms {
		if vm.ID == vmID {
			before := vm
			mutate(&s.cachedLocalVms[i])
			if reflect.DeepEqual(before, s.cachedLocalVms[i]) {
				return nil
			}
			changed := s.cachedLocalVms[i]
			return &changed
		}
	}
	return nil
}

// broadcastVmUpdated emits the VM_UPDATED PDFM event for a VM snapshot.
func (s *ParallelsService) broadcastVmUpdated(ctx basecontext.ApiContext, updated models.ParallelsVM) {
	go func() {
		if ee := eventemitter.Get(); ee != nil && ee.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypePDFM, "VM_UPDATED", models.VmUpdated{
				VmID:  updated.ID,
				NewVm: updated,
			})
			if err := ee.BroadcastMessage(msg); err != nil {
				ctx.LogErrorf("[ParallelsDesktop] [IP] Error broadcasting VM_UPDATED: %v", err)
			}
		}
	}()
}
