package models

import "encoding/json"

type VirtualMachine struct {
	ID                    string                             `json:"ID,omitempty"`
	HostUrl               string                             `json:"host_url,omitempty"`
	HostId                string                             `json:"host_id,omitempty"`
	HostName              string                             `json:"host_name,omitempty"`
	HostState             string                             `json:"host_state,omitempty"`
	User                  string                             `json:"user,omitempty"`
	HostExternalIpAddress string                             `json:"host_external_ip_address,omitempty"`
	InternalIpAddress     string                             `json:"internal_ip_address,omitempty"`
	IpStatus              string                             `json:"ip_status,omitempty"`
	IpStatusDetail        string                             `json:"ip_status_detail,omitempty"`
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
	if m.HostUrl != source.HostUrl {
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
	if m.IpStatus != source.IpStatus {
		return true
	}
	if m.IpStatusDetail != source.IpStatusDetail {
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
