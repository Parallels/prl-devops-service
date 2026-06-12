package models

import (
	"errors"
)

type PackerTemplate struct {
	ID             string            `json:"id" gorm:"column:id;primaryKey"`
	Name           string            `json:"name" gorm:"column:name;not null;type:varchar(255);"`
	Owner          string            `json:"owner" gorm:"column:owner;not null;type:varchar(255);"`
	Hostname       string            `json:"hostname" gorm:"column:hostname;not null;type:varchar(255);"`
	Description    string            `json:"description" gorm:"column:description"`
	PackerFolder   string            `json:"packer_folder" gorm:"column:packer_folder"`
	Variables      map[string]string `json:"variables" gorm:"column:variables;serializer:json"`
	Addons         []string          `json:"addons" gorm:"column:addons;serializer:json"`
	Specs          map[string]string `json:"specs" gorm:"column:specs;serializer:json"`
	Defaults       map[string]string `json:"defaults" gorm:"column:defaults;serializer:json"`
	Internal       bool              `json:"internal,omitempty" gorm:"column:internal;default:false"`
	UpdatedAt      string            `json:"updated_at,omitempty" gorm:"column:updated_at"`
	CreatedAt      string            `json:"created_at,omitempty" gorm:"column:created_at"`
	RequiredRoles  []string          `json:"required_roles,omitempty" gorm:"column:required_roles;serializer:json"`
	RequiredClaims []string          `json:"required_claims,omitempty" gorm:"column:required_claims;serializer:json"`
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
