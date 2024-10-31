package models

import "encoding/json"

type VirtualMachine struct {
	ID                    string                             `json:"ID,omitempty"`
	HostId                string                             `json:"host_id,omitempty"`
	HostState             string                             `json:"host_state,omitempty"`
	User                  string                             `json:"user,omitempty"`
	HostExternalIpAddress string                             `json:"host_external_ip_address,omitempty"`
	InternalIpAddress     string                             `json:"internal_ip_address,omitempty"`
	Host                  string                             `json:"host,omitempty"`
	Name                  string                             `json:"name,omitempty"`
	Description           string                             `json:"description,omitempty"`
	Type                  string                             `json:"type,omitempty"`
	State                 string                             `json:"state,omitempty"`
	OS                    string                             `json:"os,omitempty"`
	Template              string                             `json:"template,omitempty"`
	Uptime                string                             `json:"uptime,omitempty"`
	HomePath              string                             `json:"home_path,omitempty"`
	Home                  string                             `json:"home,omitempty"`
	RestoreImage          string                             `json:"restore_image,omitempty"`
	GuestTools            VirtualMachineGuestTools           `json:"guest_tools,omitempty"`
	MouseAndKeyboard      VirtualMachineMouseAndKeyboard     `json:"mouse_and_keyboard,omitempty"`
	USBAndBluetooth       VirtualMachineUSBAndBluetooth      `json:"usb_and_bluetooth,omitempty"`
	StartupAndShutdown    VirtualMachineStartupAndShutdown   `json:"startup_and_Shutdown,omitempty"`
	Optimization          VirtualMachineOptimization         `json:"optimization,omitempty"`
	TravelMode            VirtualMachineTravelMode           `json:"travel_mode,omitempty"`
	Security              VirtualMachineSecurity             `json:"security,omitempty"`
	SmartGuard            VirtualMachineExpiration           `json:"smart_guard,omitempty"`
	Modality              VirtualMachineModality             `json:"modality,omitempty"`
	FullScreen            VirtualMachineFullscreen           `json:"full_screen,omitempty"`
	Coherence             VirtualMachineCoherence            `json:"coherence,omitempty"`
	TimeSynchronization   VirtualMachineTimeSynchronization  `json:"time_synchronization,omitempty"`
	Expiration            VirtualMachineExpiration           `json:"expiration,omitempty"`
	BootOrder             string                             `json:"boot_order,omitempty"`
	BIOSType              string                             `json:"bios_type,omitempty"`
	EFISecureBoot         string                             `json:"efi_secure_boot,omitempty"`
	AllowSelectBootDevice string                             `json:"allow_select_boot_device,omitempty"`
	ExternalBootDevice    string                             `json:"external_boot_device,omitempty"`
	SMBIOSSettings        VirtualMachineSMBIOSSettings       `json:"smbios_settings,omitempty"`
	Hardware              VirtualMachineHardware             `json:"hardware,omitempty"`
	HostSharedFolders     map[string]interface{}             `json:"host_shared_folders,omitempty"`
	HostDefinedSharing    string                             `json:"host_defined_sharing,omitempty"`
	SharedProfile         VirtualMachineExpiration           `json:"shared_profile,omitempty"`
	SharedApplications    VirtualMachineSharedApplications   `json:"shared_applications,omitempty"`
	SmartMount            VirtualMachineSmartMount           `json:"smart_mount,omitempty"`
	MiscellaneousSharing  VirtualMachineMiscellaneousSharing `json:"miscellaneous_sharing,omitempty"`
	Advanced              VirtualMachineAdvanced             `json:"advanced,omitempty"`
	PrintManagement       VirtualMachinePrintManagement      `json:"print _management,omitempty"`
	GuestSharedFolders    VirtualMachineGuestSharedFolders   `json:"guest_shared_folders,omitempty"`
	NetworkInformation    VirtualMachineNetworkInformation   `json:"network_information,omitempty"`
	CreatedAt             string                             `json:"created_at,omitempty"`
	UpdatedAt             string                             `json:"updated_at,omitempty"`
}

