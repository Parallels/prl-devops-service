package models

type Role struct {
	ID          string  `json:"id,omitempty" gorm:"primaryKey"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Internal    bool    `json:"internal"`
	Claims      []Claim `json:"claims,omitempty" gorm:"many2many:role_claims;joinForeignKey:RoleID;joinReferences:ClaimID"`
	Users       []User  `json:"-" gorm:"many2many:user_roles;foreignkey:ID;associationForeignKey:ID"` // inverse; User.Roles owns the join
}
