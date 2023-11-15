package models

import (
	"fmt"

	"github.com/Parallels/pd-api-service/catalog/cleanupservice"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
)

type VirtualMachineCatalogManifest struct {
	ID                     string                              `json:"id"`
	CatalogId              string                              `json:"catalog_id"`
	Description            string                              `json:"description"`
	Version                string                              `json:"version"`
	Name                   string                              `json:"name"`
	Path                   string                              `json:"path,omitempty"`
	PackFile               string                              `json:"pack_path,omitempty"`
	MetadataFile           string                              `json:"metadata_path,omitempty"`
	Type                   string                              `json:"type"`
	Provider               *CatalogManifestProvider            `json:"provider"`
	Size                   int64                               `json:"size"`
	RequiredRoles          []string                            `json:"required_roles"`
	RequiredClaims         []string                            `json:"required_claims"`
	Tags                   []string                            `json:"tags"`
	CreatedAt              string                              `json:"created_at"`
	UpdatedAt              string                              `json:"updated_at"`
	LastDownloadedAt       string                              `json:"last_downloaded_at"`
	LastDownloadedUser     string                              `json:"last_downloaded_user"`
	DownloadCount          int                                 `json:"download_count"`
	CompressedPath         string                              `json:"-"`
	CompressedChecksum     string                              `json:"compressed_checksum"`
	VirtualMachineContents []VirtualMachineManifestContentItem `json:"virtual_machine_contents"`
	PackContents           []VirtualMachineManifestContentItem `json:"pack_contents"`
	Tainted                bool                                `json:"tainted"`
	TaintedBy              string                              `json:"tainted_by"`
	TaintedAt              string                              `json:"tainted_at"`
	UnTaintedBy            string                              `json:"untainted_by"`
	Revoked                bool                                `json:"revoked"`
	RevokedAt              string                              `json:"revoked_at"`
	RevokedBy              string                              `json:"revoked_by"`
	CleanupRequest         *cleanupservice.CleanupRequest      `json:"-"`
	Errors                 []error                             `json:"-"`
}

func NewVirtualMachineCatalogManifest() *VirtualMachineCatalogManifest {
	return &VirtualMachineCatalogManifest{
		Provider:               &CatalogManifestProvider{},
		VirtualMachineContents: []VirtualMachineManifestContentItem{},
		Errors:                 []error{},
		CleanupRequest:         cleanupservice.NewCleanupRequest(),
	}
}

func (m *VirtualMachineCatalogManifest) Validate() error {
	if m.ID == "" {
		m.ID = helpers.GenerateId()
	}

	if m.CatalogId == "" {
		return errors.NewWithCode("CatalogId is required", 400)
	}

	if m.Version == "" {
		m.Version = constants.LATEST_TAG
	}

	m.Name = fmt.Sprintf("%s-%s", helpers.NormalizeString(m.CatalogId), helpers.NormalizeString(m.Version))

	if m.Path == "" {
		return errors.NewWithCode("Path is required", 400)
	}
	if m.PackFile == "" {
		return errors.NewWithCode("PackFile is required", 400)
	}
	if m.MetadataFile == "" {
		return errors.NewWithCode("MetadataFile is required", 400)
	}
	if m.Provider == nil {
		return errors.NewWithCode("Provider is required", 400)
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

func (m *VirtualMachineCatalogManifest) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *VirtualMachineCatalogManifest) AddError(err error) {
	m.Errors = append(m.Errors, err)
}

func (m *VirtualMachineCatalogManifest) ClearErrors() {
	m.Errors = []error{}
}

func (m *VirtualMachineCatalogManifest) NeedsCleanup() bool {
	if m.CleanupRequest == nil {
		return false
	}
	return m.CleanupRequest.NeedsCleanup()
}

type VirtualMachineManifestContentItem struct {
	IsDir     bool   `json:"is_dir"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Checksum  string `json:"hash"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}
