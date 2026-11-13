package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"testing"
)

func TestHardwareRefreshPreservesMacVMCapacity(t *testing.T) {
	complete := true
	snapshot := models.SystemUsageResponse{TotalInUse: &models.SystemUsageItem{MacVMsRunning: []string{"one", "two"}}, PendingMacVMs: 1, MacVMInventoryComplete: &complete}
	resources := MapHostResourcesFromSystemUsageResponse(snapshot)
	if resources.TotalAppleVms != 2 || resources.TotalInUse.TotalAppleVms != 2 || resources.TotalReserved.TotalAppleVms != 1 {
		t.Fatalf("hardware mapping lost macVM capacity: %+v", resources)
	}
	roundtrip := MapSystemUsageResponseFromHostResources(resources)
	if len(roundtrip.TotalInUse.MacVMsRunning) != 2 || roundtrip.PendingMacVMs != 1 || roundtrip.MacVMInventoryComplete == nil || !*roundtrip.MacVMInventoryComplete {
		t.Fatalf("hardware response lost capacity: %+v", roundtrip)
	}
}

// The per-host VM count limit and the counts it applies to must survive both mapping
// directions: the orchestrator decides placement on the stored resources, Control
// Center reads them back from the API model.
func TestHardwareRefreshCarriesVmCountLimit(t *testing.T) {
	snapshot := models.SystemUsageResponse{
		MaxVmsPerHost:     5,
		MaxVmsPerHostMode: "running",
		TotalVms:          9,
		RunningVms:        4,
		PendingVms:        1,
	}
	resources := MapHostResourcesFromSystemUsageResponse(snapshot)
	if resources.MaxVmsPerHost != 5 || resources.MaxVmsPerHostMode != "running" {
		t.Fatalf("hardware mapping lost the VM limit: %+v", resources)
	}
	if resources.Total.TotalVms != 9 || resources.TotalInUse.TotalVms != 4 || resources.TotalReserved.TotalVms != 1 {
		t.Fatalf("hardware mapping lost the VM counts: %+v", resources)
	}

	roundtrip := MapSystemUsageResponseFromHostResources(resources)
	if roundtrip.MaxVmsPerHost != 5 || roundtrip.MaxVmsPerHostMode != "running" ||
		roundtrip.TotalVms != 9 || roundtrip.RunningVms != 4 || roundtrip.PendingVms != 1 {
		t.Fatalf("hardware response lost the VM limit: %+v", roundtrip)
	}

	// A host without a configured limit keeps reporting nothing, which is what keeps
	// the orchestrator check disabled and the legacy behaviour intact.
	empty := MapSystemUsageResponseFromHostResources(MapHostResourcesFromSystemUsageResponse(models.SystemUsageResponse{}))
	if empty.MaxVmsPerHost != 0 || empty.MaxVmsPerHostMode != "" || empty.TotalVms != 0 {
		t.Fatalf("hosts without a limit must report none: %+v", empty)
	}
}

// A limit change on an already registered host has to be seen as a resource change,
// otherwise the orchestrator would never refresh it.
func TestHostResourcesDiffDetectsVmLimit(t *testing.T) {
	current := data_models.HostResources{MaxVmsPerHost: 5, MaxVmsPerHostMode: "any"}
	if !current.Diff(data_models.HostResources{MaxVmsPerHost: 8, MaxVmsPerHostMode: "any"}) {
		t.Fatal("a changed VM limit must be reported as a diff")
	}
	if !current.Diff(data_models.HostResources{MaxVmsPerHost: 5, MaxVmsPerHostMode: "running"}) {
		t.Fatal("a changed VM limit mode must be reported as a diff")
	}
	if current.Diff(current) {
		t.Fatal("identical resources must not be reported as a diff")
	}

	counts := data_models.HostResourceItem{TotalVms: 5}
	if !counts.Diff(data_models.HostResourceItem{TotalVms: 6}) {
		t.Fatal("a changed VM count must be reported as a diff")
	}
}
