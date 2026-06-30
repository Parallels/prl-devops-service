package models

import "time"

type IpBan struct {
	BaseModel
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
