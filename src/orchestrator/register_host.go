package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) RegisterHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.GetHostHardwareInfo(host)
	if err != nil {
		return nil, err
	}

	host.Enabled = true
	dbHost, err := dbService.CreateOrchestratorHost(ctx, *host)
	if err != nil {
		return nil, err
	}

	return dbHost, nil
}
