package models

import "time"

// BaseModel contains common fields for all database models
// This follows the DRY principle and provides:
// - Unique ID for each record
// - Automatic timestamp tracking (created/updated)
// - Soft delete support (DeletedAt) - records are marked as deleted, not removed
type BaseModel struct {
	ID        string     `json:"id" gorm:"column:id;type:varchar(64);primaryKey;not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime;type:timestamp"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;type:timestamp"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;index;type:timestamp"`
}

// TableName can be overridden in each model if needed
func (BaseModel) TableName() string {
	return ""
}
