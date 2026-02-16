package parallelsdesktop

import (
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

const CacheFindNumberOfRetries = 10
const CacheFindRetryDelay = 500 * time.Millisecond

func (s *ParallelsService) findVmInCacheAndSystem(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {

	for i := 0; i < CacheFindNumberOfRetries; i++ {
		s.RLock()
		vms := s.cachedLocalVms
		s.RUnlock()
		for _, vm := range vms {
			if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
				return &vm, nil
			}
		}
		ctx.LogInfof("VM with name or id %v not found in cache, retrying... (%d/%d)", idOrName, i+1, CacheFindNumberOfRetries)
		time.Sleep(CacheFindRetryDelay)
	}
	// To fetch all virtual machines (VMs), we intentionally call `GetVms` without specifying an ID or name.
	// This process might take some time because our goal is to refresh the entire cache.
	// If we find any VMs, it means the current cache is outdated, and we will proceed to update it.
	vms, err := s.GetVms(ctx, "")
	if err == nil {
		for _, vm := range vms {
			if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
				ctx.LogWarnf("Vm is not present in cache but found in machine updating cache")
				s.Lock()
				s.cachedLocalVms = vms
				s.Unlock()
				return &vm, nil
			}
		}
	}
	return nil, ErrVirtualMachineNotFound
}

func (s *ParallelsService) findVmSync(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	return s.findVmInCacheAndSystem(ctx, idOrName)
}

// findVmSnapshotsInCacheAndSystem finds VM snapshots in cache first, then in system if not found
func (s *ParallelsService) findVmSnapshotsInCacheAndSystem(ctx basecontext.ApiContext, vmId string) ([]data_models.VirtualMachineSnapshot, error) {
	if vmId == "" {
		return nil, ErrVirtualMachineNotFound
	}

	for i := 0; i < CacheFindNumberOfRetries; i++ {
		s.RLock()
		snapshots, exists := s.cachedVmSnapshots[vmId]
		s.RUnlock()

		if exists {
			ctx.LogDebugf("Found %d snapshots for VM %s in cache", len(snapshots), vmId)
			return snapshots, nil
		}

		ctx.LogInfof("Snapshots for VM %v not found in cache, retrying... (%d/%d)", vmId, i+1, CacheFindNumberOfRetries)
		time.Sleep(CacheFindRetryDelay)
	}

	// If not found in cache, try to get from system and update cache
	ctx.LogInfof("Loading snapshots for VM %s from system", vmId)
	snapshotResponse, err := s.ListSnapshots(ctx, vmId)
	if err != nil {
		return nil, err
	}

	// Convert API response to data models
	var snapshots []data_models.VirtualMachineSnapshot
	if snapshotResponse != nil {
		for id, details := range *snapshotResponse {
			snapshot := data_models.VirtualMachineSnapshot{
				ID:      id,
				VMId:    vmId,
				Name:    details.Name,
				Date:    details.Date,
				State:   details.State,
				Current: details.Current,
				Parent:  details.Parent,
			}
			snapshots = append(snapshots, snapshot)
		}
	}

	// Update cache
	s.Lock()
	if s.cachedVmSnapshots == nil {
		s.cachedVmSnapshots = make(map[string][]data_models.VirtualMachineSnapshot)
	}
	s.cachedVmSnapshots[vmId] = snapshots
	s.Unlock()

	ctx.LogDebugf("Cached %d snapshots for VM %s", len(snapshots), vmId)
	return snapshots, nil
}

// findVmSnapshotsSync synchronously finds VM snapshots
func (s *ParallelsService) findVmSnapshotsSync(ctx basecontext.ApiContext, vmId string) ([]data_models.VirtualMachineSnapshot, error) {
	return s.findVmSnapshotsInCacheAndSystem(ctx, vmId)
}

// refreshVmSnapshotsCache refreshes the snapshot cache for a specific VM
func (s *ParallelsService) refreshVmSnapshotsCache(ctx basecontext.ApiContext, vmId string) error {
	if vmId == "" {
		return ErrVirtualMachineNotFound
	}

	ctx.LogDebugf("Refreshing snapshot cache for VM %s", vmId)

	// Clear existing cache for this VM
	s.clearVmSnapshotsCache(vmId)

	// Load fresh data from system
	_, err := s.findVmSnapshotsInCacheAndSystem(ctx, vmId)
	return err
}

// clearVmSnapshotsCache clears the snapshot cache for a specific VM
func (s *ParallelsService) clearVmSnapshotsCache(vmId string) {
	s.Lock()
	defer s.Unlock()

	if s.cachedVmSnapshots != nil {
		delete(s.cachedVmSnapshots, vmId)
	}
}

// clearAllVmSnapshotsCache clears the entire VM snapshots cache
func (s *ParallelsService) clearAllVmSnapshotsCache() {
	s.Lock()
	defer s.Unlock()

	s.cachedVmSnapshots = make(map[string][]data_models.VirtualMachineSnapshot)
}

// getCachedVmSnapshots returns a copy of cached snapshots for a specific VM
func (s *ParallelsService) getCachedVmSnapshots(vmId string) ([]data_models.VirtualMachineSnapshot, bool) {
	s.RLock()
	defer s.RUnlock()

	if s.cachedVmSnapshots == nil {
		return nil, false
	}

	snapshots, exists := s.cachedVmSnapshots[vmId]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	result := make([]data_models.VirtualMachineSnapshot, len(snapshots))
	copy(result, snapshots)
	return result, true
}

// addVmSnapshotToCache adds a snapshot to the cache for a specific VM
func (s *ParallelsService) addVmSnapshotToCache(vmId string, snapshot data_models.VirtualMachineSnapshot) {
	s.Lock()
	defer s.Unlock()

	if s.cachedVmSnapshots == nil {
		s.cachedVmSnapshots = make(map[string][]data_models.VirtualMachineSnapshot)
	}

	s.cachedVmSnapshots[vmId] = append(s.cachedVmSnapshots[vmId], snapshot)
}

// removeVmSnapshotFromCache removes a snapshot from the cache for a specific VM
func (s *ParallelsService) removeVmSnapshotFromCache(vmId string, snapshotId string) {
	s.Lock()
	defer s.Unlock()

	if s.cachedVmSnapshots == nil {
		return
	}

	snapshots, exists := s.cachedVmSnapshots[vmId]
	if !exists {
		return
	}

	// Find and remove the snapshot
	for i, snapshot := range snapshots {
		if snapshot.ID == snapshotId {
			s.cachedVmSnapshots[vmId] = append(snapshots[:i], snapshots[i+1:]...)
			break
		}
	}
}

// updateVmSnapshotInCache updates a snapshot in the cache for a specific VM
func (s *ParallelsService) updateVmSnapshotInCache(vmId string, updatedSnapshot data_models.VirtualMachineSnapshot) {
	s.Lock()
	defer s.Unlock()

	if s.cachedVmSnapshots == nil {
		return
	}

	snapshots, exists := s.cachedVmSnapshots[vmId]
	if !exists {
		return
	}

	// Find and update the snapshot
	for i, snapshot := range snapshots {
		if snapshot.ID == updatedSnapshot.ID {
			s.cachedVmSnapshots[vmId][i] = updatedSnapshot
			break
		}
	}
}
