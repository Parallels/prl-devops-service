package models

import "strings"

const MaxRunningMacVMs = 2

// IsMacVM uses VM metadata, never the user-editable display name. OS is a
// fallback for older inventories that do not report a VM type.
func IsMacVM(vm ParallelsVM) bool {
	return strings.EqualFold(vm.Type, "APPLE_VZ_VM") ||
		strings.EqualFold(vm.Type, "macvm") || strings.EqualFold(vm.OS, "macosx")
}

// OccupiesVmSlot reports whether a VM state still occupies a host slot. A slot
// becomes available only once the VM is stopped or suspended.
func OccupiesVmSlot(state string) bool {
	switch strings.ToLower(state) {
	case "running", "starting", "resuming", "paused", "pausing", "stopping", "suspending":
		return true
	}
	return false
}

// ActiveVMIDs returns the de-duplicated IDs of every virtual machine occupying a host
// slot, whatever its type. It is the basis for the "running" mode of the per-host VM
// limit as well as for the macOS VM slot limit.
func ActiveVMIDs(vms []ParallelsVM) []string {
	var ids []string
	seen := make(map[string]bool)
	for _, vm := range vms {
		if vm.ID == "" || seen[vm.ID] || !OccupiesVmSlot(vm.State) {
			continue
		}
		ids = append(ids, vm.ID)
		seen[vm.ID] = true
	}
	return ids
}

// ActiveMacVMIDs is ActiveVMIDs restricted to macOS VMs. It includes transitions that
// still occupy a host slot; a slot becomes available only once the VM is stopped or
// suspended.
func ActiveMacVMIDs(vms []ParallelsVM) []string {
	var macVms []ParallelsVM
	for _, vm := range vms {
		if IsMacVM(vm) {
			macVms = append(macVms, vm)
		}
	}
	return ActiveVMIDs(macVms)
}
