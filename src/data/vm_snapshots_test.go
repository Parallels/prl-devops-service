package data

import (
	"testing"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
	apiModels "github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestSetListVMSnapshotsByVMId(t *testing.T) {
	//1. Create a new JsonDatabase
	jsonDatabase := &JsonDatabase{
		connected: true,
	}

	//2. Set list of snapshots for a VM
	apiSnapshots := []apiModels.VMSnapshot{
		{
			Name: "Snapshot1",
			ID:   "Snapshot1",
		},
		{
			Name: "Snapshot2",
			ID:   "Snapshot2",
		},
	}

	dtoSnapshots := mappers.VMSnapshotsApiToDto(apiSnapshots)
	jsonDatabase.SetListVMSnapshotsByVMId("VM123", data_models.VMSnapshots{
		VMId:       "VM123",
		VMSnapshot: dtoSnapshots,
	})

	//3. Get list of snapshots for a VM
	snapshots, err := jsonDatabase.GetListVMSnapshotsByVMId("VM123")
	assert.NoError(t, err)
	assert.Len(t, snapshots, 2)
	assert.Equal(t, "Snapshot1", snapshots[0].Name)
	assert.Equal(t, "Snapshot2", snapshots[1].Name)
}
func TestGetListVMSnapshotsByVMId(t *testing.T) {
	//1. Create a new JsonDatabase
	jsonDatabase := &JsonDatabase{
		connected: true,
	}

	//2. Set list of snapshots for a VM
	apiSnapshots := []apiModels.VMSnapshot{
		{
			Name: "Snapshot1",
			ID:   "1",
		},
		{
			Name: "Snapshot2",
			ID:   "2",
		},
		{
			Name: "Snapshot3",
			ID:   "3",
		},
	}

	dtoSnapshots := mappers.VMSnapshotsApiToDto(apiSnapshots)
	jsonDatabase.SetListVMSnapshotsByVMId("VM123", data_models.VMSnapshots{
		VMId:       "VM123",
		VMSnapshot: dtoSnapshots,
	})

	//3. Get list of snapshots for a VM
	snapshots, err := jsonDatabase.GetListVMSnapshotsByVMId("VM123")
	assert.NoError(t, err)
	assert.Len(t, snapshots, 3)
	assert.Equal(t, "Snapshot1", snapshots[0].Name)
	assert.Equal(t, "Snapshot2", snapshots[1].Name)
	assert.Equal(t, "Snapshot3", snapshots[2].Name)
}
