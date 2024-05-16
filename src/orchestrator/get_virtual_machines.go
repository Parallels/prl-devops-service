package orchestrator

import (
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetVirtualMachines(ctx basecontext.ApiContext, filter string) ([]models.VirtualMachine, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	s.Refresh()

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	vms, err := dbService.GetOrchestratorVirtualMachines(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]models.VirtualMachine, 0)

	// Updating Host State for each VM
	for _, vm := range vms {
	hostLoop:
		for _, host := range hosts {
			if vm.HostId == host.ID {
				vm.HostState = host.State
				break hostLoop
			}
		}

		result = append(result, vm)
	}

	return result, nil
}

func (s *OrchestratorService) GetVirtualMachine(ctx basecontext.ApiContext, idOrName string) (*models.VirtualMachine, error) {
	s.Refresh()

	retryCount := 0
	var resultVm *models.VirtualMachine

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	for {
		vms, err := dbService.GetOrchestratorVirtualMachines(ctx, "")
		if err != nil {
			return nil, err
		}

		resultVm = findVmById(vms, idOrName)

		if resultVm != nil || retryCount >= 10 {
			break
		}

		s.Refresh()

		retryCount++
		time.Sleep(1 * time.Second)
	}

	if resultVm == nil {
		return nil, errors.NewWithCodef(404, "Virtual machine %s not found", idOrName)
	}

	// Updating Host State for each VM
	for _, host := range hosts {
		if strings.EqualFold(resultVm.HostId, host.ID) {
			resultVm.HostState = host.State
			break
		}
	}

	return resultVm, nil
}

func findVmById(vms []models.VirtualMachine, idOrName string) *models.VirtualMachine {
	for _, vm := range vms {
		if strings.EqualFold(vm.ID, idOrName) || strings.EqualFold(vm.Name, idOrName) {
			return &vm
		}
	}
	return nil
}
