package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) ResetVirtualMachine(ctx basecontext.ApiContext, vmId string, noCache bool) (*models.VirtualMachineOperationResponse, error) {
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

	result, err := s.ResetHostVirtualMachine(ctx, vm.HostId, vmId, false)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) ResetHostVirtualMachine(ctx basecontext.ApiContext, hostId string, vmId string, useCache bool) (*models.VirtualMachineOperationResponse, error) {
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

	result, err := s.CallResetHostVirtualMachine(host, vm.ID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *OrchestratorService) CallResetHostVirtualMachine(host *data_models.OrchestratorHost, vmId string) (*models.VirtualMachineOperationResponse, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(3 * time.Minute)
	path := "/machines/" + vmId + "/reset"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.VirtualMachineOperationResponse
	_, err = httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	s.Refresh()
	return &response, nil
}
