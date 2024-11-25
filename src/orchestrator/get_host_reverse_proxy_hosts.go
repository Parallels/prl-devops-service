package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostReverseProxyHosts(ctx basecontext.ApiContext, hostId string, filter string, noCache bool) ([]*models.ReverseProxyHost, error) {
	if noCache {
		ctx.LogDebugf("[Orchestrator] No cache set, refreshing all hosts...")
		s.Refresh()
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := dbService.GetOrchestratorReverseProxyHosts(ctx, hostId, "")
	if err != nil {
		return nil, err
	}

	return hosts, nil
}

func (s *OrchestratorService) GetHostReverseProxyHost(ctx basecontext.ApiContext, hostId string, rpHostId string, noCache bool) (*models.ReverseProxyHost, error) {
	if noCache {
		ctx.LogDebugf("[Orchestrator] No cache set, refreshing all hosts...")
		s.Refresh()
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	rpHost, err := dbService.GetOrchestratorReverseProxyHost(ctx, hostId, rpHostId)
	if err != nil {
		return nil, err
	}

	if rpHost == nil {
		return nil, errors.NewWithCodef(404, "Reverse proxy host %s not found", rpHostId)
	}

	return rpHost, nil
}
