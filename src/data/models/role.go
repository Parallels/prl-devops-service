package models

type Role struct {
	ID          string  `json:"id,omitempty" gorm:"primaryKey"`
	Name        string  `json:"name" gorm:"type:varchar(255);not null"`
	Description string  `json:"description,omitempty"`
	Internal    bool    `json:"internal" gorm:"default:false"`
	Claims      []Claim `json:"claims,omitempty" gorm:"many2many:role_claims;"`
	Users       []User  `json:"-" gorm:"many2many:user_roles;"`
}
