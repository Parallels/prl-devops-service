package parallelsdesktop

import (
	"errors"
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	serviceErrors "github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
)

func vmCapacityInventory(vms []models.ParallelsVM, err error) func(basecontext.ApiContext) ([]models.ParallelsVM, error) {
	return func(basecontext.ApiContext) ([]models.ParallelsVM, error) {
		return vms, err
	}
}

func TestCountVmsForMode(t *testing.T) {
	vms := []models.ParallelsVM{
		{ID: "one", State: "running"},
		{ID: "one", State: "running"},
		{ID: "two", State: "stopped"},
		{ID: "three", State: "paused"},
		{ID: "", State: "running"},
	}

	if got := CountVmsForMode(vms, constants.MAX_VMS_PER_HOST_MODE_ANY); got != 3 {
		t.Fatalf("any mode must count every registered VM once, got %d", got)
	}
	if got := CountVmsForMode(vms, constants.MAX_VMS_PER_HOST_MODE_RUNNING); got != 2 {
		t.Fatalf("running mode must count slot occupants only, got %d", got)
	}
	if got := CountVmsForMode(nil, constants.MAX_VMS_PER_HOST_MODE_ANY); got != 0 {
		t.Fatalf("empty inventory must count zero, got %d", got)
	}
}

func TestCountPendingVmSlots(t *testing.T) {
	creates := map[string]struct{}{"create-a": {}, "create-b": {}}
	starts := map[string]struct{}{"start-a": {}}

	if got := countPendingVmSlots(creates, starts, constants.MAX_VMS_PER_HOST_MODE_ANY); got != 2 {
		t.Fatalf("any mode must charge creates only, got %d", got)
	}
	if got := countPendingVmSlots(creates, starts, constants.MAX_VMS_PER_HOST_MODE_RUNNING); got != 1 {
		t.Fatalf("running mode must charge starts only, got %d", got)
	}
}

func TestCountPendingCapacity(t *testing.T) {
	slots := map[string]capacityCharge{
		"op-a": {cpu: 2, memory: 1024},
		"op-b": {cpu: 4, memory: 2048},
	}
	if got := countPendingCapacity(slots); got.cpu != 6 || got.memory != 3072 {
		t.Fatalf("every in-flight charge must count, got %+v", got)
	}
	if got := countPendingCapacity(nil); got.cpu != 0 || got.memory != 0 {
		t.Fatalf("no charges must sum to zero, got %+v", got)
	}
}

// No configured limit must stay completely inert: no inventory walk, no reservation,
// which is what keeps hosts without the setting behaving exactly as before.
func TestReserveVmSlotWithoutLimitIsNoOp(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{vmInventoryFn: func(basecontext.ApiContext) ([]models.ParallelsVM, error) {
		t.Fatal("the VM limit is disabled, the inventory must not be walked")
		return nil, nil
	}}

	release, err := s.reserveVmSlot(ctx, "op")
	if err != nil {
		t.Fatal(err)
	}
	release()
	if s.PendingVmSlots() != 0 {
		t.Fatalf("no reservation may be held without a limit, got %d", s.PendingVmSlots())
	}
}

func TestReserveVmSlotRejectsAtLimit(t *testing.T) {
	t.Setenv(constants.MAX_VMS_PER_HOST_ENV_VAR, "2")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{vmInventoryFn: vmCapacityInventory([]models.ParallelsVM{
		{ID: "one", State: "stopped"},
		{ID: "two", State: "running"},
	}, nil)}

	release, err := s.reserveVmSlot(ctx, "op")
	if err == nil {
		release()
		t.Fatal("a third VM admitted on a host limited to two")
	}
	sysErr, ok := err.(*serviceErrors.SystemError)
	if !ok || sysErr.Code() != 409 {
		t.Fatalf("expected a 409 system error, got %T %v", err, err)
	}
	if msg := sysErr.Message(); msg != "host VM limit reached: current 2, pending 0, maximum 2 (mode any)" {
		t.Fatalf("unexpected message: %s", msg)
	}
}

