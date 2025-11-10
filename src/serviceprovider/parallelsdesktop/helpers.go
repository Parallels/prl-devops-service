package parallelsdesktop

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *ParallelsService) findVm(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	vms, err := s.GetCachedVms(ctx, "")
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

func (s *ParallelsService) findVmSync(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	vms, err := s.GetCachedVms(ctx, "")
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
