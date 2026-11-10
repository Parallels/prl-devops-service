package parallelsdesktop

import (
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
)

// httpStatusServiceUnavailable is returned when a VM limit is configured but the host
// inventory cannot be enumerated, so the limit cannot be evaluated.
const httpStatusServiceUnavailable = 503

// CountVmsForMode returns how many virtual machines the per-host limit counts in the
// supplied inventory snapshot. In "running" mode only VMs occupying a hypervisor slot
// are counted; in "any" mode (the default) every VM registered on the host counts,
// stopped ones included. VMs are de-duplicated by ID.
func CountVmsForMode(vms []models.ParallelsVM, mode string) int64 {
	if mode == constants.MAX_VMS_PER_HOST_MODE_RUNNING {
		return int64(len(models.ActiveVMIDs(vms)))
	}

	seen := make(map[string]bool, len(vms))
	for _, vm := range vms {
		if vm.ID == "" {
			continue
		}
		seen[vm.ID] = true
	}

	return int64(len(seen))
}

// CountPendingVmSlots returns how many operations hold a slot that the inventory does
// not reflect yet. Each operation is charged once, to the thing it actually consumes:
// in "any" mode a create registers a VM that the inventory does not list yet, in
// "running" mode a start is on its way to a hypervisor slot the inventory does not
// show yet. A create in "running" mode and a start in "any" mode change nothing that is
// counted, so they consume nothing, which also keeps a create that starts its own VM
// from being charged twice.
//
// Caller holds vmAdmissionMu.
func countPendingVmSlots(createSlots, startSlots map[string]struct{}, mode string) int64 {
	slots := createSlots
	if mode == constants.MAX_VMS_PER_HOST_MODE_RUNNING {
		slots = startSlots
	}

	var pending int64
	for range slots {
		pending++
	}

	return pending
}

// liveVms returns the authoritative VM inventory, or an error when it cannot be
// enumerated. Tests inject vmInventoryFn to avoid shelling out to prlctl.
func (s *ParallelsService) liveVms(ctx basecontext.ApiContext) ([]models.ParallelsVM, error) {
	if s.vmInventoryFn != nil {
		return s.vmInventoryFn(ctx)
	}

	return s.getVmsInMachineForCurrentUser(ctx)
}

// ReserveVmSlot is the exported entry point used by the VM create and clone handlers
// to admit an operation against the per-host VM limit. See reserveVmSlot.
func (s *ParallelsService) ReserveVmSlot(ctx basecontext.ApiContext, operationID string) (func(), error) {
	return s.reserveVmSlot(ctx, operationID)
}

// reserveVmSlot admits a create or clone operation against the per-host VM limit and
// returns the func releasing the reservation. Admission and reservation are atomic
// within this host service, so a burst of concurrent creates cannot overshoot the cap.
// The release func is idempotent and must be called when the operation finishes.
//
// It is a no-op in "running" mode: a create registers a VM without occupying a
// hypervisor slot, and the start that follows passes reserveVmStartSlot, which is the
// authority for that mode. Charging the create as well would count one VM twice.
//
// When no limit is configured it is a no-op, which keeps the behaviour identical to
// hosts without the setting (and to older orchestrators).
func (s *ParallelsService) reserveVmSlot(ctx basecontext.ApiContext, operationID string) (func(), error) {
	cfg := config.Get()
	maxVms := cfg.MaxVmsPerHost()
	if maxVms <= 0 {
		return func() {}, nil
	}
	mode := cfg.MaxVmsPerHostMode()
	if mode == constants.MAX_VMS_PER_HOST_MODE_RUNNING {
		return func() {}, nil
	}

	vms, err := s.liveVms(ctx)
	if err != nil {
		// Fail closed: a limit that cannot be measured would otherwise be silently
		// unlimited, which defeats the purpose of a hard cap.
		ctx.LogErrorf("[VmCapacity] Create rejected op=%s: cannot enumerate VMs: %v", operationID, err)
		return nil, errors.NewWithCode("cannot verify host VM capacity: live host inventory failed", httpStatusServiceUnavailable)
	}

	key := operationID
	if key == "" {
		key = fmt.Sprintf("op-%d", time.Now().UnixNano())
	}

	s.vmAdmissionMu.Lock()
	current := CountVmsForMode(vms, mode)
	pending := countPendingVmSlots(s.vmCreateSlots, s.vmStartSlots, mode)
	if current+pending >= maxVms {
		s.vmAdmissionMu.Unlock()
		ctx.LogWarnf("[VmCapacity] Create rejected op=%s current=%d pending=%d maximum=%d mode=%s", key, current, pending, maxVms, mode)
		return nil, errors.NewWithCodef(409, "host VM limit reached: current %d, pending %d, maximum %d (mode %s)",
			current, pending, maxVms, mode)
	}
	if s.vmCreateSlots == nil {
		s.vmCreateSlots = make(map[string]struct{})
	}
	s.vmCreateSlots[key] = struct{}{}
	admitted := current + pending
	s.vmAdmissionMu.Unlock()

	ctx.LogInfof("[VmCapacity] Create admitted op=%s current=%d pending=%d maximum=%d mode=%s", key, current, admitted+1, maxVms, mode)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.vmAdmissionMu.Lock()
			delete(s.vmCreateSlots, key)
			s.vmAdmissionMu.Unlock()
			ctx.LogInfof("[VmCapacity] Create reservation released op=%s", key)
		})
	}, nil
}