// In "running" mode a create consumes nothing: the start that follows is the gate, so a
// create that starts its own VM must still fit into the last slot instead of being
// charged twice and rejected.
func TestReserveVmSlotRunningModeChargesStartsOnly(t *testing.T) {
	t.Setenv(constants.MAX_VMS_PER_HOST_ENV_VAR, "1")
	t.Setenv(constants.MAX_VMS_PER_HOST_MODE_ENV_VAR, constants.MAX_VMS_PER_HOST_MODE_RUNNING)
	ctx := basecontext.NewBaseContext()

	s := &ParallelsService{vmInventoryFn: func(basecontext.ApiContext) ([]models.ParallelsVM, error) {
		t.Fatal("a create consumes no slot in running mode, the inventory must not be walked")
		return nil, nil
	}}
	release, err := s.reserveVmSlot(ctx, "op")
	if err != nil {
		t.Fatal(err)
	}
	release()

	// The VM is registered but stopped: starting it takes the only slot.
	startable := &ParallelsService{vmInventoryFn: vmCapacityInventory([]models.ParallelsVM{{ID: "one", State: "stopped"}}, nil)}
	release, err = startable.reserveVmStartSlot(ctx, "one")
	if err != nil {
		t.Fatalf("create followed by its own start rejected on the last slot: %v", err)
	}
	// Until the VM is active the in-flight start is what keeps a second one out.
	if got := startable.PendingVmSlots(); got != 1 {
		t.Fatalf("the in-flight start must be published as pending, got %d", got)
	}
	if _, err := startable.reserveVmStartSlot(ctx, "one"); err == nil {
		t.Fatal("duplicate start admitted")
	}

	release()
	if got := startable.PendingVmSlots(); got != 0 {
		t.Fatalf("released start still pending: %d", got)
	}
}

// A burst of concurrent creates on an idle host must be admitted only up to the cap:
// the reservations cover the window where the new VMs are not in the inventory yet.
func TestReserveVmSlotConcurrentBurstRespectsLimit(t *testing.T) {
	t.Setenv(constants.MAX_VMS_PER_HOST_ENV_VAR, "3")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{vmInventoryFn: vmCapacityInventory([]models.ParallelsVM{{ID: "one", State: "stopped"}}, nil)}

	var wg sync.WaitGroup
	admitted := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.reserveVmSlot(ctx, ""); err != nil {
				return
			}
			admitted <- struct{}{}
			// The slot stays held: the VM is still being created.
		}()
	}
	wg.Wait()
	close(admitted)

	if got := len(admitted); got != 2 {
		t.Fatalf("expected the 2 remaining slots to be admitted, got %d", got)
	}
}

// A limit that cannot be measured must not silently become unlimited.
func TestReserveVmSlotFailsClosedWithoutInventory(t *testing.T) {
	t.Setenv(constants.MAX_VMS_PER_HOST_ENV_VAR, "1")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{vmInventoryFn: vmCapacityInventory(nil, errors.New("prlctl failed"))}

	if _, err := s.reserveVmSlot(ctx, "op"); err == nil {
		t.Fatal("create admitted while the inventory is unknown")
	} else if sysErr, ok := err.(*serviceErrors.SystemError); !ok || sysErr.Code() != httpStatusServiceUnavailable {
		t.Fatalf("expected a 503 system error, got %T %v", err, err)
	}
}

func TestReserveVmSlotReleaseFreesCapacity(t *testing.T) {
	t.Setenv(constants.MAX_VMS_PER_HOST_ENV_VAR, "1")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{vmInventoryFn: vmCapacityInventory(nil, nil)}

	release, err := s.reserveVmSlot(ctx, "op")
	if err != nil {
		t.Fatal(err)
	}
	if s.PendingVmSlots() != 1 {
		t.Fatalf("the in-flight create must be published as pending, got %d", s.PendingVmSlots())
	}
	if _, err := s.reserveVmSlot(ctx, "op-2"); err == nil {
		t.Fatal("second create admitted on a host limited to one")
	}

	release()
	release() // the release func is idempotent
	if s.PendingVmSlots() != 0 {
		t.Fatalf("released slot still counted: %d", s.PendingVmSlots())
	}
	release, err = s.reserveVmSlot(ctx, "op-3")
	if err != nil {
		t.Fatal(err)
	}
	release()
}

