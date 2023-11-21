package orchestrator

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
)

func (s *OrchestratorService) CreateHostVirtualMachine(host *data_models.OrchestratorHost, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.CreateVirtualMachineResponse
	_, err = httpClient.Post(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	response.Host = host.GetHost()

	s.Refresh()
	return &response, nil
}
