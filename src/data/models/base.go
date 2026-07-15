package models

// Legacy BaseModel for JSON database
// Uses string-based dates for compatibility with helpers.GetUtcCurrentDateTime()
type BaseModel struct {
	ID        string  `json:"id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at,omitempty"`
}
