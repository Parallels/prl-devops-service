package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineSecurityFromApi(m models.Security) data_models.VirtualMachineSecurity {
	mapped := data_models.VirtualMachineSecurity{
		Encrypted:                m.Encrypted,
		TPMEnabled:               m.TPMEnabled,
		TPMType:                  m.TPMType,
		CustomPasswordProtection: m.CustomPasswordProtection,
		ConfigurationIsLocked:    m.ConfigurationIsLocked,
		Protected:                m.Protected,
		Archived:                 m.Archived,
		Packed:                   m.Packed,
	}

	return mapped
}

func MapDtoVirtualMachineSecurityToApi(m data_models.VirtualMachineSecurity) models.Security {
	mapped := models.Security{
		Encrypted:                m.Encrypted,
		TPMEnabled:               m.TPMEnabled,
		TPMType:                  m.TPMType,
		CustomPasswordProtection: m.CustomPasswordProtection,
		ConfigurationIsLocked:    m.ConfigurationIsLocked,
		Protected:                m.Protected,
		Archived:                 m.Archived,
		Packed:                   m.Packed,
	}

	return mapped
}
