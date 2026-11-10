package interfaces

// SnapshotsService mirrors vm_snapshots method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/data/models"
)

type SnapshotsService interface {
	SetListVMSnapshotsByVMId(vmID string, newVmSnap models.VMSnapshots) error
	GetListVMSnapshotsByVMId(vmID string) ([]models.VMSnapshot, error)
}
