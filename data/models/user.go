package models

type User struct {
	ID        string  `json:"id,omitempty"`
	Username  string  `json:"username"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Password  string  `json:"password,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
	Roles     []Role  `json:"roles,omitempty"`
	Claims    []Claim `json:"claims,omitempty"`
}
