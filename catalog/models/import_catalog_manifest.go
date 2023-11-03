package models

import (
	"Parallels/pd-api-service/catalog/cleanupservice"
)

type ImportCatalogManifestRequest struct {
	ID               string            `json:"id"`
	Connection       string            `json:"connection,omitempty"`
	ProviderMetadata map[string]string `json:"provider_metadata,omitempty"`
}

func (r *ImportCatalogManifestRequest) Validate() error {
	if r.ID == "" {
		return ErrMissingId
	}
	if r.Connection == "" {
		return ErrMissingConnection
	}

	return nil
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