func (m *VirtualMachine) Diff(source VirtualMachine) bool {
	if m.ID != source.ID {
		return true
	}
	if m.HostId != source.HostId {
		return true
	}
	if m.HostState != source.HostState {
		return true
	}
	if m.HostExternalIpAddress != source.HostExternalIpAddress {
		return true
	}
	if m.InternalIpAddress != source.InternalIpAddress {
		return true
	}
	if m.User != source.User {
		return true
	}
	if m.Host != source.Host {
		return true
	}
	if m.Name != source.Name {
		return true
	}
	if m.Description != source.Description {
		return true
	}
	if m.Type != source.Type {
		return true
	}
	if m.State != source.State {
		return true
	}
	if m.OS != source.OS {
		return true
	}
	if m.Template != source.Template {
		return true
	}

	if m.HomePath != source.HomePath {
		return true
	}
	if m.Home != source.Home {
		return true
	}
	if m.RestoreImage != source.RestoreImage {
		return true
	}
	if m.GuestTools.Diff(source.GuestTools) {
		return true
	}
	if m.MouseAndKeyboard.Diff(source.MouseAndKeyboard) {
		return true
	}
	if m.USBAndBluetooth.Diff(source.USBAndBluetooth) {
		return true
	}
	if m.StartupAndShutdown.Diff(source.StartupAndShutdown) {
		return true
	}
	if m.Optimization.Diff(source.Optimization) {
		return true
	}
	if m.TravelMode.Diff(source.TravelMode) {
		return true
	}
	if m.Security.Diff(source.Security) {
		return true
	}
	if m.SmartGuard.Diff(source.SmartGuard) {
		return true
	}
	if m.Modality.Diff(source.Modality) {
		return true
	}
	if m.FullScreen.Diff(source.FullScreen) {
		return true
	}
	if m.Coherence.Diff(source.Coherence) {
		return true
	}
	if m.TimeSynchronization.Diff(source.TimeSynchronization) {
		return true
	}
	if m.Expiration.Diff(source.Expiration) {
		return true
	}
	if m.BootOrder != source.BootOrder {
		return true
	}
	if m.BIOSType != source.BIOSType {
		return true
	}
	if m.EFISecureBoot != source.EFISecureBoot {
		return true
	}
	if m.AllowSelectBootDevice != source.AllowSelectBootDevice {
		return true
	}
	if m.ExternalBootDevice != source.ExternalBootDevice {
		return true
	}
	if m.SMBIOSSettings.Diff(source.SMBIOSSettings) {
		return true
	}
	if m.Hardware.Diff(source.Hardware) {
		return true
	}

	for k, v := range m.HostSharedFolders {
		val, err := json.Marshal(v)
		if err != nil {
			return true
		}
		sourceVal := source.HostSharedFolders[k]
		sourceValBytes, err := json.Marshal(sourceVal)
		if err != nil {
			return true
		}

		if string(val) != string(sourceValBytes) {
			return true
		}
	}

	if m.HostDefinedSharing != source.HostDefinedSharing {
		return true
	}
	if m.SharedProfile.Diff(source.SharedProfile) {
		return true
	}
	if m.SharedApplications.Diff(source.SharedApplications) {
		return true
	}
	if m.SmartMount.Diff(source.SmartMount) {
		return true
	}
	if m.MiscellaneousSharing.Diff(source.MiscellaneousSharing) {
		return true
	}
	if m.Advanced.Diff(source.Advanced) {
		return true
	}
	if m.PrintManagement.Diff(source.PrintManagement) {
		return true
	}
	if m.GuestSharedFolders.Diff(source.GuestSharedFolders) {
		return true
	}
	if m.NetworkInformation.Diff(source.NetworkInformation) {
		return true
	}

	return false
}

type VirtualMachineAdvanced struct {
	VMHostnameSynchronization    string `json:"vm_hostname_synchronization"`
	PublicSSHKeysSynchronization string `json:"public_SSH_keys_synchronization"`
	ShowDeveloperTools           string `json:"show_developer_tools"`
	SwipeFromEdges               string `json:"swipe_from_edges"`
	ShareHostLocation            string `json:"share_host_location"`
	RosettaLinux                 string `json:"rosetta_linux"`
}

