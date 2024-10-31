package models

import "github.com/Parallels/prl-devops-service/errors"

type VirtualMachineUploadRequest struct {
	LocalPath  string `json:"path"`
	RemotePath string `json:"remote_path,omitempty"`
}

func (r *VirtualMachineUploadRequest) Validate() error {
	if r.LocalPath == "" || r.RemotePath == "" {
		if r.RemotePath == "" {
			return errors.NewWithCode("missing remote path", 400)
		}

		return errors.NewWithCode("missing local path", 400)
	}

	return nil
}

type VirtualMachineUploadResponse struct {
	LocalPath string `json:"path"`
	Error     string `json:"error,omitempty"`
}
