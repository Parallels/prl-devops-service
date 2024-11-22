package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) ConfigureVirtualMachine(ctx basecontext.ApiContext, vmId string, request models.VirtualMachineConfigRequest, noCache bool) (*models.VirtualMachineConfigResponse, error) {
	if noCache {
		ctx.LogDebugf("[Orchestrator] No cache set, refreshing all hosts...")
		s.Refresh()
	}

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

	if host.State != HealthyState {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", host.ID)
	}

	result, err := s.ConfigureHostVirtualMachine(ctx, vm.HostId, vmId, request, false)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) ConfigureHostVirtualMachine(ctx basecontext.ApiContext, hostId string, vmId string, request models.VirtualMachineConfigRequest, useCache bool) (*models.VirtualMachineConfigResponse, error) {
	if !useCache {
		s.Refresh()
	}

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

	result, err := s.CallConfigureHostVirtualMachine(host, vm.ID, request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) CallConfigureHostVirtualMachine(host *data_models.OrchestratorHost, vmId string, request models.VirtualMachineConfigRequest) (*models.VirtualMachineConfigResponse, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(3 * time.Minute)
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
