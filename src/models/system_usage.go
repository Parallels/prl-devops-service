package models

type SystemUsageResponse struct {
	CpuType                  string           `json:"cpu_type,omitempty"`
	CpuBrand                 string           `json:"cpu_brand,omitempty"`
	DevOpsVersion            string           `json:"devops_version,omitempty"`
	ParallelsDesktopVersion  string           `json:"parallels_desktop_version,omitempty"`
	ParallelsDesktopLicensed bool             `json:"parallels_desktop_licensed,omitempty"`
	SystemReserved           *SystemUsageItem `json:"system_reserved,omitempty"`
	Total                    *SystemUsageItem `json:"total,omitempty"`
	TotalAvailable           *SystemUsageItem `json:"total_available,omitempty"`
	TotalInUse               *SystemUsageItem `json:"total_in_use,omitempty"`
	TotalReserved            *SystemUsageItem `json:"total_reserved,omitempty"`
}

type SystemUsageItem struct {
	PhysicalCpuCount int64   `json:"physical_cpu_count,omitempty"`
	LogicalCpuCount  int64   `json:"logical_cpu_count"`
	MemorySize       float64 `json:"memory_size"`
	DiskSize         float64 `json:"disk_count"`
}

type SystemHardwareInfo struct {
	CpuType          string  `json:"cpu_type,omitempty"`
	CpuBrand         string  `json:"cpu_brand,omitempty"`
	PhysicalCpuCount int     `json:"physical_cpu_count,omitempty"`
	LogicalCpuCount  int     `json:"logical_cpu_count,omitempty"`
	MemorySize       float64 `json:"memory_size,omitempty"`
	DiskSize         float64 `json:"disk_size,omitempty"`
	FreeDiskSize     float64 `json:"free_disk_size,omitempty"`
}
