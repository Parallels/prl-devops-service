package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineSoundFromApi(m models.Sound0) data_models.VirtualMachineSound0 {
	mapped := data_models.VirtualMachineSound0{
		Enabled: m.Enabled,
		Output:  m.Output,
		Mixer:   m.Mixer,
	}

	return mapped
}

func MapDtoVirtualMachineSoundToApi(m data_models.VirtualMachineSound0) models.Sound0 {
	mapped := models.Sound0{
		Enabled: m.Enabled,
		Output:  m.Output,
		Mixer:   m.Mixer,
	}

	return mapped
}
