package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineGuestSharedFoldersFromApi(m models.GuestSharedFolders) data_models.VirtualMachineGuestSharedFolders {
	mapped := data_models.VirtualMachineGuestSharedFolders{
		Enabled:   m.Enabled,
		Automount: m.Automount,
	}

	return mapped
}

func MapDtoVirtualMachineGuestSharedFoldersToApi(m data_models.VirtualMachineGuestSharedFolders) models.GuestSharedFolders {
	mapped := models.GuestSharedFolders{
		Enabled:   m.Enabled,
		Automount: m.Automount,
	}

	return mapped
}
