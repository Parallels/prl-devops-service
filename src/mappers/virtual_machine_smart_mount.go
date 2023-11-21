package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineSmartMountFromApi(m models.SmartMount) data_models.VirtualMachineSmartMount {
	mapped := data_models.VirtualMachineSmartMount{
		Enabled:         m.Enabled,
		RemovableDrives: m.RemovableDrives,
		CDDVDDrives:     m.CDDVDDrives,
		NetworkShares:   m.NetworkShares,
	}

	return mapped
}

func MapDtoVirtualMachineSmartMountToApi(m data_models.VirtualMachineSmartMount) models.SmartMount {
	mapped := models.SmartMount{
		Enabled:         m.Enabled,
		RemovableDrives: m.RemovableDrives,
		CDDVDDrives:     m.CDDVDDrives,
		NetworkShares:   m.NetworkShares,
	}

	return mapped
}
