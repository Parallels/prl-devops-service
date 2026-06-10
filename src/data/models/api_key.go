package models

type ApiKey struct {
	ID        string `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"not null"`
	Key       string `json:"key" gorm:"unique;not null"`
	Secret    string `json:"secret" gorm:"not null"`
	Revoked   bool   `json:"revoked" gorm:"default:false"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	RevokedAt string `json:"revoked_at"`
	ExpiresAt string `json:"expires_at"`
	UserID    string `json:"user_id,omitempty" gorm:"index"`
	*DbRecord `json:"db_record" gorm:"embedded"`
}

func (a ApiKey) GetUserID() string {
	return a.UserID
}
