package models

import (
	"errors"
)

type PackerTemplate struct {
	ID             string            `json:"id" gorm:"primaryKey"`
	Name           string            `json:"name" gorm:"not null;type:varchar(255);"`
	Owner          string            `json:"owner" gorm:"not null;type:varchar(255);"`
	Hostname       string            `json:"hostname" gorm:"not null;type:varchar(255);"`
	Description    string            `json:"description"`
	PackerFolder   string            `json:"packer_folder"`
	Variables      map[string]string `json:"variables" gorm:"serializer:json"`
	Addons         []string          `json:"addons" gorm:"serializer:json"`
	Specs          map[string]string `json:"specs" gorm:"serializer:json"`
	Defaults       map[string]string `json:"defaults" gorm:"serializer:json"`
	Internal       bool              `json:"internal,omitempty" gorm:"default:false"`
	UpdatedAt      string            `json:"updated_at,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	RequiredRoles  []string          `json:"required_roles,omitempty" gorm:"serializer:json"`
	RequiredClaims []string          `json:"required_claims,omitempty" gorm:"serializer:json"`
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
