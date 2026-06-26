package entities

import "github.com/Parallels/prl-devops-service/database/common"

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailTemplate represents an email template in the database
type EmailTemplate struct {
	common.BaseModelWithTenant
	Name     string `json:"name" gorm:"size:100;not null"`
	Slug     string `json:"slug" gorm:"size:100;not null;uniqueIndex"`
	Subject  string `json:"subject" gorm:"size:255;not null"`
	Body     string `json:"body" gorm:"type:text;not null"`
	IsSystem bool   `json:"is_system" gorm:"default:false"`
}

func (e *EmailTemplate) TableName() string {
	return "email_templates"
}

func (e *EmailTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return
}
