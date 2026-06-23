package models

type ApiKey struct {
	ID        string `json:"id" gorm:"column:id;primaryKey;not null;type:varchar(64)"`
	Name      string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Key       string `json:"key" gorm:"column:key;unique;not null;type:varchar(65)"`
	Secret    string `json:"secret" gorm:"column:secret;not null;type:varchar(65)"`
	Revoked   bool   `json:"revoked" gorm:"column:revoked;type:boolean;default:false;not null"`
	CreatedAt string `json:"created_at" gorm:"column:created_at;type:timestamp"`
	UpdatedAt string `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	RevokedAt string `json:"revoked_at" gorm:"column:revoked_at;type:timestamp"`
	ExpiresAt string `json:"expires_at" gorm:"column:expires_at;type:timestamp"`
	UserID    string `json:"user_id,omitempty" gorm:"column:user_id;index;type:varchar(32)"`
	*DbRecord `json:"db_record" gorm:"embedded"`
}

func (a ApiKey) GetUserID() string {
	return a.UserID
}
