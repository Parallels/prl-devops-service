package models

type HostResourceOverviewResponseItem struct {
	CpuType           string           `json:"cpu_type,omitempty"`
	CpuBrand          string           `json:"cpu_brand,omitempty"`
	DevOpsVersion     string           `json:"devops_version,omitempty"`
	OSName            string           `json:"os_name,omitempty"`
	OsVersion         string           `json:"os_version,omitempty"`
	ExternalIpAddress string           `json:"external_ip_address,omitempty"`
	Total             HostResourceItem `json:"total,omitempty"`
	TotalAvailable    HostResourceItem `json:"total_available,omitempty"`
	TotalInUse        HostResourceItem `json:"total_in_use,omitempty"`
	TotalReserved     HostResourceItem `json:"total_reserved,omitempty"`
}

type HostResources struct {
	CpuType           string           `json:"cpu_type,omitempty"`
	CpuBrand          string           `json:"cpu_brand,omitempty"`
	DevOpsVersion     string           `json:"devops_version,omitempty"`
	OsName            string           `json:"os_name,omitempty"`
	OsVersion         string           `json:"os_version,omitempty"`
	ExternalIpAddress string           `json:"external_ip_address,omitempty"`
	Total             HostResourceItem `json:"total,omitempty"`
	TotalAvailable    HostResourceItem `json:"total_available,omitempty"`
	TotalInUse        HostResourceItem `json:"total_in_use,omitempty"`
	TotalReserved     HostResourceItem `json:"total_reserved,omitempty"`
}

func (c *HostResources) Diff(source HostResources) bool {
	if c.CpuType != source.CpuType {
		return true
	}
	if c.Total.Diff(source.Total) {
		return true
	}
	if c.TotalAvailable.Diff(source.TotalAvailable) {
		return true
	}
	if c.TotalInUse.Diff(source.TotalInUse) {
		return true
	}
	if c.TotalReserved.Diff(source.TotalReserved) {
		return true
	}

	return false
}

type HostResourceItem struct {
	CpuType          string  `json:"cpu_type,omitempty"`
	PhysicalCpuCount int64   `json:"physical_cpu_count,omitempty"`
	LogicalCpuCount  int64   `json:"logical_cpu_count"`
	MemorySize       float64 `json:"memory_size,omitempty"`
	DiskSize         float64 `json:"disk_size,omitempty"`
	FreeDiskSize     float64 `json:"free_disk_size,omitempty"`
}

func (c *HostResourceItem) Diff(source HostResourceItem) bool {
	if c.PhysicalCpuCount != source.PhysicalCpuCount {
		return true
	}
	if c.LogicalCpuCount != source.LogicalCpuCount {
		return true
	}
	if c.MemorySize != source.MemorySize {
		return true
	}
	if c.DiskSize != source.DiskSize {
		return true
	}
	if c.FreeDiskSize != source.FreeDiskSize {
		return true
	}

	return false
}
