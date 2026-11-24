package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func hardwareFixture() VirtualMachineHardware {
	return VirtualMachineHardware{
		CPU:    VirtualMachineCPU{Cpus: 2},
		Memory: VirtualMachineMemory{Size: "4096Mb"},
		Hdds: map[string]VirtualMachineHdd{
			"hdd0": {Enabled: true, Size: "40Gb", Type: "virtual"},
			"hdd1": {Enabled: true, Size: "20Gb", Type: "virtual"},
		},
		CdRoms: map[string]VirtualMachineCdrom{
			"cdrom0": {Enabled: true, Port: "sata:1:0"},
		},
		Nets: map[string]VirtualMachineNet{
			"net0": {Enabled: true, MAC: "00:1C:42:AA:BB:CC"},
			"net1": {Enabled: true, MAC: "00:1C:42:11:22:33"},
		},
		Sounds: map[string]VirtualMachineSound{
			"sound0": {Enabled: true, Output: "coreaudio"},
		},
		Fdds: map[string]VirtualMachineFdd{
			"fdd0": {Enabled: false, Port: "floppy"},
		},
		Extras: map[string]json.RawMessage{
			"com0": json.RawMessage(`{"enabled":true}`),
		},
	}
}

func TestVirtualMachineHardwareDiffDetectsDeviceChanges(t *testing.T) {
	base := hardwareFixture()

	if base.Diff(base) {
		t.Fatal("identical hardware must not diff")
	}

	addedDisk := hardwareFixture()
	addedDisk.Hdds["hdd2"] = VirtualMachineHdd{Enabled: true, Size: "10Gb"}
	if !base.Diff(addedDisk) {
		t.Fatal("adding hdd2 (--device-add hdd) must be reported as a change")
	}

	removedDisk := hardwareFixture()
	delete(removedDisk.Hdds, "hdd1")
	if !base.Diff(removedDisk) {
		t.Fatal("removing hdd1 must be reported as a change")
	}

	grownDisk := hardwareFixture()
	disk := grownDisk.Hdds["hdd1"]
	disk.Size = "30Gb"
	grownDisk.Hdds["hdd1"] = disk
	if !base.Diff(grownDisk) {
		t.Fatal("resizing hdd1 must be reported as a change")
	}

	addedFloppy := hardwareFixture()
	delete(addedFloppy.Fdds, "fdd0")
	if !base.Diff(addedFloppy) {
		t.Fatal("removing fdd0 must be reported as a change")
	}

	addedNet := hardwareFixture()
	delete(addedNet.Nets, "net1")
	if !base.Diff(addedNet) {
		t.Fatal("removing net1 must be reported as a change")
	}

	changedExtras := hardwareFixture()
	changedExtras.Extras["com0"] = json.RawMessage(`{"enabled":false}`)
	if !base.Diff(changedExtras) {
		t.Fatal("changed unknown device payload must be reported as a change")
	}

	nilExtras := hardwareFixture()
	nilExtras.Extras = nil
	emptyExtras := hardwareFixture()
	emptyExtras.Extras = map[string]json.RawMessage{}
	if nilExtras.Diff(emptyExtras) || emptyExtras.Diff(nilExtras) {
		t.Fatal("nil and empty extras must be considered equal")
	}
}

func TestVirtualMachineHardwareCodecFlatKeys(t *testing.T) {
	base := hardwareFixture()

	out, err := json.Marshal(base)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	// The serializer column keeps the flat prlctl-style keys.
	for _, key := range []string{`"hdd0"`, `"hdd1"`, `"cdrom0"`, `"net0"`, `"net1"`, `"sound0"`, `"fdd0"`, `"com0"`} {
		if !containsJSONKey(t, out, key) {
			t.Fatalf("flat key %s missing from stored hardware: %s", key, out)
		}
	}

	var again VirtualMachineHardware
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("re-unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(base, again) {
		t.Fatalf("round trip lost devices:\n base=%+v\nagain=%+v", base, again)
	}
}

// Rows written by the previous model stored hdd0/cdrom0/net0/sound0 as fixed
// fields — same flat keys, so the new codec must read them unchanged.
func TestVirtualMachineHardwareUnmarshalLegacyRow(t *testing.T) {
	legacy := `{"cpu":{"cpus":1},"memory":{"size":"1024Mb"},"video":{},"memory_quota":{},"hdd0":{"enabled":true,"port":"sata:0:0","image":"/vms/base.hdd","type":"expand","size":"20Gb","online-compact":"auto"},"cdrom0":{"enabled":false,"port":"sata:1:0","image":""},"usb":{"enabled":false},"net0":{"enabled":true,"type":"shared","mac":"00:1C:42:00:00:01","card":"e1000"},"sound0":{"enabled":false,"output":"coreaudio","mixer":"on"}}`

	var hardware VirtualMachineHardware
	if err := json.Unmarshal([]byte(legacy), &hardware); err != nil {
		t.Fatalf("legacy row must still parse: %v", err)
	}
	if hardware.Hdds["hdd0"].Size != "20Gb" || hardware.Hdds["hdd0"].OnlineCompact != "auto" {
		t.Fatalf("hdd0 not decoded from legacy row: %+v", hardware.Hdds)
	}
	if len(hardware.Nets) != 1 || hardware.Nets["net0"].MAC != "00:1C:42:00:00:01" {
		t.Fatalf("net0 not decoded from legacy row: %+v", hardware.Nets)
	}
	if len(hardware.CdRoms) != 1 || len(hardware.Sounds) != 1 || len(hardware.Fdds) != 0 || len(hardware.Extras) != 0 {
		t.Fatalf("legacy buckets wrong: %+v", hardware)
	}
}

func TestVirtualMachineHardwareZeroMarshalsWithoutDeviceKeys(t *testing.T) {
	out, err := json.Marshal(VirtualMachineHardware{})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	for _, key := range []string{"hdd0", "cdrom0", "net0", "sound0", "fdd0"} {
		if _, ok := decoded[key]; ok {
			t.Fatalf("zero hardware must not emit %s: %s", key, out)
		}
	}
}

func containsJSONKey(t *testing.T, data []byte, key string) bool {
	t.Helper()
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	trimmed := key[1 : len(key)-1]
	_, ok := decoded[trimmed]
	return ok
}
