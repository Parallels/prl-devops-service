package entities

import "github.com/Parallels/prl-devops-service/database/common"

import "time"

type User struct {
	common.BaseModel
	Name                   string    `json:"name" gorm:"not null;type:text"`
	Username               string    `json:"username" gorm:"not null;unique;type:text"`
	Password               string    `json:"password" gorm:"not null;type:text"`
	Email                  string    `json:"email" gorm:"not null;unique;type:text"`
	Roles                  []Role    `json:"roles" gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;"`
	Claims                 []Claim   `json:"claims" gorm:"many2many:user_claims;constraint:OnDelete:CASCADE;"`
	Status                 string    `json:"status" gorm:"not null;type:text;default:'active'"`
	TenantID               string    `json:"tenant_id" gorm:"type:text"`
	TwoFactorEnabled       bool      `json:"two_factor_enabled" gorm:"type:boolean;not null;default:false"`
	TwoFactorSecret        string    `json:"two_factor_secret" gorm:"type:text;not null;default:''"`
	TwoFactorVerified      bool      `json:"two_factor_verified" gorm:"type:boolean;not null;default:false"`
	TwoFactorRecoveryCodes string    `json:"two_factor_recovery_codes" gorm:"type:text;not null;default:''"`
	TwoFactorAuthType      string    `json:"two_factor_auth_type" gorm:"type:text;not null;default:''"`
	Blocked                bool      `json:"blocked" gorm:"type:boolean;not null;default:false"`
	RefreshToken           string    `json:"refresh_token" gorm:"type:text;not null;default:''"`
	RefreshTokenExpiresAt  time.Time `json:"refresh_token_expires_at" gorm:"type:timestamp"`
}

func (User) TableName() string {
	return "users"
}
