package models

import (

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebAuthnCredential struct {
	BaseModel
	UserID          string `json:"user_id" gorm:"size:36;not null;index"`
	CredentialID    []byte `json:"credential_id" gorm:"uniqueIndex;not null"` // Binary data
	PublicKey       []byte `json:"public_key" gorm:"not null"`                // Binary data
	AttestationType string `json:"attestation_type" gorm:"size:50"`
	Transport       string `json:"transport" gorm:"size:255"` // Comma-separated or JSON
	AAGUID          string `json:"aaguid" gorm:"size:36"`
	SignCount       uint32 `json:"sign_count" gorm:"default:0"`
	CloneWarning    bool   `json:"clone_warning" gorm:"default:false"`
}

func (e *WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

func (e *WebAuthnCredential) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return
}
