package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineOptimizationFromApi(m models.Optimization) data_models.VirtualMachineOptimization {
	mapped := data_models.VirtualMachineOptimization{
		FasterVirtualMachine:     m.FasterVirtualMachine,
		HypervisorType:           m.HypervisorType,
		AdaptiveHypervisor:       m.AdaptiveHypervisor,
		DisabledWindowsLogo:      m.DisabledWindowsLogo,
		AutoCompressVirtualDisks: m.AutoCompressVirtualDisks,
		NestedVirtualization:     m.NestedVirtualization,
		PMUVirtualization:        m.PMUVirtualization,
		LongerBatteryLife:        m.LongerBatteryLife,
		ShowBatteryStatus:        m.ShowBatteryStatus,
		ResourceQuota:            m.ResourceQuota,
	}

	return mapped
}

func MapDtoVirtualMachineOptimizationToApi(m data_models.VirtualMachineOptimization) models.Optimization {
	mapped := models.Optimization{
		FasterVirtualMachine:     m.FasterVirtualMachine,
		HypervisorType:           m.HypervisorType,
		AdaptiveHypervisor:       m.AdaptiveHypervisor,
		DisabledWindowsLogo:      m.DisabledWindowsLogo,
		AutoCompressVirtualDisks: m.AutoCompressVirtualDisks,
		NestedVirtualization:     m.NestedVirtualization,
		PMUVirtualization:        m.PMUVirtualization,
		LongerBatteryLife:        m.LongerBatteryLife,
		ShowBatteryStatus:        m.ShowBatteryStatus,
		ResourceQuota:            m.ResourceQuota,
	}

	return mapped
}
