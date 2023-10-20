package models

type ApiKey struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	Secret    string `json:"secret"`
	Revoked   bool   `json:"revoked"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	RevokedAt string `json:"revoked_at,omitempty"`
}
