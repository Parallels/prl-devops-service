package models

type VirtualMachineSpec struct {
	Name         string `json:"name"`
	Memory       string `json:"memory"`
	Cpu          string `json:"cpu"`
	Disk         string `json:"disk"`
	Distribution string `json:"distribution"`
	Version      string `json:"version"`
	OsType       string `json:"os_type"`
}
