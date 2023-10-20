package models

type User struct {
	ID        string      `json:"id,omitempty"`
	Username  string      `json:"username"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Password  string      `json:"password,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
	UpdatedAt string      `json:"updated_at,omitempty"`
	Roles     []UserRole  `json:"roles,omitempty"`
	Claims    []UserClaim `json:"claims,omitempty"`
}
