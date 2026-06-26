package entities


import (
	"time"

	"github.com/Parallels/prl-devops-service/database/common"
)

type Tenant struct {
	common.BaseModel
	Name          string                                   `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Description   string                                   `json:"description" gorm:"column:description;type:text"`
	Domain        string                                   `json:"domain" gorm:"column:domain;type:varchar(255);not null;unique"`
	OwnerID       string                                   `json:"owner_id" gorm:"column:owner_id;type:varchar(255);"`
	ContactEmail  string                                   `json:"contact_email" gorm:"column:contact_email;type:varchar(255);"`
	Status        RecordStatus                   `json:"status" gorm:"column:status;type:varchar(50);default:'active'"`
	Country       string                                   `json:"country" gorm:"column:country;type:varchar(255);"`
	State         string                                   `json:"state" gorm:"column:state;type:varchar(255);"`
	City          string                                   `json:"city" gorm:"column:city;type:varchar(255);"`
	ActivatedAt   *time.Time                               `json:"activated_at" gorm:"column:activated_at;type:timestamp;"`
	DeactivatedAt *time.Time                               `json:"deactivated_at" gorm:"column:deactivated_at;type:timestamp;"`
	Metadata      common.JSONObject[map[string]interface{}] `json:"metadata" gorm:"column:metadata;type:text"`
	LogoURL       string                                   `json:"logo_url" gorm:"column:logo_url;type:varchar(255);"`
	Require2FA    bool                                     `json:"require_2fa" gorm:"column:require_2fa;type:boolean;default:false"`
}
