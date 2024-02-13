package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineMemoryQuotaFromApi(m models.MemoryQuota) data_models.VirtualMachineMemoryQuota {
	mapped := data_models.VirtualMachineMemoryQuota{
		Auto: m.Auto,
	}

	return mapped
}

func MapDtoVirtualMachineMemoryQuotaToApi(m data_models.VirtualMachineMemoryQuota) models.MemoryQuota {
	mapped := models.MemoryQuota{
		Auto: m.Auto,
	}

	return mapped
}
