package parallelsdesktop

import (
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

const CacheFindNumberOfRetries = 10
const CacheFindRetryDelay = 500 * time.Millisecond

func (s *ParallelsService) findVmInCacheAndSystem(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {

	for i := 0; i < CacheFindNumberOfRetries; i++ {
		s.RLock()
		vms := s.cachedLocalVms
		s.RUnlock()
		for _, vm := range vms {
			if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
				return &vm, nil
			}
		}
		ctx.LogInfof("VM with name or id %v not found in cache, retrying... (%d/%d)", idOrName, i+1, CacheFindNumberOfRetries)
		time.Sleep(CacheFindRetryDelay)
	}
	// To fetch all virtual machines (VMs), we intentionally call `GetVms` without specifying an ID or name.
	// This process might take some time because our goal is to refresh the entire cache.
	// If we find any VMs, it means the current cache is outdated, and we will proceed to update it.
	vms, err := s.GetVms(ctx, "")
	if err == nil {
		for _, vm := range vms {
			if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
				ctx.LogWarnf("Vm is not present in cache but found in machine updating cache")
				s.Lock()
				s.cachedLocalVms = vms
				s.Unlock()
				return &vm, nil
			}
		}
	}
	return nil, ErrVirtualMachineNotFound
}

func (s *ParallelsService) findVmSync(ctx basecontext.ApiContext, idOrName string) (*models.ParallelsVM, error) {
	return s.findVmInCacheAndSystem(ctx, idOrName)
}

func (s *ParallelsService) findVmWithStateInCacheAndSystem(ctx basecontext.ApiContext, idOrName string, state string) (*models.ParallelsVM, error) {
	for i := 0; i < CacheFindNumberOfRetries; i++ {
		s.RLock()
		vms := s.cachedLocalVms
		s.RUnlock()
		for _, vm := range vms {
			if (strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName)) && strings.EqualFold(vm.State, state) {
				return &vm, nil
			}
		}
		ctx.LogInfof("VM with name or id %v and state %v not found in cache, retrying... (%d/%d)", idOrName, state, i+1, CacheFindNumberOfRetries)
		time.Sleep(CacheFindRetryDelay)
	}
	vm, err := s.getVmForCurrentUser(ctx, idOrName)
	if err == nil {
		if (strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName)) && strings.EqualFold(vm.State, state) {
			ctx.LogWarnf("Vm with desired state is not present in cache but found in machine, updating cache")
			go func() {
				s.refreshCache(ctx)
			}()
			return &vm, nil
		}
	}
	return nil, ErrVirtualMachineNotFound
}
