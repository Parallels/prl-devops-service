package models

type DbRecord struct {
	IsLocked bool   `json:"is_locked" gorm:"column:is_locked;default:false"`
	LockedBy string `json:"locked_by" gorm:"column:locked_by;type:varchar(255)"`
	LockedAt string `json:"locked_at" gorm:"column:locked_at"`
}
