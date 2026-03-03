package models

type RoleClaims struct {
	RoleID  string `json:"role_id"`
	ClaimID string `json:"claim_id"`
}

func (RoleClaims) TableName() string {
	return "role_claims"
}
