package entities

import "github.com/Parallels/prl-devops-service/database/common"

type Role struct {
	common.BaseModelWithTenant
	Name          string        `json:"name" gorm:"not null;type:text"`
	Description   string        `json:"description" gorm:"not null;type:text"`
	SecurityLevel SecurityLevel `json:"security_level" gorm:"not null;type:text"`
	Claims        []Claim       `json:"claims" gorm:"many2many:role_claims;constraint:OnDelete:CASCADE;"`
}

func (Role) TableName() string {
	return "roles"
}
