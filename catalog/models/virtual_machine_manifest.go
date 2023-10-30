package models

type VirtualMachineManifest struct {
	Name               string                              `json:"name"`
	Path               string                              `json:"path,omitempty"`
	MetadataPath       string                              `json:"metadata_path,omitempty"`
	ID                 string                              `json:"id"`
	Type               string                              `json:"type"`
	Provider           CatalogManifestProvider             `json:"provider"`
	RequiredRoles      []string                            `json:"required_roles"`
	RequiredClaims     []string                            `json:"required_claims"`
	Tags               []string                            `json:"tags"`
	CreatedAt          string                              `json:"created_at"`
	UpdatedAt          string                              `json:"updated_at"`
	LastDownloadedAt   string                              `json:"last_downloaded_at"`
	LastDownloadedUser string                              `json:"last_downloaded_user"`
	Size               int64                               `json:"size"`
	Contents           []VirtualMachineManifestContentItem `json:"files"`
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
