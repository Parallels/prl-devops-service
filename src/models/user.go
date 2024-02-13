package models

import (
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
)

type UserCreateRequest struct {
	Username string   `json:"username"`
	Name     string   `json:"name,omitempty"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles,omitempty"`
	Claims   []string `json:"claims,omitempty"`
}

func (r *UserCreateRequest) Validate() error {
	if r.Username == "" {
		return errors.NewWithCode("Username is required", 400)
	}
	if r.Name == "" {
		return errors.NewWithCode("Name is required", 400)
	}
	if r.Email == "" {
		return errors.NewWithCode("Email is required", 400)
	}
	if len(r.Roles) == 0 {
		r.Roles = append(r.Roles, constants.USER_ROLE)
	}

	if len(r.Claims) == 0 {
		r.Claims = append(r.Claims,
			constants.READ_ONLY_CLAIM,
			constants.LIST_CATALOG_MANIFEST_CLAIM,
			constants.LIST_PACKER_TEMPLATE_CLAIM,
			constants.LIST_VM_CLAIM,
			constants.LIST_CLAIM)
	}

	return nil
}

type ApiUser struct {
	ID       string   `json:"id,omitempty"`
	Username string   `json:"username"`
	Name     string   `json:"name,omitempty"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles,omitempty"`
	Claims   []string `json:"claims,omitempty"`
}
