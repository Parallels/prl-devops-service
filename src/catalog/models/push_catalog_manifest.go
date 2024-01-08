package models

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/serviceprovider/system"
)

var (
	ErrPushMissingLocalPath    = errors.NewWithCode("missing local path", 400)
	ErrPushMissingCatalogId    = errors.NewWithCode("missing catalog_id", 400)
	ErrPushMissingVersion      = errors.NewWithCode("missing version", 400)
	ErrPushVersionInvalidChars = errors.NewWithCode("version contains invalid characters", 400)
	ErrMissingArchitecture     = errors.NewWithCode("missing architecture", 400)
	ErrInvalidArchitecture     = errors.NewWithCode("invalid architecture, needs to be either x86_64 or arm64", 400)
)

type PushCatalogManifestRequest struct {
	LocalPath               string                 `json:"local_path"`
	CatalogId               string                 `json:"catalog_id"`
	Description             string                 `json:"description"`
	Version                 string                 `json:"version"`
	Architecture            string                 `json:"architecture"`
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

	if r.Architecture == "" {
		ctx := basecontext.NewRootBaseContext()
		sysCtl := system.Get(ctx)
		arch, err := sysCtl.GetArchitecture(ctx)
		if err != nil {
			return errors.NewWithCode("unable to determine architecture and none was set", 400)
		}
		r.Architecture = arch
	} else {
		if r.Architecture == "amd64" {
			r.Architecture = "x86_64"
		}
		if r.Architecture == "arm" {
			r.Architecture = "arm64"
		}
		if r.Architecture == "aarch64" {
			r.Architecture = "arm64"
		}

		if r.Architecture != "x86_64" && r.Architecture != "arm64" {
			return ErrInvalidArchitecture
		}
	}

	if helpers.ContainsIllegalChars(r.Version) {
		return ErrPushVersionInvalidChars
	}

	return nil
}
