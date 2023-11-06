package models

import (
	"errors"
	"os"

	"github.com/Parallels/pd-api-service/constants"
)

type RegisterVirtualMachineRequest struct {
	Path                      string `json:"path"`
	Owner                     string `json:"owner"`
	Uuid                      string `json:"uuid,omitempty"`
	MachineName               string `json:"machine_name,omitempty"`
	RegenerateSourceUuid      bool   `json:"regenerate_source_uuid,omitempty"`
	Force                     bool   `json:"force,omitempty"`
	DelayApplyingRestrictions bool   `json:"delay_applying_restrictions,omitempty"`
}

func (r *RegisterVirtualMachineRequest) Validate() error {
	if r.Path == "" {
		return errors.New("missing path")
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	return nil
}
