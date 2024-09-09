package models

type PdFileMinimumSpecRequirement struct {
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
	Disk   int `json:"disk"`
}
