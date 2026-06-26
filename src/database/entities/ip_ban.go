package entities

import "github.com/Parallels/prl-devops-service/database/common"

import (
	"time"
)

type IpBan struct {
	common.BaseModelWithTenant
	IP        string     `json:"ip" gorm:"type:varchar(50);uniqueIndex;not null"`
	Reason    string     `json:"reason" gorm:"type:text"`
	BanLevel  string     `json:"ban_level" gorm:"type:varchar(50);default:'global'"` // global, tenant
	Enabled   bool       `json:"enabled" gorm:"default:true"`
	BannedAt  time.Time  `json:"banned_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (IpBan) TableName() string {
	return "ip_bans"
}
