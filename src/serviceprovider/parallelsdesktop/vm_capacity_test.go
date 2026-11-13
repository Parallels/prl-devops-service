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