// Starts only consume a slot in "running" mode; in "any" mode the VM is already
// registered and therefore already counted.
func TestReserveVmStartSlotAppliesToRunningModeOnly(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	inventory := vmCapacityInventory([]models.ParallelsVM{{ID: "one", State: "stopped"}}, nil)

	t.Setenv(constants.MAX_VMS_PER_HOST_ENV_VAR, "1")
	t.Setenv(constants.MAX_VMS_PER_HOST_MODE_ENV_VAR, constants.MAX_VMS_PER_HOST_MODE_ANY)
	s := &ParallelsService{vmInventoryFn: inventory}
	release, err := s.reserveVmStartSlot(ctx, "one")
	if err != nil {
		t.Fatal(err)
	}
	release()

	t.Setenv(constants.MAX_VMS_PER_HOST_MODE_ENV_VAR, constants.MAX_VMS_PER_HOST_MODE_RUNNING)
	running := &ParallelsService{vmInventoryFn: vmCapacityInventory([]models.ParallelsVM{
		{ID: "one", State: "running"},
		{ID: "two", State: "stopped"},
	}, nil)}
	if _, err := running.reserveVmStartSlot(ctx, "two"); err == nil {
		t.Fatal("start admitted over the running-mode limit")
	}

	// Starting an already active VM consumes nothing new.
	active := &ParallelsService{vmInventoryFn: vmCapacityInventory([]models.ParallelsVM{{ID: "one", State: "running"}}, nil)}
	release, err = active.reserveVmStartSlot(ctx, "one")
	if err != nil {
		t.Fatal(err)
	}
	release()
}

// --- CPU/memory admission gate -------------------------------------------------------------------

// hostCapacity injects the machine totals the gate compares against, which is the same
// call GetHardwareUsage takes its Total from.
func hostCapacity(hardware *models.SystemHardwareInfo, err error) func(basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
	return func(basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
		return hardware, err
	}
}

// idleHost is a host with cpuCount logical CPUs and memoryMiB of RAM, no VMs on it.
func idleHost(cpuCount int, memoryMiB float64) func(basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
	return hostCapacity(&models.SystemHardwareInfo{LogicalCpuCount: cpuCount, MemorySize: memoryMiB}, nil)
}

func vmWithHardware(id, state string, cpus int64, memory string) models.ParallelsVM {
	return models.ParallelsVM{
		ID:    id,
		State: state,
		Hardware: models.Hardware{
			CPU:    models.CPU{Cpus: cpus},
			Memory: models.Memory{Size: memory},
		},
	}
}

// enableVmCapacity turns the opt-in gate on and pins the host's own share so the
// arithmetic in the assertions stays readable: 1 CPU and 1 GiB kept for the host, so a
// host built with 5 CPUs has exactly 4 free. The config getters treat 0 as "unset" and
// fall back to the defaults, which is why a test cannot simply reserve nothing.
func enableVmCapacity(t *testing.T, ratio string) {
	t.Helper()
	t.Setenv(constants.VMCPUAdmissionEnabledEnvVar, "true")
	t.Setenv(constants.SYSTEM_RESERVED_CPU_ENV_VAR, "1")
	t.Setenv(constants.SYSTEM_RESERVED_MEMORY_ENV_VAR, "1024")
	if ratio != "" {
		t.Setenv(constants.VMCPUOvercommitRatioEnvVar, ratio)
	}
}

// With the gate off the host must read nothing, hold nothing and publish nothing: an
// upgrade has to leave a deployment that oversubscribes on purpose untouched.
func TestReserveVmCapacityDisabledIsNoOp(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{
		vmInventoryFn: func(basecontext.ApiContext) ([]models.ParallelsVM, error) {
			t.Fatal("the gate is disabled, the VM inventory must not be walked")
			return nil, nil
		},
		hostCapacityFn: func(basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
			t.Fatal("the gate is disabled, the machine totals must not be read")
			return nil, nil
		},
		// Charges could only be left behind by a gate that was on and got switched
		// off; they must never surface as pending capacity.
		vmCapacitySlots: map[string]capacityCharge{"stale": {cpu: 4, memory: 2048}},
	}

	release, err := s.ReserveVmCapacity(ctx, "op", 64, 131072)
	if err != nil {
		t.Fatal(err)
	}
	release()
	if got := s.PendingCpu(); got != 0 {
		t.Fatalf("a disabled gate must publish no pending CPU, got %d", got)
	}
	if got := s.PendingMemory(); got != 0 {
		t.Fatalf("a disabled gate must publish no pending memory, got %v", got)
	}
}

// A create whose specs cannot be read is admitted on the count gate alone, without even
// walking the inventory: refusing a create over a number we cannot measure would turn a
// safeguard into an outage.
func TestReserveVmCapacityWithoutSpecsAdmitsCountOnly(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{
		vmInventoryFn: func(basecontext.ApiContext) ([]models.ParallelsVM, error) {
			t.Fatal("an unmeasurable charge admits count-only, the inventory must not be walked")
			return nil, nil
		},
		hostCapacityFn: func(basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
			t.Fatal("an unmeasurable charge admits count-only, the machine totals must not be read")
			return nil, nil
		},
	}

	release, err := s.ReserveVmCapacity(ctx, "op", 0, 0)
	if err != nil {
		t.Fatalf("a create without readable specs was blocked: %v", err)
	}
	release()
}

