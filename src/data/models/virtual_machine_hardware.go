package models

import (
	"encoding/json"
	"regexp"
)

type VirtualMachineGuestTools struct {
	State   string `json:"state"`
	Version string `json:"version,omitempty"`
}

func (m *VirtualMachineGuestTools) Diff(source VirtualMachineGuestTools) bool {
	if m.State != source.State {
		return true
	}
	if m.Version != source.Version {
		return true
	}

	return false
}

// VirtualMachineHardware mirrors the prlctl hardware section into the
// virtual_machines.hardware serializer column. Devices are dynamic keyed
// entries (hdd0..n, cdrom0..n, net0..n, sound0..n, fdd0..n) held in typed
// maps keyed by device name, so --device-add results are persisted and
// surfaced instead of being dropped by fixed hdd0/cdrom0/net0/sound0 slots.
// Unknown keys (serial/parallel devices, future additions) survive verbatim
// in Extras. The custom JSON codec below keeps the stored/wire format flat
// and identical to prlctl output (old rows written with the fixed slots
// parse unchanged).
type VirtualMachineHardware struct {
	CPU         VirtualMachineCPU              `json:"cpu"`
	Memory      VirtualMachineMemory           `json:"memory"`
	Video       VirtualMachineVideo            `json:"video"`
	MemoryQuota VirtualMachineMemoryQuota      `json:"memory_quota"`
	USB         VirtualMachineExpiration       `json:"usb"`
	Hdds        map[string]VirtualMachineHdd   `json:"-"`
	CdRoms      map[string]VirtualMachineCdrom `json:"-"`
	Nets        map[string]VirtualMachineNet   `json:"-"`
	Sounds      map[string]VirtualMachineSound `json:"-"`
	Fdds        map[string]VirtualMachineFdd   `json:"-"`
	Extras      map[string]json.RawMessage     `json:"-"`
}

// hardwareDeviceKey matches the dynamic prlctl device keys (hdd0, cdrom1,
// net0, sound0, fdd0, ...).
var hardwareDeviceKey = regexp.MustCompile(`^(hdd|cdrom|net|sound|fdd)[0-9]+$`)

// UnmarshalJSON decodes the flat prlctl hardware object: the known scalar
// groups (cpu, memory, video, memory_quota, usb) fill their typed fields,
// every device key lands in the matching typed map keyed by the device name,
// and anything else is preserved as raw JSON in Extras.
func (m *VirtualMachineHardware) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*m = VirtualMachineHardware{}
	for key, value := range raw {
		switch key {
		case "cpu":
			if err := json.Unmarshal(value, &m.CPU); err != nil {
				return err
			}
		case "memory":
			if err := json.Unmarshal(value, &m.Memory); err != nil {
				return err
			}
		case "video":
			if err := json.Unmarshal(value, &m.Video); err != nil {
				return err
			}
		case "memory_quota":
			if err := json.Unmarshal(value, &m.MemoryQuota); err != nil {
				return err
			}
		case "usb":
			if err := json.Unmarshal(value, &m.USB); err != nil {
				return err
			}
		default:
			if match := hardwareDeviceKey.FindStringSubmatch(key); match != nil {
				switch match[1] {
				case "hdd":
					var device VirtualMachineHdd
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if m.Hdds == nil {
						m.Hdds = map[string]VirtualMachineHdd{}
					}
					m.Hdds[key] = device
				case "cdrom":
					var device VirtualMachineCdrom
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if m.CdRoms == nil {
						m.CdRoms = map[string]VirtualMachineCdrom{}
					}
					m.CdRoms[key] = device
				case "net":
					var device VirtualMachineNet
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if m.Nets == nil {
						m.Nets = map[string]VirtualMachineNet{}
					}
					m.Nets[key] = device
				case "sound":
					var device VirtualMachineSound
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if m.Sounds == nil {
						m.Sounds = map[string]VirtualMachineSound{}
					}
					m.Sounds[key] = device
				case "fdd":
					var device VirtualMachineFdd
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if m.Fdds == nil {
						m.Fdds = map[string]VirtualMachineFdd{}
					}
					m.Fdds[key] = device
				}
				continue
			}
			if m.Extras == nil {
				m.Extras = map[string]json.RawMessage{}
			}
			m.Extras[key] = append(json.RawMessage(nil), value...)
		}
	}

	return nil
}

