package entities


type UserClaims struct {
	UserID  string `json:"user_id" gorm:"not null;type:text;index;default:'00000000-0000-0000-0000-000000000000'"`
	ClaimID string `json:"claim_id" gorm:"not null;type:text;index;default:'00000000-0000-0000-0000-000000000000'"`
}

func (UserClaims) TableName() string {
	return "user_claims"
}
