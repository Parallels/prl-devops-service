package models

import "Parallels/pd-api-service/errors"

type UnregisterVirtualMachineRequest struct {
	ID              string `json:"id"`
	Owner           string `json:"owner"`
	CleanSourceUuid bool   `json:"clean_source_uuid,omitempty"`
}

func (r *UnregisterVirtualMachineRequest) Validate() error {
	if r.ID == "" {
		return errors.ErrMissingId()
	}
	if r.Owner == "" {
		r.Owner = "root"
	}

	return nil
}
