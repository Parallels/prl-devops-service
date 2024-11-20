package models

type CatalogManifest struct {
	Name                    string                        `json:"name" yaml:"name"`
	ID                      string                        `json:"id,omitempty" yaml:"id,omitempty"`
	CatalogId               string                        `json:"catalog_id,omitempty" yaml:"catalog_id,omitempty"`
	Description             string                        `json:"description,omitempty" yaml:"description,omitempty"`
	Architecture            string                        `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	Version                 string                        `json:"version,omitempty" yaml:"version,omitempty"`
	Type                    string                        `json:"type,omitempty" yaml:"type,omitempty"`
	Tags                    []string                      `json:"tags,omitempty" yaml:"tags,omitempty"`
	Size                    int64                         `json:"size,omitempty" yaml:"size,omitempty"`
	Path                    string                        `json:"path,omitempty" yaml:"path,omitempty"`
	PackFilename            string                        `json:"pack_filename,omitempty" yaml:"pack_filename,omitempty"`
	MetadataFilename        string                        `json:"metadata_filename,omitempty" yaml:"metadata_filename,omitempty"`
	Provider                *RemoteVirtualMachineProvider `json:"provider,omitempty" yaml:"provider,omitempty"`
	CreatedAt               string                        `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt               string                        `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	RequiredClaims          []string                      `json:"required_claims,omitempty" yaml:"required_claims,omitempty"`
	RequiredRoles           []string                      `json:"required_roles,omitempty" yaml:"required_roles,omitempty"`
	LastDownloadedAt        string                        `json:"last_downloaded_at,omitempty" yaml:"last_downloaded_at,omitempty"`
	LastDownloadedUser      string                        `json:"last_downloaded_user,omitempty" yaml:"last_downloaded_user,omitempty"`
	IsCompressed            bool                          `json:"is_compressed,omitempty" yaml:"is_compressed,omitempty"`
	PackRelativePath        string                        `json:"pack_relative_path,omitempty" yaml:"pack_relative_path,omitempty"`
	DownloadCount           int                           `json:"download_count,omitempty" yaml:"download_count,omitempty"`
	Tainted                 bool                          `json:"tainted,omitempty" yaml:"tainted,omitempty"`
	TaintedBy               string                        `json:"tainted_by,omitempty" yaml:"tainted_by,omitempty"`
	TaintedAt               string                        `json:"tainted_at,omitempty" yaml:"tainted_at,omitempty"`
	UnTaintedBy             string                        `json:"untainted_by,omitempty" yaml:"untainted_by,omitempty"`
	Revoked                 bool                          `json:"revoked,omitempty" yaml:"revoked,omitempty"`
	RevokedAt               string                        `json:"revoked_at,omitempty" yaml:"revoked_at,omitempty"`
	RevokedBy               string                        `json:"revoked_by,omitempty" yaml:"revoked_by,omitempty"`
	PackContents            []CatalogManifestPackItem     `json:"pack_contents,omitempty" yaml:"pack_contents,omitempty"`
	PackSize                int64                         `json:"pack_size,omitempty" yaml:"pack_size,omitempty"`
	MinimumSpecRequirements *MinimumSpecRequirement       `json:"minimum_requirements,omitempty" yaml:"minimum_requirements,omitempty"`
	CacheDate               string                        `json:"cache_date,omitempty"`
	CacheLocalFullPath      string                        `json:"cache_local_path,omitempty"`
	CacheMetadataName       string                        `json:"cache_metadata_name,omitempty"`
	CacheFileName           string                        `json:"cache_file_name,omitempty"`
	CacheType               string                        `json:"cache_type,omitempty"`
	CacheSize               int64                         `json:"cache_size,omitempty"`
}

type MinimumSpecRequirement struct {
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
	Disk   int `json:"disk"`
}

type RemoteVirtualMachineProvider struct {
	Type     string            `json:"type,omitempty" yaml:"type,omitempty"`
	Host     string            `json:"host,omitempty" yaml:"host,omitempty"`
	Port     string            `json:"port,omitempty" yaml:"port,omitempty"`
	Username string            `json:"user,omitempty" yaml:"user,omitempty"`
	Password string            `json:"password,omitempty" yaml:"password,omitempty"`
	ApiKey   string            `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	Meta     map[string]string `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type CatalogManifestPackItem struct {
	IsDir     bool   `json:"is_dir,omitempty" yaml:"is_dir,omitempty"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	Checksum  string `json:"hash,omitempty" yaml:"hash,omitempty"`
	Size      int64  `json:"size,omitempty" yaml:"size,omitempty"`
	CreatedAt string `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	DeletedAt string `json:"deleted_at,omitempty" yaml:"deleted_at,omitempty"`
}

type PullCatalogManifestResponse struct {
	ID          string           `json:"id,omitempty" yaml:"id,omitempty"`
	MachineID   string           `json:"machine_id,omitempty" yaml:"machine_id,omitempty"`
	LocalPath   string           `json:"local_path,omitempty" yaml:"local_path,omitempty"`
	MachineName string           `json:"machine_name,omitempty" yaml:"machine_name,omitempty"`
	Manifest    *CatalogManifest `json:"manifest,omitempty" yaml:"manifest,omitempty"`
}

type ImportCatalogManifestResponse struct {
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
}

type ImportVmResponse struct {
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
}

type VirtualMachineCatalogManifestList struct {
	TotalSize int64             `json:"total_size"`
	Manifests []CatalogManifest `json:"manifests"`
}
