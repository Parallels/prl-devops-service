package orchestrator

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
)

func (s *OrchestratorService) ConfigureHostVirtualMachine(host *data_models.OrchestratorHost, vmId string, request models.VirtualMachineConfigRequest) (*models.VirtualMachineConfigResponse, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines/" + vmId + "/set"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.VirtualMachineConfigResponse
	_, err = httpClient.Put(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	s.Refresh()
	return &response, nil
}
