package models

import (
	"Parallels/pd-api-service/common"
	"errors"
)

type CreateVirtualMachineRequest struct {
	Template     string `json:"template"`
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	Memory       int    `json:"memory"`
	Cpu          int    `json:"cpu"`
	Disk         int    `json:"disk"`
	DesiredState string `json:"desiredState"`
}

func (r *CreateVirtualMachineRequest) Validate() error {
	if r.Template == "" {
		return errors.New("Template cannot be empty")
	}

	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Memory <= 0 {
		common.Logger.Info("Memory is less than 0, setting to 2048")
		r.Memory = 2048
	}

	if r.Owner == "" {
		common.Logger.Info("Owner is empty, setting to 'root'")
		r.Owner = "root"
	}

	if r.Cpu <= 0 {
		common.Logger.Info("CPU is less than 0, setting to 2")
		r.Cpu = 2
	}

	if r.Disk <= 0 {
		common.Logger.Info("Disk is less than 0, setting to 20480")
		r.Disk = 20480
	}

	if r.DesiredState == "" {
		common.Logger.Info("DesiredState is empty, setting to 'running'")
		r.DesiredState = "running"
	}

	return nil
}
