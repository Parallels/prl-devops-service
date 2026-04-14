package models

import (
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
)

type VirtualMachineCatalogManifestPatch struct {
	Description    *string                        `json:"description,omitempty"`
	RequiredRoles  []string                       `json:"required_roles"`
	RequiredClaims []string                       `json:"required_claims"`
	Tags           []string                       `json:"tags"`
	Provider       *CatalogManifestProvider       `json:"-"`
	Connection     string                         `json:"connection"`
	CleanupRequest *cleanupservice.CleanupService `json:"-"`
	Errors         []error                        `json:"-"`
}

func NewVirtualMachineCatalogPatch() *VirtualMachineCatalogManifestPatch {
	return &VirtualMachineCatalogManifestPatch{
		RequiredRoles:  []string{},
		RequiredClaims: []string{},
		Tags:           []string{},
		Errors:         []error{},
		CleanupRequest: cleanupservice.NewCleanupService(),
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

// UpdateCatalogManifestMetadataRequest is used for the PUT /metadata endpoint.
// Pointer fields allow partial updates: nil means "do not change", non-nil means "replace with this value".
type UpdateCatalogManifestMetadataRequest struct {
	Description    *string  `json:"description,omitempty"`
	RequiredRoles  []string `json:"required_roles,omitempty"`
	RequiredClaims []string `json:"required_claims,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}
