package models

type DbRecord struct {
	IsLocked bool   `json:"is_locked"`
	LockedBy string `json:"locked_by"`
	LockedAt string `json:"locked_at"`
}
