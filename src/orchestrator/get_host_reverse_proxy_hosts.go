package orchestrator

import (
	"fmt"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

func (s *OrchestratorService) GetHostReverseProxyHosts(ctx basecontext.ApiContext, hostId string, filter string, useCache bool) ([]*models.ReverseProxyHost, error) {
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

	proxyHosts, err := s.db.GetOrchestratorReverseProxyHosts(ctx, hostId, "")
	if err != nil {
		return nil, err
	}

	return proxyHosts, nil
}

func (s *OrchestratorService) CallGetHostReverseProxyHosts(host *data_models.OrchestratorHost) ([]*models.ReverseProxyHost, error) {
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
	path := "/reverse-proxy/hosts"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response []models.ReverseProxyHost
	_, err = httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	var proxyHosts []*models.ReverseProxyHost
	for _, rpHost := range response {
		proxyHosts = append(proxyHosts, &rpHost)
	}

	return proxyHosts, nil
}

func (s *OrchestratorService) GetHostReverseProxyHost(ctx basecontext.ApiContext, hostId string, rpHostId string, useCache bool) (*models.ReverseProxyHost, error) {
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

	rpHost, err := s.db.GetOrchestratorReverseProxyHost(ctx, hostId, rpHostId)
	if err != nil {
		return nil, err
	}

	if rpHost == nil {
		return nil, errors.NewWithCodef(404, "Reverse proxy host %s not found", rpHostId)
	}

	return rpHost, nil
}

func (s *OrchestratorService) CallGetHostReverseProxyHost(host *data_models.OrchestratorHost, rpHostId string) (*models.ReverseProxyHost, error) {
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
	path := fmt.Sprintf("/reverse-proxy/hosts/%s", rpHostId)
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.ReverseProxyHost
	_, err = httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
