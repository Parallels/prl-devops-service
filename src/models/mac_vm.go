package models

import "strings"

const MaxRunningMacVMs = 2

// IsMacVM uses VM metadata, never the user-editable display name. OS is a
// fallback for older inventories that do not report a VM type.
func IsMacVM(vm ParallelsVM) bool {
	return strings.EqualFold(vm.Type, "APPLE_VZ_VM") ||
		strings.EqualFold(vm.Type, "macvm") || strings.EqualFold(vm.OS, "macosx")
}

// ActiveMacVMIDs includes transitions that still occupy a host slot. A slot
// becomes available only once the VM is stopped or suspended.
func ActiveMacVMIDs(vms []ParallelsVM) []string {
	var ids []string
	seen := make(map[string]bool)
	for _, vm := range vms {
		if !IsMacVM(vm) || vm.ID == "" || seen[vm.ID] {
			continue
		}
		switch strings.ToLower(vm.State) {
		case "running", "starting", "resuming", "paused", "pausing", "stopping", "suspending":
			ids = append(ids, vm.ID)
			seen[vm.ID] = true
		}
	}
	return ids
}
