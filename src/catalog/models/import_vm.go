package models

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/helpers"
)

type ImportVmRequest struct {
	CatalogId         string            `json:"catalog_id"`
	Version           string            `json:"version"`
	Architecture      string            `json:"architecture"`
	Connection        string            `json:"connection,omitempty"`
	Description       string            `json:"description,omitempty"`
	IsCompressed      bool              `json:"is_compressed,omitempty"`
	Type              string            `json:"type,omitempty"`
	Force             bool              `json:"force,omitempty"`
	MachineRemotePath string            `json:"machine_remote_path,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	RequiredClaims    []string          `json:"required_claims,omitempty"`
	RequiredRoles     []string          `json:"required_roles,omitempty"`
	ProviderMetadata  map[string]string `json:"provider_metadata,omitempty"`
}

func (r *ImportVmRequest) Validate() error {
	if r.CatalogId == "" {
		return ErrPullMissingCatalogId
	}
	if r.Version == "" {
		return ErrPushMissingVersion
	}

	if r.Connection == "" {
		return ErrMissingConnection
	}

	if r.Architecture == "" {
		return ErrMissingArchitecture
	}

	if r.MachineRemotePath == "" {
		return ErrMissingMachineRemotePath
	}

	return nil
}

func (r *ImportVmRequest) Name() string {
	return fmt.Sprintf("%s-%s-%s", helpers.NormalizeString(r.CatalogId), helpers.NormalizeString(r.Architecture), helpers.NormalizeString(r.Version))
}

type ImportVmResponse struct {
	ID             string                         `json:"id"`
	LocalPath      string                         `json:"local_path"`
	MachineName    string                         `json:"machine_name"`
	Manifest       *VirtualMachineCatalogManifest `json:"manifest"`
	CleanupRequest *cleanupservice.CleanupRequest `json:"-"`
	Errors         []error                        `json:"-"`
}

func NewImportVmRequestResponse() *ImportVmResponse {
	return &ImportVmResponse{
		CleanupRequest: cleanupservice.NewCleanupRequest(),
	}
}

func (m *ImportVmResponse) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *ImportVmResponse) AddError(err error) {
	m.Errors = append(m.Errors, err)
}
