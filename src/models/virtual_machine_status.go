package models

type VirtualMachineStatus struct {
	UUID         string `json:"uuid"`
	Status       string `json:"status"`
	IPConfigured string `json:"ip_configured"`
	Name         string `json:"name"`
}
