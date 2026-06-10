package models

type Role struct {
	ID          string  `json:"id,omitempty" gorm:"primaryKey"`
	Name        string  `json:"name" gorm:"not null"`
	Description string  `json:"description,omitempty"`
	Internal    bool    `json:"internal"`
	Claims      []Claim `json:"claims,omitempty"`
	Users       []User  `json:"-"`
}
