package entities

import "github.com/Parallels/prl-devops-service/database/common"

type Configuration struct {
	common.BaseModelWithTenant
	Vault   string `json:"vault" yaml:"vault" gorm:"column:vault;type:varchar(255);not null"`
	Key     string `json:"key" yaml:"key" gorm:"column:key;type:varchar(255);not null;unique"`
	Value   string `json:"value" yaml:"value" gorm:"column:value;type:text;not null"`
	Version int    `json:"version" yaml:"version" gorm:"column:version;type:int;not null"`
}

func (c *Configuration) TableName() string {
	return "configuration"
}
