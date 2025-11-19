package parallelsdesktop

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *ParallelsService) findVm(ctx basecontext.ApiContext, idOrName string, cached bool) (*models.ParallelsVM, error) {
	var err error
	var vms []models.ParallelsVM
	if cached {
		vms, err = s.GetCachedVms(ctx, "")
	} else {
		vms, err = s.GetVms(ctx, "")
	}
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
			return &vm, nil
		}
	}

	return nil, ErrVirtualMachineNotFound
}

func (s *ParallelsService) findVmSync(ctx basecontext.ApiContext, idOrName string, cached bool) (*models.ParallelsVM, error) {
	var err error
	var vms []models.ParallelsVM
	if cached {
		vms, err = s.GetCachedVms(ctx, "")
	} else {
		vms, err = s.GetVms(ctx, "")
	}

	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
			return &vm, nil
		}
	}

	return nil, ErrVirtualMachineNotFound
}
