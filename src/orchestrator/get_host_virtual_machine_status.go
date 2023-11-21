package orchestrator

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
)

func (s *OrchestratorService) GetHostVirtualMachineStatus(host *data_models.OrchestratorHost, vmId string) (*models.VirtualMachineStatusResponse, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines/" + vmId + "/status"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.VirtualMachineStatusResponse
	_, err = httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
