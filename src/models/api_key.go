package models

import (
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

type ApiKeyRequest struct {
	Name      string `json:"name"`
	Key       string `json:"key"`
	Secret    string `json:"secret"`
	Revoked   bool   `json:"revoked,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	RevokedAt string `json:"revoked_at,omitempty"`
}

func (r *ApiKeyRequest) Validate() error {
	if r.Name == "" {
		return errors.NewWithCode("Name is required", 400)
	}
	if r.Key == "" {
		return errors.NewWithCode("Key is required", 400)
	}
	if r.Secret == "" {
		return errors.NewWithCode("Secret is required", 400)
	}

	r.Key = strings.ToUpper(helpers.NormalizeString(r.Key))

	return nil
}

type ApiKeyResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Key     string `json:"key"`
	Encoded string `json:"encoded,omitempty"`
	Revoked bool   `json:"revoked,omitempty"`
}
