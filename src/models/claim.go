package models

import "github.com/Parallels/prl-devops-service/errors"

type ClaimRequest struct {
	Name string `json:"name"`
}

func (r *ClaimRequest) Validate() error {
	if r.Name == "" {
		return errors.NewWithCode("Claim name cannot be empty", 400)
	}

	return nil
}

type ClaimResponse struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}
