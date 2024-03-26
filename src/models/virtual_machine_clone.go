package models

import "github.com/Parallels/prl-devops-service/errors"

type VirtualMachineCloneCommandRequest struct {
	CloneName string `json:"clone_name"`
}

func (r *VirtualMachineCloneCommandRequest) Validate() error {
	if r.CloneName == "" {
		return errors.NewWithCode("missing clone name", 400)
	}

	return nil
}

type VirtualMachineCloneCommandResponse struct {
	Id     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}
