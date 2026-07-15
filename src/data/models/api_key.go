package models

// Legacy ApiKey for JSON database
type ApiKey struct {
	ID        string  `json:"id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at,omitempty"`
	Name      string  `json:"name"`
	Key       string  `json:"key"`
	Secret    string  `json:"secret"`
	UserID    string  `json:"user_id,omitempty"`
	Revoked   bool    `json:"revoked"`
	RevokedAt string  `json:"revoked_at,omitempty"`
	ExpiresAt string  `json:"expires_at,omitempty"`
	*DbRecord `json:"db_record"`
}

func (a ApiKey) GetUserID() string {
	return a.UserID
}
