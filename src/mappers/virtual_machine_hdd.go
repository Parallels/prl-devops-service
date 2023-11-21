package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineHddFromApi(m models.Hdd0) data_models.VirtualMachineHdd0 {
	mapped := data_models.VirtualMachineHdd0{
		Enabled:       m.Enabled,
		Port:          m.Port,
		Image:         m.Image,
		Type:          m.Type,
		Size:          m.Size,
		OnlineCompact: m.OnlineCompact,
	}

	return mapped
}

func MapDtoVirtualMachineHddToApi(m data_models.VirtualMachineHdd0) models.Hdd0 {
	mapped := models.Hdd0{
		Enabled:       m.Enabled,
		Port:          m.Port,
		Image:         m.Image,
		Type:          m.Type,
		Size:          m.Size,
		OnlineCompact: m.OnlineCompact,
	}

	return mapped
}
