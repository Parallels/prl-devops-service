package models

import "errors"

type RenameVirtualMachineRequest struct {
	ID          string `json:"id"`
	CurrentName string `json:"current_name"`
	Description string `json:"description"`
	NewName     string `json:"new_name"`
}

func (r *RenameVirtualMachineRequest) Validate() error {
	if r.ID == "" && r.CurrentName == "" {
		if r.ID == "" {
			return errors.New("missing id")
		}
		if r.CurrentName == "" {
			return errors.New("missing current_name")
		}
	}

	if r.NewName == "" {
		return errors.New("missing new_name")
	}

	return nil
}

func (r *RenameVirtualMachineRequest) GetId() string {
	if r.ID != "" {
		return r.ID
	} else {
		return r.CurrentName
	}
}