func (m *VirtualMachineAdvanced) Diff(source VirtualMachineAdvanced) bool {
	if m.VMHostnameSynchronization != source.VMHostnameSynchronization {
		return true
	}
	if m.PublicSSHKeysSynchronization != source.PublicSSHKeysSynchronization {
		return true
	}
	if m.ShowDeveloperTools != source.ShowDeveloperTools {
		return true
	}
	if m.SwipeFromEdges != source.SwipeFromEdges {
		return true
	}
	if m.ShareHostLocation != source.ShareHostLocation {
		return true
	}
	if m.RosettaLinux != source.RosettaLinux {
		return true
	}

	return false
}

type VirtualMachineCoherence struct {
	ShowWindowsSystrayInMACMenu string `json:"show_windows_systray_in_mac_menu"`
	AutoSwitchToFullScreen      string `json:"auto-switch_to_full_screen"`
	DisableAero                 string `json:"disable_aero"`
	HideMinimizedWindows        string `json:"hide_minimized_windows"`
}

func (m *VirtualMachineCoherence) Diff(source VirtualMachineCoherence) bool {
	if m.ShowWindowsSystrayInMACMenu != source.ShowWindowsSystrayInMACMenu {
		return true
	}
	if m.AutoSwitchToFullScreen != source.AutoSwitchToFullScreen {
		return true
	}
	if m.DisableAero != source.DisableAero {
		return true
	}
	if m.HideMinimizedWindows != source.HideMinimizedWindows {
		return true
	}

	return false
}

type VirtualMachineExpiration struct {
	Enabled bool `json:"enabled"`
}

func (m *VirtualMachineExpiration) Diff(source VirtualMachineExpiration) bool {
	return m.Enabled != source.Enabled
}

type VirtualMachineFullscreen struct {
	UseAllDisplays        string `json:"use_all_displays"`
	ActivateSpacesOnClick string `json:"activate_spaces_on_click"`
	OptimizeForGames      string `json:"optimize_for_games"`
	GammaControl          string `json:"gamma_control"`
	ScaleViewMode         string `json:"scale_view_mode"`
}

func (m *VirtualMachineFullscreen) Diff(source VirtualMachineFullscreen) bool {
	if m.UseAllDisplays != source.UseAllDisplays {
		return true
	}
	if m.ActivateSpacesOnClick != source.ActivateSpacesOnClick {
		return true
	}
	if m.OptimizeForGames != source.OptimizeForGames {
		return true
	}
	if m.GammaControl != source.GammaControl {
		return true
	}
	if m.ScaleViewMode != source.ScaleViewMode {
		return true
	}

	return false
}

type VirtualMachineGuestSharedFolders struct {
	Enabled   bool   `json:"enabled"`
	Automount string `json:"automount"`
}

func (m *VirtualMachineGuestSharedFolders) Diff(source VirtualMachineGuestSharedFolders) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.Automount != source.Automount {
		return true
	}

	return false
}

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

type VirtualMachineMiscellaneousSharing struct {
	SharedClipboard string `json:"shared_clipboard"`
	SharedCloud     string `json:"shared_cloud"`
}

func (m *VirtualMachineMiscellaneousSharing) Diff(source VirtualMachineMiscellaneousSharing) bool {
	if m.SharedClipboard != source.SharedClipboard {
		return true
	}
	if m.SharedCloud != source.SharedCloud {
		return true
	}

	return false
}

type VirtualMachineModality struct {
	OpacityPercentage  int64  `json:"opacity_(percentage)"`
	StayOnTop          string `json:"stay_on_top"`
	ShowOnAllSpaces    string `json:"show_on_all_spaces"`
	CaptureMouseClicks string `json:"capture_mouse_clicks"`
}

func (m *VirtualMachineModality) Diff(source VirtualMachineModality) bool {
	if m.OpacityPercentage != source.OpacityPercentage {
		return true
	}
	if m.StayOnTop != source.StayOnTop {
		return true
	}
	if m.ShowOnAllSpaces != source.ShowOnAllSpaces {
		return true
	}
	if m.CaptureMouseClicks != source.CaptureMouseClicks {
		return true
	}

	return false
}

