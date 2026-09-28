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
