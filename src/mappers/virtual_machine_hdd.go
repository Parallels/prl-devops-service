package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineHddFromApi(m models.Hdd) data_models.VirtualMachineHdd {
	mapped := data_models.VirtualMachineHdd{
		Enabled:       m.Enabled,
		Port:          m.Port,
		Image:         m.Image,
		Type:          m.Type,
		Size:          m.Size,
		OnlineCompact: m.OnlineCompact,
	}

	return mapped
}

func MapDtoVirtualMachineHddToApi(m data_models.VirtualMachineHdd) models.Hdd {
	mapped := models.Hdd{
		Enabled:       m.Enabled,
		Port:          m.Port,
		Image:         m.Image,
		Type:          m.Type,
		Size:          m.Size,
		OnlineCompact: m.OnlineCompact,
	}

	return mapped
}
