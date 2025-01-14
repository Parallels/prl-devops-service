package models

type SystemUsageResponse struct {
	CpuType                  string              `json:"cpu_type,omitempty"`
	CpuBrand                 string              `json:"cpu_brand,omitempty"`
	DevOpsVersion            string              `json:"devops_version,omitempty"`
	OsName                   string              `json:"os_name,omitempty"`
	OsVersion                string              `json:"os_version,omitempty"`
	ParallelsDesktopVersion  string              `json:"parallels_desktop_version,omitempty"`
	ParallelsDesktopLicensed bool                `json:"parallels_desktop_licensed,omitempty"`
	ExternalIpAddress        string              `json:"external_ip_address,omitempty"`
	IsReverseProxyEnabled    bool                `json:"is_reverse_proxy_enabled"`
	IsLogStreamingEnabled    bool                `json:"is_log_streaming_enabled"`
	ReverseProxy             *SystemReverseProxy `json:"reverse_proxy,omitempty"`
	SystemReserved           *SystemUsageItem    `json:"system_reserved,omitempty"`
	Total                    *SystemUsageItem    `json:"total,omitempty"`
	TotalAvailable           *SystemUsageItem    `json:"total_available,omitempty"`
	TotalInUse               *SystemUsageItem    `json:"total_in_use,omitempty"`
	TotalReserved            *SystemUsageItem    `json:"total_reserved,omitempty"`
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

type SystemReverseProxy struct {
	Enabled bool               `json:"enabled,omitempty"`
	Host    string             `json:"host,omitempty"`
	Port    string             `json:"port,omitempty"`
	Hosts   []ReverseProxyHost `json:"hosts,omitempty"`
}
