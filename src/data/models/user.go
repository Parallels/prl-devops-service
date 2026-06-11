package models

type User struct {
	ID                  string  `json:"id,omitempty" gorm:"primaryKey"`
	Username            string  `json:"username" gorm:"type:varchar(255);unique;not null"`
	Name                string  `json:"name" gorm:"type:varchar(255);not null"`
	Email               string  `json:"email" gorm:"type:varchar(255);not null;unique"`
	Password            string  `json:"password,omitempty" gorm:"type:varchar(255);not null"`
	CreatedAt           string  `json:"created_at,omitempty"`
	UpdatedAt           string  `json:"updated_at,omitempty"`
	Roles               []Role  `json:"roles,omitempty" gorm:"many2many:user_roles"`
	Claims              []Claim `json:"claims,omitempty" gorm:"many2many:user_claims"`
	FailedLoginAttempts int     `json:"failed_login_attempts,omitempty" gorm:"default:0"`
	Blocked             bool    `json:"blocked,omitempty"`
	BlockedSince        string  `json:"blocked_since,omitempty"`
	BlockedReason       string  `json:"blocked_reason,omitempty"`
}
