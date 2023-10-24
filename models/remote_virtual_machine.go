package models

type RemoteVirtualMachineManifest struct {
	Name           string                     `json:"name"`
	Path           string                     `json:"path"`
	ID             string                     `json:"id"`
	Type           string                     `json:"type"`
	Remote         RemoteVirtualMachineHost   `json:"remote"`
	RequiredRoles  []string                   `json:"required_roles"`
	RequiredClaims []string                   `json:"required_claims"`
	Tags           []string                   `json:"tags"`
	CreatedAt      string                     `json:"created_at"`
	UpdatedAt      string                     `json:"updated_at"`
	Files          []RemoteVirtualMachineFile `json:"files"`
}

type RemoteVirtualMachineHost struct {
	Type RemoteVirtualMachineHostType `json:"type"`
}

type RemoteVirtualMachineHostType string

const (
	RemoteVirtualMachineHostTypeAws   RemoteVirtualMachineHostType = "aws"
	RemoteVirtualMachineHostTypeAzure RemoteVirtualMachineHostType = "azure"
)

type RemoteVirtualMachineFile struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Checksum  string `json:"hash"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}
