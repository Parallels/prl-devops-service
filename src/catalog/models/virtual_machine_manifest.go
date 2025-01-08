package models

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

type VirtualMachineCatalogManifest struct {
	ID                      string                              `json:"id"`
	CatalogId               string                              `json:"catalog_id"`
	Description             string                              `json:"description"`
	Version                 string                              `json:"version"`
	Name                    string                              `json:"name"`
	Architecture            string                              `json:"architecture"`
	Path                    string                              `json:"path,omitempty"`
	PackFile                string                              `json:"pack_path,omitempty"`
	MetadataFile            string                              `json:"metadata_path,omitempty"`
	Type                    string                              `json:"type"`
	Provider                *CatalogManifestProvider            `json:"provider"`
	Size                    int64                               `json:"size"`
	RequiredRoles           []string                            `json:"required_roles"`
	RequiredClaims          []string                            `json:"required_claims"`
	Tags                    []string                            `json:"tags"`
	CreatedAt               string                              `json:"created_at"`
	UpdatedAt               string                              `json:"updated_at"`
	LastDownloadedAt        string                              `json:"last_downloaded_at"`
	LastDownloadedUser      string                              `json:"last_downloaded_user"`
	IsCompressed            bool                                `json:"is_compressed"`
	PackRelativePath        string                              `json:"pack_relative_path"`
	DownloadCount           int                                 `json:"download_count"`
	CompressedPath          string                              `json:"-"`
	CompressedChecksum      string                              `json:"compressed_checksum"`
	VirtualMachineContents  []VirtualMachineManifestContentItem `json:"virtual_machine_contents"`
	PackContents            []VirtualMachineManifestContentItem `json:"pack_contents"`
	PackSize                int64                               `json:"pack_size,omitempty"`
	Tainted                 bool                                `json:"tainted"`
	TaintedBy               string                              `json:"tainted_by"`
	TaintedAt               string                              `json:"tainted_at"`
	UnTaintedBy             string                              `json:"untainted_by"`
	Revoked                 bool                                `json:"revoked"`
	RevokedAt               string                              `json:"revoked_at"`
	RevokedBy               string                              `json:"revoked_by"`
	MinimumSpecRequirements *MinimumSpecRequirement             `json:"minimum_requirements,omitempty"`
	CachedDate              string                              `json:"cached_date,omitempty"`
	CacheLastUsed           string                              `json:"cache_last_used,omitempty"`
	CacheUsedCount          int64                               `json:"cache_used_count,omitempty"`
	CacheLocalFullPath      string                              `json:"cache_local_path,omitempty"`
	CacheMetadataName       string                              `json:"cache_metadata_name,omitempty"`
	CacheFileName           string                              `json:"cache_file_name,omitempty"`
	CacheType               string                              `json:"cache_type,omitempty"`
	CacheSize               int64                               `json:"cache_size,omitempty"`
	CacheCompleted          bool                                `json:"cache_completed,omitempty"`
	CleanupRequest          *cleanupservice.CleanupService      `json:"-"`
	Errors                  []error                             `json:"-"`
}

func NewVirtualMachineCatalogManifest() *VirtualMachineCatalogManifest {
	return &VirtualMachineCatalogManifest{
		Provider:               &CatalogManifestProvider{},
		VirtualMachineContents: []VirtualMachineManifestContentItem{},
		Errors:                 []error{},
		CleanupRequest:         cleanupservice.NewCleanupService(),
	}
}

func (m *VirtualMachineCatalogManifest) Validate(importVm bool) error {
	if m.ID == "" {
		m.ID = helpers.GenerateId()
	}

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

	m.Name = fmt.Sprintf("%s-%s-%s", helpers.NormalizeString(m.CatalogId), helpers.NormalizeString(m.Architecture), helpers.NormalizeString(m.Version))

	if !importVm {
		if m.Path == "" {
			return errors.NewWithCode("Path is required", 400)
		}
		if m.PackFile == "" {
			return errors.NewWithCode("PackFile is required", 400)
		}
		if m.MetadataFile == "" {
			return errors.NewWithCode("MetadataFile is required", 400)
		}
	}

	if m.RequiredClaims == nil {
		m.RequiredClaims = []string{}
	}
	if m.RequiredRoles == nil {
		m.RequiredRoles = []string{}
	}
	if m.Tags == nil {
		m.Tags = []string{}
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

type VirtualMachineManifestArchitectureType string

const (
	VirtualMachineManifestArchitectureTypeX86_64 VirtualMachineManifestArchitectureType = "x86_64"
	VirtualMachineManifestArchitectureTypeArm64  VirtualMachineManifestArchitectureType = "arm64"
)

func (t VirtualMachineManifestArchitectureType) String() string {
	return string(t)
}

func (t VirtualMachineManifestArchitectureType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *VirtualMachineManifestArchitectureType) UnmarshalJSON(b []byte) error {
	*t = VirtualMachineManifestArchitectureType(b)
	return nil
}
