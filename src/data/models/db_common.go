package models

type DbRecord struct {
	IsLocked bool   `json:"is_locked" gorm:"default:false"`
	LockedBy string `json:"locked_by" gorm:"type:varchar(255)"`
	LockedAt string `json:"locked_at"`
}
