package mappers

import (
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
