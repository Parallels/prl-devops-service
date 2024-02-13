package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineExpirationFromApi(m models.Expiration) data_models.VirtualMachineExpiration {
	mapped := data_models.VirtualMachineExpiration{
		Enabled: m.Enabled,
	}

	return mapped
}

func MapDtoVirtualMachineExpirationToApi(m data_models.VirtualMachineExpiration) models.Expiration {
	mapped := models.Expiration{
		Enabled: m.Enabled,
	}

	return mapped
}
