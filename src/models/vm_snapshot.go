package models

type CreateVMSnapshotRequest struct {
	SnapshotName        string `json:"snapshot_name,omitempty"`
	SnapshotDescription string `json:"snapshot_description,omitempty"`
}

type CreateVMSnapshotResponse struct {
	SnapshotName string `json:"snapshot_name,omitempty"`
	SnapshotId   string `json:"snapshot_id,omitempty"`
}

type DeleteVMSnapshotRequest struct {
	DeleteChildren bool `json:"delete_children,omitempty"`
}

type VMSnapshot struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	State   string `json:"state"`
	Current bool   `json:"current"`
	Parent  string `json:"parent"`
  Children []VMSnapshot `json:"children,omitempty"`
}

type VMSnapshots []VMSnapshot

type ListVMSnapshotResponse struct {
	Snapshots VMSnapshots `json:"snapshots,omitempty"`
}

type RevertVMSnapshotRequest struct {
	SkipResume bool `json:"skip_resume,omitempty"`
}
