package models


// NotificationChannel represents a logical notification channel (topic)
type NotificationChannel struct {
	BaseModel
	Name        string `json:"name" gorm:"type:text;not null;uniqueIndex:idx_name_tenant"`
	Description string `json:"description" gorm:"type:text"`
}

func (NotificationChannel) TableName() string {
	return "notification_channels"
}