// Two concurrent 4-CPU creates on a host with 4 free CPUs: only one fits. The charge
// covers the window where neither VM is in the inventory yet, which is the race the VM
// count gate cannot see.
func TestReserveVmCapacityConcurrentBurstRespectsCpuLimit(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()
	// 5 logical CPUs with 1 kept for the host = 4 free; memory is ample so CPU binds.
	s := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, nil),
		hostCapacityFn: idleHost(5, 32768),
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	admitted := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := s.ReserveVmCapacity(ctx, "", 4, 0); err != nil {
				return
			}
			admitted <- struct{}{}
			// The charge stays held: the VM is still being created.
		}()
	}
	close(start)
	wg.Wait()
	close(admitted)

	if got := len(admitted); got != 1 {
		t.Fatalf("expected exactly one 4-CPU create to fit into 4 free CPUs, got %d", got)
	}
}

// The overcommit ratio is the knob for a deployment that oversubscribes on purpose: at
// 2 the host admits twice its free CPUs, and both of the creates above land.
func TestReserveVmCapacityOvercommitRatioAdmitsBoth(t *testing.T) {
	enableVmCapacity(t, "2")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, nil),
		hostCapacityFn: idleHost(5, 32768),
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	admitted := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := s.ReserveVmCapacity(ctx, "", 4, 0); err != nil {
				return
			}
			admitted <- struct{}{}
		}()
	}
	close(start)
	wg.Wait()
	close(admitted)

	if got := len(admitted); got != 2 {
		t.Fatalf("an overcommit ratio of 2 must admit both 4-CPU creates, got %d", got)
	}
}

// Only the VMs on the hypervisor are charged: a running VM's CPUs are gone from the
// pool, a stopped one's are not, which is what makes the gate agree with the
// TotalAvailable the orchestrator places from.
func TestReserveVmCapacityChargesRunningVmsOnly(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{
		vmInventoryFn: vmCapacityInventory([]models.ParallelsVM{
			vmWithHardware("running-one", "running", 2, "1024 MB"),
			vmWithHardware("stopped-one", "stopped", 4, "4096 MB"),
		}, nil),
		hostCapacityFn: idleHost(5, 32768),
	}

	release, err := s.ReserveVmCapacity(ctx, "op", 2, 0)
	if err != nil {
		t.Fatalf("the 2 CPUs left by the running VM were refused: %v", err)
	}
	release()

	if _, err := s.ReserveVmCapacity(ctx, "op", 3, 0); err == nil {
		t.Fatal("3 CPUs admitted while the running VM holds 2 of the 4 free")
	}
}

// The refusal has to name what was asked for and what was there, the way the VM count
// refusal does, or an operator cannot tell an over-sized request from a broken gate.
func TestReserveVmCapacityRefusalNamesRequiredAndAvailable(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{
		vmInventoryFn: vmCapacityInventory([]models.ParallelsVM{
			vmWithHardware("running-one", "running", 2, "1024 MB"),
		}, nil),
		hostCapacityFn: idleHost(5, 32768),
	}

	_, err := s.ReserveVmCapacity(ctx, "op", 3, 0)
	sysErr, ok := err.(*serviceErrors.SystemError)
	if !ok || sysErr.Code() != 409 {
		t.Fatalf("expected a 409 system error, got %T %v", err, err)
	}
	if msg := sysErr.Message(); msg != "cpu admission refused: required 3 logical CPUs, available 2" {
		t.Fatalf("unexpected message: %s", msg)
	}
}

// Memory is the second dimension and is refused in the unit the host publishes it in,
// MiB, so a refusal can be read straight off GET /v1/config/hardware.
func TestReserveVmCapacityMemoryRefusalReportsMiB(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()
	// 5120 MiB with 1024 kept for the host = 4096 free; CPUs are ample so memory binds.
	s := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, nil),
		hostCapacityFn: idleHost(64, 5120),
	}

	release, err := s.ReserveVmCapacity(ctx, "op", 2, 4096)
	if err != nil {
		t.Fatalf("the first create filled the host exactly and was refused: %v", err)
	}

	_, err = s.ReserveVmCapacity(ctx, "op-2", 2, 4096)
	sysErr, ok := err.(*serviceErrors.SystemError)
	if !ok || sysErr.Code() != 409 {
		release()
		t.Fatalf("expected a 409 system error, got %T %v", err, err)
	}
	if msg := sysErr.Message(); msg != "memory admission refused: required 4096 MiB, available 0 MiB" {
		t.Fatalf("unexpected message: %s", msg)
	}
	release()
}

