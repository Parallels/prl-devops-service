package models

import (
	"os"

	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
)

type CreateVirtualMachineRequest struct {
	Name            string                              `json:"name"`
	Owner           string                              `json:"owner,omitempty"`
	PackerTemplate  *CreatePackerVirtualMachineRequest  `json:"packer_template,omitempty"`
	VagrantBox      *CreateVagrantMachineRequest        `json:"vagrant_box,omitempty"`
	CatalogManifest *CreateCatalogVirtualMachineRequest `json:"catalog_manifest,omitempty"`
}

func (r *CreateVirtualMachineRequest) Validate() error {
	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	if r.PackerTemplate != nil {
		if r.VagrantBox != nil || r.CatalogManifest != nil {
			return errors.New("Only one of packer_template, vagrant_box or catalog_manifest can be specified")
		}
		r.PackerTemplate.Name = r.Name
		r.PackerTemplate.Owner = r.Owner
		return r.PackerTemplate.Validate()
	}

	if r.VagrantBox != nil {
		if r.PackerTemplate != nil || r.CatalogManifest != nil {
			return errors.New("Only one of packer_template, vagrant_box or catalog_manifest can be specified")
		}
		r.VagrantBox.Name = r.Name
		r.VagrantBox.Owner = r.Owner
		return r.VagrantBox.Validate()
	}

	if r.CatalogManifest != nil {
		if r.PackerTemplate != nil || r.VagrantBox != nil {
			return errors.New("Only one of packer_template, vagrant_box or catalog_manifest can be specified")
		}
		r.CatalogManifest.Name = r.Name
		r.CatalogManifest.Owner = r.Owner
		return r.CatalogManifest.Validate()
	}

	return nil
}

type CreateVirtualMachineResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	CurrentState string `json:"current_state"`
}
