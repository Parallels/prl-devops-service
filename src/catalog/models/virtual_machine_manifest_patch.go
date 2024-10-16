package models

import (
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
)

type VirtualMachineCatalogManifestPatch struct {
	RequiredRoles  []string                       `json:"required_roles"`
	RequiredClaims []string                       `json:"required_claims"`
	Tags           []string                       `json:"tags"`
	Provider       *CatalogManifestProvider       `json:"-"`
	Connection     string                         `json:"connection"`
	CleanupRequest *cleanupservice.CleanupRequest `json:"-"`
	Errors         []error                        `json:"-"`
}

func NewVirtualMachineCatalogPatch() *VirtualMachineCatalogManifestPatch {
	return &VirtualMachineCatalogManifestPatch{
		RequiredRoles:  []string{},
		RequiredClaims: []string{},
		Tags:           []string{},
		Errors:         []error{},
		CleanupRequest: cleanupservice.NewCleanupRequest(),
	}
}

func (m *VirtualMachineCatalogManifestPatch) Validate() error {
	if m.RequiredClaims == nil {
		m.RequiredClaims = []string{}
	}
	if m.RequiredRoles == nil {
		m.RequiredRoles = []string{}
	}
	if m.Tags == nil {
		m.Tags = []string{}
	}
	if m.Connection != "" {
		m.Provider = &CatalogManifestProvider{}
		if err := m.Provider.Parse(m.Connection); err != nil {
			return err
		}
	}

	return nil
}

func (m *VirtualMachineCatalogManifestPatch) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *VirtualMachineCatalogManifestPatch) AddError(err error) {
	m.Errors = append(m.Errors, err)
}

func (m *VirtualMachineCatalogManifestPatch) ClearErrors() {
	m.Errors = []error{}
}

func (m *VirtualMachineCatalogManifestPatch) NeedsCleanup() bool {
	if m.CleanupRequest == nil {
		return false
	}
	return m.CleanupRequest.NeedsCleanup()
}
