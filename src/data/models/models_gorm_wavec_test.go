package models_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
	appmodels "github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/require"
)

// Wave-C golden fixtures (PBI #419 task 04): orchestrator host, snapshots
// (incl. wrapper shape), virtual machine (the wire-typo museum: capital "ID",
// "startup_and_Shutdown", "print _management" — all pinned byte-exact).

func waveCTargets() map[string]func() interface{} {
	vmFull := func() interface{} {
		return models.VirtualMachine{
			ID: "vm-101", Name: "prod-vm", HostId: "h-1", HostName: "host-a", User: "alice",
			HostUrl: "https://h:18443", HostState: "running", Type: "vm", State: "running",
			OS: "ubuntu", BootOrder: "hdd0.cdrom0", BIOSType: "efi", EFISecureBoot: "yes",
			ExternalBootDevice: "", HostSharedFolders: map[string]interface{}{"share1": "x"},
			HostDefinedSharing: "read_only", IpStatus: "up",
			GuestTools: models.VirtualMachineGuestTools{}, MouseAndKeyboard: models.VirtualMachineMouseAndKeyboard{},
			USBAndBluetooth: models.VirtualMachineUSBAndBluetooth{}, StartupAndShutdown: models.VirtualMachineStartupAndShutdown{},
			Optimization: models.VirtualMachineOptimization{}, TravelMode: models.VirtualMachineTravelMode{},
			Security: models.VirtualMachineSecurity{}, SmartGuard: models.VirtualMachineExpiration{},
			Modality: models.VirtualMachineModality{}, FullScreen: models.VirtualMachineFullscreen{},
			Coherence: models.VirtualMachineCoherence{}, TimeSynchronization: models.VirtualMachineTimeSynchronization{},
			Expiration: models.VirtualMachineExpiration{}, SMBIOSSettings: models.VirtualMachineSMBIOSSettings{},
			Hardware: models.VirtualMachineHardware{}, SharedProfile: models.VirtualMachineExpiration{},
			SharedApplications: models.VirtualMachineSharedApplications{}, SmartMount: models.VirtualMachineSmartMount{},
			MiscellaneousSharing: models.VirtualMachineMiscellaneousSharing{}, Advanced: models.VirtualMachineAdvanced{},
			PrintManagement: models.VirtualMachinePrintManagement{}, GuestSharedFolders: models.VirtualMachineGuestSharedFolders{},
			NetworkInformation: models.VirtualMachineNetworkInformation{},
			CreatedAt:          "2024-01-01 00:00:00", UpdatedAt: "2024-01-02 00:00:00",
		}
	}
	return map[string]func() interface{}{
		"orchestrator_host.json": func() interface{} {
			return models.OrchestratorHost{
				ID: "h-1", Enabled: true, Host: "10.0.0.9", Architecture: "arm64",
				CpuModel: "M3", OsVersion: "14.5", OsName: "macOS", ExternalIpAddress: "1.2.3.4",
				DevOpsVersion: "1.9.0", Description: "edge host", Tags: []string{"gpu", "edge"},
				Port: "18443", Schema: "https", PathPrefix: "", ParallelsDesktopVersion: "19.4",
				HasWebsocketEvents: true, Authentication: &models.OrchestratorHostAuthentication{},
				Resources: &models.HostResources{}, State: "online",
				LastUnhealthy: "2024-06-01 00:00:00", LastUnhealthyErrorMessage: "timeout",
				HealthCheck:           &appmodels.ApiHealthCheck{}, // json:"-" — golden asserts ABSENCE
				VirtualMachines:       []models.VirtualMachine{{ID: "vm-1", Name: "vm-a"}},
				IsReverseProxyEnabled: true, EnabledModules: []string{"vm", "packer"},
				ReverseProxy:      &models.ReverseProxy{ID: "rp-1", HostID: "h-1", Enabled: true},
				ReverseProxyHosts: []*models.ReverseProxyHost{{ID: "rph-1", HostID: "h-1", Host: "app.x", Port: "443"}},
				CacheConfig:       &appmodels.CatalogCacheConfig{}, CacheItems: []appmodels.HostCatalogCacheItem{},
				CreatedAt: "2024-01-01 00:00:00", UpdatedAt: "2024-01-02 00:00:00",
				RequiredClaims: []string{"host.read"}, RequiredRoles: []string{"host_admin"},
			}
		},
		"vm_snapshot.json": func() interface{} {
			return models.VMSnapshot{ID: "snap-1", Name: "pre-upgrade", Date: "2024-05-01 10:00:00",
				State: "valid", Current: true, Parent: "snap-0"}
		},
		"vm_snapshots_wrapper.json": func() interface{} {
			return models.VMSnapshots{VMId: "vm-101", VMSnapshot: []models.VMSnapshot{
				{ID: "snap-1", Name: "pre-upgrade", Date: "2024-05-01 10:00:00", State: "valid", Current: true},
				{Name: "auto", Date: "2024-05-02 10:00:00", State: "valid"},
			}}
		},
		"virtual_machine.json":         vmFull,
		"virtual_machine_minimal.json": func() interface{} { return models.VirtualMachine{ID: "vm-1", Name: "vm-a"} },
	}
}

// TestCaptureGoldenWaveC — env-gated capture (waves A/B protocol).
func TestCaptureGoldenWaveC(t *testing.T) {
	if os.Getenv("GOLDEN_CAPTURE") != "1" {
		t.Skip("set GOLDEN_CAPTURE=1 to (re)generate fixtures")
	}
	require.NoError(t, os.MkdirAll("testdata/golden", 0o755))
	for name, build := range waveCTargets() {
		require.NoError(t, os.WriteFile(filepath.Join("testdata", "golden", name), marshalForGolden(t, build()), 0o644))
	}
}

func TestOrchestratorHost_JSONGolden(t *testing.T) {
	assertGoldenB(t, "orchestrator_host.json", waveCTargets()["orchestrator_host.json"]())
}
func TestVMSnapshot_JSONGolden(t *testing.T) {
	assertGoldenB(t, "vm_snapshot.json", waveCTargets()["vm_snapshot.json"]())
}
func TestVMSnapshotsWrapper_JSONGolden(t *testing.T) {
	assertGoldenB(t, "vm_snapshots_wrapper.json", waveCTargets()["vm_snapshots_wrapper.json"]())
}
func TestVirtualMachine_JSONGolden(t *testing.T) {
	assertGoldenB(t, "virtual_machine.json", waveCTargets()["virtual_machine.json"]())
}
func TestVirtualMachineMinimal_JSONGolden(t *testing.T) {
	assertGoldenB(t, "virtual_machine_minimal.json", waveCTargets()["virtual_machine_minimal.json"]())
}

// TestVMSnapshot_CarriersLeakNothing — the normalisation carriers (HostID/
// VMID, added WITH the tags) must never reach the wire: with carriers set,
// marshal output must still equal the PRE-TAG golden byte-for-byte.
func TestVMSnapshot_CarriersLeakNothing(t *testing.T) {
	v := models.VMSnapshot{ID: "snap-1", Name: "pre-upgrade", Date: "2024-05-01 10:00:00",
		State: "valid", Current: true, Parent: "snap-0", HostID: "h-1", VMID: "vm-101"}
	assertGoldenB(t, "vm_snapshot.json", v)
}
