package models

import (
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
)

var (
	ErrPushMissingLocalPath    = errors.NewWithCode("missing local path", 400)
	ErrPushMissingCatalogId    = errors.NewWithCode("missing catalog_id", 400)
	ErrPushMissingVersion      = errors.NewWithCode("missing version", 400)
	ErrPushVersionInvalidChars = errors.NewWithCode("version contains invalid characters", 400)
)

type PushCatalogManifestRequest struct {
	LocalPath               string                 `json:"local_path"`
	CatalogId               string                 `json:"catalog_id"`
	Description             string                 `json:"description"`
	Version                 string                 `json:"version"`
	Connection              string                 `json:"connection"`
	Uuid                    string                 `json:"uuid,omitempty"`
	RequiredRoles           []string               `json:"required_roles,omitempty"`
	RequiredClaims          []string               `json:"required_claims,omitempty"`
	Tags                    []string               `json:"tags,omitempty"`
	MinimumSpecRequirements MinimumSpecRequirement `json:"minimum_requirements,omitempty"`
}

type ImportRemoteMachineRequest struct {
	LocalPath  string `json:"local_path"`
	Name       string `json:"name"`
	Connection string `json:"connection"`
}

type MinimumSpecRequirement struct {
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
	Disk   int `json:"disk"`
}

func (r *PushCatalogManifestRequest) Validate() error {
	if r.LocalPath == "" {
		return ErrPushMissingLocalPath
	}

	if r.CatalogId == "" {
		return ErrPushMissingCatalogId
	}

	if r.Connection == "" {
		return ErrMissingConnection
	}

	if r.Version == "" {
		return ErrPushMissingVersion
	}
	if helpers.ContainsIllegalChars(r.Version) {
		return ErrPushVersionInvalidChars
	}

	return nil
}
