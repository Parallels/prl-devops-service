package mappers

import (
	"encoding/json"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineHardwareFromApi(m models.Hardware) data_models.VirtualMachineHardware {
	mapped := data_models.VirtualMachineHardware{
		CPU:         MapDtoVirtualMachineCpuFromApi(m.CPU),
		Memory:      MapDtoVirtualMachineMemoryFromApi(m.Memory),
		Video:       MapDtoVirtualMachineVideoFromApi(m.Video),
		MemoryQuota: MapDtoVirtualMachineMemoryQuotaFromApi(m.MemoryQuota),
		USB:         MapDtoVirtualMachineExpirationFromApi(m.USB),
	}

	if len(m.Hdds) > 0 {
		mapped.Hdds = make(map[string]data_models.VirtualMachineHdd, len(m.Hdds))
		for key, device := range m.Hdds {
			mapped.Hdds[key] = MapDtoVirtualMachineHddFromApi(device)
		}
	}
	if len(m.CdRoms) > 0 {
		mapped.CdRoms = make(map[string]data_models.VirtualMachineCdrom, len(m.CdRoms))
		for key, device := range m.CdRoms {
			mapped.CdRoms[key] = MapDtoVirtualMachineCdromFromApi(device)
		}
	}
	if len(m.Nets) > 0 {
		mapped.Nets = make(map[string]data_models.VirtualMachineNet, len(m.Nets))
		for key, device := range m.Nets {
			mapped.Nets[key] = MapDtoVirtualMachineNetFromApi(device)
		}
	}
	if len(m.Sounds) > 0 {
		mapped.Sounds = make(map[string]data_models.VirtualMachineSound, len(m.Sounds))
		for key, device := range m.Sounds {
			mapped.Sounds[key] = MapDtoVirtualMachineSoundFromApi(device)
		}
	}
	if len(m.Fdds) > 0 {
		mapped.Fdds = make(map[string]data_models.VirtualMachineFdd, len(m.Fdds))
		for key, device := range m.Fdds {
			mapped.Fdds[key] = MapDtoVirtualMachineFddFromApi(device)
		}
	}
	// Unknown hardware keys are carried over verbatim so a prlctl-reported
	// device type this service does not model is never lost on persist.
	if len(m.Extras) > 0 {
		mapped.Extras = make(map[string]json.RawMessage, len(m.Extras))
		for key, raw := range m.Extras {
			mapped.Extras[key] = append(json.RawMessage(nil), raw...)
		}
	}

	return mapped
}

func MapDtoVirtualMachineHardwareToApi(m data_models.VirtualMachineHardware) models.Hardware {
	mapped := models.Hardware{
		CPU:         MapDtoVirtualMachineCpuToApi(m.CPU),
		Memory:      MapDtoVirtualMachineMemoryToApi(m.Memory),
		Video:       MapDtoVirtualMachineVideoToApi(m.Video),
		MemoryQuota: MapDtoVirtualMachineMemoryQuotaToApi(m.MemoryQuota),
		USB:         MapDtoVirtualMachineExpirationToApi(m.USB),
	}

	if len(m.Hdds) > 0 {
		mapped.Hdds = make(map[string]models.Hdd, len(m.Hdds))
		for key, device := range m.Hdds {
			mapped.Hdds[key] = MapDtoVirtualMachineHddToApi(device)
		}
	}
	if len(m.CdRoms) > 0 {
		mapped.CdRoms = make(map[string]models.Cdrom, len(m.CdRoms))
		for key, device := range m.CdRoms {
			mapped.CdRoms[key] = MapDtoVirtualMachineCdromToApi(device)
		}
	}
	if len(m.Nets) > 0 {
		mapped.Nets = make(map[string]models.Net, len(m.Nets))
		for key, device := range m.Nets {
			mapped.Nets[key] = MapDtoVirtualMachineNetToApi(device)
		}
	}
	if len(m.Sounds) > 0 {
		mapped.Sounds = make(map[string]models.Sound, len(m.Sounds))
		for key, device := range m.Sounds {
			mapped.Sounds[key] = MapDtoVirtualMachineSoundToApi(device)
		}
	}
	if len(m.Fdds) > 0 {
		mapped.Fdds = make(map[string]models.Fdd, len(m.Fdds))
		for key, device := range m.Fdds {
			mapped.Fdds[key] = MapDtoVirtualMachineFddToApi(device)
		}
	}
	if len(m.Extras) > 0 {
		mapped.Extras = make(map[string]json.RawMessage, len(m.Extras))
		for key, raw := range m.Extras {
			mapped.Extras[key] = append(json.RawMessage(nil), raw...)
		}
	}

	return mapped
}
