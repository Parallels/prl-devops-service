package models

type HostsVMSnapshotsRecord struct {
	HostId string `json:"host_id"`
	// map of vmId to list of snapshots
	VMSnapshots map[string][]VMSnapshot `json:"vm_snapshots"`
}
