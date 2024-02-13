package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineMouseAndKeyboardFromApi(m models.MouseAndKeyboard) data_models.VirtualMachineMouseAndKeyboard {
	mapped := data_models.VirtualMachineMouseAndKeyboard{
		SmartMouseOptimizedForGames: m.SmartMouseOptimizedForGames,
		StickyMouse:                 m.StickyMouse,
		SmoothScrolling:             m.SmoothScrolling,
		KeyboardOptimizationMode:    m.KeyboardOptimizationMode,
	}

	return mapped
}

func MapDtoVirtualMachineMouseAndKeyboardToApi(m data_models.VirtualMachineMouseAndKeyboard) models.MouseAndKeyboard {
	mapped := models.MouseAndKeyboard{
		SmartMouseOptimizedForGames: m.SmartMouseOptimizedForGames,
		StickyMouse:                 m.StickyMouse,
		SmoothScrolling:             m.SmoothScrolling,
		KeyboardOptimizationMode:    m.KeyboardOptimizationMode,
	}

	return mapped
}
