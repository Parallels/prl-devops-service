package entities


type RoleClaims struct {
	RoleID  string `json:"role_id" gorm:"not null;type:text;index;default:'00000000-0000-0000-0000-000000000000'"`
	ClaimID string `json:"claim_id" gorm:"not null;type:text;index;default:'00000000-0000-0000-0000-000000000000'"`
	// Add unique constraint to prevent duplicates
	// This will be handled by GORM's many2many relationship
}

func (RoleClaims) TableName() string {
	return "role_claims"
}
