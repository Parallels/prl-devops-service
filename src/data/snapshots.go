package data

import (
	apiModels "github.com/Parallels/prl-devops-service/models"
)

func (j *JsonDatabase) SetListSnapshotsByVMId(vmID string, snapshots []apiModels.Snapshot) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	if j.data.Snapshots == nil {
		j.data.Snapshots = make(map[string][]apiModels.Snapshot)
	}

	j.data.Snapshots[vmID] = snapshots
	return nil
}

func (j *JsonDatabase) GetListSnapshotsByVMId(vmID string) ([]apiModels.Snapshot, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	if j.data.Snapshots == nil {
		return nil, nil
	}

	snapshots, ok := j.data.Snapshots[vmID]
	if !ok {
		return nil, nil
	}

	return snapshots, nil
}
