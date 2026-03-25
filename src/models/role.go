package models

import "github.com/Parallels/prl-devops-service/errors"

type RoleRequest struct {
	Name   string   `json:"name"`
	Claims []string `json:"claims,omitempty"`
}

func (r *RoleRequest) Validate() error {
	if r.Name == "" {
		return errors.NewWithCode("Role name cannot be empty", 400)
	}

	return nil
}

type RoleClaimRequest struct {
	Name string `json:"name"`
}

func (r *RoleClaimRequest) Validate() error {
	if r.Name == "" {
		return errors.NewWithCode("Claim name cannot be empty", 400)
	}

	return nil
}

type RoleResponse struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Claims      []ClaimResponse `json:"claims"`
	Users       []ApiUser       `json:"users"`
}
