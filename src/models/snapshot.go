package models

type CreateSnapShotRequest struct {
	SnapshotName        string `json:"snapshot_name,omitempty"`
	SnapshotDescription string `json:"snapshot_description,omitempty"`
}

type CreateSnapShotResponse struct {
	SnapshotName string `json:"snapshot_name,omitempty"`
	SnapshotId   string `json:"snapshot_id,omitempty"`
}

type DeleteSnapshotRequest struct {
	DeleteChildren bool `json:"delete_children,omitempty"`
}

type ListSnapshotRequest struct {
}
type Snapshot struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	State   string `json:"state"`
	Current bool   `json:"current"`
	Parent  string `json:"parent"`
}

type ListSnapshotResponse struct {
	Snapshots []Snapshot `json:"snapshots,omitempty"`
}

type RevertSnapshotRequest struct {
	SkipResume bool `json:"skip_resume,omitempty"`
}
