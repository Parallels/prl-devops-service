package orchestrator

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) UnregisterHostVirtualMachine(host *data_models.OrchestratorHost, vmId string, request models.UnregisterVirtualMachineRequest) (*models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines/" + vmId + "/unregister"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.ParallelsVM
	_, err = httpClient.Post(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	s.Refresh()
	return &response, nil
}
