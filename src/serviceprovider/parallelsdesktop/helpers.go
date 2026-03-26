package parallelsdesktop

import (
	"reflect"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
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
		ctx.LogInfof("[ParallelsDesktop] VM with name or id %v not found in cache, retrying... (%d/%d)", idOrName, i+1, CacheFindNumberOfRetries)
		time.Sleep(CacheFindRetryDelay)
	}
	// To fetch all virtual machines (VMs), we intentionally call `GetVms` without specifying an ID or name.
	// This process might take some time because our goal is to refresh the entire cache.
	// If we find any VMs, it means the current cache is outdated, and we will proceed to update it.
	vms, err := s.GetVms(ctx, "")
	if err == nil {
		for _, vm := range vms {
			if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
				ctx.LogWarnf("[ParallelsDesktop] VM is not present in cache but found in machine updating cache")
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
