package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func MapDtoVirtualMachineVideoFromApi(m models.Video) data_models.VirtualMachineVideo {
	mapped := data_models.VirtualMachineVideo{
		AdapterType:           m.AdapterType,
		Size:                  m.Size,
		The3DAcceleration:     m.The3DAcceleration,
		VerticalSync:          m.VerticalSync,
		HighResolution:        m.HighResolution,
		HighResolutionInGuest: m.HighResolutionInGuest,
		NativeScalingInGuest:  m.NativeScalingInGuest,
		AutomaticVideoMemory:  m.AutomaticVideoMemory,
	}

	return mapped
}

func MapDtoVirtualMachineVideoToApi(m data_models.VirtualMachineVideo) models.Video {
	mapped := models.Video{
		AdapterType:           m.AdapterType,
		Size:                  m.Size,
		The3DAcceleration:     m.The3DAcceleration,
		VerticalSync:          m.VerticalSync,
		HighResolution:        m.HighResolution,
		HighResolutionInGuest: m.HighResolutionInGuest,
		NativeScalingInGuest:  m.NativeScalingInGuest,
		AutomaticVideoMemory:  m.AutomaticVideoMemory,
	}

	return mapped
}
