package models

// HostsVMSnapshotsRecord is the legacy wrapper (orchestrator_vm_snapshots
// file): a wire/aggregation carrier only — NEVER an AutoMigrate root. Its
// host_id + map key normalise into the vm_snapshots HostID/VMID carrier
// columns (VMSnapshot); VMSnapshot owns the table (mapping row 21).
type HostsVMSnapshotsRecord struct {
	HostId string `json:"host_id"`
	// map of vmId to list of snapshots
	VMSnapshots map[string][]VMSnapshot `json:"vm_snapshots"`
}
