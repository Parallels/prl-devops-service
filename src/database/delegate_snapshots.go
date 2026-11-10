package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/snapshots delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) SetListVMSnapshotsByVMId(vmID string, newVmSnap models.VMSnapshots) error {
	return g.legacy.SetListVMSnapshotsByVMId(vmID, newVmSnap)
}

func (g *GormDatabase) GetListVMSnapshotsByVMId(vmID string) ([]models.VMSnapshot, error) {
	return g.legacy.GetListVMSnapshotsByVMId(vmID)
}
