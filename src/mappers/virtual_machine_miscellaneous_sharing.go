package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineMiscellaneousSharingFromApi(m models.MiscellaneousSharing) data_models.VirtualMachineMiscellaneousSharing {
	mapped := data_models.VirtualMachineMiscellaneousSharing{
		SharedClipboard: m.SharedClipboard,
		SharedCloud:     m.SharedCloud,
	}

	return mapped
}

func MapDtoVirtualMachineMiscellaneousSharingToApi(m data_models.VirtualMachineMiscellaneousSharing) models.MiscellaneousSharing {
	mapped := models.MiscellaneousSharing{
		SharedClipboard: m.SharedClipboard,
		SharedCloud:     m.SharedCloud,
	}

	return mapped
}
