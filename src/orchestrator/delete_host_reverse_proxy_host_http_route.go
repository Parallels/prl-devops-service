package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) DeleteHostReverseProxyHostHttpRoute(ctx basecontext.ApiContext, hostId string, rpHostId string, httpRouteId string) error {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
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

	err = s.CallHostReverseProxyHostHttpRoute(host, rpHostId, httpRouteId)
	if err != nil {
		return err
	}

	s.Refresh()
	return nil
}

func (s *OrchestratorService) CallHostReverseProxyHostHttpRoute(host *data_models.OrchestratorHost, rpHostId string, httpRouteId string) error {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(1 * time.Minute)

	path := "/reverse-proxy/hosts/" + rpHostId + "/http_routes/" + httpRouteId
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	api_response, err := httpClient.Delete(url.String(), nil)
	if err != nil {
		return err
	}

	if api_response.StatusCode != 202 {
		return errors.NewWithCodef(400, "Error deleting reverse proxy host %s http route %v: %v", rpHostId, httpRouteId, api_response.StatusCode)
	}

	return nil
}
