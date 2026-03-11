package models

type Snapshot struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	State   string `json:"state"`
	Current bool   `json:"current"`
	Parent  string `json:"parent"`
}

type VMSnapshot struct {
	VMId      string     `json:"vm_id"`
	Snapshots []Snapshot `json:"snapshots"`
}
