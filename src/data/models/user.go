package models

type User struct {
	ID                  string  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Username            string  `json:"username" gorm:"column:username;type:varchar(255);unique;not null"`
	Name                string  `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Email               string  `json:"email" gorm:"column:email;type:varchar(255);not null;unique"`
	Password            string  `json:"password,omitempty" gorm:"column:password;type:varchar(255);not null"`
	CreatedAt           string  `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt           string  `json:"updated_at,omitempty" gorm:"column:updated_at"`
	Roles               []Role  `json:"roles,omitempty" gorm:"many2many:user_roles"`
	Claims              []Claim `json:"claims,omitempty" gorm:"many2many:user_claims"`
	FailedLoginAttempts int     `json:"failed_login_attempts,omitempty" gorm:"column:failed_login_attempts;default:0"`
	Blocked             bool    `json:"blocked,omitempty" gorm:"column:blocked"`
	BlockedSince        string  `json:"blocked_since,omitempty" gorm:"column:blocked_since"`
	BlockedReason       string  `json:"blocked_reason,omitempty" gorm:"column:blocked_reason"`
}
