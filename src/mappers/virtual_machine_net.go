package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineNetFromApi(m models.Net) data_models.VirtualMachineNet {
	mapped := data_models.VirtualMachineNet{
		Enabled: m.Enabled,
		Type:    m.Type,
		MAC:     m.MAC,
		Card:    m.Card,
	}

	return mapped
}

func MapDtoVirtualMachineNetToApi(m data_models.VirtualMachineNet) models.Net {
	mapped := models.Net{
		Enabled: m.Enabled,
		Type:    m.Type,
		MAC:     m.MAC,
		Card:    m.Card,
	}

	return mapped
}
