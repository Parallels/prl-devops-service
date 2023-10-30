package parallels_desktop

import (
	"Parallels/pd-api-service/models"
	"strings"
)

func (s *ParallelsService) findVm(idOrName string) (*models.ParallelsVM, error) {
	vms, err := s.GetVms()
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
			return &vm, nil
		}
	}

	return nil, nil
}