type VirtualMachineMouseAndKeyboard struct {
	SmartMouseOptimizedForGames string `json:"smart_mouse_optimized_for_games"`
	StickyMouse                 string `json:"sticky_mouse"`
	SmoothScrolling             string `json:"smooth_scrolling"`
	KeyboardOptimizationMode    string `json:"keyboard_optimization_mode"`
}

func (m *VirtualMachineMouseAndKeyboard) Diff(source VirtualMachineMouseAndKeyboard) bool {
	if m.SmartMouseOptimizedForGames != source.SmartMouseOptimizedForGames {
		return true
	}
	if m.StickyMouse != source.StickyMouse {
		return true
	}
	if m.SmoothScrolling != source.SmoothScrolling {
		return true
	}
	if m.KeyboardOptimizationMode != source.KeyboardOptimizationMode {
		return true
	}

	return false
}

type VirtualMachineOptimization struct {
	FasterVirtualMachine     string `json:"faster_virtual_machine"`
	HypervisorType           string `json:"hypervisor_type"`
	AdaptiveHypervisor       string `json:"adaptive_hypervisor"`
	DisabledWindowsLogo      string `json:"disabled_Windows_logo"`
	AutoCompressVirtualDisks string `json:"auto_compress_virtual_disks"`
	NestedVirtualization     string `json:"nested_virtualization"`
	PMUVirtualization        string `json:"PMU_virtualization"`
	LongerBatteryLife        string `json:"longer_battery_life"`
	ShowBatteryStatus        string `json:"show_battery_status"`
	ResourceQuota            string `json:"resource_quota"`
}

func (m *VirtualMachineOptimization) Diff(source VirtualMachineOptimization) bool {
	if m.FasterVirtualMachine != source.FasterVirtualMachine {
		return true
	}
	if m.HypervisorType != source.HypervisorType {
		return true
	}
	if m.AdaptiveHypervisor != source.AdaptiveHypervisor {
		return true
	}
	if m.DisabledWindowsLogo != source.DisabledWindowsLogo {
		return true
	}
	if m.AutoCompressVirtualDisks != source.AutoCompressVirtualDisks {
		return true
	}
	if m.NestedVirtualization != source.NestedVirtualization {
		return true
	}
	if m.PMUVirtualization != source.PMUVirtualization {
		return true
	}
	if m.LongerBatteryLife != source.LongerBatteryLife {
		return true
	}
	if m.ShowBatteryStatus != source.ShowBatteryStatus {
		return true
	}
	if m.ResourceQuota != source.ResourceQuota {
		return true
	}

	return false
}

type VirtualMachinePrintManagement struct {
	SynchronizeWithHostPrinters string `json:"synchronize_with_host_printers"`
	SynchronizeDefaultPrinter   string `json:"synchronize_default_printer"`
	ShowHostPrinterUI           string `json:"show_host_printer_ui"`
}

func (m *VirtualMachinePrintManagement) Diff(source VirtualMachinePrintManagement) bool {
	if m.SynchronizeWithHostPrinters != source.SynchronizeWithHostPrinters {
		return true
	}
	if m.SynchronizeDefaultPrinter != source.SynchronizeDefaultPrinter {
		return true
	}
	if m.ShowHostPrinterUI != source.ShowHostPrinterUI {
		return true
	}

	return false
}

type VirtualMachineSMBIOSSettings struct {
	BIOSVersion        string `json:"bios_version"`
	SystemSerialNumber string `json:"system_serial_number"`
	BoardManufacturer  string `json:"board_manufacturer"`
}

func (m *VirtualMachineSMBIOSSettings) Diff(source VirtualMachineSMBIOSSettings) bool {
	if m.BIOSVersion != source.BIOSVersion {
		return true
	}
	if m.SystemSerialNumber != source.SystemSerialNumber {
		return true
	}
	if m.BoardManufacturer != source.BoardManufacturer {
		return true
	}

	return false
}

