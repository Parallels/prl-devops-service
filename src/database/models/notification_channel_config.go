package models


// NotificationChannelConfig represents the configuration for a provider on a channel
type NotificationChannelConfig struct {
	BaseModel
	ChannelID     string              `json:"channel_id" gorm:"type:text;not null;index"`
	Channel       NotificationChannel `json:"channel" gorm:"foreignKey:ChannelID;references:ID;constraint:OnDelete:CASCADE"`
	ProviderType  string              `json:"provider_type" gorm:"type:text;not null"` // e.g., "webhook", "in_app"
	Configuration string              `json:"configuration" gorm:"type:text"`          // JSON payload
}

func (NotificationChannelConfig) TableName() string {
	return "notification_channel_configs"
}
