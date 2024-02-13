package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineHardwareFromApi(m models.Hardware) data_models.VirtualMachineHardware {
	mapped := data_models.VirtualMachineHardware{
		CPU:         MapDtoVirtualMachineCpuFromApi(m.CPU),
		Memory:      MapDtoVirtualMachineMemoryFromApi(m.Memory),
		Video:       MapDtoVirtualMachineVideoFromApi(m.Video),
		MemoryQuota: MapDtoVirtualMachineMemoryQuotaFromApi(m.MemoryQuota),
		Hdd0:        MapDtoVirtualMachineHddFromApi(m.Hdd0),
		Cdrom0:      MapDtoVirtualMachineCdromFromApi(m.Cdrom0),
		USB:         MapDtoVirtualMachineExpirationFromApi(m.USB),
		Net0:        MapDtoVirtualMachineNetFromApi(m.Net0),
		Sound0:      MapDtoVirtualMachineSoundFromApi(m.Sound0),
	}

	return mapped
}

func MapDtoVirtualMachineHardwareToApi(m data_models.VirtualMachineHardware) models.Hardware {
	mapped := models.Hardware{
		CPU:         MapDtoVirtualMachineCpuToApi(m.CPU),
		Memory:      MapDtoVirtualMachineMemoryToApi(m.Memory),
		Video:       MapDtoVirtualMachineVideoToApi(m.Video),
		MemoryQuota: MapDtoVirtualMachineMemoryQuotaToApi(m.MemoryQuota),
		Hdd0:        MapDtoVirtualMachineHddToApi(m.Hdd0),
		Cdrom0:      MapDtoVirtualMachineCdromToApi(m.Cdrom0),
		USB:         MapDtoVirtualMachineExpirationToApi(m.USB),
		Net0:        MapDtoVirtualMachineNetToApi(m.Net0),
		Sound0:      MapDtoVirtualMachineSoundToApi(m.Sound0),
	}

	return mapped
}
