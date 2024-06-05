package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) UpdateHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hw, err := s.GetHostHardwareInfo(host)
	if err != nil {
		return nil, err
	}

	host.Enabled = true
	if host.Architecture != hw.CpuType {
		if host.Resources == nil {
			host.Resources = &models.HostResources{}
		}

		dtoResources := mappers.MapHostResourcesFromSystemUsageResponse(*hw)
		host.Resources = &dtoResources
		host.Architecture = hw.CpuType
		host.CpuModel = hw.CpuBrand
	}

	dbHost, err := dbService.UpdateOrchestratorHost(ctx, host)
	if err != nil {
		return nil, err
	}

	return dbHost, nil
}
