package models

// UserConfigValueType enum for legacy JSON database
type UserConfigValueType string

const (
	UserConfigValueTypeString UserConfigValueType = "string"
	UserConfigValueTypeBool   UserConfigValueType = "bool"
	UserConfigValueTypeInt    UserConfigValueType = "int"
	UserConfigValueTypeJSON   UserConfigValueType = "json"
)

// Legacy UserConfig for JSON database
type UserConfig struct {
	ID        string              `json:"id"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	DeletedAt *string             `json:"deleted_at,omitempty"`
	UserID    string              `json:"user_id"`
	Slug      string              `json:"slug"`
	Name      string              `json:"name"`
	Type      UserConfigValueType `json:"type"`
	Value     string              `json:"value"`
	*DbRecord `json:"db_record"`
}
