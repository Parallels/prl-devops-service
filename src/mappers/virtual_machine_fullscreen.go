package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineFullscreenFromApi(m models.Fullscreen) data_models.VirtualMachineFullscreen {
	mapped := data_models.VirtualMachineFullscreen{
		UseAllDisplays:        m.UseAllDisplays,
		ActivateSpacesOnClick: m.ActivateSpacesOnClick,
		OptimizeForGames:      m.OptimizeForGames,
		GammaControl:          m.GammaControl,
		ScaleViewMode:         m.ScaleViewMode,
	}

	return mapped
}

func MapDtoVirtualMachineFullscreenToApi(m data_models.VirtualMachineFullscreen) models.Fullscreen {
	mapped := models.Fullscreen{
		UseAllDisplays:        m.UseAllDisplays,
		ActivateSpacesOnClick: m.ActivateSpacesOnClick,
		OptimizeForGames:      m.OptimizeForGames,
		GammaControl:          m.GammaControl,
		ScaleViewMode:         m.ScaleViewMode,
	}

	return mapped
}
