package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineSmbiosSettingsFromApi(m models.SMBIOSSettings) data_models.VirtualMachineSMBIOSSettings {
	mapped := data_models.VirtualMachineSMBIOSSettings{
		BIOSVersion:        m.BIOSVersion,
		SystemSerialNumber: m.SystemSerialNumber,
		BoardManufacturer:  m.BoardManufacturer,
	}

	return mapped
}

func MapDtoVirtualMachineSmbiosSettingsToApi(m data_models.VirtualMachineSMBIOSSettings) models.SMBIOSSettings {
	mapped := models.SMBIOSSettings{
		BIOSVersion:        m.BIOSVersion,
		SystemSerialNumber: m.SystemSerialNumber,
		BoardManufacturer:  m.BoardManufacturer,
	}

	return mapped
}
