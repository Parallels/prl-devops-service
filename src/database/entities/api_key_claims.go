package entities


type ApiKeyClaims struct {
	ApiKeyID string `json:"api_key_id" gorm:"not null;type:text;index"`
	ClaimID  string `json:"claim_id" gorm:"not null;type:text;index"`
}

func (ApiKeyClaims) TableName() string {
	return "api_key_claims"
}
