package orchestrator

import (
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	api_models "github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) GetHostVirtualMachinesInfo(host *models.OrchestratorHost) ([]api_models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	path := "/v1/machines"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response []api_models.ParallelsVM
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting hardware info for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return response, nil
}
