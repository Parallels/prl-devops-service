package data

import (
	"github.com/Parallels/prl-devops-service/data/models"
	apiModels "github.com/Parallels/prl-devops-service/models"
)

func (j *JsonDatabase) SetListSnapshotsByVMId(vmID string, listSnapshotResponse *apiModels.ListSnapshotResponse) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	if j.data.VMSnapshots == nil {
		j.data.VMSnapshots = make(map[string][]models.Snapshot)
	}
	snaps := []models.Snapshot{}
	for _, snapshot := range listSnapshotResponse.Snapshots {
		snaps = append(snaps, models.Snapshot{
			ID:      snapshot.ID,
			Name:    snapshot.Name,
			Date:    snapshot.Date,
			State:   snapshot.State,
			Current: snapshot.Current,
			Parent:  snapshot.Parent,
		})
	}
	j.data.VMSnapshots[vmID] = snaps
	return nil
}

func (j *JsonDatabase) GetListSnapshotsByVMId(vmID string) (*apiModels.ListSnapshotResponse, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	if j.data.VMSnapshots == nil {
		return nil, nil
	}

	snaps, ok := j.data.VMSnapshots[vmID]
	if !ok {
		return nil, nil
	}
	resp := apiModels.ListSnapshotResponse{}
	for i := range snaps {
		resp.Snapshots = append(resp.Snapshots, apiModels.Snapshot{
			ID:      snaps[i].ID,
			Name:    snaps[i].Name,
			Date:    snaps[i].Date,
			State:   snaps[i].State,
			Current: snaps[i].Current,
			Parent:  snaps[i].Parent,
		})
	}
	return &resp, nil
}
