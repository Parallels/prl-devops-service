package models

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
