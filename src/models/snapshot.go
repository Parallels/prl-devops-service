package models

import (
	"github.com/Parallels/prl-devops-service/errors"
)

type CreateSnapShotRequest struct {
	SnapshotName        string `json:"snapshot_name,omitempty"`
	SnapshotDescription string `json:"snapshot_description,omitempty"`
}

type CreateSnapShotResponse struct {
	SnapshotName string `json:"snapshot_name,omitempty"`
	SnapshotId   string `json:"snapshot_id,omitempty"`
}

type DeleteSnapshotRequest struct {
	VMId      string `json:"vm_id"`
	VMName    string `json:"vm_name"`
	ChildName string `json:"child_name,omitempty"`
}

func (r *DeleteSnapshotRequest) Validate() error {

	if r.VMId == "" && r.VMName == "" {
		return errors.New("vm_id and vm_name cannot be empty")
	}

	return nil
}

type ListSnapshotRequest struct {
	VMId   string `json:"vm_id"`
	VMName string `json:"vm_name"`
}

func (r *ListSnapshotRequest) Validate() error {

	if r.VMId == "" && r.VMName == "" {
		return errors.New("vm_id and vm_name cannot be empty")
	}

	return nil
}

type RevertSnapshotRequest struct {
	VMId   string `json:"vm_id"`
	VMName string `json:"vm_name"`
}

func (r *RevertSnapshotRequest) Validate() error {

	if r.VMId == "" && r.VMName == "" {
		return errors.New("vm_id and vm_name cannot be empty")
	}

	return nil
}

type SwitchSnapshotRequest struct {
	VMId       string `json:"vm_id"`
	VMName     string `json:"vm_name"`
	SkipResume bool   `json:"skip_resume,omitempty"`
}

func (r *SwitchSnapshotRequest) Validate() error {

	if r.VMId == "" && r.VMName == "" {
		return errors.New("vm_id and vm_name cannot be empty")
	}

	return nil
}
