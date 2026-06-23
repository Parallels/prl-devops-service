package models

type Role struct {
	ID          string  `json:"id,omitempty" gorm:"column:id;primaryKey;not null;type:varchar(64)"`
	Name        string  `json:"name" gorm:"column:name;type:varchar(64);unique;not null"`
	Description string  `json:"description,omitempty" gorm:"column:description;type:text"`
	Internal    bool    `json:"internal" gorm:"column:internal;type:boolean;default:false;not null"`
	Claims      []Claim `json:"claims,omitempty" gorm:"many2many:role_claims;"`
	Users       []User  `json:"-" gorm:"many2many:user_roles;"`
}
