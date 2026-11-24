package parallelsdesktop

import (
	"testing"

	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func usageVm(state string, disks map[string]models.Hdd) models.ParallelsVM {
	vm := models.ParallelsVM{State: state}
	vm.Hardware.CPU.Cpus = 2
	vm.Hardware.Memory.Size = "4096Mb"
	vm.Hardware.Hdds = disks
	return vm
}

func megabytes(t *testing.T, size string) float64 {
	t.Helper()
	bytes, err := helpers.GetSizeByteFromString(size)
	if err != nil {
		t.Fatalf("bad size %q: %v", size, err)
	}
	return helpers.ConvertByteToMegabyte(bytes)
}

// The capacity regression: a VM whose second disk was added with
// --device-add hdd must charge its full disk capacity, not only hdd0.
func TestAccumulateVmHardwareUsageSumsAllDisks(t *testing.T) {
	result := &models.SystemUsageResponse{}

	vm := usageVm("running", map[string]models.Hdd{
		"hdd0": {Enabled: true, Size: "40Gb"},
		"hdd1": {Enabled: true, Size: "20Gb"},
	})
	if err := accumulateVmHardwareUsage(result, vm); err != nil {
		t.Fatalf("accumulate failed: %v", err)
	}

	wantDisk := megabytes(t, "40Gb") + megabytes(t, "20Gb")
	if result.TotalInUse == nil {
		t.Fatal("running VM must fill TotalInUse")
	}
	if result.TotalInUse.DiskSize != wantDisk {
		t.Fatalf("disk capacity undercounted: got %v, want %v", result.TotalInUse.DiskSize, wantDisk)
	}
	if result.TotalInUse.LogicalCpuCount != 2 {
		t.Fatalf("cpus: got %v, want 2", result.TotalInUse.LogicalCpuCount)
	}
	if result.TotalInUse.MemorySize != megabytes(t, "4096Mb") {
		t.Fatalf("memory: got %v, want %v", result.TotalInUse.MemorySize, megabytes(t, "4096Mb"))
	}
	if result.TotalReserved != nil {
		t.Fatal("running VM must not touch TotalReserved")
	}
}

func TestAccumulateVmHardwareUsageStoppedUsesReserved(t *testing.T) {
	result := &models.SystemUsageResponse{}

	vm := usageVm("stopped", map[string]models.Hdd{"hdd0": {Enabled: true, Size: "40Gb"}})
	if err := accumulateVmHardwareUsage(result, vm); err != nil {
		t.Fatalf("accumulate failed: %v", err)
	}
	if result.TotalReserved == nil || result.TotalReserved.DiskSize != megabytes(t, "40Gb") {
		t.Fatalf("stopped VM must charge TotalReserved, got %+v", result.TotalReserved)
	}
	if result.TotalInUse != nil {
		t.Fatal("stopped VM must not touch TotalInUse")
	}
}

func TestAccumulateVmHardwareUsageSkipsEmptySizes(t *testing.T) {
	result := &models.SystemUsageResponse{}

	vm := usageVm("running", map[string]models.Hdd{
		"hdd0": {Enabled: true, Size: "40Gb"},
		"hdd1": {Enabled: false, Size: ""},
	})
	if err := accumulateVmHardwareUsage(result, vm); err != nil {
		t.Fatalf("empty disk size must be skipped, not fatal: %v", err)
	}
	if result.TotalInUse.DiskSize != megabytes(t, "40Gb") {
		t.Fatalf("empty-size disk must contribute 0, got %v", result.TotalInUse.DiskSize)
	}
}

func TestAccumulateVmHardwareUsageInvalidDiskSize(t *testing.T) {
	result := &models.SystemUsageResponse{}

	vm := usageVm("running", map[string]models.Hdd{"hdd0": {Enabled: true, Size: "lots"}})
	if err := accumulateVmHardwareUsage(result, vm); err == nil {
		t.Fatal("invalid disk size must surface as an error")
	}
}
