package models

import "testing"

func TestActiveMacVMIDs(t *testing.T) {
	vms := []ParallelsVM{
		{ID: "one", Type: "APPLE_VZ_VM", State: "running"},
		{ID: "one", Type: "APPLE_VZ_VM", State: "running"},
		{ID: "two", OS: "macosx", State: "starting"},
		{ID: "three", Type: "macvm", State: "paused"},
		{ID: "stopped", Type: "macvm", State: "stopped"},
		{ID: "suspended", Type: "macvm", State: "suspended"},
		{ID: "linux", Name: "mac build", OS: "linux", State: "running"},
	}
	ids := ActiveMacVMIDs(vms)
	if len(ids) != 3 {
		t.Fatalf("expected 3 unique active macVMs, got %v", ids)
	}
}

// ActiveVMIDs is the counting basis of the "running" mode of the per-host VM limit:
// every VM type counts, not only macOS ones.
func TestActiveVMIDsCountsEveryVmType(t *testing.T) {
	vms := []ParallelsVM{
		{ID: "mac", Type: "macvm", State: "running"},
		{ID: "linux", Type: "vm", OS: "linux", State: "paused"},
		{ID: "linux", Type: "vm", OS: "linux", State: "paused"},
		{ID: "stopped", Type: "vm", OS: "linux", State: "stopped"},
		{ID: "", Type: "vm", OS: "linux", State: "running"},
	}

	ids := ActiveVMIDs(vms)
	if len(ids) != 2 || ids[0] != "mac" || ids[1] != "linux" {
		t.Fatalf("expected the two slot-occupying VMs, got %v", ids)
	}
	if len(ActiveMacVMIDs(vms)) != 1 {
		t.Fatalf("macVM subset must stay restricted to macOS VMs: %v", ActiveMacVMIDs(vms))
	}
}

func TestOccupiesVmSlot(t *testing.T) {
	for _, state := range []string{"running", "starting", "resuming", "paused", "pausing", "stopping", "suspending", "RUNNING"} {
		if !OccupiesVmSlot(state) {
			t.Fatalf("state %s must occupy a slot", state)
		}
	}
	for _, state := range []string{"stopped", "suspended", "installing", ""} {
		if OccupiesVmSlot(state) {
			t.Fatalf("state %s must not occupy a slot", state)
		}
	}
}
