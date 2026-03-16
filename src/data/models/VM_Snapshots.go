package models

type VMSnapshot struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	State   string `json:"state"`
	Current bool   `json:"current"`
	Parent  string `json:"parent"`
}

type VMSnapshots struct {
	VMId       string       `json:"vm_id"`
	VMSnapshot []VMSnapshot `json:"vm_snapshots"`
}
