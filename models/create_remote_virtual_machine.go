package models

import (
	"Parallels/pd-api-service/errors"
)

type CreateRemoteVirtualMachineRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	ApiKey   string `json:"api_key,omitempty"`
	Name     string `json:"name"`
	Owner    string `json:"owner,omitempty"`
}

func (r *CreateRemoteVirtualMachineRequest) Validate() error {
	if r.Host == "" {
		return errors.New("Host cannot be empty")
	}

	if r.Port == "" {
		return errors.New("Port cannot be empty")
	}

	if r.Username != "" && r.Password == "" {
		return errors.New("Username password cannot be empty")
	}

	if r.ApiKey == "" && r.Username == "" {
		return errors.New("ApiKey or Username cannot be empty")
	}

	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Owner == "" {
		return errors.New("Owner cannot be empty")
	}

	return nil
}
