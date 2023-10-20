package models

type User struct {
	ID       string      `json:"id,omitempty"`
	Username string      `json:"username"`
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	Roles    []UserRole  `json:"roles"`
	Claims   []UserClaim `json:"claims"`
}

type UserClaim struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type UserRole struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}