// A capacity that cannot be measured must not silently become unlimited: both the
// machine totals and the VM inventory fail closed with a 503, the posture the VM count
// gate already takes.
func TestReserveVmCapacityFailsClosedWithoutInventory(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()

	unknownMachine := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, nil),
		hostCapacityFn: hostCapacity(nil, errors.New("sysctl failed")),
	}
	if _, err := unknownMachine.ReserveVmCapacity(ctx, "op", 2, 1024); err == nil {
		t.Fatal("create admitted while the machine totals are unknown")
	} else if sysErr, ok := err.(*serviceErrors.SystemError); !ok || sysErr.Code() != httpStatusServiceUnavailable {
		t.Fatalf("expected a 503 system error, got %T %v", err, err)
	}
	if got := unknownMachine.PendingCpu(); got != 0 {
		t.Fatalf("a refused create must hold no capacity, got %d", got)
	}

	unknownVms := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, errors.New("prlctl failed")),
		hostCapacityFn: idleHost(5, 32768),
	}
	if _, err := unknownVms.ReserveVmCapacity(ctx, "op", 2, 1024); err == nil {
		t.Fatal("create admitted while the VM inventory is unknown")
	} else if sysErr, ok := err.(*serviceErrors.SystemError); !ok || sysErr.Code() != httpStatusServiceUnavailable {
		t.Fatalf("expected a 503 system error, got %T %v", err, err)
	}

	// A zeroed snapshot is the same failure wearing a different face: it would refuse
	// every create as "host full" and send the operator looking in the wrong place.
	zeroed := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, nil),
		hostCapacityFn: idleHost(0, 0),
	}
	if _, err := zeroed.ReserveVmCapacity(ctx, "op", 1, 0); err == nil {
		t.Fatal("create admitted against a host that reported no capacity")
	} else if sysErr, ok := err.(*serviceErrors.SystemError); !ok || sysErr.Code() != httpStatusServiceUnavailable {
		t.Fatalf("expected a 503 system error, got %T %v", err, err)
	}
}

// The release func goes back exactly what the admission took, once: calling it twice
// must not under-count the host and let a burst through that does not fit.
func TestReserveVmCapacityReleaseIsIdempotent(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, nil),
		hostCapacityFn: idleHost(5, 32768),
	}

	release, err := s.ReserveVmCapacity(ctx, "op", 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	release()
	release() // the release func is idempotent
	if got := s.PendingCpu(); got != 0 {
		t.Fatalf("released charge still counted: %d", got)
	}

	// The whole 4 CPUs are usable again, and only the 4 CPUs.
	if _, err := s.ReserveVmCapacity(ctx, "op-2", 4, 0); err != nil {
		t.Fatalf("a double release took capacity off the host: %v", err)
	}
	if _, err := s.ReserveVmCapacity(ctx, "op-3", 1, 0); err == nil {
		t.Fatal("a double release gave the host capacity it does not have")
	}
}

// The in-flight charge is what makes the published snapshot honest between probes: it
// has to be visible while the create runs and gone the moment it is released.
func TestReserveVmCapacityPendingIsPublishedWhileInFlight(t *testing.T) {
	enableVmCapacity(t, "")
	ctx := basecontext.NewBaseContext()
	s := &ParallelsService{
		vmInventoryFn:  vmCapacityInventory(nil, nil),
		hostCapacityFn: idleHost(8, 16384),
	}

	release, err := s.ReserveVmCapacity(ctx, "op", 3, 1536)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.PendingCpu(); got != 3 {
		t.Fatalf("the in-flight create must be published as pending CPU, got %d", got)
	}
	if got := s.PendingMemory(); got != 1536 {
		t.Fatalf("the in-flight create must be published as pending memory, got %v", got)
	}

	release()
	if got := s.PendingCpu(); got != 0 {
		t.Fatalf("released charge still published: %d", got)
	}
	if got := s.PendingMemory(); got != 0 {
		t.Fatalf("released charge still published: %v", got)
	}
}
