package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetResources(ctx basecontext.ApiContext) ([]models.HostResourceOverviewResponseItem, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	s.Refresh()

	totalResources := dbService.GetOrchestratorTotalResources(ctx)
	inUseResources := dbService.GetOrchestratorInUseResources(ctx)
	availableResources := dbService.GetOrchestratorAvailableResources(ctx)
	reservedResources := dbService.GetOrchestratorReservedResources(ctx)

	result := make([]models.HostResourceOverviewResponseItem, 0)
	for key, value := range totalResources {
		item := models.HostResourceOverviewResponseItem{}
		item.Total = value
		item.TotalAvailable = availableResources[key]
		item.TotalInUse = inUseResources[key]
		item.TotalReserved = reservedResources[key]
		item.CpuType = key
		result = append(result, item)
	}

	return result, nil
}
