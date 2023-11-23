package models

import (
	"os"

	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
)

type CreatePackerVirtualMachineRequest struct {
	Template     string                     `json:"template"`
	Owner        string                     `json:"owner"`
	Architecture string                     `json:"architecture"`
	Name         string                     `json:"name"`
	Specs        *CreateVirtualMachineSpecs `json:"specs,omitempty"`
	DesiredState string                     `json:"desiredState"`
}

func (r *CreatePackerVirtualMachineRequest) Validate() error {
	if r.Template == "" {
		return errors.New("Template cannot be empty")
	}

	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Specs == nil {
		r.Specs = &CreateVirtualMachineSpecs{
			Cpu:    "2",
			Disk:   "20480",
			Memory: "2048",
		}
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	if r.DesiredState == "" {
		r.DesiredState = "running"
	}

	return nil
}
