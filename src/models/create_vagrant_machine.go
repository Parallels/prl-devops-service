package models

import (
	"os"

	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
)

type CreateVagrantMachineRequest struct {
	Box                   string                     `json:"box"`
	Version               string                     `json:"version"`
	Owner                 string                     `json:"owner"`
	Name                  string                     `json:"name"`
	Specs                 *CreateVirtualMachineSpecs `json:"specs,omitempty"`
	VagrantFilePath       string                     `json:"vagrant_file_path"`
	CustomVagrantConfig   string                     `json:"custom_vagrant_config"`
	CustomParallelsConfig string                     `json:"custom_parallels_config"`
}

func (r *CreateVagrantMachineRequest) Validate() error {
	if r.Box == "" && r.VagrantFilePath == "" {
		if r.Box == "" {
			return errors.New("Box cannot be empty")
		}

		if r.VagrantFilePath == "" {
			return errors.New("VagrantFilePath cannot be empty")
		}
	}

	if r.Box != "" && r.VagrantFilePath != "" {
		return errors.New("Only one of box or vagrant_file_path can be specified")
	}

	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	return nil
}