// MarshalJSON re-emits the flat prlctl-shaped hardware object from the typed
// device maps, including the unknown keys captured in Extras. encoding/json
// sorts map keys, so the output is deterministic (golden tests depend on it).
func (m VirtualMachineHardware) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"cpu":          m.CPU,
		"memory":       m.Memory,
		"video":        m.Video,
		"memory_quota": m.MemoryQuota,
		"usb":          m.USB,
	}
	for key, device := range m.Hdds {
		out[key] = device
	}
	for key, device := range m.CdRoms {
		out[key] = device
	}
	for key, device := range m.Nets {
		out[key] = device
	}
	for key, device := range m.Sounds {
		out[key] = device
	}
	for key, device := range m.Fdds {
		out[key] = device
	}
	for key, raw := range m.Extras {
		if _, taken := out[key]; !taken {
			out[key] = raw
		}
	}

	return json.Marshal(out)
}

func (m *VirtualMachineHardware) Diff(source VirtualMachineHardware) bool {
	if m.CPU.Diff(source.CPU) {
		return true
	}
	if m.Memory.Diff(source.Memory) {
		return true
	}
	if m.Video.Diff(source.Video) {
		return true
	}
	if m.MemoryQuota.Diff(source.MemoryQuota) {
		return true
	}
	if m.USB.Diff(source.USB) {
		return true
	}
	if len(m.Hdds) != len(source.Hdds) {
		return true
	}
	for key, device := range m.Hdds {
		sourceDevice, ok := source.Hdds[key]
		if !ok || device.Diff(sourceDevice) {
			return true
		}
	}
	if len(m.CdRoms) != len(source.CdRoms) {
		return true
	}
	for key, device := range m.CdRoms {
		sourceDevice, ok := source.CdRoms[key]
		if !ok || device.Diff(sourceDevice) {
			return true
		}
	}
	if len(m.Nets) != len(source.Nets) {
		return true
	}
	for key, device := range m.Nets {
		sourceDevice, ok := source.Nets[key]
		if !ok || device.Diff(sourceDevice) {
			return true
		}
	}
	if len(m.Sounds) != len(source.Sounds) {
		return true
	}
	for key, device := range m.Sounds {
		sourceDevice, ok := source.Sounds[key]
		if !ok || device.Diff(sourceDevice) {
			return true
		}
	}
	if len(m.Fdds) != len(source.Fdds) {
		return true
	}
	for key, device := range m.Fdds {
		sourceDevice, ok := source.Fdds[key]
		if !ok || device.Diff(sourceDevice) {
			return true
		}
	}
	if len(m.Extras) != len(source.Extras) {
		return true
	}
	for key, raw := range m.Extras {
		sourceRaw, ok := source.Extras[key]
		if !ok || string(raw) != string(sourceRaw) {
			return true
		}
	}

	return false
}

