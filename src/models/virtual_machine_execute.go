package models

import "github.com/Parallels/prl-devops-service/errors"

type VirtualMachineExecuteCommandRequest struct {
	EnvironmentVariables map[string]string                   `json:"environment_variables,omitempty"`
	Command              string                              `json:"command"`
	Script               *VirtualMachineExecuteCommandScript `json:"script,omitempty"`
	UseSudo              bool                                `json:"use_sudo,omitempty"`
	UseSSH               bool                                `json:"use_ssh,omitempty"`
	User                 string                              `json:"user,omitempty"`
}

func (r *VirtualMachineExecuteCommandRequest) Validate() error {
	if r.Command == "" || r.Script == nil {
		if r.Script == nil {
			return errors.NewWithCode("missing script", 400)
		}

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

type VirtualMachineExecuteCommandScript struct {
	LocalPath  string `json:"path"`
	RemotePath string `json:"remote_path,omitempty"`
	Parameters string `json:"parameters,omitempty"`
}
