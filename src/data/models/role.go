package models

type Role struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Internal    bool    `json:"internal"`
	Claims      []Claim `json:"claims,omitempty"`
	Users       []User  `json:"-"`
}
