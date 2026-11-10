package models

type Claim struct {
	ID          string `json:"id,omitempty" gorm:"primaryKey"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Internal    bool   `json:"internal"`
	Group       string `json:"group,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Action      string `json:"action,omitempty"`
	Users       []User `json:"-" gorm:"many2many:user_claims;foreignkey:ID;associationForeignKey:ID"` // inverse; User.Claims owns the join
}
