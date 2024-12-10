package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostReverseProxyConfig(ctx basecontext.ApiContext, hostId string, filter string, noCache bool) (*models.ReverseProxy, error) {
	if noCache {
		ctx.LogDebugf("[Orchestrator] No cache set, refreshing all hosts...")
		s.Refresh()
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := dbService.GetOrchestratorReverseProxyConfig(ctx, hostId)
	if err != nil {
		return nil, err
	}

	return hosts, nil
}
