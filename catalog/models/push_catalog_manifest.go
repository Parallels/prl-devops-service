package models

import "Parallels/pd-api-service/errors"

var (
	ErrMissingLocalPath = errors.NewWithCode("missing local path", 400)
	ErrMissingName      = errors.NewWithCode("missing name", 400)
)

type PushCatalogManifestRequest struct {
	LocalPath      string   `json:"local_path"`
	Name           string   `json:"name"`
	Connection     string   `json:"connection"`
	Uuid           string   `json:"uuid,omitempty"`
	RequiredRoles  []string `json:"required_roles,omitempty"`
	RequiredClaims []string `json:"required_claims,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type ImportRemoteMachineRequest struct {
	LocalPath  string `json:"local_path"`
	Name       string `json:"name"`
	Connection string `json:"connection"`
}

func (r *PushCatalogManifestRequest) Validate() error {
	if r.LocalPath == "" {
		return ErrMissingLocalPath
	}
	if r.Name == "" {
		return ErrMissingName
	}
	if r.Connection == "" {
		return ErrMissingConnection
	}
	return nil
}
