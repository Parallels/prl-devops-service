package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineTravelModeFromApi(m models.TravelMode) data_models.VirtualMachineTravelMode {
	mapped := data_models.VirtualMachineTravelMode{
		EnterCondition: m.EnterCondition,
		EnterThreshold: m.EnterThreshold,
		QuitCondition:  m.QuitCondition,
	}

	return mapped
}

func MapDtoVirtualMachineTravelModeToApi(m data_models.VirtualMachineTravelMode) models.TravelMode {
	mapped := models.TravelMode{
		EnterCondition: m.EnterCondition,
		EnterThreshold: m.EnterThreshold,
		QuitCondition:  m.QuitCondition,
	}

	return mapped
}
