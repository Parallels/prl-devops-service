package models

import (
	"os"

	"github.com/Parallels/pd-api-service/catalog/cleanupservice"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
)

var (
	ErrPullMissingPath        = errors.NewWithCode("missing path", 400)
	ErrPullMissingCatalogId   = errors.NewWithCode("missing catalog id", 400)
	ErrPullMissingMachineName = errors.NewWithCode("missing machine name", 400)
	ErrMissingConnection      = errors.NewWithCode("missing connection", 400)
)

type PullCatalogManifestRequest struct {
	CatalogId          string            `json:"catalog_id"`
	Version            string            `json:"version,omitempty"`
	Owner              string            `json:"owner,omitempty"`
	MachineName        string            `json:"machine_name,omitempty"`
	Path               string            `json:"path,omitempty"`
	Connection         string            `json:"connection,omitempty"`
	ProviderMetadata   map[string]string `json:"provider_metadata,omitempty"`
	StartAfterPull     bool              `json:"start_after_pull,omitempty"`
	LocalMachineFolder string            `json:"-"`
}

func (r *PullCatalogManifestRequest) Validate() error {
	if r.Path == "" {
		return ErrPullMissingPath
	}
	if r.CatalogId == "" {
		return ErrPullMissingCatalogId
	}
	if r.Version == "" {
		r.Version = constants.LATEST_TAG
	}
	if r.MachineName == "" {
		return ErrPullMissingMachineName
	}
	if r.Connection == "" {
		return ErrMissingConnection
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	return nil
}

type PullCatalogManifestResponse struct {
	ID             string                         `json:"id"`
	CatalogId      string                         `json:"catalog_id"`
	Version        string                         `json:"version"`
	LocalPath      string                         `json:"local_path"`
	MachineName    string                         `json:"machine_name"`
	Manifest       *VirtualMachineCatalogManifest `json:"manifest"`
	CleanupRequest *cleanupservice.CleanupRequest `json:"-"`
	Errors         []error                        `json:"-"`
}

func NewPullCatalogManifestResponse() *PullCatalogManifestResponse {
	return &PullCatalogManifestResponse{
		CleanupRequest: cleanupservice.NewCleanupRequest(),
		Errors:         []error{},
	}
}

func (m *PullCatalogManifestResponse) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *PullCatalogManifestResponse) AddError(err error) {
	m.Errors = append(m.Errors, err)
}
