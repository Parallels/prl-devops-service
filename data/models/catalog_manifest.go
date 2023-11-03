package models

type CatalogManifest struct {
	ID                     string                       `json:"id"`
	Name                   string                       `json:"name"`
	Path                   string                       `json:"path,omitempty"`
	PackFile               string                       `json:"pack_path,omitempty"`
	MetadataFile           string                       `json:"metadata_path,omitempty"`
	Type                   string                       `json:"type"`
	Provider               *CatalogManifestProvider     `json:"provider"`
	Size                   int64                        `json:"size"`
	RequiredRoles          []string                     `json:"required_roles"`
	RequiredClaims         []string                     `json:"required_claims"`
	Tags                   []string                     `json:"tags"`
	CreatedAt              string                       `json:"created_at"`
	UpdatedAt              string                       `json:"updated_at"`
	LastDownloadedAt       string                       `json:"last_downloaded_at"`
	LastDownloadedUser     string                       `json:"last_downloaded_user"`
	VirtualMachineContents []CatalogManifestContentItem `json:"virtual_machine_contents"`
	PackContents           []CatalogManifestContentItem `json:"pack_contents"`
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
