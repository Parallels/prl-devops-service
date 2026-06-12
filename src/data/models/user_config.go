package models

type UserConfigValueType string

const (
	UserConfigValueTypeString UserConfigValueType = "string"
	UserConfigValueTypeBool   UserConfigValueType = "bool"
	UserConfigValueTypeInt    UserConfigValueType = "int"
	UserConfigValueTypeJSON   UserConfigValueType = "json"
)

type UserConfig struct {
	ID        string              `json:"id" gorm:"column:id;primaryKey"`
	UserID    string              `json:"user_id" gorm:"column:user_id;index;type:varchar(255);not null"`
	Slug      string              `json:"slug" gorm:"column:slug;type:varchar(255);not null"`
	Name      string              `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Type      UserConfigValueType `json:"type" gorm:"column:type;type:varchar(255);not null" `
	Value     string              `json:"value" gorm:"column:value"`
	CreatedAt string              `json:"created_at" gorm:"column:created_at"`
	UpdatedAt string              `json:"updated_at" gorm:"column:updated_at"`
	*DbRecord `json:"db_record" gorm:"embedded"`
}
