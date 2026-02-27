package data

import (
	"testing"

	apiModels "github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestSetListSnapshotsByVMId(t *testing.T) {
	//1. Create a new JsonDatabase
	jsonDatabase := &JsonDatabase{
		connected: true,
	}

	//2. Set list of snapshots for a VM
	jsonDatabase.SetListSnapshotsByVMId("VM123", []apiModels.Snapshot{
		{
			Name: "Snapshot1",
			ID:   "Snapshot1",
		},
		{
			Name: "Snapshot2",
			ID:   "Snapshot2",
		},
	})
	//3. Get list of snapshots for a VM
	snapshots, err := jsonDatabase.GetListSnapshotsByVMId("VM123")
	assert.NoError(t, err)
	assert.Len(t, snapshots, 2)
	assert.Equal(t, "Snapshot1", snapshots[0].Name)
	assert.Equal(t, "Snapshot2", snapshots[1].Name)
}
func TestGetListSnapshotsByVMId(t *testing.T) {
	//1. Create a new JsonDatabase
	jsonDatabase := &JsonDatabase{
		connected: true,
	}

	//2. Set list of snapshots for a VM
	jsonDatabase.SetListSnapshotsByVMId("VM123", []apiModels.Snapshot{
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
	})
	//3. Get list of snapshots for a VM
	snapshots, err := jsonDatabase.GetListSnapshotsByVMId("VM123")
	assert.NoError(t, err)
	assert.Len(t, snapshots, 3)
	assert.Equal(t, "Snapshot1", snapshots[0].Name)
	assert.Equal(t, "Snapshot2", snapshots[1].Name)
	assert.Equal(t, "Snapshot3", snapshots[2].Name)
}
