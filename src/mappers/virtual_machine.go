package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineFromApi(m models.ParallelsVM) data_models.VirtualMachine {
	mapped := data_models.VirtualMachine{
		HostId:                m.HostId,
		Host:                  m.Host,
		User:                  m.User,
		ID:                    m.ID,
		Name:                  m.Name,
		Description:           m.Description,
		Type:                  m.Type,
		State:                 m.State,
		OS:                    m.OS,
		Template:              m.Template,
		Uptime:                m.Uptime,
		HomePath:              m.HomePath,
		Home:                  m.Home,
		RestoreImage:          m.RestoreImage,
		GuestTools:            MapDtoVirtualMachineGuestToolsFromApi(m.GuestTools),
		MouseAndKeyboard:      MapDtoVirtualMachineMouseAndKeyboardFromApi(m.MouseAndKeyboard),
		StartupAndShutdown:    MapDtoVirtualMachineStartupAndShutdownFromApi(m.StartupAndShutdown),
		Optimization:          MapDtoVirtualMachineOptimizationFromApi(m.Optimization),
		TravelMode:            MapDtoVirtualMachineTravelModeFromApi(m.TravelMode),
		Security:              MapDtoVirtualMachineSecurityFromApi(m.Security),
		SmartGuard:            MapDtoVirtualMachineExpirationFromApi(m.SmartGuard),
		Modality:              MapDtoVirtualMachineModalityFromApi(m.Modality),
		FullScreen:            MapDtoVirtualMachineFullscreenFromApi(m.Fullscreen),
		Coherence:             MapDtoVirtualMachineCoherenceFromApi(m.Coherence),
		TimeSynchronization:   MapDtoVirtualMachineTimeSynchronizationFromApi(m.TimeSynchronization),
		Expiration:            MapDtoVirtualMachineExpirationFromApi(m.Expiration),
		BootOrder:             m.BootOrder,
		BIOSType:              m.BIOSType,
		EFISecureBoot:         m.EFISecureBoot,
		AllowSelectBootDevice: m.AllowSelectBootDevice,
		ExternalBootDevice:    m.ExternalBootDevice,
		SMBIOSSettings:        MapDtoVirtualMachineSmbiosSettingsFromApi(m.SMBIOSSettings),
		Hardware:              MapDtoVirtualMachineHardwareFromApi(m.Hardware),
		HostSharedFolders:     m.HostSharedFolders,
		HostDefinedSharing:    m.HostDefinedSharing,
		SharedProfile:         MapDtoVirtualMachineExpirationFromApi(m.SharedProfile),
		SharedApplications:    MapDtoVirtualMachineSharedApplicationsFromApi(m.SharedApplications),
		SmartMount:            MapDtoVirtualMachineSmartMountFromApi(m.SmartMount),
		MiscellaneousSharing:  MapDtoVirtualMachineMiscellaneousSharingFromApi(m.MiscellaneousSharing),
		Advanced:              MapDtoVirtualMachineAdvancedFromApi(m.Advanced),
		PrintManagement:       MapDtoVirtualMachinePrintManagementFromApi(m.PrintManagement),
		GuestSharedFolders:    MapDtoVirtualMachineGuestSharedFoldersFromApi(m.GuestSharedFolders),
	}

	return mapped
}

func MapDtoVirtualMachineToApi(m data_models.VirtualMachine) models.ParallelsVM {
	mapped := models.ParallelsVM{
		HostId:                m.HostId,
		Host:                  m.Host,
		User:                  m.User,
		ID:                    m.ID,
		Name:                  m.Name,
		Description:           m.Description,
		Type:                  m.Type,
		State:                 m.State,
		OS:                    m.OS,
		Template:              m.Template,
		Uptime:                m.Uptime,
		HomePath:              m.HomePath,
		Home:                  m.Home,
		RestoreImage:          m.RestoreImage,
		GuestTools:            MapDtoVirtualMachineGuestToolsToApi(m.GuestTools),
		MouseAndKeyboard:      MapDtoVirtualMachineMouseAndKeyboardToApi(m.MouseAndKeyboard),
		StartupAndShutdown:    MapDtoVirtualMachineStartupAndShutdownToApi(m.StartupAndShutdown),
		Optimization:          MapDtoVirtualMachineOptimizationToApi(m.Optimization),
		TravelMode:            MapDtoVirtualMachineTravelModeToApi(m.TravelMode),
		Security:              MapDtoVirtualMachineSecurityToApi(m.Security),
		SmartGuard:            MapDtoVirtualMachineExpirationToApi(m.SmartGuard),
		Modality:              MapDtoVirtualMachineModalityToApi(m.Modality),
		Fullscreen:            MapDtoVirtualMachineFullscreenToApi(m.FullScreen),
		Coherence:             MapDtoVirtualMachineCoherenceToApi(m.Coherence),
		TimeSynchronization:   MapDtoVirtualMachineTimeSynchronizationToApi(m.TimeSynchronization),
		Expiration:            MapDtoVirtualMachineExpirationToApi(m.Expiration),
		BootOrder:             m.BootOrder,
		BIOSType:              m.BIOSType,
		EFISecureBoot:         m.EFISecureBoot,
		AllowSelectBootDevice: m.AllowSelectBootDevice,
		ExternalBootDevice:    m.ExternalBootDevice,
		SMBIOSSettings:        MapDtoVirtualMachineSmbiosSettingsToApi(m.SMBIOSSettings),
		Hardware:              MapDtoVirtualMachineHardwareToApi(m.Hardware),
		HostSharedFolders:     m.HostSharedFolders,
		HostDefinedSharing:    m.HostDefinedSharing,
		SharedProfile:         MapDtoVirtualMachineExpirationToApi(m.SharedProfile),
		SharedApplications:    MapDtoVirtualMachineSharedApplicationsToApi(m.SharedApplications),
		SmartMount:            MapDtoVirtualMachineSmartMountToApi(m.SmartMount),
		MiscellaneousSharing:  MapDtoVirtualMachineMiscellaneousSharingToApi(m.MiscellaneousSharing),
		Advanced:              MapDtoVirtualMachineAdvancedToApi(m.Advanced),
		PrintManagement:       MapDtoVirtualMachinePrintManagementToApi(m.PrintManagement),
		GuestSharedFolders:    MapDtoVirtualMachineGuestSharedFoldersToApi(m.GuestSharedFolders),
	}

	return mapped
}
