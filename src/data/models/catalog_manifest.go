package models

import "strings"

type CatalogManifest struct {
	ID                      string                       `json:"id"`
	CatalogId               string                       `json:"catalog_id"`
	Name                    string                       `json:"name"`
	Version                 string                       `json:"version"`
	Architecture            string                       `json:"architecture"`
	Description             string                       `json:"description"`
	Path                    string                       `json:"path,omitempty"`
	PackFile                string                       `json:"pack_path,omitempty"`
	MetadataFile            string                       `json:"metadata_path,omitempty"`
	Type                    string                       `json:"type"`
	Provider                *CatalogManifestProvider     `json:"provider"`
	Size                    int64                        `json:"size"`
	RequiredRoles           []string                     `json:"required_roles"`
	RequiredClaims          []string                     `json:"required_claims"`
	Tags                    []string                     `json:"tags"`
	CreatedAt               string                       `json:"created_at"`
	UpdatedAt               string                       `json:"updated_at"`
	LastDownloadedAt        string                       `json:"last_downloaded_at"`
	LastDownloadedUser      string                       `json:"last_downloaded_user"`
	DownloadCount           int                          `json:"download_count"`
	VirtualMachineContents  []CatalogManifestContentItem `json:"virtual_machine_contents"`
	PackContents            []CatalogManifestContentItem `json:"pack_contents"`
	PackSize                int64                        `json:"pack_size,omitempty"`
	MinimumSpecRequirements *MinimumSpecRequirement      `json:"minimum_requirements,omitempty"`
	Tainted                 bool                         `json:"tainted"`
	TaintedBy               string                       `json:"tainted_by"`
	UnTaintedBy             string                       `json:"untainted_by"`
	TaintedAt               string                       `json:"tainted_at"`
	Revoked                 bool                         `json:"revoked"`
	RevokedBy               string                       `json:"revoked_by"`
	RevokedAt               string                       `json:"revoked_at"`
}

func (r *CatalogManifest) AddTag(tag string) {
	exists := false
	if r.Tags == nil {
		r.Tags = []string{}
	}

	for _, t := range r.Tags {
		if t == tag {
			exists = true
			break
		}
	}

	if !exists {
		r.Tags = append(r.Tags, tag)
	}
}

func (r *CatalogManifest) RemoveTag(tag string) {
	if r.Tags == nil {
		return
	}

	for i, t := range r.Tags {
		if t == tag {
			r.Tags = append(r.Tags[:i], r.Tags[i+1:]...)
			break
		}
	}
}

func (r *CatalogManifest) HasTag(tag string) bool {
	if r.Tags == nil {
		return false
	}

	for _, t := range r.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}

	return false
}

func (r CatalogManifest) GetRequiredClaims() []string {
	return r.RequiredClaims
}

func (r CatalogManifest) GetRequiredRoles() []string {
	return r.RequiredRoles
}

type CatalogManifestContentItem struct {
	IsDir     bool   `json:"is_dir"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Checksum  string `json:"hash"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}

type CatalogManifestProvider struct {
	Type     string            `json:"type"`
	Host     string            `json:"host"`
	Port     string            `json:"port"`
	Username string            `json:"user"`
	Password string            `json:"password"`
	ApiKey   string            `json:"api_key"`
	Meta     map[string]string `json:"meta"`
}

func (m *CatalogManifestProvider) String() string {
	r := "provider=" + m.Type
	if m.Host == "" {
		r += ";host=" + m.Host
	}
	if m.Port == "" {
		r += ";port=" + m.Port
	}
	if m.Username == "" {
		r += ";user=" + m.Username
	}
	if m.Password == "" {
		r += ";password=" + m.Password
	}
	if m.ApiKey == "" {
		r += ";api_key=" + m.ApiKey
	}

	for k, v := range m.Meta {
		r += ";" + k + "=" + v
	}

	return strings.TrimRight(r, ";")
}

type MinimumSpecRequirement struct {
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
	Disk   int `json:"disk"`
}
