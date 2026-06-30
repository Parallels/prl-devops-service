package models

import "time"

type DbRecord struct {
	IsLocked bool       `json:"is_locked" gorm:"column:is_locked;default:false;not null;type:boolean"`
	LockedBy string     `json:"locked_by" gorm:"column:locked_by;type:varchar(255)"`
	LockedAt *time.Time `json:"locked_at,omitempty" gorm:"column:locked_at;type:timestamp"`
}
