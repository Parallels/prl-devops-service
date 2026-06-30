package models

type Role struct {
	BaseModel           // Embeds ID, CreatedAt, UpdatedAt, DeletedAt
	Name        string  `json:"name" gorm:"column:name;type:varchar(64);unique;not null"`
	Description string  `json:"description,omitempty" gorm:"column:description;type:text"`
	Internal    bool    `json:"internal" gorm:"column:internal;type:boolean;default:false"`
	Claims      []Claim `json:"claims,omitempty" gorm:"many2many:role_claims;"`
	Users       []User  `json:"-" gorm:"many2many:user_roles;"`
}

func (Role) TableName() string {
	return "roles"
}
