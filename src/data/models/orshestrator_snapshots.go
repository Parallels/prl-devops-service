package models

type OrchestratorSnapshot struct {
	HostId string `json:"host_id"`
	// map of vmId to list of snapshots
	Snapshots map[string][]Snapshot `json:"snapshots"`
}
