package models

type Claim struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Internal    bool   `json:"internal"`
	Group       string `json:"group,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Action      string `json:"action,omitempty"`
	Users       []User `json:"-"`
}
