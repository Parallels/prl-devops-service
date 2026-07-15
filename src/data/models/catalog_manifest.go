package models

import "strings"

type CatalogManifest struct {
	ID                      string                       `json:"id" gorm:"primaryKey;type:varchar(64);column:id;not null"`
	CatalogId               string                       `json:"catalog_id" gorm:"column:catalog_id;not null;type:varchar(64);index"`
	Name                    string                       `json:"name" gorm:"column:name;not null;type:varchar(255);index"`
	Version                 string                       `json:"version" gorm:"column:version;not null;type:varchar(64)"`
	Architecture            string                       `json:"architecture" gorm:"column:architecture;not null;type:varchar(64)"`
	Description             string                       `json:"description" gorm:"column:description;not null;type:text"`
	Path                    string                       `json:"path,omitempty" gorm:"column:path;type:varchar(255)"`
	PackFile                string                       `json:"pack_path,omitempty" gorm:"column:pack_path;type:varchar(255)"`
	MetadataFile            string                       `json:"metadata_path,omitempty" gorm:"column:metadata_path;type:varchar(255)"`
	Type                    string                       `json:"type" gorm:"column:type;not null;type:varchar(64)"`
	Provider                *CatalogManifestProvider     `json:"provider" gorm:"column:provider;type:json;serializer:json"`
	Size                    int64                        `json:"size" gorm:"column:size;not null;type:bigint"`
	RequiredRoles           []string                     `json:"required_roles" gorm:"column:required_roles;type:json;serializer:json"`
	RequiredClaims          []string                     `json:"required_claims" gorm:"column:required_claims;type:json;serializer:json"`
	Tags                    []string                     `json:"tags" gorm:"column:tags;type:json;serializer:json"`
	CreatedAt               string                       `json:"created_at" gorm:"column:created_at;not null;type:timestamp"`
	UpdatedAt               string                       `json:"updated_at" gorm:"column:updated_at;not null;type:timestamp"`
	LastDownloadedAt        string                       `json:"last_downloaded_at" gorm:"column:last_downloaded_at;type:timestamp"`
	LastDownloadedUser      string                       `json:"last_downloaded_user" gorm:"column:last_downloaded_user;type:varchar(255)"`
	IsCompressed            bool                         `json:"is_compressed" gorm:"column:is_compressed;type:boolean;default:0"`
	PackRelativePath        string                       `json:"pack_relative_path" gorm:"column:pack_relative_path;type:varchar(255)"`
	DownloadCount           int                          `json:"download_count" gorm:"column:download_count;type:int"`
	VirtualMachineContents  []CatalogManifestContentItem `json:"virtual_machine_contents" gorm:"column:virtual_machine_contents;type:json;serializer:json"`
	PackContents            []CatalogManifestContentItem `json:"pack_contents" gorm:"column:pack_contents;type:json;serializer:json"`
	PackSize                int64                        `json:"pack_size,omitempty" gorm:"column:pack_size;type:bigint"`
	MinimumSpecRequirements *MinimumSpecRequirement      `json:"minimum_requirements,omitempty" gorm:"column:minimum_requirements;type:json;serializer:json"`
	Tainted                 bool                         `json:"tainted" gorm:"column:tainted;type:boolean;default:0"`
	TaintedBy               string                       `json:"tainted_by" gorm:"column:tainted_by;type:varchar(255)"`
	UnTaintedBy             string                       `json:"untainted_by" gorm:"column:untainted_by;type:varchar(255)"`
	TaintedAt               string                       `json:"tainted_at" gorm:"column:tainted_at;type:timestamp"`
	Revoked                 bool                         `json:"revoked" gorm:"column:revoked;type:boolean;default:0"`
	RevokedBy               string                       `json:"revoked_by" gorm:"column:revoked_by;type:varchar(255)"`
	RevokedAt               string                       `json:"revoked_at" gorm:"column:revoked_at;type:timestamp"`
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
