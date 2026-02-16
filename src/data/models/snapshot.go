package models

// VirtualMachineSnapshot represents a snapshot record in the database
type VirtualMachineSnapshot struct {
	ID      string `json:"id"`               // Snapshot ID (with curly braces)
	VMId    string `json:"vm_id"`            // Virtual Machine ID this snapshot belongs to
	Name    string `json:"name,omitempty"`   // Snapshot name
	Date    string `json:"date"`             // Creation date/time
	State   string `json:"state"`            // VM state when snapshot was taken (poweroff, poweron, etc.)
	Current bool   `json:"current"`          // Whether this is the current snapshot
	Parent  string `json:"parent,omitempty"` // Parent snapshot ID (with curly braces), empty for root snapshot
}