type VirtualMachineCPU struct {
	Cpus    int64  `json:"cpus"`
	Auto    string `json:"auto"`
	VTX     bool   `json:"VT-x"`
	Hotplug bool   `json:"hotplug"`
	Accl    string `json:"accl"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
}

func (m *VirtualMachineCPU) Diff(source VirtualMachineCPU) bool {
	if m.Cpus != source.Cpus {
		return true
	}
	if m.Auto != source.Auto {
		return true
	}
	if m.VTX != source.VTX {
		return true
	}
	if m.Hotplug != source.Hotplug {
		return true
	}
	if m.Accl != source.Accl {
		return true
	}
	if m.Mode != source.Mode {
		return true
	}
	if m.Type != source.Type {
		return true
	}

	return false
}

type VirtualMachineCdrom struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Image   string `json:"image"`
	State   string `json:"state,omitempty"`
}

func (m *VirtualMachineCdrom) Diff(source VirtualMachineCdrom) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.Port != source.Port {
		return true
	}
	if m.Image != source.Image {
		return true
	}
	if m.State != source.State {
		return true
	}

	return false
}

type VirtualMachineHdd struct {
	Enabled       bool   `json:"enabled"`
	Port          string `json:"port"`
	Image         string `json:"image"`
	Type          string `json:"type"`
	Size          string `json:"size"`
	OnlineCompact string `json:"online-compact"`
}

func (m *VirtualMachineHdd) Diff(source VirtualMachineHdd) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.Port != source.Port {
		return true
	}
	if m.Image != source.Image {
		return true
	}
	if m.Type != source.Type {
		return true
	}
	if m.Size != source.Size {
		return true
	}
	if m.OnlineCompact != source.OnlineCompact {
		return true
	}

	return false
}

type VirtualMachineFdd struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Image   string `json:"image"`
	State   string `json:"state,omitempty"`
}

func (m *VirtualMachineFdd) Diff(source VirtualMachineFdd) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.Port != source.Port {
		return true
	}
	if m.Image != source.Image {
		return true
	}
	if m.State != source.State {
		return true
	}

	return false
}

type VirtualMachineMemory struct {
	Size    string `json:"size"`
	Auto    string `json:"auto"`
	Hotplug bool   `json:"hotplug"`
}

func (m *VirtualMachineMemory) Diff(source VirtualMachineMemory) bool {
	if m.Size != source.Size {
		return true
	}
	if m.Auto != source.Auto {
		return true
	}
	if m.Hotplug != source.Hotplug {
		return true
	}

	return false
}

type VirtualMachineMemoryQuota struct {
	Auto string `json:"auto"`
}

func (m *VirtualMachineMemoryQuota) Diff(source VirtualMachineMemoryQuota) bool {
	return m.Auto != source.Auto
}

type VirtualMachineNet struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	MAC     string `json:"mac"`
	Card    string `json:"card"`
}

func (m *VirtualMachineNet) Diff(source VirtualMachineNet) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.Type != source.Type {
		return true
	}
	if m.MAC != source.MAC {
		return true
	}
	if m.Card != source.Card {
		return true
	}

	return false
}

type VirtualMachineSound struct {
	Enabled bool   `json:"enabled"`
	Output  string `json:"output"`
	Mixer   string `json:"mixer"`
}

func (m *VirtualMachineSound) Diff(source VirtualMachineSound) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.Output != source.Output {
		return true
	}
	if m.Mixer != source.Mixer {
		return true
	}

	return false
}

type VirtualMachineVideo struct {
	AdapterType           string `json:"adapter-type"`
	Size                  string `json:"size"`
	The3DAcceleration     string `json:"3d-acceleration"`
	VerticalSync          string `json:"vertical-sync"`
	HighResolution        string `json:"high-resolution"`
	HighResolutionInGuest string `json:"high-resolution-in-guest"`
	NativeScalingInGuest  string `json:"native-scaling-in-guest"`
	AutomaticVideoMemory  string `json:"automatic-video-memory"`
}

func (m *VirtualMachineVideo) Diff(source VirtualMachineVideo) bool {
	if m.AdapterType != source.AdapterType {
		return true
	}
	if m.Size != source.Size {
		return true
	}
	if m.The3DAcceleration != source.The3DAcceleration {
		return true
	}
	if m.VerticalSync != source.VerticalSync {
		return true
	}
	if m.HighResolution != source.HighResolution {
		return true
	}
	if m.HighResolutionInGuest != source.HighResolutionInGuest {
		return true
	}
	if m.NativeScalingInGuest != source.NativeScalingInGuest {
		return true
	}
	if m.AutomaticVideoMemory != source.AutomaticVideoMemory {
		return true
	}

	return false
}
