package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) DeleteVirtualMachine(ctx basecontext.ApiContext, vmId string) error {
	vm, err := s.GetVirtualMachine(ctx, vmId)
	if err != nil {
		return err
	}

	if vm == nil {
		return errors.NewWithCodef(404, "Virtual machine %s not found", vmId)
	}

	host, err := s.GetHost(ctx, vm.HostId)
	if err != nil {
		return err
	}

	if host == nil {
		return errors.NewWithCodef(404, "Host %s not found", vm.HostId)
	}

	if !host.Enabled {
		return errors.NewWithCodef(400, "Host %s is disabled", vm.HostId)
	}

	if host.State != "healthy" {
		return errors.NewWithCodef(400, "Host %s is not healthy", vm.HostId)
	}

	err = s.DeleteHostVirtualMachine(ctx, vm.HostId, vmId)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrchestratorService) DeleteHostVirtualMachine(ctx basecontext.ApiContext, hostId string, vmId string) error {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return err
	}

	host, err := s.GetHost(ctx, hostId)
	if err != nil {
		return err
	}

	if host == nil {
		return errors.NewWithCodef(404, "Host %s not found", hostId)
	}

	if !host.Enabled {
		return errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}

	if host.State != "healthy" {
		return errors.NewWithCodef(400, "Host %s is not healthy", hostId)
	}

	vm, err := s.GetHostVirtualMachine(ctx, hostId, vmId)
	if err != nil {
		return err
	}

	if vm == nil {
		return errors.NewWithCodef(404, "Virtual machine %s not found on host %s", vmId, hostId)
	}

	err = s.CallDeleteHostVirtualMachine(host, vmId)
	if err != nil {
		return err
	}

	err = dbService.DeleteOrchestratorVirtualMachine(ctx, hostId, vmId)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrchestratorService) CallDeleteHostVirtualMachine(host *models.OrchestratorHost, vmId string) error {
	httpClient := s.getApiClient(*host)
	path := "/v1/machines/" + vmId
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	apiResponse, err := httpClient.Delete(url.String(), nil)
	if err != nil {
		return err
	}

	if apiResponse.StatusCode != 202 {
		return errors.NewWithCodef(400, "Error deleting virtual machine for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	s.Refresh()
	return nil
}
