package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineMemoryFromApi(m models.Memory) data_models.VirtualMachineMemory {
	mapped := data_models.VirtualMachineMemory{
		Size:    m.Size,
		Auto:    m.Auto,
		Hotplug: m.Hotplug,
	}

	return mapped
}

func MapDtoVirtualMachineMemoryToApi(m data_models.VirtualMachineMemory) models.Memory {
	mapped := models.Memory{
		Size:    m.Size,
		Auto:    m.Auto,
		Hotplug: m.Hotplug,
	}

	return mapped
}
