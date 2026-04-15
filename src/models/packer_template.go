package models

import (
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/errors"
)

type CreatePackerTemplateRequest struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	PackerFolder   string            `json:"packer_folder"`
	Variables      map[string]string `json:"variables,omitempty"`
	Addons         []string          `json:"addons,omitempty"`
	Specs          map[string]string `json:"specs,omitempty"`
	Defaults       map[string]string `json:"defaults,omitempty"`
	Internal       bool              `json:"internal,omitempty"`
	UpdatedAt      string            `json:"updated_at,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	RequiredRoles  []string          `json:"required_roles,omitempty"`
	RequiredClaims []string          `json:"required_claims,omitempty"`
}

func (m *CreatePackerTemplateRequest) Validate() *errors.Diagnostics {
	diagnostics := errors.NewDiagnostics("CreatePackerTemplateRequest validation")
	if m.ID == "" {
		diagnostics.AddError(strconv.Itoa(http.StatusBadRequest), "id cannot be empty", "CreatePackerTemplateRequest")
	}
	if m.Name == "" {
		diagnostics.AddError(strconv.Itoa(http.StatusBadRequest), "name cannot be empty", "CreatePackerTemplateRequest")
	}

	if m.Specs == nil {
		m.Specs = make(map[string]string)
		m.Specs["memory"] = "2048"
		m.Specs["cpu"] = "2"
		m.Specs["disk"] = "20480"
	}

	if m.Defaults == nil {
		m.Defaults = make(map[string]string)
	}

	if m.Variables == nil {
		m.Variables = make(map[string]string)
	}

	if m.RequiredClaims == nil {
		m.RequiredClaims = make([]string, 0)
	}

	if m.RequiredRoles == nil {
		m.RequiredRoles = make([]string, 0)
	}

	return diagnostics
}

type PackerTemplateResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	PackerFolder   string            `json:"packer_folder"`
	Variables      map[string]string `json:"variables,omitempty"`
	Addons         []string          `json:"addons,omitempty"`
	Specs          map[string]string `json:"specs,omitempty"`
	Defaults       map[string]string `json:"defaults,omitempty"`
	Internal       bool              `json:"internal,omitempty"`
	UpdatedAt      string            `json:"updated_at,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	RequiredRoles  []string          `json:"required_roles,omitempty"`
	RequiredClaims []string          `json:"required_claims,omitempty"`
}
