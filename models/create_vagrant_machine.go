package models

import (
	"os"

	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
)

type CreateVagrantMachineRequest struct {
	Box                   string `json:"box"`
	Version               string `json:"version"`
	Owner                 string `json:"owner"`
	Name                  string `json:"name"`
	CustomVagrantConfig   string `json:"custom_vagrant_config"`
	CustomParallelsConfig string `json:"custom_parallels_config"`
}

func (r *CreateVagrantMachineRequest) Validate() error {
	if r.Box == "" {
		return errors.New("Box cannot be empty")
	}

	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	return nil
}
