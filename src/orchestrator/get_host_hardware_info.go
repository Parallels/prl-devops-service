package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	api_models "github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) GetHostHardwareInfo(host *models.OrchestratorHost) (*api_models.SystemUsageResponse, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(10 * time.Second)
	path := "/v1/config/hardware"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response api_models.SystemUsageResponse
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting hardware info for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return &response, nil
}
