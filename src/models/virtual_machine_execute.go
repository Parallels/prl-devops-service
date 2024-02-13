package models

import "github.com/Parallels/prl-devops-service/errors"

type VirtualMachineExecuteCommandRequest struct {
	Command string `json:"command"`
}

func (r *VirtualMachineExecuteCommandRequest) Validate() error {
	if r.Command == "" {
		return errors.NewWithCode("missing command", 400)
	}

	return nil
}

type VirtualMachineExecuteCommandResponse struct {
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}
