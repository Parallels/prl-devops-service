package models

import "time"

type ApiKey struct {
	BaseModel            // Embeds ID, CreatedAt, UpdatedAt, DeletedAt
	Name      string     `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Key       string     `json:"key" gorm:"column:key;unique;not null;type:varchar(65)"`
	Secret    string     `json:"secret" gorm:"column:secret;not null;type:varchar(65)"`
	Revoked   bool       `json:"revoked" gorm:"column:revoked;type:boolean;default:false"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" gorm:"column:revoked_at;type:timestamp"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"column:expires_at;type:timestamp"`
	*DbRecord `json:"db_record" gorm:"embedded"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
