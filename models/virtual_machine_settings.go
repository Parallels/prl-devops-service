package models

import (
	"errors"
	"fmt"
)

type VirtualMachineSetRequest struct {
	Owner      string                        `json:"owner"`
	Operations []*VirtualMachineSetOperation `json:"operations"`
}

type VirtualMachineSetOperation struct {
	Owner     string                              `json:"owner"`
	Group     string                              `json:"group"`
	Operation string                              `json:"operation"`
	Value     string                              `json:"value"`
	Options   []*VirtualMachineSetOperationOption `json:"options"`
	Error     error                               `json:"error"`
}

type VirtualMachineSetOperationOption struct {
	Flag  string `json:"flag"`
	Value string `json:"value"`
}

type VirtualMachineSetResponse struct {
	Operations []VirtualMachineSetOperationResponse `json:"operations"`
}

type VirtualMachineSetOperationResponse struct {
	Group     string `json:"group"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

func (r *VirtualMachineSetRequest) Validate() error {
	if len(r.Operations) == 0 {
		return errors.New("Operations cannot be empty")
	}

	if r.Owner == "" {
		r.Owner = "root"
	}

	for _, op := range r.Operations {
		if err := op.Validate(); err != nil {
			return err
		}
		if op.Owner == "" {
			op.Owner = r.Owner
		}
	}
	return nil
}

func (r *VirtualMachineSetRequest) HasErrors() bool {
	for _, op := range r.Operations {
		if op.Error != nil {
			return true
		}
	}

	return false
}

func (r *VirtualMachineSetOperation) Validate() error {
	if r.Owner == "" {
		r.Owner = "root"
	}
	if r.Group == "" {
		return errors.New("Group cannot be empty")
	}
	if r.Operation == "" {
		return errors.New("Operation cannot be empty")
	}
	switch r.Group {
	case "state":
		if r.Operation != "start" &&
			r.Operation != "stop" &&
			r.Operation != "restart" &&
			r.Operation != "pause" &&
			r.Operation != "resume" &&
			r.Operation != "reset" &&
			r.Operation != "suspend" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	case "machine":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "rename" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	case "cpu":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "set" &&
			r.Operation != "set_type" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	case "memory":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "set" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	case "device":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "add" &&
			r.Operation != "remove" &&
			r.Operation != "set" &&
			r.Operation != "disconnect" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	case "shared_folder":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "add" &&
			r.Operation != "remove" &&
			r.Operation != "set" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	case "network":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "add" &&
			r.Operation != "remove" &&
			r.Operation != "set" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	case "rosetta":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "set" {
			return fmt.Errorf("Operation %s not supported on group %s", r.Operation, r.Group)
		}
	default:
		return fmt.Errorf("Group %s not supported", r.Group)
	}

	return nil
}

func (r *VirtualMachineSetOperation) GetOption(key string) *VirtualMachineSetOperationOption {
	for _, opt := range r.Options {
		if opt.Flag == key {
			return opt
		}
	}

	return nil
}
