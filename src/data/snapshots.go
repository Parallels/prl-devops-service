package data

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	api_models "github.com/Parallels/prl-devops-service/models"
)

var (
	ErrSnapshotVMIdCannotBeEmpty = errors.NewWithCode("VM ID cannot be empty", 400)
)

// GetSnapshotsByVMId returns all snapshots for a specific VM
func (j *JsonDatabase) GetSnapshotsByVMId(ctx basecontext.ApiContext, vmId string) ([]models.VirtualMachineSnapshot, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if vmId == "" {
		return nil, ErrSnapshotVMIdCannotBeEmpty
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	if j.data.VMSnapshots == nil {
		return []models.VirtualMachineSnapshot{}, nil
	}

	snapshots, exists := j.data.VMSnapshots[vmId]
	if !exists {
		return []models.VirtualMachineSnapshot{}, nil
	}

	return snapshots, nil
}

// GetSnapshotList returns snapshots for a VM in API-compatible map format
func (j *JsonDatabase) GetSnapshotList(ctx basecontext.ApiContext, vmId string) (map[string]api_models.SnapshotDetails, error) {
	snapshots, err := j.GetSnapshotsByVMId(ctx, vmId)
	if err != nil {
		return nil, err
	}

	snapshotMap := make(map[string]api_models.SnapshotDetails)
	for _, snapshot := range snapshots {
		snapshotMap[snapshot.ID] = api_models.SnapshotDetails{
			Name:    snapshot.Name,
			Date:    snapshot.Date,
			State:   snapshot.State,
			Current: snapshot.Current,
			Parent:  snapshot.Parent,
		}
	}

	return snapshotMap, nil
}

// UpdateSnapshotsByVMId replaces all snapshots for a VM (creates or updates)
func (j *JsonDatabase) UpdateSnapshotsByVMId(ctx basecontext.ApiContext, vmId string, snapshots []models.VirtualMachineSnapshot) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if vmId == "" {
		return ErrSnapshotVMIdCannotBeEmpty
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	if j.data.VMSnapshots == nil {
		j.data.VMSnapshots = make(map[string][]models.VirtualMachineSnapshot)
	}

	j.data.VMSnapshots[vmId] = snapshots
	return nil
}

// DeleteSnapshotsByVMId removes all snapshots for a specific VM
func (j *JsonDatabase) DeleteSnapshotsByVMId(ctx basecontext.ApiContext, vmId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if vmId == "" {
		return ErrSnapshotVMIdCannotBeEmpty
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	if j.data.VMSnapshots != nil {
		delete(j.data.VMSnapshots, vmId)
	}

	return nil
}
