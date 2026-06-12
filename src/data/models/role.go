package models

type Role struct {
	ID          string  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name        string  `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Description string  `json:"description,omitempty" gorm:"column:description"`
	Internal    bool    `json:"internal" gorm:"column:internal;default:false"`
	Claims      []Claim `json:"claims,omitempty" gorm:"many2many:role_claims;"`
	Users       []User  `json:"-" gorm:"many2many:user_roles;"`
}
