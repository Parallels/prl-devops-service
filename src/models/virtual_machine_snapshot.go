package models

import "github.com/Parallels/prl-devops-service/errors"

// SnapshotCreateRequest represents a request to create a new snapshot
type SnapshotCreateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Validate validates the snapshot create request
func (r *SnapshotCreateRequest) Validate() error {
	if r.Name == "" {
		return errors.NewWithCode("name is required", 400)
	}
	return nil
}

type SnapshotSwitchRequest struct {
	SnapshotId string `json:"snapshot_id"`
	SkipResume bool   `json:"skip_resume,omitempty"` // VM will not be started if it was running when snapshot was taken
}

// Validate validates the snapshot switch request
func (r *SnapshotSwitchRequest) Validate() error {
	if r.SnapshotId == "" {
		return errors.NewWithCode("snapshot_id is required", 400)
	}
	return nil
}

// SnapshotDetails represents information about a single snapshot
type SnapshotDetails struct {
	Name    string `json:"name,omitempty"`
	Date    string `json:"date"`
	State   string `json:"state"`
	Current bool   `json:"current"`
	Parent  string `json:"parent,omitempty"`
}

type SnapshotListResponse map[string]SnapshotDetails

// SnapshotDeleteRequest represents a request to delete a snapshot
type SnapshotDeleteRequest struct {
	SnapshotId     string `json:"snapshot_id"`
	DeleteChildren bool   `json:"delete_children,omitempty"` // Flag indicating action on child snapshots
}

// Validate validates the snapshot delete request
func (r *SnapshotDeleteRequest) Validate() error {
	if r.SnapshotId == "" {
		return errors.NewWithCode("snapshot_id is required", 400)
	}
	return nil
}
