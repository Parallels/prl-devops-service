package data

import (
	apiModels "github.com/Parallels/prl-devops-service/models"
)

func (j *JsonDatabase) SetListSnapshotsByVMId(vmID string, listSnapshotResponse *apiModels.ListSnapshotResponse) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	if j.data.Snapshots == nil {
		j.data.Snapshots = make(map[string]apiModels.ListSnapshotResponse)
	}

	j.data.Snapshots[vmID] = *listSnapshotResponse
	return nil
}

func (j *JsonDatabase) GetListSnapshotsByVMId(vmID string) (*apiModels.ListSnapshotResponse, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	if j.data.Snapshots == nil {
		return nil, nil
	}

	listSnapshotResponse, ok := j.data.Snapshots[vmID]
	if !ok {
		return nil, nil
	}

	return &listSnapshotResponse, nil
}
