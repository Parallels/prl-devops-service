package models

type Claim struct {
	ID          string `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name        string `json:"name" gorm:"column:name;type:varchar(255);unique;not null"`
	Description string `json:"description,omitempty" gorm:"column:description"`
	Internal    bool   `json:"internal" gorm:"column:internal;default:false"`
	Group       string `json:"group,omitempty" gorm:"column:group;type:varchar(255)"`
	Resource    string `json:"resource,omitempty" gorm:"column:resource;type:varchar(255)"`
	Action      string `json:"action,omitempty" gorm:"column:action;type:varchar(255)"`
	Users       []User `json:"-" gorm:"many2many:user_claims"`
}
