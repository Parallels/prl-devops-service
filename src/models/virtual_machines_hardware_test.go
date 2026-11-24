package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// realisticHardwareJSON mirrors prlctl list --json hardware output once a
// second disk, a second NIC and a floppy have been added with --device-add.
const realisticHardwareJSON = `{
	"cpu": {"cpus": 4, "auto": "no", "VT-x": true, "hotplug": false, "accl": "default", "mode": "host", "type": "virtualization"},
	"memory": {"size": "4096Mb", "auto": "no", "hotplug": false},
	"video": {"adapter-type": "vvga2", "size": "128Mb"},
	"memory_quota": {"auto": "no"},
	"hdd0": {"enabled": true, "port": "sata:0:0", "image": "/vms/ubuntu-0.hdd", "type": "virtual", "size": "40 Gb", "online-compact": "auto"},
	"hdd1": {"enabled": true, "port": "sata:0:1", "image": "/vms/extra-1.hdd", "type": "virtual", "size": "20 Gb", "online-compact": "auto"},
	"cdrom0": {"enabled": true, "port": "sata:1:0", "image": "", "state": "disconnected"},
	"net0": {"enabled": true, "type": "shared", "mac": "00:1C:42:AA:BB:CC", "card": "e1000"},
	"net1": {"enabled": true, "type": "host-only", "mac": "00:1C:42:11:22:33", "card": "e1000"},
	"sound0": {"enabled": false, "output": "coreaudio", "mixer": "on"},
	"fdd0": {"enabled": false, "port": "floppy", "image": "", "state": "disconnected"},
	"usb": {"enabled": true},
	"com0": {"enabled": true, "port": "1"}
}`

func TestHardwareUnmarshalKeepsEveryDevice(t *testing.T) {
	var hardware Hardware
	if err := json.Unmarshal([]byte(realisticHardwareJSON), &hardware); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(hardware.Hdds) != 2 {
		t.Fatalf("expected 2 disks (hdd0+hdd1), got %d", len(hardware.Hdds))
	}
	if hardware.Hdds["hdd1"].Size != "20 Gb" {
		t.Fatalf("newly added disk hdd1 lost its size: %+v", hardware.Hdds["hdd1"])
	}
	if hardware.Hdds["hdd0"].Size != "40 Gb" {
		t.Fatalf("primary disk hdd0 lost its size: %+v", hardware.Hdds["hdd0"])
	}
	if len(hardware.CdRoms) != 1 || !hardware.CdRoms["cdrom0"].Enabled {
		t.Fatalf("cdrom0 missing: %+v", hardware.CdRoms)
	}
	if len(hardware.Nets) != 2 || hardware.Nets["net1"].Type != "host-only" {
		t.Fatalf("net1 missing: %+v", hardware.Nets)
	}
	if len(hardware.Sounds) != 1 {
		t.Fatalf("sound0 missing: %+v", hardware.Sounds)
	}
	if len(hardware.Fdds) != 1 || hardware.Fdds["fdd0"].Port != "floppy" {
		t.Fatalf("fdd0 missing: %+v", hardware.Fdds)
	}
	if !hardware.USB.Enabled {
		t.Fatal("usb not decoded")
	}
	if hardware.CPU.Cpus != 4 || hardware.Memory.Size != "4096Mb" {
		t.Fatalf("scalar groups not decoded: %+v %+v", hardware.CPU, hardware.Memory)
	}
	rawCom0, ok := hardware.Extras["com0"]
	if !ok {
		t.Fatalf("unknown device com0 was dropped; extras=%v", hardware.Extras)
	}
	if !strings.Contains(string(rawCom0), "\"port\"") {
		t.Fatalf("com0 payload not preserved verbatim: %s", rawCom0)
	}
}

func TestHardwareMarshalKeepsFlatPrlctlKeys(t *testing.T) {
	var hardware Hardware
	if err := json.Unmarshal([]byte(realisticHardwareJSON), &hardware); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	out, err := json.Marshal(hardware)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	for _, key := range []string{`"hdd0"`, `"hdd1"`, `"cdrom0"`, `"net0"`, `"net1"`, `"sound0"`, `"fdd0"`, `"com0"`, `"cpu"`, `"memory"`, `"memory_quota"`, `"usb"`} {
		if !strings.Contains(string(out), key) {
			t.Fatalf("flat wire key %s missing from marshalled hardware: %s", key, out)
		}
	}

	// encoding/json sorts map keys, so the output must be byte-stable —
	// the golden JSON fixtures depend on it.
	again, err := json.Marshal(hardware)
	if err != nil {
		t.Fatalf("second marshal failed: %v", err)
	}
	if string(out) != string(again) {
		t.Fatalf("marshal not deterministic:\n%s\n%s", out, again)
	}
}

func TestHardwareRoundTripPreservesDevicesAndExtras(t *testing.T) {
	var first Hardware
	if err := json.Unmarshal([]byte(realisticHardwareJSON), &first); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	out, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var second Hardware
	if err := json.Unmarshal(out, &second); err != nil {
		t.Fatalf("re-unmarshal failed: %v", err)
	}

	// json.Marshal compacts the whole document, so raw Extras bytes are only
	// byte-stable once compacted — normalise the first decode the same way
	// and require a lossless structural comparison.
	firstCopy := first
	if len(firstCopy.Extras) > 0 {
		firstCopy.Extras = map[string]json.RawMessage{}
		for key, raw := range first.Extras {
			var buf bytes.Buffer
			if err := json.Compact(&buf, raw); err != nil {
				t.Fatalf("compact %s failed: %v", key, err)
			}
			firstCopy.Extras[key] = buf.Bytes()
		}
	}
	if !reflect.DeepEqual(firstCopy, second) {
		t.Fatalf("round trip lost devices:\n first=%+v\nsecond=%+v", firstCopy, second)
	}
}

// Rows written before this change stored the fixed-slot flat JSON (hdd0 etc.)
// — they must still parse into the typed maps.
func TestHardwareUnmarshalLegacySingleSlotRow(t *testing.T) {
	legacy := `{"cpu":{"cpus":2},"memory":{"size":"2048Mb"},"video":{},"memory_quota":{},"hdd0":{"enabled":true,"size":"10Gb"},"cdrom0":{"enabled":false,"port":"sata:1:0"},"usb":{"enabled":false},"net0":{"enabled":true,"mac":"00:1C:42:00:00:01"},"sound0":{"enabled":true}}`

	var hardware Hardware
	if err := json.Unmarshal([]byte(legacy), &hardware); err != nil {
		t.Fatalf("legacy row must still parse: %v", err)
	}
	if len(hardware.Hdds) != 1 || hardware.Hdds["hdd0"].Size != "10Gb" {
		t.Fatalf("hdd0 not decoded from legacy row: %+v", hardware.Hdds)
	}
	if len(hardware.Nets) != 1 || len(hardware.CdRoms) != 1 || len(hardware.Sounds) != 1 || len(hardware.Fdds) != 0 {
		t.Fatalf("legacy device buckets wrong: %+v", hardware)
	}
	if len(hardware.Extras) != 0 {
		t.Fatalf("unexpected extras: %v", hardware.Extras)
	}
}

func TestHardwareUnmarshalInvalidJSON(t *testing.T) {
	var hardware Hardware
	if err := json.Unmarshal([]byte(`{"hdd0":`), &hardware); err == nil {
		t.Fatal("expected error for truncated hardware JSON")
	}
}
