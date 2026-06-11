package models

type UserConfigValueType string

const (
	UserConfigValueTypeString UserConfigValueType = "string"
	UserConfigValueTypeBool   UserConfigValueType = "bool"
	UserConfigValueTypeInt    UserConfigValueType = "int"
	UserConfigValueTypeJSON   UserConfigValueType = "json"
)

type UserConfig struct {
	ID        string              `json:"id" gorm:"primaryKey"`
	UserID    string              `json:"user_id" gorm:"index;type:varchar(255);not null"`
	Slug      string              `json:"slug" gorm:"type:varchar(255);not null"`
	Name      string              `json:"name" gorm:"type:varchar(255);not null"`
	Type      UserConfigValueType `json:"type" gorm:"type:varchar(255);not null" `
	Value     string              `json:"value"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	*DbRecord `json:"db_record" gorm:"embedded"`
}
