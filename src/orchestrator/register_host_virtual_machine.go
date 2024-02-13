package orchestrator

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) RegisterHostVirtualMachine(host *data_models.OrchestratorHost, request models.RegisterVirtualMachineRequest) (*models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines/register"
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
