package models

import (
	"os"

	"github.com/Parallels/pd-api-service/catalog/cleanupservice"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
)

var (
	ErrMissingPath        = errors.NewWithCode("missing path", 400)
	ErrMissingId          = errors.NewWithCode("missing id", 400)
	ErrMissingMachineName = errors.NewWithCode("missing machine name", 400)
	ErrMissingConnection  = errors.NewWithCode("missing connection", 400)
)

type PullCatalogManifestRequest struct {
	ID                 string            `json:"id"`
	Owner              string            `json:"owner,omitempty"`
	MachineName        string            `json:"machine_name,omitempty"`
	Path               string            `json:"path,omitempty"`
	Connection         string            `json:"connection,omitempty"`
	ProviderMetadata   map[string]string `json:"provider_metadata,omitempty"`
	LocalMachineFolder string            `json:"-"`
}

func (r *PullCatalogManifestRequest) Validate() error {
	if r.Path == "" {
		return ErrMissingPath
	}
	if r.ID == "" {
		return ErrMissingId
	}
	if r.MachineName == "" {
		return ErrMissingMachineName
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
