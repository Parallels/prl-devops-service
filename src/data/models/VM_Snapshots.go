package models

// VMSnapshot owns table vm_snapshots (default-name collision resolved in its
// favour per D-2). HostID/VMID are normalisation carriers (D-3): the legacy
// wrapper keys land in columns, wire JSON is unchanged (json:"-").
// rows(vm_snapshots) = Σ_hosts Σ_vm |VMSnapshots[vm_id]| (#424 formula);
// empty ID gets helpers.GenerateId at migration time.
type VMSnapshot struct {
	ID      string `json:"id,omitempty" gorm:"primaryKey"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	State   string `json:"state"`
	Current bool   `json:"current"`
	Parent  string `json:"parent"`
	HostID  string `json:"-" gorm:"column:host_id;index:idx_vmsnapshots_host_vm"` // from HostsVMSnapshotsRecord key
	VMID    string `json:"-" gorm:"column:vm_id;index:idx_vmsnapshots_host_vm"`   // from map key / wrapper VMId
}

// VMSnapshots is a wire/aggregation wrapper — never an AutoMigrate root
// (normalised into vm_snapshots rows; stays untagged).
type VMSnapshots struct {
	VMId       string       `json:"vm_id"`
	VMSnapshot []VMSnapshot `json:"vm_snapshots"`
}
