package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) GetVirtualMachineStatus(ctx basecontext.ApiContext, vmId string) (*models.VirtualMachineStatusResponse, error) {
	vm, err := s.GetVirtualMachine(ctx, vmId)
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

	result, err := s.GetHostVirtualMachineStatus(ctx, vm.HostId, vmId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) GetHostVirtualMachineStatus(ctx basecontext.ApiContext, hostId string, vmId string) (*models.VirtualMachineStatusResponse, error) {
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

	vm, err := s.GetHostVirtualMachine(ctx, hostId, vmId)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, errors.NewWithCodef(404, "Virtual machine %s not found on host %s", vmId, hostId)
	}

	result, err := s.CallGetHostVirtualMachineStatus(host, vmId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) CallGetHostVirtualMachineStatus(host *data_models.OrchestratorHost, vmId string) (*models.VirtualMachineStatusResponse, error) {
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
