package models

type ParallelsDesktopInfo struct {
	ID                           string                                  `json:"ID"`
	Hostname                     string                                  `json:"Hostname"`
	OS                           string                                  `json:"OS"`
	Version                      string                                  `json:"Version"`
	BuildFlags                   string                                  `json:"Build flags"`
	StartedAsService             string                                  `json:"Started as service"`
	VMHome                       string                                  `json:"VM home"`
	MemoryLimit                  ParallelsDesktopMemoryLimit             `json:"Memory limit"`
	MinimalSecurityLevel         string                                  `json:"Minimal security level"`
	ManageSettingsForNewUsers    string                                  `json:"Manage settings for new users"`
	CEPMechanism                 string                                  `json:"CEP mechanism"`
	VerboseLog                   string                                  `json:"Verbose log"`
	LogRotation                  string                                  `json:"Log rotation"`
	ExternalDeviceAutoConnect    string                                  `json:"External device auto connect"`
	WebPortalDomain              string                                  `json:"Web portal domain"`
	HostID                       string                                  `json:"Host ID"`
	AllowAttachScreenshots       string                                  `json:"Allow attach screenshots"`
	CustomPasswordProtection     string                                  `json:"Custom password protection"`
	PreferencesAreLocked         string                                  `json:"Preferences are locked"`
	DisableDeletingReportArchive string                                  `json:"Disable deleting report archive"`
	AddSerialPortOutputToReport  string                                  `json:"Add serial port output to report"`
	License                      HardwareInfoParallelsDesktopLicense     `json:"License"`
	HardwareID                   string                                  `json:"Hardware Id"`
	SignedIn                     string                                  `json:"Signed In"`
	HardwareInfo                 map[string]ParallelsDesktopHardwareInfo `json:"Hardware info"`
}

type ParallelsDesktopHardwareInfo struct {
	Name string                       `json:"name"`
	Type ParallelsDesktopHardwareType `json:"type"`
}

type HardwareInfoParallelsDesktopLicense struct {
	State      string `json:"state"`
	Key        string `json:"key"`
	Restricted string `json:"restricted"`
}

type ParallelsDesktopMemoryLimit struct {
	Mode string `json:"mode"`
}

type ParallelsDesktopHardwareType string

const (
	HDD     ParallelsDesktopHardwareType = "hdd"
	HDDPart ParallelsDesktopHardwareType = "hdd-part"
	Serial  ParallelsDesktopHardwareType = "serial"
	USB     ParallelsDesktopHardwareType = "usb"
)