// reserveVmStartSlot admits a start or resume against the per-host VM limit. It only
// applies in "running" mode: in "any" mode the VM is already registered and counted,
// so starting it consumes nothing new.
//
// Every VM that reaches a hypervisor slot passes this gate with a live inventory read,
// which makes it the authority for the mode; creates are deliberately not counted
// there, so a create that starts its own VM is charged exactly once.
func (s *ParallelsService) reserveVmStartSlot(ctx basecontext.ApiContext, id string) (func(), error) {
	cfg := config.Get()
	maxVms := cfg.MaxVmsPerHost()
	if maxVms <= 0 || cfg.MaxVmsPerHostMode() != constants.MAX_VMS_PER_HOST_MODE_RUNNING {
		return func() {}, nil
	}

	vms, err := s.liveVms(ctx)
	if err != nil {
		ctx.LogErrorf("[VmCapacity] Start rejected vm=%s: cannot enumerate VMs: %v", id, err)
		return nil, errors.NewWithCode("cannot verify host VM capacity: live host inventory failed", httpStatusServiceUnavailable)
	}

	// Already active (or mid-transition to active): its slot is already counted.
	for _, vm := range vms {
		if vm.ID == id && models.OccupiesVmSlot(vm.State) {
			return func() {}, nil
		}
	}

	s.vmAdmissionMu.Lock()
	current := CountVmsForMode(vms, constants.MAX_VMS_PER_HOST_MODE_RUNNING)
	pending := countPendingVmSlots(s.vmCreateSlots, s.vmStartSlots, constants.MAX_VMS_PER_HOST_MODE_RUNNING)
	if current+pending >= maxVms {
		s.vmAdmissionMu.Unlock()
		ctx.LogWarnf("[VmCapacity] Start rejected vm=%s current=%d pending=%d maximum=%d mode=%s",
			id, current, pending, maxVms, constants.MAX_VMS_PER_HOST_MODE_RUNNING)
		return nil, errors.NewWithCodef(409, "host VM limit reached: current %d, pending %d, maximum %d (mode %s)",
			current, pending, maxVms, constants.MAX_VMS_PER_HOST_MODE_RUNNING)
	}
	if s.vmStartSlots == nil {
		s.vmStartSlots = make(map[string]struct{})
	}
	s.vmStartSlots[id] = struct{}{}
	s.vmAdmissionMu.Unlock()

	ctx.LogInfof("[VmCapacity] Start admitted vm=%s current=%d pending=%d maximum=%d", id, current, pending+1, maxVms)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.vmAdmissionMu.Lock()
			delete(s.vmStartSlots, id)
			s.vmAdmissionMu.Unlock()
			ctx.LogInfof("[VmCapacity] Start reservation released vm=%s", id)
		})
	}, nil
}

// PendingVmSlots returns the number of operations holding a VM slot reservation that
// the inventory does not reflect yet. It is published to the orchestrator as the
// pending VM count so placement does not overbook a host between refreshes.
func (s *ParallelsService) PendingVmSlots() int64 {
	cfg := config.Get()
	if cfg.MaxVmsPerHost() <= 0 {
		return 0
	}

	s.vmAdmissionMu.Lock()
	defer s.vmAdmissionMu.Unlock()

	return countPendingVmSlots(s.vmCreateSlots, s.vmStartSlots, cfg.MaxVmsPerHostMode())
}
