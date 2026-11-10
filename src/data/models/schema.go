package models

// DatabaseSchema is the single-row meta root (mapping row 26; ≤1 row —
// legacy meta record; #424 may redirect to migration_meta). Gap-pass of
// PBI #419 task 06: rows 25/26 must both exist as roots; Configuration
// landed in wave B, this straggler lands here.
type DatabaseSchema struct {
	Version string `json:"version" gorm:"primaryKey"`
}
