package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineTimeSynchronizationFromApi(m models.TimeSynchronization) data_models.VirtualMachineTimeSynchronization {
	mapped := data_models.VirtualMachineTimeSynchronization{
		Enabled:                         m.Enabled,
		SmartMode:                       m.SmartMode,
		IntervalInSeconds:               m.IntervalInSeconds,
		TimezoneSynchronizationDisabled: m.TimezoneSynchronizationDisabled,
	}

	return mapped
}

func MapDtoVirtualMachineTimeSynchronizationToApi(m data_models.VirtualMachineTimeSynchronization) models.TimeSynchronization {
	mapped := models.TimeSynchronization{
		Enabled:                         m.Enabled,
		SmartMode:                       m.SmartMode,
		IntervalInSeconds:               m.IntervalInSeconds,
		TimezoneSynchronizationDisabled: m.TimezoneSynchronizationDisabled,
	}

	return mapped
}
