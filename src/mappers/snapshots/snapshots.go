package snapshots

import (
	"github.com/Parallels/prl-devops-service/data/models"
	apiModels "github.com/Parallels/prl-devops-service/models"
)

func DtoToApi(dSnapshots []models.Snapshot) []apiModels.Snapshot {
	var apiSnapshots []apiModels.Snapshot
	for _, dSnapshot := range dSnapshots {
		apiSnapshots = append(apiSnapshots, apiModels.Snapshot{
			ID:      dSnapshot.ID,
			Name:    dSnapshot.Name,
			Date:    dSnapshot.Date,
			State:   dSnapshot.State,
			Current: dSnapshot.Current,
			Parent:  dSnapshot.Parent,
		})
	}
	return apiSnapshots
}

func ApiToDto(apiSnapshots []apiModels.Snapshot) []models.Snapshot {
	var dSnapshots []models.Snapshot
	for _, apiSnapshot := range apiSnapshots {
		dSnapshots = append(dSnapshots, models.Snapshot{
			ID:      apiSnapshot.ID,
			Name:    apiSnapshot.Name,
			Date:    apiSnapshot.Date,
			State:   apiSnapshot.State,
			Current: apiSnapshot.Current,
			Parent:  apiSnapshot.Parent,
		})
	}
	return dSnapshots
}
