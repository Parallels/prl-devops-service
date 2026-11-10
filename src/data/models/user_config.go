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
	UserID    string              `json:"user_id" gorm:"index;uniqueIndex:ux_userconfig_user_slug"`
	Slug      string              `json:"slug" gorm:"uniqueIndex:ux_userconfig_user_slug"`
	Name      string              `json:"name"`
	Type      UserConfigValueType `json:"type"`
	Value     string              `json:"value"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	*DbRecord `json:"db_record" gorm:"-:all"`
}
