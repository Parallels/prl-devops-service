package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineUsbAndBluetoothFromApi(m models.USBAndBluetooth) data_models.VirtualMachineUSBAndBluetooth {
	mapped := data_models.VirtualMachineUSBAndBluetooth{
		AutomaticSharingCameras:    m.AutomaticSharingCameras,
		AutomaticSharingBluetooth:  m.AutomaticSharingBluetooth,
		AutomaticSharingSmartCards: m.AutomaticSharingSmartCards,
		AutomaticSharingGamepads:   m.AutomaticSharingGamepads,
		SupportUSB30:               m.SupportUSB30,
	}

	return mapped
}

func MapDtoVirtualMachineUsbAndBluetoothToApi(m data_models.VirtualMachineUSBAndBluetooth) models.USBAndBluetooth {
	mapped := models.USBAndBluetooth{
		AutomaticSharingCameras:    m.AutomaticSharingCameras,
		AutomaticSharingBluetooth:  m.AutomaticSharingBluetooth,
		AutomaticSharingSmartCards: m.AutomaticSharingSmartCards,
		AutomaticSharingGamepads:   m.AutomaticSharingGamepads,
		SupportUSB30:               m.SupportUSB30,
	}

	return mapped
}
