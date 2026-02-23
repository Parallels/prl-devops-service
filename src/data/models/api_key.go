package models

type ApiKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	Secret    string `json:"secret"`
	Revoked   bool   `json:"revoked"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	RevokedAt string `json:"revoked_at"`
	ExpiresAt string `json:"expires_at"`
	*DbRecord `json:"db_record"`
}
