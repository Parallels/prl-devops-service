package models

type HostResourceOverviewResponseItem struct {
	CpuType        string            `json:"cpu_type,omitempty"`
	CpuBrand       string            `json:"cpu_brand,omitempty"`
	ReverseProxy   *HostReverseProxy `json:"reverse_proxy,omitempty"`
	TotalAppleVms  int64             `json:"total_apple_vms,omitempty"`
	SystemReserved HostResourceItem  `json:"system_reserved,omitempty"`
	Total          HostResourceItem  `json:"total,omitempty"`
	TotalAvailable HostResourceItem  `json:"total_available,omitempty"`
	TotalInUse     HostResourceItem  `json:"total_in_use,omitempty"`
	TotalReserved  HostResourceItem  `json:"total_reserved,omitempty"`
}

type HostResources struct {
	CpuType        string            `json:"cpu_type,omitempty"`
	CpuBrand       string            `json:"cpu_brand,omitempty"`
	ReverseProxy   *HostReverseProxy `json:"reverse_proxy,omitempty"`
	TotalAppleVms  int64             `json:"total_apple_vms,omitempty"`
	SystemReserved HostResourceItem  `json:"system_reserved,omitempty"`
	Total          HostResourceItem  `json:"total,omitempty"`
	TotalAvailable HostResourceItem  `json:"total_available,omitempty"`
	TotalInUse     HostResourceItem  `json:"total_in_use,omitempty"`
	TotalReserved  HostResourceItem  `json:"total_reserved,omitempty"`
}

func (c *HostResources) Diff(source HostResources) bool {
	if c.CpuType != source.CpuType {
		return true
	}
	if c.Total.Diff(source.Total) {
		return true
	}
	if c.TotalAppleVms != source.TotalAppleVms {
		return true
	}
	if c.ReverseProxy == nil && source.ReverseProxy != nil {
		return true
	}
	if c.ReverseProxy != nil && source.ReverseProxy == nil {
		return true
	}
	if c.ReverseProxy != nil && source.ReverseProxy != nil {
		c.ReverseProxy.Diff(*source.ReverseProxy)
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
	if c.TotalAppleVms != source.TotalAppleVms {
		return true
	}
	if c.SystemReserved.Diff(source.SystemReserved) {
		return true
	}

	return false
}

type HostResourceItem struct {
	CpuType          string  `json:"cpu_type,omitempty"`
	TotalAppleVms    int64   `json:"total_apple_vms,omitempty"`
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
	if c.TotalAppleVms != source.TotalAppleVms {
		return true
	}

	return false
}

type HostReverseProxy struct {
	Enabled bool               `json:"enabled,omitempty"`
	Host    string             `json:"host,omitempty"`
	Port    string             `json:"port,omitempty"`
	Hosts   []ReverseProxyHost `json:"hosts,omitempty"`
}

func (c *HostReverseProxy) Diff(source HostReverseProxy) bool {
	if c.Enabled != source.Enabled {
		return true
	}
	if c.Host != source.Host {
		return true
	}
	if c.Port != source.Port {
		return true
	}
	if len(c.Hosts) != len(source.Hosts) {
		return true
	}

	for i, host := range c.Hosts {
		if host.Diff(source.Hosts[i]) {
			return true
		}
	}

	return false
}
