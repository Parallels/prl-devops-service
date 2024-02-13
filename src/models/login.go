package models

import "github.com/Parallels/prl-devops-service/errors"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.NewWithCode("Email is required", 400)
	}
	if r.Password == "" {
		return errors.NewWithCode("Password is required", 400)
	}

	return nil
}

type LoginResponse struct {
	Email     string `json:"email,omitempty"`
	Token     string `json:"token,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

func (r *ValidateTokenRequest) Validate() error {
	if r.Token == "" {
		return errors.NewWithCode("Token is required", 400)
	}

	return nil
}

type ValidateTokenResponse struct {
	Valid bool `json:"valid"`
}
