package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

func (s *OrchestratorService) RestartHostReverseProxy(ctx basecontext.ApiContext, hostId string) error {
	host, err := s.GetHost(ctx, hostId)
	if err != nil {
		return err
	}

	if host == nil {
		return errors.NewWithCodef(404, "Host %s not found", hostId)
	}

	if !host.Enabled {
		return errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}

	if host.State != "healthy" {
		return errors.NewWithCodef(400, "Host %s is not healthy", hostId)
	}

	err = s.CallRestartHostReverseProxy(host)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrchestratorService) CallRestartHostReverseProxy(host *models.OrchestratorHost) error {
	httpClient := s.getApiClient(*host)
	path := "/reverse-proxy/restart"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	apiResponse, err := httpClient.Put(url.String(), nil, nil)
	if err != nil {
		return err
	}

	if apiResponse.StatusCode != 202 {
		return errors.NewWithCodef(400, "Error restarting host %s reverse proxy: %v", host.Host, apiResponse.StatusCode)
	}

	s.Refresh()
	return nil
}
