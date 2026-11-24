package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineCdromFromApi(m models.Cdrom) data_models.VirtualMachineCdrom {
	mapped := data_models.VirtualMachineCdrom{
		Enabled: m.Enabled,
		Port:    m.Port,
		Image:   m.Image,
		State:   m.State,
	}

	return mapped
}

func MapDtoVirtualMachineCdromToApi(m data_models.VirtualMachineCdrom) models.Cdrom {
	mapped := models.Cdrom{
		Enabled: m.Enabled,
		Port:    m.Port,
		Image:   m.Image,
		State:   m.State,
	}

	return mapped
}
