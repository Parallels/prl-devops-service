package models

import "errors"

type PullCatalogManifestRequest struct {
	ID               string            `json:"id"`
	MachineName      string            `json:"machine_name,omitempty"`
	Path             string            `json:"path,omitempty"`
	Connection       string            `json:"connection,omitempty"`
	ProviderMetadata map[string]string `json:"provider_metadata,omitempty"`
}

var ErrMissingPath = errors.New("missing path")
var ErrMissingId = errors.New("missing id")

func (r *PullCatalogManifestRequest) Validate() error {
	if r.Path == "" {
		return ErrMissingPath
	}
	if r.ID == "" {
		return ErrMissingId
	}

	return nil
}

type PullCatalogManifestResponse struct {
	ID          string                  `json:"id"`
	LocalPath   string                  `json:"local_path"`
	MachineName string                  `json:"machine_name"`
	Manifest    *VirtualMachineManifest `json:"manifest"`
}
