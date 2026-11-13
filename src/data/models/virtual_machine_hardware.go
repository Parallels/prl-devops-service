package models

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

type VirtualMachineHardware struct {
	CPU         VirtualMachineCPU         `json:"cpu"`
	Memory      VirtualMachineMemory      `json:"memory"`
	Video       VirtualMachineVideo       `json:"video"`
	MemoryQuota VirtualMachineMemoryQuota `json:"memory_quota"`
	Hdd0        VirtualMachineHdd0        `json:"hdd0"`
	Cdrom0      VirtualMachineCdrom0      `json:"cdrom0"`
	USB         VirtualMachineExpiration  `json:"usb"`
	Net0        VirtualMachineNet0        `json:"net0"`
	Sound0      VirtualMachineSound0      `json:"sound0"`
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
	if m.Hdd0.Diff(source.Hdd0) {
		return true
	}
	if m.Cdrom0.Diff(source.Cdrom0) {
		return true
	}
	if m.USB.Diff(source.USB) {
		return true
	}
	if m.Net0.Diff(source.Net0) {
		return true
	}
	if m.Sound0.Diff(source.Sound0) {
		return true
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

type VirtualMachineCdrom0 struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Image   string `json:"image"`
	State   string `json:"state,omitempty"`
}

func (m *VirtualMachineCdrom0) Diff(source VirtualMachineCdrom0) bool {
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

type VirtualMachineHdd0 struct {
	Enabled       bool   `json:"enabled"`
	Port          string `json:"port"`
	Image         string `json:"image"`
	Type          string `json:"type"`
	Size          string `json:"size"`
	OnlineCompact string `json:"online-compact"`
}

func (m *VirtualMachineHdd0) Diff(source VirtualMachineHdd0) bool {
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

type VirtualMachineNet0 struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	MAC     string `json:"mac"`
	Card    string `json:"card"`
}

func (m *VirtualMachineNet0) Diff(source VirtualMachineNet0) bool {
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

type VirtualMachineSound0 struct {
	Enabled bool   `json:"enabled"`
	Output  string `json:"output"`
	Mixer   string `json:"mixer"`
}

func (m *VirtualMachineSound0) Diff(source VirtualMachineSound0) bool {
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
