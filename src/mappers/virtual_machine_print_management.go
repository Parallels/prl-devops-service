package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachinePrintManagementFromApi(m models.PrintManagement) data_models.VirtualMachinePrintManagement {
	mapped := data_models.VirtualMachinePrintManagement{
		SynchronizeWithHostPrinters: m.SynchronizeWithHostPrinters,
		SynchronizeDefaultPrinter:   m.SynchronizeDefaultPrinter,
		ShowHostPrinterUI:           m.ShowHostPrinterUI,
	}

	return mapped
}

func MapDtoVirtualMachinePrintManagementToApi(m data_models.VirtualMachinePrintManagement) models.PrintManagement {
	mapped := models.PrintManagement{
		SynchronizeWithHostPrinters: m.SynchronizeWithHostPrinters,
		SynchronizeDefaultPrinter:   m.SynchronizeDefaultPrinter,
		ShowHostPrinterUI:           m.ShowHostPrinterUI,
	}

	return mapped
}
