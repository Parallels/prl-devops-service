package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineNetworkInformationFromApi(m models.NetworkInformation) data_models.VirtualMachineNetworkInformation {
	mapped := data_models.VirtualMachineNetworkInformation{
		Conditioned: m.Conditioned,
		Inbound:     MapDtoVirtualMachineNetworkInformationBoundFromApi(m.Inbound),
		Outbound:    MapDtoVirtualMachineNetworkInformationBoundFromApi(m.Outbound),
		IPAddresses: make([]data_models.VirtualMachineNetworkInformationIPAddress, 0),
	}
	if len(m.IPAddresses) > 0 {
		for _, ip := range m.IPAddresses {
			mapped.IPAddresses = append(mapped.IPAddresses, MapDtoVirtualMachineNetworkInformationIPAddressFromApi(ip))
		}
	}

	return mapped
}

func MapDtoVirtualNetworkInformationToApi(m data_models.VirtualMachineNetworkInformation) models.NetworkInformation {
	mapped := models.NetworkInformation{
		Conditioned: m.Conditioned,
		Inbound:     MapDtoVirtualMachineNetworkInformationBoundToApi(m.Inbound),
		Outbound:    MapDtoVirtualMachineNetworkInformationBoundToApi(m.Outbound),
		IPAddresses: make([]models.NetworkInformationIPAddress, 0),
	}

	if len(m.IPAddresses) > 0 {
		for _, ip := range m.IPAddresses {
			mapped.IPAddresses = append(mapped.IPAddresses, MapDtoVirtualMachineNetworkInformationIPAddressToApi(ip))
		}
	}

	return mapped
}

func MapDtoVirtualMachineNetworkInformationBoundFromApi(m models.NetworkInformationBound) data_models.VirtualMachineNetworkInformationBound {
	mapped := data_models.VirtualMachineNetworkInformationBound{
		Bandwidth:  m.Bandwidth,
		PacketLoss: m.PacketLoss,
		Delay:      m.Delay,
	}

	return mapped
}

func MapDtoVirtualMachineNetworkInformationBoundToApi(m data_models.VirtualMachineNetworkInformationBound) models.NetworkInformationBound {
	mapped := models.NetworkInformationBound{
		Bandwidth:  m.Bandwidth,
		PacketLoss: m.PacketLoss,
		Delay:      m.Delay,
	}

	return mapped
}

func MapDtoVirtualMachineNetworkInformationIPAddressFromApi(m models.NetworkInformationIPAddress) data_models.VirtualMachineNetworkInformationIPAddress {
	mapped := data_models.VirtualMachineNetworkInformationIPAddress{
		Type: m.Type,
		IP:   m.IP,
	}

	return mapped
}

func MapDtoVirtualMachineNetworkInformationIPAddressToApi(m data_models.VirtualMachineNetworkInformationIPAddress) models.NetworkInformationIPAddress {
	mapped := models.NetworkInformationIPAddress{
		Type: m.Type,
		IP:   m.IP,
	}

	return mapped
}
