package parallelsdesktop

import (
	"fmt"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

// pendingMacVMs avoids counting a reservation twice once the VM becomes active.
func (s *ParallelsService) pendingMacVMs(active []string) int64 {
	s.macAdmissionMu.Lock()
	defer s.macAdmissionMu.Unlock()
	return pendingMacVMCount(s.macReservations, active)
}

func pendingMacVMCount(reservations map[string]string, active []string) int64 {
	seen := make(map[string]bool, len(active))
	for _, id := range active {
		seen[id] = true
	}
	var count int64
	for _, id := range reservations {
		if !seen[id] {
			count++
			seen[id] = true
		}
	}
	return count
}

// Admission and reservation are atomic within this host service, including requests
// from different orchestrators. Always enumerate live VMs rather than the cache.
func (s *ParallelsService) reserveMacVMStart(ctx basecontext.ApiContext, id string) (func(), error) {
	s.macAdmissionMu.Lock()
	defer s.macAdmissionMu.Unlock()
	user, err := system.Get().GetCurrentUser(ctx)
	if err != nil || user != "root" {
		ctx.LogWarnf("[MacVMCapacity] Start rejected vm=%s: cannot enumerate all users", id)
		return nil, fmt.Errorf("cannot verify macOS VM capacity: host service must run as root to enumerate all users")
	}
	vms, err := s.getVmsInMachineForCurrentUser(ctx)
	if err != nil {
		ctx.LogErrorf("[MacVMCapacity] Live inventory failed vm=%s: %v", id, err)
		return nil, fmt.Errorf("cannot verify macOS VM capacity: live host inventory failed")
	}
	active := models.ActiveMacVMIDs(vms)
	if err := s.claimMacVMSlot(id, active); err != nil {
		ctx.LogWarnf("[MacVMCapacity] Start rejected vm=%s active_ids=%v: %v", id, active, err)
		return nil, err
	}
	ctx.LogInfof("[MacVMCapacity] Start admitted vm=%s active=%d pending=%d active_ids=%v", id, len(active), pendingMacVMCount(s.macReservations, active), active)
	var once sync.Once
	return func() {
		once.Do(func() {
			s.macAdmissionMu.Lock()
			delete(s.macReservations, id)
			s.macAdmissionMu.Unlock()
			ctx.LogInfof("[MacVMCapacity] Start reservation released vm=%s", id)
		})
	}, nil
}

// Caller holds macAdmissionMu.
func (s *ParallelsService) claimMacVMSlot(id string, active []string) error {
	if _, exists := s.macReservations[id]; exists {
		return fmt.Errorf("macOS VM %s already has a start in progress", id)
	}
	alreadyActive := false
	for _, activeID := range active {
		if activeID == id {
			alreadyActive = true
		}
	}
	pending := pendingMacVMCount(s.macReservations, active)
	if !alreadyActive && int64(len(active))+pending >= models.MaxRunningMacVMs {
		return fmt.Errorf("insufficient macOS VM capacity: active %d, pending starts %d, maximum %d", len(active), pending, models.MaxRunningMacVMs)
	}
	if s.macReservations == nil {
		s.macReservations = make(map[string]string)
	}
	s.macReservations[id] = id
	return nil
}
