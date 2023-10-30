package models

import "errors"

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

var ErrMissingLocalPath = errors.New("missing local path")
var ErrMissingName = errors.New("missing name")
var ErrMissingConnection = errors.New("missing connection")

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
