package orchestrator

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) RenameHostVirtualMachine(host *data_models.OrchestratorHost, vmId string, request models.RenameVirtualMachineRequest) (*models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines/" + vmId + "/rename"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.ParallelsVM
	_, err = httpClient.Put(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	s.Refresh()
	return &response, nil
}
