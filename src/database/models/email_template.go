package models

import (

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailTemplate struct {
	BaseModel
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
