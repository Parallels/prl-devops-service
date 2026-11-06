package parallelsdesktop

import (
	"reflect"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

const (
	CacheFindNumberOfRetries = 10
	// CacheFindRetryBaseDelay is the first delay between cache probes in
	// findVmInCacheAndSystem. Each following delay doubles until
	// CacheFindRetryMaxDelay, so repeated misses back off instead of
	// hammering the cache every fixed interval.
	CacheFindRetryBaseDelay = 250 * time.Millisecond
	CacheFindRetryMaxDelay  = 2 * time.Second
	// CacheFindRetryBudget caps the total time spent waiting for the async
	// cache refresh before we fall through to the full system scan.
	CacheFindRetryBudget = 5 * time.Second
	// CacheFindNotFoundTTL is how long a lookup that was confirmed missing
	// (absent from cache AND from the system scan) keeps short-circuiting
	// subsequent lookups. Without it, consumers such as the reverse proxy
	// re-run the whole retry + full-scan cycle for a deleted VM at a very
	// fast pace. The entry is cleared early when the VM reappears in cache
	// or is registered again (VM_ADDED).
	CacheFindNotFoundTTL = 15 * time.Second
)

// vmNotFoundKey normalizes a name-or-id lookup key for the not-found backoff.
func vmNotFoundKey(idOrName string) string {
	return strings.ToLower(strings.TrimSpace(idOrName))
}

// findVmInCache probes the in-memory VM cache for a name or ID (case-insensitive).
func (s *ParallelsService) findVmInCache(idOrName string) *models.ParallelsVM {
	s.RLock()
	vms := s.cachedLocalVms
	s.RUnlock()
	for _, vm := range vms {
		if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
			return &vm
		}
	}
	return nil
}

// vmNotFoundBackoffActive reports whether idOrName was confirmed missing from
// both cache and system less than CacheFindNotFoundTTL ago. Expired entries are
// evicted on sight.
func (s *ParallelsService) vmNotFoundBackoffActive(idOrName string) bool {
	key := vmNotFoundKey(idOrName)
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	at, ok := s.notFoundVm[key]
	if !ok {
		return false
	}
	if time.Since(at) >= CacheFindNotFoundTTL {
		delete(s.notFoundVm, key)
		return false
	}
	return true
}

// markVmNotFound records that idOrName was just confirmed missing from the
// whole system, starting the not-found backoff window.
func (s *ParallelsService) markVmNotFound(idOrName string) {
	key := vmNotFoundKey(idOrName)
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	if s.notFoundVm == nil {
		s.notFoundVm = make(map[string]time.Time)
	}
	s.notFoundVm[key] = time.Now()
}

// clearVmNotFoundBackoff forgets the confirmed-miss record for idOrName.
func (s *ParallelsService) clearVmNotFoundBackoff(idOrName string) {
	key := vmNotFoundKey(idOrName)
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	delete(s.notFoundVm, key)
}

// waitCacheFindRetry sleeps the current backoff step. Tests may override
// s.findProbeWait to observe the schedule without waiting in real time.
func (s *ParallelsService) waitCacheFindRetry(d time.Duration) {
	if s.findProbeWait != nil {
		s.findProbeWait(d)
		return
	}
	time.Sleep(d)
}

// GetVmFast resolves a VM without waiting for the async cache refresh: one
// cache probe, the confirmed-miss backoff check, then a single full system
// scan. Callers on latency-sensitive paths — such as reverse proxy route
// resolution running inside a start operation with a timeout — use this
// instead of GetVm so a missing VM costs one scan, not the retry loop too.
func (s *ParallelsService) GetVmFast(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	return s.findVmWithCacheRefresh(ctx, idOrName, 0)
}

func (s *ParallelsService) findVmInCacheAndSystem(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	return s.findVmWithCacheRefresh(ctx, idOrName, CacheFindRetryBudget)
}

