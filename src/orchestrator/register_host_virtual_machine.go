package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) RegisterHostVirtualMachine(ctx basecontext.ApiContext, hostId string, request models.RegisterVirtualMachineRequest) (*models.ParallelsVM, error) {
	host, err := s.GetHost(ctx, hostId)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", hostId)
	}

	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", host.Host)
	}
	if host.State != "healthy" {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	return s.CallRegisterHostVirtualMachine(host, request)
}

func (s *OrchestratorService) CallRegisterHostVirtualMachine(host *data_models.OrchestratorHost, request models.RegisterVirtualMachineRequest) (*models.ParallelsVM, error) {
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
