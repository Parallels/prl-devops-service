package models

type ApiKey struct {
	ID        string `json:"id" gorm:"column:id;primaryKey"`
	Name      string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Key       string `json:"key" gorm:"column:key;unique;not null"`
	Secret    string `json:"secret" gorm:"column:secret;not null"`
	Revoked   bool   `json:"revoked" gorm:"column:revoked;default:false"`
	CreatedAt string `json:"created_at" gorm:"column:created_at"`
	UpdatedAt string `json:"updated_at" gorm:"column:updated_at"`
	RevokedAt string `json:"revoked_at" gorm:"column:revoked_at"`
	ExpiresAt string `json:"expires_at" gorm:"column:expires_at"`
	UserID    string `json:"user_id,omitempty" gorm:"column:user_id;index"`
	*DbRecord `json:"db_record" gorm:"embedded"`
}

func (a ApiKey) GetUserID() string {
	return a.UserID
}
