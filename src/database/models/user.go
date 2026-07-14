package models

import "time"

type User struct {
	BaseModel                      // Embeds ID, CreatedAt, UpdatedAt, DeletedAt
	Username            string     `json:"username" gorm:"column:username;type:varchar(255);unique;not null"`
	Name                string     `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Email               string     `json:"email" gorm:"column:email;type:varchar(255);not null;unique"`
	Password            string     `json:"password,omitempty" gorm:"column:password;type:varchar(255);not null"`
	Roles               []Role     `json:"roles,omitempty" gorm:"many2many:user_roles;"`
	Claims              []Claim    `json:"claims,omitempty" gorm:"many2many:user_claims"`
	FailedLoginAttempts int        `json:"failed_login_attempts,omitempty" gorm:"column:failed_login_attempts;default:0;type:integer"`
	Blocked             bool       `json:"blocked,omitempty" gorm:"column:blocked;type:boolean;default:false"`
	BlockedSince        *time.Time `json:"blocked_since,omitempty" gorm:"column:blocked_since;type:timestamp"`
	BlockedReason       string     `json:"blocked_reason,omitempty" gorm:"column:blocked_reason;type:text"`
}

func (User) TableName() string {
	return "users"
}
