package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) EnableHost(ctx basecontext.ApiContext, hostIdOrHost string) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}
	host, err := dbService.GetOrchestratorHost(ctx, hostIdOrHost)
	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", hostIdOrHost)
	}

	host.Enabled = true
	updatedHost, err := dbService.UpdateOrchestratorHost(ctx, host)
	if err != nil {
		return nil, err
	}

	s.Refresh()

	manager := GetHostWebSocketManager()
	if manager != nil {
		manager.ProbeAndConnect(*updatedHost)
	}

	return updatedHost, nil
}

func (s *OrchestratorService) DisableHost(ctx basecontext.ApiContext, hostIdOrHost string) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}
	host, err := dbService.GetOrchestratorHost(ctx, hostIdOrHost)
	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", hostIdOrHost)
	}

	host.Enabled = false
	updatedHost, err := dbService.UpdateOrchestratorHost(ctx, host)
	if err != nil {
		return nil, err
	}

	s.Refresh()

	manager := GetHostWebSocketManager()
	if manager != nil {
		manager.DisconnectHost(updatedHost.ID)
	}

	return updatedHost, nil
}
