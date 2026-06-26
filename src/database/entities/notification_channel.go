package entities

import "github.com/Parallels/prl-devops-service/database/common"

// NotificationChannel represents a logical notification channel (topic)
type NotificationChannel struct {
	common.BaseModelWithTenant
	Name        string `json:"name" gorm:"type:text;not null;uniqueIndex:idx_name_tenant"`
	Description string `json:"description" gorm:"type:text"`
}

func (NotificationChannel) TableName() string {
	return "notification_channels"
}
