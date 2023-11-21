package orchestrator

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
)

func (s *OrchestratorService) ExecuteOnHostVirtualMachine(host *data_models.OrchestratorHost, vmId string, request models.VirtualMachineExecuteCommandRequest) (*models.VirtualMachineExecuteCommandResponse, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines/" + vmId + "/execute"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.VirtualMachineExecuteCommandResponse
	_, err = httpClient.Put(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
