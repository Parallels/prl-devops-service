package models

import (
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/errors"
	"os"
)

type CreatePackerVirtualMachineRequest struct {
	Template     string `json:"template"`
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	Memory       string `json:"memory"`
	Cpu          string `json:"cpu"`
	Disk         string `json:"disk"`
	DesiredState string `json:"desiredState"`
}

func (r *CreatePackerVirtualMachineRequest) Validate() error {
	if r.Template == "" {
		return errors.New("Template cannot be empty")
	}

	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Memory == "" {
		r.Memory = "2048"
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	if r.Cpu == "" {
		r.Cpu = "2"
	}

	if r.Disk == "" {
		r.Disk = "20480"
	}

	if r.DesiredState == "" {
		r.DesiredState = "running"
	}

	return nil
}
