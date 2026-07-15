package models

// Legacy Role for JSON database
type Role struct {
	ID          string  `json:"id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Internal    bool    `json:"internal"`
	Claims      []Claim `json:"claims,omitempty"`
	Users       []User  `json:"-"`
}
