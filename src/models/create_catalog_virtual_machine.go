package models

import (
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
)

type CreateCatalogVirtualMachineRequest struct {
	CatalogId        string                     `json:"catalog_id"`
	Version          string                     `json:"version,omitempty"`
	Architecture     string                     `json:"architecture,omitempty"`
	Owner            string                     `json:"owner,omitempty"`
	MachineName      string                     `json:"machine_name,omitempty"`
	Connection       string                     `json:"connection,omitempty"`
	CatalogManagerId string                     `json:"catalog_manager_id,omitempty"`
	Path             string                     `json:"path,omitempty"`
	ProviderMetadata map[string]string          `json:"provider_metadata,omitempty"`
	StartAfterPull   bool                       `json:"start_after_pull,omitempty"`
	Specs            *CreateVirtualMachineSpecs `json:"specs,omitempty"`
}

func (r *CreateCatalogVirtualMachineRequest) Validate() error {
	if r.CatalogId == "" {
		return errors.NewWithCode("missing catalog id", 400)
	}
	if r.Version == "" {
		r.Version = constants.LATEST_TAG
	}

	if r.MachineName == "" {
		return errors.NewWithCode("missing machine name", 400)
	}

	connection := strings.TrimSpace(r.Connection)
	catalogManagerId := strings.TrimSpace(r.CatalogManagerId)
	if connection != "" && catalogManagerId != "" {
		return errors.NewWithCode("connection and catalog_manager_id cannot both be provided", 400)
	}
	if r.Owner == "" {
		owner := os.Getenv(constants.CURRENT_USER_ENV_VAR)
		if owner == "" {
			owner = "root"
		}
		r.Owner = owner
	}

	return nil
}
