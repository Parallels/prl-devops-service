package models

type VirtualMachineSpec struct {
	Name         string `json:"name"`
	Memory       int    `json:"memory"`
	Cpu          int    `json:"cpu"`
	Disk         int    `json:"disk"`
	Distribution string `json:"distribution"`
	Version      string `json:"version"`
	OsType       string `json:"os_type"`
}
