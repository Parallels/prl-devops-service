package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineStartupAndShutdownFromApi(m models.StartupAndShutdown) data_models.VirtualMachineStartupAndShutdown {
	mapped := data_models.VirtualMachineStartupAndShutdown{
		Autostart:      m.Autostart,
		AutostartDelay: m.AutostartDelay,
		Autostop:       m.Autostop,
		StartupView:    m.StartupView,
		OnShutdown:     m.OnShutdown,
		OnWindowClose:  m.OnWindowClose,
		PauseIdle:      m.PauseIdle,
		UndoDisks:      m.UndoDisks,
	}

	return mapped
}

func MapDtoVirtualMachineStartupAndShutdownToApi(m data_models.VirtualMachineStartupAndShutdown) models.StartupAndShutdown {
	mapped := models.StartupAndShutdown{
		Autostart:      m.Autostart,
		AutostartDelay: m.AutostartDelay,
		Autostop:       m.Autostop,
		StartupView:    m.StartupView,
		OnShutdown:     m.OnShutdown,
		OnWindowClose:  m.OnWindowClose,
		PauseIdle:      m.PauseIdle,
		UndoDisks:      m.UndoDisks,
	}

	return mapped
}
