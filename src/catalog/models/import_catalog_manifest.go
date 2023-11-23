package models

import (
	"fmt"

	"github.com/Parallels/pd-api-service/catalog/cleanupservice"
	"github.com/Parallels/pd-api-service/helpers"
)

type ImportCatalogManifestRequest struct {
	CatalogId        string            `json:"catalog_id"`
	Version          string            `json:"version"`
	Architecture     string            `json:"architecture"`
	Connection       string            `json:"connection,omitempty"`
	ProviderMetadata map[string]string `json:"provider_metadata,omitempty"`
}

func (r *ImportCatalogManifestRequest) Validate() error {
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

	return nil
}

func (r *ImportCatalogManifestRequest) Name() string {
	return fmt.Sprintf("%s-%s-%s", helpers.NormalizeString(r.CatalogId), helpers.NormalizeString(r.Architecture), helpers.NormalizeString(r.Version))
}

type ImportCatalogManifestResponse struct {
	ID             string                         `json:"id"`
	LocalPath      string                         `json:"local_path"`
	MachineName    string                         `json:"machine_name"`
	Manifest       *VirtualMachineCatalogManifest `json:"manifest"`
	CleanupRequest *cleanupservice.CleanupRequest `json:"-"`
	Errors         []error                        `json:"-"`
}

func NewImportCatalogManifestResponse() *ImportCatalogManifestResponse {
	return &ImportCatalogManifestResponse{
		CleanupRequest: cleanupservice.NewCleanupRequest(),
	}
}

func (m *ImportCatalogManifestResponse) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *ImportCatalogManifestResponse) AddError(err error) {
	m.Errors = append(m.Errors, err)
}
