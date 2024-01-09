package orchestrator

import (
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	api_models "github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
)

func (s *OrchestratorService) GetHostHealthProbeCheck(host *models.OrchestratorHost) (*restapi.HealthProbeResponse, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(s.healthCheckTimeout)

	path := "/health/probe"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}
	var response restapi.HealthProbeResponse
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting health check for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return &response, nil
}

func (s *OrchestratorService) GetHostSystemHealthCheck(host *models.OrchestratorHost) (*api_models.ApiHealthCheck, error) {
	httpClient := s.getApiClient(*host)

	path := "/health/system"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}
	var response api_models.ApiHealthCheck
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting health check for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return &response, nil
}

func (s *OrchestratorService) GetHostHealthCheckState(host *models.OrchestratorHost) string {
	healthCheck, err := s.GetHostHealthProbeCheck(host)
	if err != nil {
		return "unhealthy"
	}
	if healthCheck.Status == "OK" {
		return "healthy"
	} else {
		return "unhealthy"
	}
}
