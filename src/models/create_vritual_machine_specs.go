package models

import "strconv"

type CreateVirtualMachineSpecs struct {
	Type   string `json:"type,omitempty"`
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	Disk   string `json:"disk,omitempty"`
}

func (c *CreateVirtualMachineSpecs) GetCpuCount() int64 {
	count, err := strconv.Atoi(c.Cpu)
	if err != nil {
		return 1
	}

	return int64(count)
}

func (c *CreateVirtualMachineSpecs) GetMemorySize() float64 {
	size, err := strconv.ParseFloat(c.Memory, 64)
	if err != nil {
		return 1
	}

	return size
}

func (c *CreateVirtualMachineSpecs) GetDiskSize() float64 {
	size, err := strconv.ParseFloat(c.Disk, 64)
	if err != nil {
		return 1
	}

	return size
}
