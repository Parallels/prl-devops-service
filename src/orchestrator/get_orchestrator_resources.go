package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetResources(ctx basecontext.ApiContext, noCache bool) ([]models.HostResourceOverviewResponseItem, error) {
	if noCache {
		ctx.LogDebugf("[Orchestrator] No cache set, refreshing all hosts...")
		s.Refresh()
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	totalResources := dbService.GetOrchestratorTotalResources(ctx)
	inUseResources := dbService.GetOrchestratorInUseResources(ctx)
	availableResources := dbService.GetOrchestratorAvailableResources(ctx)
	reservedResources := dbService.GetOrchestratorReservedResources(ctx)
	systemReservedResources := dbService.GetOrchestratorSystemReservedResources(ctx)

	result := make([]models.HostResourceOverviewResponseItem, 0)
	for key, value := range totalResources {
		item := models.HostResourceOverviewResponseItem{}
		item.SystemReserved = systemReservedResources[key]
		item.Total = value
		item.TotalAvailable = availableResources[key]
		item.TotalInUse = inUseResources[key]
		item.TotalReserved = reservedResources[key]
		item.CpuType = key
		result = append(result, item)
	}

	return result, nil
}
