package entities

import "github.com/Parallels/prl-devops-service/database/common"

import (
	"fmt"

)

type Claim struct {
	common.BaseModelWithTenant
	Service       string               `json:"service" gorm:"not null;type:text"`
	Module        string               `json:"module" gorm:"not null;type:text"`
	Action        AccessLevel   `json:"action" gorm:"not null;type:text"`
	SecurityLevel SecurityLevel `json:"security_level" gorm:"not null;type:text"`
}

func (c *Claim) GetSlug() string {
	return fmt.Sprintf("%s::%s::%s", c.Service, c.Module, c.Action)
}

func (Claim) TableName() string {
	return "claims"
}
