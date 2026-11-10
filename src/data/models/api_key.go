package models

type ApiKey struct {
	ID        string `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Key       string `json:"key" gorm:"uniqueIndex"`
	Secret    string `json:"secret"`
	Revoked   bool   `json:"revoked"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	RevokedAt string `json:"revoked_at"`
	ExpiresAt string `json:"expires_at"`
	UserID    string `json:"user_id,omitempty" gorm:"index"`
	Type      string `json:"type,omitempty"` // "internal" or "external". Default: "external"
	*DbRecord `json:"db_record" gorm:"-:all"`
}

func (a ApiKey) GetUserID() string {
	return a.UserID
}
