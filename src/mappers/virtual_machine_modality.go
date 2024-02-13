package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineModalityFromApi(m models.Modality) data_models.VirtualMachineModality {
	mapped := data_models.VirtualMachineModality{
		OpacityPercentage:  m.OpacityPercentage,
		StayOnTop:          m.StayOnTop,
		ShowOnAllSpaces:    m.ShowOnAllSpaces,
		CaptureMouseClicks: m.CaptureMouseClicks,
	}

	return mapped
}

func MapDtoVirtualMachineModalityToApi(m data_models.VirtualMachineModality) models.Modality {
	mapped := models.Modality{
		OpacityPercentage:  m.OpacityPercentage,
		StayOnTop:          m.StayOnTop,
		ShowOnAllSpaces:    m.ShowOnAllSpaces,
		CaptureMouseClicks: m.CaptureMouseClicks,
	}

	return mapped
}
