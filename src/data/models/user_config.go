package models

type UserConfigValueType string

const (
	UserConfigValueTypeString UserConfigValueType = "string"
	UserConfigValueTypeBool   UserConfigValueType = "bool"
	UserConfigValueTypeInt    UserConfigValueType = "int"
	UserConfigValueTypeJSON   UserConfigValueType = "json"
)

type UserConfig struct {
	ID        string              `json:"id" gorm:"column:id;primaryKey;not null;type:varchar(64)"`
	UserID    string              `json:"user_id" gorm:"column:user_id;index;type:varchar(64);not null"`
	Slug      string              `json:"slug" gorm:"column:slug;type:varchar(64);not null"`
	Name      string              `json:"name" gorm:"column:name;type:varchar(64);not null"`
	Type      UserConfigValueType `json:"type" gorm:"column:type;type:varchar(64);not null" `
	Value     string              `json:"value" gorm:"column:value;type:text;not null"`
	CreatedAt string              `json:"created_at" gorm:"column:created_at;type:timestamp"`
	UpdatedAt string              `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	*DbRecord `json:"db_record" gorm:"embedded"`
}
