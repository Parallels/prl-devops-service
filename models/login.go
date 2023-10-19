package models

type ApiLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
