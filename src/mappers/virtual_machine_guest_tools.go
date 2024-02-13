package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineGuestToolsFromApi(m models.GuestTools) data_models.VirtualMachineGuestTools {
	mapped := data_models.VirtualMachineGuestTools{
		State:   m.State,
		Version: m.Version,
	}

	return mapped
}

func MapDtoVirtualMachineGuestToolsToApi(m data_models.VirtualMachineGuestTools) models.GuestTools {
	mapped := models.GuestTools{
		State:   m.State,
		Version: m.Version,
	}

	return mapped
}
