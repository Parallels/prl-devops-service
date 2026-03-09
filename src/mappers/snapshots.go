package mappers

import (
	"github.com/Parallels/prl-devops-service/data/models"
	apiModels "github.com/Parallels/prl-devops-service/models"
)

func DtoSnapshotToApi(snapshot []models.Snapshot) []apiModels.Snapshot {
	var snapshots []apiModels.Snapshot
	for _, snapshot := range snapshot {
		snapshots = append(snapshots, apiModels.Snapshot{
			ID:      snapshot.ID,
			Name:    snapshot.Name,
			Date:    snapshot.Date,
			State:   snapshot.State,
			Current: snapshot.Current,
			Parent:  snapshot.Parent,
		})
	}
	return snapshots
}

func SnapshotsToDto(snapshot []apiModels.Snapshot) []models.Snapshot {
	var snapshots []models.Snapshot
	for _, snapshot := range snapshot {
		snapshots = append(snapshots, models.Snapshot{
			ID:      snapshot.ID,
			Name:    snapshot.Name,
			Date:    snapshot.Date,
			State:   snapshot.State,
			Current: snapshot.Current,
			Parent:  snapshot.Parent,
		})
	}
	return snapshots
}
