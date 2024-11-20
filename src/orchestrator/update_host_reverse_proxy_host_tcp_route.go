package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) UpdateHostReverseProxyHostTcpRoute(ctx basecontext.ApiContext, hostId string, rpHostId string, r models.ReverseProxyHostTcpRouteCreateRequest) (*models.ReverseProxyHost, *models.ApiErrorResponse) {
	var api_error *models.ApiErrorResponse

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		api_error = &models.ApiErrorResponse{
			Message: "There was an error getting the database",
			Code:    500,
		}
		return nil, api_error
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		api_error = &models.ApiErrorResponse{
			Message: "There was an error getting the host from the database",
			Code:    500,
		}
		return nil, api_error
	}

	if host == nil {
		api_error = &models.ApiErrorResponse{
			Message: "Host not found",
			Code:    404,
		}
		return nil, api_error
	}

	if !host.Enabled {
		api_error = &models.ApiErrorResponse{
			Message: "Host is disabled",
			Code:    400,
		}
		return nil, api_error
	}

	if host.State != "healthy" {
		api_error = &models.ApiErrorResponse{
			Message: "Host is not healthy",
			Code:    400,
		}
		return nil, api_error
	}

	resp, err := s.CallUpdateHostReverseProxyHostTcpRoute(host, rpHostId, r)
	if err != nil {
		api_error = &models.ApiErrorResponse{
			Message: err.Error(),
			Code:    400,
		}
		return nil, api_error
	}

	s.Refresh()
	return resp, nil
}

func (s *OrchestratorService) CallUpdateHostReverseProxyHostTcpRoute(host *data_models.OrchestratorHost, rpHostId string, r models.ReverseProxyHostTcpRouteCreateRequest) (*models.ReverseProxyHost, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(1 * time.Minute)

	path := "/reverse-proxy/hosts/" + rpHostId + "/tcp_route"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var rep models.ReverseProxyHost
	_, err = httpClient.Post(url.String(), r, &rep)
	if err != nil {
		return nil, err
	}

	s.Refresh()
	return &rep, nil
}
