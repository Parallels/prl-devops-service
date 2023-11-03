package models

import (
	"errors"
)

type PackerTemplate struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Owner          string            `json:"owner"`
	Hostname       string            `json:"hostname"`
	Description    string            `json:"description"`
	PackerFolder   string            `json:"packer_folder"`
	Variables      map[string]string `json:"variables"`
	Addons         []string          `json:"addons"`
	Specs          map[string]string `json:"specs"`
	Defaults       map[string]string `json:"defaults"`
	Internal       bool              `json:"internal,omitempty"`
	UpdatedAt      string            `json:"updated_at,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	RequiredRoles  []string          `json:"required_roles,omitempty"`
	RequiredClaims []string          `json:"required_claims,omitempty"`
}

func (m *PackerTemplate) Validate() error {
	if m.ID == "" {
		return errors.New("ID cannot be empty")
	}

	if m.Name == "" {
		return errors.New("name cannot be empty")
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

	return nil
}

func (r PackerTemplate) GetRequiredClaims() []string {
	return r.RequiredClaims
}

func (r PackerTemplate) GetRequiredRoles() []string {
	return r.RequiredRoles
}
