package models

import "time"

type Notification struct {
	BaseModel
	UserID  string     `json:"user_id" gorm:"type:text;not null;index"`
	Subject string     `json:"subject" gorm:"type:text;not null"`
	Content string     `json:"content" gorm:"type:text;not null"`
	Type    string     `json:"type" gorm:"type:text;not null"`
	Channel string     `json:"channel" gorm:"type:text"`
	Read    bool       `json:"read" gorm:"type:boolean;not null;default:false"`
	ReadAt  *time.Time `json:"read_at" gorm:"type:timestamp"`
}

func (Notification) TableName() string {
	return "notifications"
}
