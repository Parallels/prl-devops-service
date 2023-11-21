package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineCpuFromApi(m models.CPU) data_models.VirtualMachineCPU {
	mapped := data_models.VirtualMachineCPU{
		Cpus:    m.Cpus,
		Auto:    m.Auto,
		VTX:     m.VTX,
		Hotplug: m.Hotplug,
		Accl:    m.Accl,
		Mode:    m.Mode,
		Type:    m.Type,
	}

	return mapped
}

func MapDtoVirtualMachineCpuToApi(m data_models.VirtualMachineCPU) models.CPU {
	mapped := models.CPU{
		Cpus:    m.Cpus,
		Auto:    m.Auto,
		VTX:     m.VTX,
		Hotplug: m.Hotplug,
		Accl:    m.Accl,
		Mode:    m.Mode,
		Type:    m.Type,
	}

	return mapped
}
