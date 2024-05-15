package models

import (
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
)

type VirtualMachineCatalogPatch struct {
	CatalogId      string                         `json:"catalog_id"`
	Version        string                         `json:"version"`
	Architecture   string                         `json:"architecture"`
	RequiredRoles  []string                       `json:"required_roles"`
	RequiredClaims []string                       `json:"required_claims"`
	Tags           []string                       `json:"tags"`
	CleanupRequest *cleanupservice.CleanupRequest `json:"-"`
	Errors         []error                        `json:"-"`
}

func NewVirtualMachineCatalogPatch() *VirtualMachineCatalogPatch {
	return &VirtualMachineCatalogPatch{
		RequiredRoles:  []string{},
		RequiredClaims: []string{},
		Tags:           []string{},
		Errors:         []error{},
		CleanupRequest: cleanupservice.NewCleanupRequest(),
	}
}

func (m *VirtualMachineCatalogPatch) Validate() error {
	if m.CatalogId == "" {
		return errors.NewWithCode("CatalogId is required", 400)
	}

	if m.Version == "" {
		m.Version = constants.LATEST_TAG
	}

	if m.Architecture == "" {
		return errors.NewWithCode("Architecture is required", 400)
	}

	if m.Architecture != "x86_64" && m.Architecture != "arm64" {
		return errors.NewWithCode("Architecture must be either x86_64 or arm64", 400)
	}

	if m.RequiredClaims == nil {
		m.RequiredClaims = []string{}
	}
	if m.RequiredRoles == nil {
		m.RequiredRoles = []string{}
	}
	if m.Tags == nil {
		m.RequiredRoles = []string{}
	}

	return nil
}

func (m *VirtualMachineCatalogPatch) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *VirtualMachineCatalogPatch) AddError(err error) {
	m.Errors = append(m.Errors, err)
}

func (m *VirtualMachineCatalogPatch) ClearErrors() {
	m.Errors = []error{}
}

func (m *VirtualMachineCatalogPatch) NeedsCleanup() bool {
	if m.CleanupRequest == nil {
		return false
	}
	return m.CleanupRequest.NeedsCleanup()
}
