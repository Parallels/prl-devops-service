package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineSoundFromApi(m models.Sound) data_models.VirtualMachineSound {
	mapped := data_models.VirtualMachineSound{
		Enabled: m.Enabled,
		Output:  m.Output,
		Mixer:   m.Mixer,
	}

	return mapped
}

func MapDtoVirtualMachineSoundToApi(m data_models.VirtualMachineSound) models.Sound {
	mapped := models.Sound{
		Enabled: m.Enabled,
		Output:  m.Output,
		Mixer:   m.Mixer,
	}

	return mapped
}