type VirtualMachineSecurity struct {
	Encrypted                string `json:"encrypted"`
	TPMEnabled               string `json:"tpm_enabled"`
	TPMType                  string `json:"tpm_type"`
	CustomPasswordProtection string `json:"custom_password_protection"`
	ConfigurationIsLocked    string `json:"configuration_is_locked"`
	Protected                string `json:"protected"`
	Archived                 string `json:"archived"`
	Packed                   string `json:"packed"`
}

func (m *VirtualMachineSecurity) Diff(source VirtualMachineSecurity) bool {
	if m.Encrypted != source.Encrypted {
		return true
	}
	if m.TPMEnabled != source.TPMEnabled {
		return true
	}
	if m.TPMType != source.TPMType {
		return true
	}
	if m.CustomPasswordProtection != source.CustomPasswordProtection {
		return true
	}
	if m.ConfigurationIsLocked != source.ConfigurationIsLocked {
		return true
	}
	if m.Protected != source.Protected {
		return true
	}
	if m.Archived != source.Archived {
		return true
	}
	if m.Packed != source.Packed {
		return true
	}

	return false
}

type VirtualMachineSharedApplications struct {
	Enabled                      bool   `json:"enabled"`
	HostToGuestAppsSharing       string `json:"host-to-guest_apps_sharing"`
	GuestToHostAppsSharing       string `json:"guest-to-host_apps_sharing"`
	ShowGuestAppsFolderInDock    string `json:"show_guest_apps_folder_in_Dock"`
	ShowGuestNotifications       string `json:"show_guest_notifications"`
	BounceDockIconWhenAppFlashes string `json:"bounce_dock_icon_when_app_flashes"`
}

func (m *VirtualMachineSharedApplications) Diff(source VirtualMachineSharedApplications) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.HostToGuestAppsSharing != source.HostToGuestAppsSharing {
		return true
	}
	if m.GuestToHostAppsSharing != source.GuestToHostAppsSharing {
		return true
	}
	if m.ShowGuestAppsFolderInDock != source.ShowGuestAppsFolderInDock {
		return true
	}
	if m.ShowGuestNotifications != source.ShowGuestNotifications {
		return true
	}
	if m.BounceDockIconWhenAppFlashes != source.BounceDockIconWhenAppFlashes {
		return true
	}

	return false
}

type VirtualMachineSmartMount struct {
	Enabled         bool   `json:"enabled"`
	RemovableDrives string `json:"removable drives,omitempty"`
	CDDVDDrives     string `json:"cd_dvd_drives,omitempty"`
	NetworkShares   string `json:"network_shares,omitempty"`
}

func (m *VirtualMachineSmartMount) Diff(source VirtualMachineSmartMount) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.RemovableDrives != source.RemovableDrives {
		return true
	}
	if m.CDDVDDrives != source.CDDVDDrives {
		return true
	}
	if m.NetworkShares != source.NetworkShares {
		return true
	}

	return false
}

type VirtualMachineStartupAndShutdown struct {
	Autostart      string `json:"autostart"`
	AutostartDelay int64  `json:"autostart_delay"`
	Autostop       string `json:"autostop"`
	StartupView    string `json:"startup_view"`
	OnShutdown     string `json:"on_shutdown"`
	OnWindowClose  string `json:"on_window_close"`
	PauseIdle      string `json:"pause_idle"`
	UndoDisks      string `json:"undo_disks"`
}

func (m *VirtualMachineStartupAndShutdown) Diff(source VirtualMachineStartupAndShutdown) bool {
	if m.Autostart != source.Autostart {
		return true
	}
	if m.AutostartDelay != source.AutostartDelay {
		return true
	}
	if m.Autostop != source.Autostop {
		return true
	}
	if m.StartupView != source.StartupView {
		return true
	}
	if m.OnShutdown != source.OnShutdown {
		return true
	}
	if m.OnWindowClose != source.OnWindowClose {
		return true
	}
	if m.PauseIdle != source.PauseIdle {
		return true
	}
	if m.UndoDisks != source.UndoDisks {
		return true
	}

	return false
}

