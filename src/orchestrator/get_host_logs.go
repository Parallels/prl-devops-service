package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostLogs(ctx basecontext.ApiContext, hostId string) (string, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return "", err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return "", err
	}
	if host == nil {
		return "", errors.NewWithCodef(404, "Host %s not found", hostId)
	}
	if !host.Enabled {
		return "", errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}
	if host.State != "healthy" {
		return "", errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	result, err := s.CallGetHostLogs(host)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *OrchestratorService) CallGetHostLogs(host *models.OrchestratorHost) (string, error) {
	httpClient := s.getApiClient(*host)
	path := "/v1/logs"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return "", err
	}

	var response string
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return "", err
	}

	if apiResponse.StatusCode != 200 {
		return "", errors.NewWithCodef(400, "Error getting logs for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return response, nil
}
