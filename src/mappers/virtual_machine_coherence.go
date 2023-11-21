package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineCoherenceFromApi(m models.Coherence) data_models.VirtualMachineCoherence {
	mapped := data_models.VirtualMachineCoherence{
		ShowWindowsSystrayInMACMenu: m.ShowWindowsSystrayInMACMenu,
		AutoSwitchToFullScreen:      m.AutoSwitchToFullScreen,
		DisableAero:                 m.DisableAero,
		HideMinimizedWindows:        m.HideMinimizedWindows,
	}

	return mapped
}

func MapDtoVirtualMachineCoherenceToApi(m data_models.VirtualMachineCoherence) models.Coherence {
	mapped := models.Coherence{
		ShowWindowsSystrayInMACMenu: m.ShowWindowsSystrayInMACMenu,
		AutoSwitchToFullScreen:      m.AutoSwitchToFullScreen,
		DisableAero:                 m.DisableAero,
		HideMinimizedWindows:        m.HideMinimizedWindows,
	}

	return mapped
}
