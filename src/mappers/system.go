package mappers

import (
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapHostResourcesFromSystemUsageResponse(m models.SystemUsageResponse) data_models.HostResources {
	result := data_models.HostResources{
		CpuType:        m.CpuType,
		Total:          MapHostResourceItemFromSystemUsageItem(m.Total),
		TotalAvailable: MapHostResourceItemFromSystemUsageItem(m.TotalAvailable),
		TotalInUse:     MapHostResourceItemFromSystemUsageItem(m.TotalInUse),
		TotalReserved:  MapHostResourceItemFromSystemUsageItem(m.TotalReserved),
	}

	return result
}

func MapHostResourceItemFromSystemUsageItem(m models.SystemUsageItem) data_models.HostResourceItem {
	result := data_models.HostResourceItem{
		PhysicalCpuCount: m.PhysicalCpuCount,
		LogicalCpuCount:  m.LogicalCpuCount,
		MemorySize:       m.MemorySize,
		DiskSize:         m.DiskSize,
	}

	return result
}

func MapSystemUsageResponseFromHostResources(m data_models.HostResources) models.SystemUsageResponse {
	result := models.SystemUsageResponse{
		CpuType:        m.CpuType,
		Total:          MapSystemUsageItemFromHostResourceItem(m.Total),
		TotalAvailable: MapSystemUsageItemFromHostResourceItem(m.TotalAvailable),
		TotalInUse:     MapSystemUsageItemFromHostResourceItem(m.TotalInUse),
		TotalReserved:  MapSystemUsageItemFromHostResourceItem(m.TotalReserved),
	}

	return result
}

func MapSystemUsageItemFromHostResourceItem(m data_models.HostResourceItem) models.SystemUsageItem {
	result := models.SystemUsageItem{
		PhysicalCpuCount: m.PhysicalCpuCount,
		LogicalCpuCount:  m.LogicalCpuCount,
		MemorySize:       m.MemorySize,
		DiskSize:         m.DiskSize,
	}

	return result
}

func MapApiHostResourceItemFromHostResourceItem(m data_models.HostResourceItem) models.HostResourceItem {
	result := models.HostResourceItem{
		PhysicalCpuCount: m.PhysicalCpuCount,
		LogicalCpuCount:  m.LogicalCpuCount,
		MemorySize:       m.MemorySize,
		DiskSize:         m.DiskSize,
	}

	return result
}
