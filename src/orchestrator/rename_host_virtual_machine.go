package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) RenameVirtualMachine(ctx basecontext.ApiContext, vmId string, request models.RenameVirtualMachineRequest) (*models.ParallelsVM, error) {
	vm, err := s.GetVirtualMachine(ctx, vmId, false)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.NewWithCodef(404, "Virtual machine %s not found", vmId)
	}

	host, err := s.GetHost(ctx, vm.HostId)
	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", vm.HostId)
	}

	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", host.ID)
	}

	if host.State != "healthy" {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", host.ID)
	}

	result, err := s.RenameHostVirtualMachine(ctx, vm.HostId, vmId, request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) RenameHostVirtualMachine(ctx basecontext.ApiContext, hostId string, vmId string, request models.RenameVirtualMachineRequest) (*models.ParallelsVM, error) {
	host, err := s.GetHost(ctx, hostId)
	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", hostId)
	}

	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}

	if host.State != "healthy" {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", hostId)
	}

	vm, err := s.GetHostVirtualMachine(ctx, hostId, vmId, false)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, errors.NewWithCodef(404, "Virtual machine %s not found on host %s", vmId, hostId)
	}

	result, err := s.CallRenameHostVirtualMachine(host, vmId, request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) CallRenameHostVirtualMachine(host *data_models.OrchestratorHost, vmId string, request models.RenameVirtualMachineRequest) (*models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(2 * time.Minute)
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