// findVmWithCacheRefresh looks the VM up in cache, waiting up to retryBudget
// for the event-driven cache refresh to land (exponential backoff between
// probes), then falls back to one full system scan. A retryBudget of 0 skips
// the waiting entirely and goes straight to the scan.
func (s *ParallelsService) findVmWithCacheRefresh(ctx basecontext.ApiContext, idOrName string, retryBudget time.Duration) (*models.ParallelsVM, error) {
	// Cheap first probe against the current cache.
	if vm := s.findVmInCache(idOrName); vm != nil {
		return vm, nil
	}
	// The VM was recently confirmed gone (e.g. deleted): don't repeat the
	// expensive retry + full-scan cycle at a fast pace. Fail fast; the TTL
	// or a VM_ADDED event reopens the lookup.
	if s.vmNotFoundBackoffActive(idOrName) {
		ctx.LogDebugf("[ParallelsDesktop] VM %v recently confirmed missing, backing off (skip cache retries)", idOrName)
		return nil, ErrVirtualMachineNotFound
	}

	delay := CacheFindRetryBaseDelay
	deadline := time.Now().Add(retryBudget)
	for i := 0; retryBudget > 0 && i < CacheFindNumberOfRetries; i++ {
		if vm := s.findVmInCache(idOrName); vm != nil {
			s.clearVmNotFoundBackoff(idOrName)
			return vm, nil
		}
		if i == CacheFindNumberOfRetries-1 {
			break
		}
		if delay > CacheFindRetryMaxDelay {
			delay = CacheFindRetryMaxDelay
		}
		left := deadline.Sub(time.Now())
		if delay > left {
			ctx.LogInfof("[ParallelsDesktop] Retry budget (%v) for VM %v exhausted after %v probes, refreshing from system", retryBudget, idOrName, i+1)
			break
		}
		ctx.LogInfof("[ParallelsDesktop] VM with name or id %v not found in cache, retrying in %v... (probe %v, retry budget %v left)", idOrName, delay, i+1, left.Round(100*time.Millisecond))
		s.waitCacheFindRetry(delay)
		delay *= 2
	}
	// To fetch all virtual machines (VMs), we intentionally call `GetVms` without specifying an ID or name.
	// This process might take some time because our goal is to refresh the entire cache.
	// If we find any VMs, it means the current cache is outdated, and we will proceed to update it.
	scanVms := s.GetVms
	if s.findSystemVms != nil {
		scanVms = s.findSystemVms
	}
	vms, err := scanVms(ctx, "")
	if err == nil {
		for _, vm := range vms {
			if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
				ctx.LogWarnf("[ParallelsDesktop] VM is not present in cache but found in machine updating cache")
				s.Lock()
				s.cachedLocalVms = vms
				s.Unlock()
				s.clearVmNotFoundBackoff(idOrName)
				return &vm, nil
			}
		}
		// The system scan succeeded and the VM is genuinely absent — remember
		// it so repeated lookups (reverse proxy routes, polling clients) back
		// off instead of re-running this whole cycle. A failed scan (err != nil)
		// is *not* evidence of absence, so it never marks the VM missing.
		s.markVmNotFound(idOrName)
	}
	return nil, ErrVirtualMachineNotFound
}

func (s *ParallelsService) findVmSync(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	return s.findVmInCacheAndSystem(ctx, idOrName)
}

type ChangeType int

const (
	NoChange ChangeType = iota
	OnlyUptimeChanged
	MeaningfulChange
)

// evaluateVmChanges classifies the difference between two VM states.
func (s *ParallelsService) evaluateVmChanges(oldVm, newVm models.ParallelsVM) ChangeType {
	// 1. If they are exactly identical, do nothing.
	if reflect.DeepEqual(oldVm, newVm) {
		return NoChange
	}

	// 2. Make copies to strip out volatile fields
	oldCopy := oldVm
	newCopy := newVm

	// BLANK OUT NOISY FIELDS (Adjust 'Uptime' to your actual field name)
	oldCopy.Uptime = ""
	newCopy.Uptime = ""
	// oldCopy.CpuUsage = 0
	// newCopy.CpuUsage = 0

	// 3. If stripping the uptime makes them identical, ONLY the uptime changed!
	if reflect.DeepEqual(oldCopy, newCopy) {
		return OnlyUptimeChanged
	}

	// 4. Otherwise, something real changed (IP, State, RAM, etc.)
	return MeaningfulChange
}

// unflattenVMSnapshots takes a flat list of snapshots and builds a tree structure based on Parent field.
func unflattenVMSnapshots(input []models.VMSnapshot) []models.VMSnapshot {
	// 1. Group snapshots by Parent ID for O(N) access
	childrenMap := make(map[string][]models.VMSnapshot)
	for _, snap := range input {
		childrenMap[snap.Parent] = append(childrenMap[snap.Parent], snap)
	}

	// 2. Recursive function to build the tree
	var buildTree func(parentID string) []models.VMSnapshot
	buildTree = func(parentID string) []models.VMSnapshot {
		children := childrenMap[parentID]
		// For each child in this list, populate its own children recursively
		for i := range children {
			children[i].Children = buildTree(children[i].ID)
		}
		return children
	}

	// 3. Start building from root nodes (those with empty Parent)
	return buildTree("")
}
