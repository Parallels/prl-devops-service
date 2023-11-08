package models

import (
	"fmt"
	"os"
	"strings"

	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
)

type VirtualMachineConfigRequest struct {
	Owner      string                                  `json:"owner"`
	Operations []*VirtualMachineConfigRequestOperation `json:"operations"`
}

type VirtualMachineConfigRequestOperation struct {
	Owner     string                                        `json:"owner"`
	Group     string                                        `json:"group"`
	Operation string                                        `json:"operation"`
	Value     string                                        `json:"value"`
	Options   []*VirtualMachineConfigRequestOperationOption `json:"options"`
	Flags     []string                                      `json:"flags"`
	Error     error                                         `json:"error"`
}

type VirtualMachineConfigRequestOperationOption struct {
	Flag  string `json:"flag"`
	Value string `json:"value"`
}

type VirtualMachineConfigResponse struct {
	Operations []VirtualMachineConfigResponseOperation `json:"operations"`
}

type VirtualMachineConfigResponseOperation struct {
	Group     string `json:"group"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

func (r *VirtualMachineConfigRequest) Validate() error {
	if len(r.Operations) == 0 {
		return errors.ErrConfigOperationEmpty()
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	for _, op := range r.Operations {
		if err := op.Validate(); err != nil {
			return err
		}
		if r.Owner == "" {
			r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
		}
	}
	return nil
}

func (r *VirtualMachineConfigRequest) HasErrors() bool {
	for _, op := range r.Operations {
		if op.Error != nil {
			return true
		}
	}

	return false
}

func (r *VirtualMachineConfigRequestOperation) Validate() error {
	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}
	if r.Group == "" {
		return errors.ErrConfigGroupEmpty()
	}
	if r.Operation == "" {
		return errors.ErrConfigOperationEmpty()
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
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "machine":
		if r.Value == "" {
			return errors.ErrValueEmpty()
		}
		if r.Operation != "rename" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "cpu":
		if r.Value == "" {
			return errors.ErrValueEmpty()
		}
		if r.Operation != "set" &&
			r.Operation != "set_type" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "memory":
		if r.Value == "" {
			return errors.ErrValueEmpty()
		}
		if r.Operation != "set" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "boot-order":
		if r.Value == "" {
			return errors.ErrValueEmpty()
		}
	case "efi-secure-boot":
		if r.Value == "" {
			return errors.ErrValueEmpty()
		}
	case "time":
		if r.Value == "" {
			return errors.ErrValueEmpty()
		}
	case "device":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "add" &&
			r.Operation != "remove" &&
			r.Operation != "set" &&
			r.Operation != "disconnect" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "shared-folder":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "add" &&
			r.Operation != "delete" &&
			r.Operation != "set" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "network":
		if r.Value == "" {
			return errors.New("Value cannot be empty")
		}
		if r.Operation != "add" &&
			r.Operation != "remove" &&
			r.Operation != "set" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "rosetta":
		if r.Value == "" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
		if r.Operation != "set" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	case "cmd":
		if r.Operation == "" {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
		if r.Flags == nil || len(r.Flags) == 0 {
			return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
		}
	default:
		return errors.ErrConfigOperationNotSupported(r.Group, r.Operation)
	}

	return nil
}

func (r *VirtualMachineConfigRequestOperation) GetOption(key string) *VirtualMachineConfigRequestOperationOption {
	for _, opt := range r.Options {
		if opt.Flag == key {
			return opt
		}
	}

	return nil
}

func (r *VirtualMachineConfigRequestOperation) GetCmdArgs() []string {
	args := make([]string, 0)
	for _, flag := range r.Flags {
		args = append(args, fmt.Sprintf("--%s", flag))
	}
	for _, option := range r.Options {
		args = append(args, fmt.Sprintf("--%s", r.CleanString(option.Flag)), r.CleanString(option.Value))
	}

	return args
}

func (r *VirtualMachineConfigRequestOperation) GetRawFlagsArgs() []string {
	args := make([]string, 0)
	for _, flag := range r.Flags {
		if !strings.HasPrefix(flag, "--") {
			flag = fmt.Sprintf("--%s", flag)
		}
		args = append(args, r.CleanString(flag))
	}

	return args
}

func (r *VirtualMachineConfigRequestOperation) GetRawOptionsArgs() []string {
	args := make([]string, 0)
	for _, option := range r.Options {
		if !strings.HasPrefix(option.Flag, "--") {
			option.Flag = fmt.Sprintf("--%s", option.Flag)
		}
		args = append(args, fmt.Sprintf("%s %s", r.CleanString(option.Flag), r.CleanString(option.Value)))
	}

	return args
}

func (r *VirtualMachineConfigRequestOperation) CleanString(value string) string {
	value = strings.ReplaceAll(value, "\n\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\t", "")
	value = strings.ReplaceAll(value, "\v", "")

	return value
}
