package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineAdvancedFromApi(m models.Advanced) data_models.VirtualMachineAdvanced {
	mapped := data_models.VirtualMachineAdvanced{
		VMHostnameSynchronization:    m.VMHostnameSynchronization,
		PublicSSHKeysSynchronization: m.PublicSSHKeysSynchronization,
		ShowDeveloperTools:           m.ShowDeveloperTools,
		SwipeFromEdges:               m.SwipeFromEdges,
		ShareHostLocation:            m.ShareHostLocation,
		RosettaLinux:                 m.RosettaLinux,
	}

	return mapped
}

func MapDtoVirtualMachineAdvancedToApi(m data_models.VirtualMachineAdvanced) models.Advanced {
	mapped := models.Advanced{
		VMHostnameSynchronization:    m.VMHostnameSynchronization,
		PublicSSHKeysSynchronization: m.PublicSSHKeysSynchronization,
		ShowDeveloperTools:           m.ShowDeveloperTools,
		SwipeFromEdges:               m.SwipeFromEdges,
		ShareHostLocation:            m.ShareHostLocation,
		RosettaLinux:                 m.RosettaLinux,
	}

	return mapped
}
