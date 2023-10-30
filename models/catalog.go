package models

type CatalogVirtualMachineManifest struct {
	Name               string                       `json:"name"`
	ID                 string                       `json:"id"`
	Type               string                       `json:"type"`
	Tags               []string                     `json:"tags,omitempty"`
	Size               string                       `json:"size,omitempty"`
	Provider           RemoteVirtualMachineProvider `json:"provider,omitempty"`
	CreatedAt          string                       `json:"created_at,omitempty"`
	UpdatedAt          string                       `json:"updated_at,omitempty"`
	LastDownloadedAt   string                       `json:"last_downloaded_at,omitempty"`
	LastDownloadedUser string                       `json:"last_downloaded_user,omitempty"`
}

type RemoteVirtualMachineProvider struct {
	Type string            `json:"type"`
	Meta map[string]string `json:"meta"`
}

type PullCatalogManifestResponse struct {
	ID          string                         `json:"id"`
	LocalPath   string                         `json:"local_path"`
	MachineName string                         `json:"machine_name"`
	Manifest    *CatalogVirtualMachineManifest `json:"manifest"`
}
