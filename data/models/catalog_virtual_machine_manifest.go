package models

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/constants"
)

type CatalogVirtualMachineManifest struct {
	Name               string                            `json:"name"`
	Path               string                            `json:"path,omitempty"`
	MetadataPath       string                            `json:"metadata_path,omitempty"`
	ID                 string                            `json:"id"`
	Type               string                            `json:"type"`
	Provider           RemoteVirtualMachineProvider      `json:"provider"`
	RequiredRoles      []string                          `json:"required_roles"`
	RequiredClaims     []string                          `json:"required_claims"`
	Tags               []string                          `json:"tags"`
	CreatedAt          string                            `json:"created_at"`
	UpdatedAt          string                            `json:"updated_at"`
	LastDownloadedAt   string                            `json:"last_downloaded_at"`
	LastDownloadedUser string                            `json:"last_downloaded_user"`
	Size               int64                             `json:"size"`
	Contents           []RemoteVirtualMachineContentItem `json:"files"`
}

type RemoteVirtualMachineContentItem struct {
	IsDir     bool   `json:"is_dir"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Checksum  string `json:"hash"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}

type RemoteVirtualMachineProvider struct {
	Type string            `json:"type"`
	Meta map[string]string `json:"meta"`
}

func (r *CatalogVirtualMachineManifest) IsAuthorized(ctx *basecontext.AuthorizationContext) bool {
	if ctx.AuthorizedBy == "ApiKeyAuthorization" {
		return true
	}
	if ctx.User == nil || !ctx.IsAuthorized {
		return false
	}
	if ctx.IsUserInRole(constants.SUPER_USER_ROLE) {
		return true
	}

	isAuthorized := false
	hasClaims := false
	if len(r.RequiredRoles) == 0 {
		isAuthorized = true
	} else {
		for _, role := range r.RequiredRoles {
			if ctx.IsUserInRole(role) {
				isAuthorized = true
			}
		}
	}

	if len(r.RequiredClaims) == 0 {
		hasClaims = true
	} else {
		for _, claim := range r.RequiredClaims {
			if ctx.UserHasClaim(claim) {
				hasClaims = true
			}
		}
	}

	if isAuthorized && hasClaims {
		return true
	}

	return false
}
