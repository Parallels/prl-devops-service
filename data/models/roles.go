package models

type Role struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Internal bool   `json:"internal"`
}
