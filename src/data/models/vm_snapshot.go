package models

type VMSnapshot struct {
	VMId      string     `json:"vm_id"`
	Snapshots []Snapshot `json:"snapshots"`
}
