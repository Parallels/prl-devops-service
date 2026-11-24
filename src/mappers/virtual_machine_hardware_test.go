package mappers

import (
	"encoding/json"
	"reflect"
	"testing"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func hardwareApiFixture() models.Hardware {
	return models.Hardware{
		CPU:    models.CPU{Cpus: 4},
		Memory: models.Memory{Size: "4096Mb"},
		USB:    models.Expiration{Enabled: true},
		Hdds: map[string]models.Hdd{
			"hdd0": {Enabled: true, Size: "40Gb", Type: "virtual", OnlineCompact: "auto"},
			"hdd1": {Enabled: true, Size: "20Gb", Type: "virtual", OnlineCompact: "auto"},
		},
		CdRoms: map[string]models.Cdrom{
			"cdrom0": {Enabled: true, Port: "sata:1:0", State: "disconnected"},
			"cdrom1": {Enabled: false, Image: "/isos/tool.iso"},
		},
		Nets: map[string]models.Net{
			"net0": {Enabled: true, Type: "shared", MAC: "00:1C:42:AA:BB:CC", Card: "e1000"},
			"net1": {Enabled: true, Type: "host-only", MAC: "00:1C:42:11:22:33", Card: "e1000"},
		},
		Sounds: map[string]models.Sound{
			"sound0": {Enabled: true, Output: "coreaudio", Mixer: "on"},
		},
		Fdds: map[string]models.Fdd{
			"fdd0": {Enabled: false, Port: "floppy", State: "disconnected"},
		},
		Extras: map[string]json.RawMessage{
			"com0": json.RawMessage(`{"enabled":true}`),
		},
	}
}

func TestMapDtoVirtualMachineHardwareRoundTripKeepsAllDevices(t *testing.T) {
	original := hardwareApiFixture()

	stored := MapDtoVirtualMachineHardwareFromApi(original)
	if len(stored.Hdds) != 2 || stored.Hdds["hdd1"].Size != "20Gb" {
		t.Fatalf("hdd1 lost on FromApi: %+v", stored.Hdds)
	}
	if len(stored.CdRoms) != 2 || stored.CdRoms["cdrom1"].Image != "/isos/tool.iso" {
		t.Fatalf("cdrom1 lost on FromApi: %+v", stored.CdRoms)
	}
	if len(stored.Nets) != 2 || stored.Nets["net1"].Type != "host-only" {
		t.Fatalf("net1 lost on FromApi: %+v", stored.Nets)
	}
	if len(stored.Sounds) != 1 || len(stored.Fdds) != 1 || stored.Fdds["fdd0"].Port != "floppy" {
		t.Fatalf("fdd0 lost on FromApi: %+v", stored.Fdds)
	}
	if raw, ok := stored.Extras["com0"]; !ok || string(raw) != `{"enabled":true}` {
		t.Fatalf("extras lost on FromApi: %v", stored.Extras)
	}

	restored := MapDtoVirtualMachineHardwareToApi(stored)
	if !reflect.DeepEqual(original, restored) {
		t.Fatalf("hardware did not survive API->data->API:\n original=%+v\nrestored=%+v", original, restored)
	}
}

// The byte-pinned golden fixtures serialise a zero VirtualMachineHardware;
// FromApi must not fabricate empty device maps for a zero hardware.
func TestMapDtoVirtualMachineHardwareZeroStaysNil(t *testing.T) {
	stored := MapDtoVirtualMachineHardwareFromApi(models.Hardware{})
	if stored.Hdds != nil || stored.CdRoms != nil || stored.Nets != nil || stored.Sounds != nil || stored.Fdds != nil || stored.Extras != nil {
		t.Fatalf("zero hardware must map to nil device maps, got %+v", stored)
	}

	restored := MapDtoVirtualMachineHardwareToApi(data_models.VirtualMachineHardware{})
	if restored.Hdds != nil || restored.CdRoms != nil || restored.Nets != nil || restored.Sounds != nil || restored.Fdds != nil || restored.Extras != nil {
		t.Fatalf("zero stored hardware must map back to nil device maps, got %+v", restored)
	}
}
