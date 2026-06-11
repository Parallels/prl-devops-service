package models

type Claim struct {
	ID          string `json:"id,omitempty" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"unique;not null"`
	Description string `json:"description,omitempty"`
	Internal    bool   `json:"internal" gorm:"default:false"`
	Group       string `json:"group,omitempty" gorm:"type:varchar(255)"`
	Resource    string `json:"resource,omitempty" gorm:"type:varchar(255)"`
	Action      string `json:"action,omitempty" gorm:"type:varchar(255)"`
	Users       []User `json:"-" gorm:"many2many:user_claims"`
}
