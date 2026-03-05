package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

func (s *OrchestratorService) GetHostReverseProxyConfig(ctx basecontext.ApiContext, hostId string, filter string, useCache bool) (*models.ReverseProxy, error) {
	if !useCache {
		ctx.LogDebugf("[Orchestrator] No cache set, refreshing all hosts...")
		s.Refresh()
	}

	host, err := s.GetHost(ctx, hostId)
	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", hostId)
	}

	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}

	if host.State != HealthyState {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", hostId)
	}

	hosts, err := s.db.GetOrchestratorReverseProxyConfig(ctx, hostId)
	if err != nil {
		return nil, err
	}

	return hosts, nil
}

func (s *OrchestratorService) CallGetHostReverseProxyConfig(host *data_models.OrchestratorHost) (*models.ReverseProxy, error) {
	if host == nil {
		return nil, errors.NewWithCodef(404, "Host not found")
	}

	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", host.ID)
	}

	if host.State != HealthyState {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", host.ID)
	}

	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(3 * time.Minute)
	path := "/reverse-proxy"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.ReverseProxy
	_, err = httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
