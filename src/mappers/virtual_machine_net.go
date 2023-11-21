package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineNetFromApi(m models.Net0) data_models.VirtualMachineNet0 {
	mapped := data_models.VirtualMachineNet0{
		Enabled: m.Enabled,
		Type:    m.Type,
		MAC:     m.MAC,
		Card:    m.Card,
	}

	return mapped
}

func MapDtoVirtualMachineNetToApi(m data_models.VirtualMachineNet0) models.Net0 {
	mapped := models.Net0{
		Enabled: m.Enabled,
		Type:    m.Type,
		MAC:     m.MAC,
		Card:    m.Card,
	}

	return mapped
}
