package mappers

import (
	"github.com/Parallels/prl-devops-service/data/models"
	apiModels "github.com/Parallels/prl-devops-service/models"
)

func VMSnapshotsDtoToApi(dVMSnapshots []models.VMSnapshot) []apiModels.VMSnapshot {
	var apiVMSnapshots []apiModels.VMSnapshot
	for _, dVMSnapshot := range dVMSnapshots {
		apiVMSnapshots = append(apiVMSnapshots, apiModels.VMSnapshot{
			ID:      dVMSnapshot.ID,
			Name:    dVMSnapshot.Name,
			Date:    dVMSnapshot.Date,
			State:   dVMSnapshot.State,
			Current: dVMSnapshot.Current,
			Parent:  dVMSnapshot.Parent,
		})
	}
	return apiVMSnapshots
}

func VMSnapshotsApiToDto(apiVMSnapshots []apiModels.VMSnapshot) []models.VMSnapshot {
	var dVMSnapshots []models.VMSnapshot
	for _, apiVMSnapshot := range apiVMSnapshots {
		dVMSnapshots = append(dVMSnapshots, models.VMSnapshot{
			ID:      apiVMSnapshot.ID,
			Name:    apiVMSnapshot.Name,
			Date:    apiVMSnapshot.Date,
			State:   apiVMSnapshot.State,
			Current: apiVMSnapshot.Current,
			Parent:  apiVMSnapshot.Parent,
		})
	}
	return dVMSnapshots
}
