package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineFddFromApi(m models.Fdd) data_models.VirtualMachineFdd {
	mapped := data_models.VirtualMachineFdd{
		Enabled: m.Enabled,
		Port:    m.Port,
		Image:   m.Image,
		State:   m.State,
	}

	return mapped
}

func MapDtoVirtualMachineFddToApi(m data_models.VirtualMachineFdd) models.Fdd {
	mapped := models.Fdd{
		Enabled: m.Enabled,
		Port:    m.Port,
		Image:   m.Image,
		State:   m.State,
	}

	return mapped
}