type VirtualMachineTimeSynchronization struct {
	Enabled                         bool   `json:"enabled"`
	SmartMode                       string `json:"Smart_mode"`
	IntervalInSeconds               int64  `json:"interval"`
	TimezoneSynchronizationDisabled string `json:"timezone_synchronization_disabled"`
}

func (m *VirtualMachineTimeSynchronization) Diff(source VirtualMachineTimeSynchronization) bool {
	if m.Enabled != source.Enabled {
		return true
	}
	if m.SmartMode != source.SmartMode {
		return true
	}
	if m.IntervalInSeconds != source.IntervalInSeconds {
		return true
	}
	if m.TimezoneSynchronizationDisabled != source.TimezoneSynchronizationDisabled {
		return true
	}

	return false
}

type VirtualMachineTravelMode struct {
	EnterCondition string `json:"enter_condition"`
	EnterThreshold int64  `json:"enter_threshold"`
	QuitCondition  string `json:"quit_condition"`
}

func (m *VirtualMachineTravelMode) Diff(source VirtualMachineTravelMode) bool {
	if m.EnterCondition != source.EnterCondition {
		return true
	}
	if m.EnterThreshold != source.EnterThreshold {
		return true
	}
	if m.QuitCondition != source.QuitCondition {
		return true
	}

	return false
}

type VirtualMachineUSBAndBluetooth struct {
	AutomaticSharingCameras    string `json:"automatic_sharing_cameras"`
	AutomaticSharingBluetooth  string `json:"automatic_sharing_bluetooth"`
	AutomaticSharingSmartCards string `json:"automatic_sharing_smart_cards"`
	AutomaticSharingGamepads   string `json:"automatic_sharing_gamepads"`
	SupportUSB30               string `json:"support_usb_3_0"`
}

func (m *VirtualMachineUSBAndBluetooth) Diff(source VirtualMachineUSBAndBluetooth) bool {
	if m.AutomaticSharingCameras != source.AutomaticSharingCameras {
		return true
	}
	if m.AutomaticSharingBluetooth != source.AutomaticSharingBluetooth {
		return true
	}
	if m.AutomaticSharingSmartCards != source.AutomaticSharingSmartCards {
		return true
	}
	if m.AutomaticSharingGamepads != source.AutomaticSharingGamepads {
		return true
	}
	if m.SupportUSB30 != source.SupportUSB30 {
		return true
	}

	return false
}

type VirtualMachineNetworkInformation struct {
	Conditioned string                                      `json:"conditioned"`
	Inbound     VirtualMachineNetworkInformationBound       `json:"inbound"`
	Outbound    VirtualMachineNetworkInformationBound       `json:"outbound"`
	IPAddresses []VirtualMachineNetworkInformationIPAddress `json:"ip_addresses"`
}

func (m *VirtualMachineNetworkInformation) Diff(source VirtualMachineNetworkInformation) bool {
	if m.Conditioned != source.Conditioned {
		return true
	}

	if m.Inbound.Diff(source.Inbound) {
		return true
	}

	if m.Outbound.Diff(source.Outbound) {
		return true
	}

	for i, v := range m.IPAddresses {
		if v.Diff(source.IPAddresses[i]) {
			return true
		}
	}

	return false
}

type VirtualMachineNetworkInformationIPAddress struct {
	Type string `json:"type"`
	IP   string `json:"ip"`
}

func (m *VirtualMachineNetworkInformationIPAddress) Diff(source VirtualMachineNetworkInformationIPAddress) bool {
	if m.Type != source.Type {
		return true
	}
	if m.IP != source.IP {
		return true
	}

	return false
}

type VirtualMachineNetworkInformationBound struct {
	Bandwidth  string `json:"bandwidth"`
	PacketLoss string `json:"packet_loss"`
	Delay      string `json:"delay"`
}

func (m *VirtualMachineNetworkInformationBound) Diff(source VirtualMachineNetworkInformationBound) bool {
	if m.Bandwidth != source.Bandwidth {
		return true
	}
	if m.PacketLoss != source.PacketLoss {
		return true
	}
	if m.Delay != source.Delay {
		return true
	}

	return false
}
