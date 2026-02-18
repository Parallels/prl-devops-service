package parallelsdesktop

import (
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *ParallelsService) findVmInCacheAndSystem(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	vm, err := s.findVmSync(ctx, idOrName, true)
	if err == nil {
		return vm, nil
	}
	return s.findVmSync(ctx, idOrName, false)
}

func (s *ParallelsService) findVmSync(ctx basecontext.ApiContext, idOrName string, cached bool) (*models.ParallelsVM, error) {
	var err error
	var vms []models.ParallelsVM

	findVM := func(vms []models.ParallelsVM, idOrName string) (*models.ParallelsVM, error) {
		for _, vm := range vms {
			if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
				return &vm, nil
			}
		}
		return nil, ErrVirtualMachineNotFound
	}
	if cached {
		vms, err = s.GetCachedVms(ctx, "")
		if err == nil {
			for i := 0; i < 10; i++ {
				vm, err := findVM(vms, idOrName)
				if err == nil {
					return vm, nil
				}
				time.Sleep(500 * time.Millisecond)
			}
			ctx.LogInfof("VM with name or id %v not found after 10 attempts in cache", idOrName)
			return nil, ErrVirtualMachineNotFoundInCache
		}
	} else {
		vms, err = s.GetVms(ctx, "")
		if err == nil {
			return findVM(vms, idOrName)
		}
	}
	ctx.LogInfof("VM with name or id %v not found with cached=%v", idOrName, cached)
	return nil, ErrVirtualMachineNotFound
}
