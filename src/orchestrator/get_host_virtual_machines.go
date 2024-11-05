package orchestrator

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostVirtualMachines(ctx basecontext.ApiContext, hostId string, filter string) ([]*models.VirtualMachine, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	vms, err := dbService.GetOrchestratorHostVirtualMachines(ctx, hostId, "")
	if err != nil {
		return nil, err
	}

	result := make([]*models.VirtualMachine, 0)

	// Updating Host State for each VM
	for _, vm := range vms {
	hostLoop:
		for _, host := range hosts {
			if strings.EqualFold(vm.HostId, host.ID) {
				vm.HostState = host.State
				break hostLoop
			}
		}

		result = append(result, &vm)
	}

	return result, nil
}

func (s *OrchestratorService) GetHostVirtualMachine(ctx basecontext.ApiContext, hostId string, vmId string) (*models.VirtualMachine, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	vm, err := dbService.GetOrchestratorHostVirtualMachine(ctx, hostId, vmId)
	if err != nil {
		return nil, err
	}

	// Updating Host State for each VM

	for _, host := range hosts {
		if vm.HostId == host.ID {
			vm.HostState = host.State
			break
		}
	}

	return vm, nil
}

func (s *OrchestratorService) GetHostAppleVirtualMachines(ctx basecontext.ApiContext, hostId string, vmId string) ([]*models.VirtualMachine, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	appleVms := make([]*models.VirtualMachine, 0)

	vms, err := dbService.GetOrchestratorHostVirtualMachines(ctx, hostId, "")
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Type == "APPLE_VZ_VM" {
			appleVms = append(appleVms, &vm)
		}
	}

	// Updating Host State for each VM
	for _, host := range hosts {
		for _, vm := range appleVms {
			if vm.HostId == host.ID {
				vm.HostState = host.State
				break
			}
		}
	}

	return appleVms, nil
}

func (s *OrchestratorService) GetHostVirtualMachinesInfo(host *models.OrchestratorHost) ([]api_models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	path := "/v1/machines"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response []api_models.ParallelsVM
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting virtual machines for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return response, nil
}
