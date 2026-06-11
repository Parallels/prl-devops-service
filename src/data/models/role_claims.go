package models

type RoleClaims struct {
	RoleID  string `json:"role_id" gorm:"primaryKey;type:varchar(255)"`
	ClaimID string `json:"claim_id" gorm:"primaryKey;type:varchar(255)"`
}

func (RoleClaims) TableName() string {
	return "role_claims"
}
