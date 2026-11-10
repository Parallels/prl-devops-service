package models

type User struct {
	ID                  string  `json:"id,omitempty" gorm:"primaryKey"`
	Username            string  `json:"username" gorm:"uniqueIndex"`
	Name                string  `json:"name"`
	Email               string  `json:"email" gorm:"index"` // NOT unique — empty-string legacy rows (plan risk list)
	Password            string  `json:"password,omitempty"`
	CreatedAt           string  `json:"created_at,omitempty"` // D-1: string dates
	UpdatedAt           string  `json:"updated_at,omitempty"`
	Roles               []Role  `json:"roles,omitempty" gorm:"many2many:user_roles;"`
	Claims              []Claim `json:"claims,omitempty" gorm:"many2many:user_claims;"`
	FailedLoginAttempts int     `json:"failed_login_attempts,omitempty"`
	Blocked             bool    `json:"blocked,omitempty"`
	BlockedSince        string  `json:"blocked_since,omitempty"`
	BlockedReason       string  `json:"blocked_reason,omitempty"`
}
