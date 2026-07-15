package models

// Legacy Claim for JSON database
type Claim struct {
	ID          string  `json:"id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Internal    bool    `json:"internal"`
	Group       string  `json:"group,omitempty"`
	Resource    string  `json:"resource,omitempty"`
	Action      string  `json:"action,omitempty"`
	Users       []User  `json:"-"`
}
