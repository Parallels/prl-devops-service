package models

type VirtualMachineStatusResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	IpConfigured string `json:"ip_configured"`
}
