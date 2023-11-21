package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineCdromFromApi(m models.Cdrom0) data_models.VirtualMachineCdrom0 {
	mapped := data_models.VirtualMachineCdrom0{
		Enabled: m.Enabled,
		Port:    m.Port,
		Image:   m.Image,
		State:   m.State,
	}

	return mapped
}

func MapDtoVirtualMachineCdromToApi(m data_models.VirtualMachineCdrom0) models.Cdrom0 {
	mapped := models.Cdrom0{
		Enabled: m.Enabled,
		Port:    m.Port,
		Image:   m.Image,
		State:   m.State,
	}

	return mapped
}
