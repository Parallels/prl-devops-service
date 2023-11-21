package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapDtoVirtualMachineSharedApplicationsFromApi(m models.SharedApplications) data_models.VirtualMachineSharedApplications {
	mapped := data_models.VirtualMachineSharedApplications{
		Enabled:                      m.Enabled,
		HostToGuestAppsSharing:       m.HostToGuestAppsSharing,
		GuestToHostAppsSharing:       m.GuestToHostAppsSharing,
		ShowGuestAppsFolderInDock:    m.ShowGuestAppsFolderInDock,
		ShowGuestNotifications:       m.ShowGuestNotifications,
		BounceDockIconWhenAppFlashes: m.BounceDockIconWhenAppFlashes,
	}

	return mapped
}

func MapDtoVirtualMachineSharedApplicationsToApi(m data_models.VirtualMachineSharedApplications) models.SharedApplications {
	mapped := models.SharedApplications{
		Enabled:                      m.Enabled,
		HostToGuestAppsSharing:       m.HostToGuestAppsSharing,
		GuestToHostAppsSharing:       m.GuestToHostAppsSharing,
		ShowGuestAppsFolderInDock:    m.ShowGuestAppsFolderInDock,
		ShowGuestNotifications:       m.ShowGuestNotifications,
		BounceDockIconWhenAppFlashes: m.BounceDockIconWhenAppFlashes,
	}

	return mapped
}
