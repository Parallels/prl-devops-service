package models

import "github.com/Parallels/prl-devops-service/errors"

type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	ApiKey   string `json:"api_key,omitempty"`
}

func (r *LoginRequest) Validate() error {
	// Accept either Email+Password, Username+Password, or ApiKey (with optional Password)
	if r.ApiKey != "" {
		// API key login, password optional
		return nil
	}
	if r.Email != "" {
		if r.Password == "" {
			return errors.NewWithCode("Password is required for email login", 400)
		}
		return nil
	}
	if r.Username != "" {
		if r.Password == "" {
			return errors.NewWithCode("Password is required for username login", 400)
		}
		return nil
	}
	return errors.NewWithCode("Either email, username or api_key must be provided", 400)
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
