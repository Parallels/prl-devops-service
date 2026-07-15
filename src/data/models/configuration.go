package models

// Legacy Configuration for JSON database
type Configuration struct {
	ID        string  `json:"id" yaml:"id"`
	CreatedAt string  `json:"created_at" yaml:"created_at"`
	UpdatedAt string  `json:"updated_at" yaml:"updated_at"`
	DeletedAt *string `json:"deleted_at,omitempty" yaml:"deleted_at,omitempty"`
	Vault     string  `json:"vault" yaml:"vault"`
	Key       string  `json:"key" yaml:"key"`
	Value     string  `json:"value" yaml:"value"`
	Version   int     `json:"version" yaml:"version"`
}
