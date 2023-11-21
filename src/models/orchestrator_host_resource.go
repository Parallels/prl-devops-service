package models

type HostResourceOverviewResponse struct {
	CpuType        string           `json:"cpu_type,omitempty"`
	Total          HostResourceItem `json:"total"`
	TotalAvailable HostResourceItem `json:"total_available"`
	TotalInUse     HostResourceItem `json:"total_in_use"`
	TotalReserved  HostResourceItem `json:"total_reserved"`
}
