package models

import (
	"encoding/json"
	"regexp"
)

type ParallelsVMs []ParallelsVM

type ParallelsVM struct {
	ID                    string                 `json:"ID"`
	Host                  string                 `json:"host,omitempty"`
	HostUrl               string                 `json:"host_url,omitempty"`
	HostId                string                 `json:"host_id,omitempty"`
	HostName              string                 `json:"host_name,omitempty"`
	HostState             string                 `json:"host_state,omitempty"`
	HostExternalIpAddress string                 `json:"host_external_ip_address,omitempty"`
	InternalIpAddress     string                 `json:"internal_ip_address,omitempty"`
	IpStatus              string                 `json:"ip_status,omitempty"`        // "", pending, resolving, resolved, failed
	IpStatusDetail        string                 `json:"ip_status_detail,omitempty"` // human-readable detail for the status
	User                  string                 `json:"user"`
	Name                  string                 `json:"Name"`
	Description           string                 `json:"Description"`
	Type                  string                 `json:"Type"`
	State                 string                 `json:"State"`
	OS                    string                 `json:"OS"`
	Template              string                 `json:"Template"`
	Uptime                string                 `json:"Uptime"`
	HomePath              string                 `json:"Home path"`
	Home                  string                 `json:"Home"`
	RestoreImage          string                 `json:"Restore Image"`
	GuestTools            GuestTools             `json:"GuestTools"`
	MouseAndKeyboard      MouseAndKeyboard       `json:"Mouse and Keyboard"`
	USBAndBluetooth       USBAndBluetooth        `json:"USB and Bluetooth"`
	StartupAndShutdown    StartupAndShutdown     `json:"Startup and Shutdown"`
	Optimization          Optimization           `json:"Optimization"`
	TravelMode            TravelMode             `json:"Travel mode"`
	Security              Security               `json:"Security"`
	SmartGuard            Expiration             `json:"Smart Guard"`
	Modality              Modality               `json:"Modality"`
	Fullscreen            Fullscreen             `json:"Fullscreen"`
	Coherence             Coherence              `json:"Coherence"`
	TimeSynchronization   TimeSynchronization    `json:"Time Synchronization"`
	Expiration            Expiration             `json:"Expiration"`
	BootOrder             string                 `json:"Boot order"`
	BIOSType              string                 `json:"BIOS type"`
	EFISecureBoot         string                 `json:"EFI Secure boot"`
	AllowSelectBootDevice string                 `json:"Allow select boot device"`
	ExternalBootDevice    string                 `json:"External boot device"`
	SMBIOSSettings        SMBIOSSettings         `json:"SMBIOS settings"`
	Hardware              Hardware               `json:"Hardware"`
	HostSharedFolders     map[string]interface{} `json:"Host Shared Folders"`
	HostDefinedSharing    string                 `json:"Host defined sharing"`
	SharedProfile         Expiration             `json:"Shared Profile"`
	SharedApplications    SharedApplications     `json:"Shared Applications"`
	SmartMount            SmartMount             `json:"SmartMount"`
	MiscellaneousSharing  MiscellaneousSharing   `json:"Miscellaneous Sharing"`
	Advanced              Advanced               `json:"Advanced"`
	PrintManagement       PrintManagement        `json:"Print Management,omitempty"`
	GuestSharedFolders    GuestSharedFolders     `json:"Guest Shared Folders,omitempty"`
	NetworkInformation    NetworkInformation     `json:"Network,omitempty"`
}

type Advanced struct {
	VMHostnameSynchronization    string `json:"VM hostname synchronization"`
	PublicSSHKeysSynchronization string `json:"Public SSH keys synchronization"`
	ShowDeveloperTools           string `json:"Show developer tools"`
	SwipeFromEdges               string `json:"Swipe from edges"`
	ShareHostLocation            string `json:"Share host location"`
	RosettaLinux                 string `json:"Rosetta Linux"`
}

type Coherence struct {
	ShowWindowsSystrayInMACMenu string `json:"Show Windows systray in Mac menu"`
	AutoSwitchToFullScreen      string `json:"Auto-switch to full screen"`
	DisableAero                 string `json:"Disable aero"`
	HideMinimizedWindows        string `json:"Hide minimized windows"`
}

type Expiration struct {
	Enabled bool `json:"enabled"`
}

type Fullscreen struct {
	UseAllDisplays        string `json:"Use all displays"`
	ActivateSpacesOnClick string `json:"Activate spaces on click"`
	OptimizeForGames      string `json:"Optimize for games"`
	GammaControl          string `json:"Gamma control"`
	ScaleViewMode         string `json:"Scale view mode"`
}

type GuestSharedFolders struct {
	Enabled   bool   `json:"enabled"`
	Automount string `json:"Automount"`
}

type GuestTools struct {
	State   string `json:"state"`
	Version string `json:"version,omitempty"`
}

// Hardware is the "Hardware" section of the prlctl JSON. Devices are dynamic
// keyed entries (hdd0..n, cdrom0..n, net0..n, sound0..n, fdd0..n), so they are
// held in typed maps keyed by the device name instead of fixed single-slot
// fields — devices added with prlctl --device-add are never dropped. Unknown
// keys (serial/parallel devices, future prlctl additions) are preserved
// verbatim in Extras. The custom JSON codec below keeps the wire format flat
// and identical to prlctl output.
type Hardware struct {
	CPU         CPU                        `json:"cpu"`
	Memory      Memory                     `json:"memory"`
	Video       Video                      `json:"video"`
	MemoryQuota MemoryQuota                `json:"memory_quota"`
	USB         Expiration                 `json:"usb"`
	Hdds        map[string]Hdd             `json:"-"`
	CdRoms      map[string]Cdrom           `json:"-"`
	Nets        map[string]Net             `json:"-"`
	Sounds      map[string]Sound           `json:"-"`
	Fdds        map[string]Fdd             `json:"-"`
	Extras      map[string]json.RawMessage `json:"-"`
}

// hardwareDeviceKey matches the dynamic prlctl device keys (hdd0, cdrom1,
// net0, sound0, fdd0, ...).
var hardwareDeviceKey = regexp.MustCompile(`^(hdd|cdrom|net|sound|fdd)[0-9]+$`)

// UnmarshalJSON decodes the flat prlctl hardware object: the known scalar
// groups (cpu, memory, video, memory_quota, usb) fill their typed fields,
// every device key lands in the matching typed map keyed by the device name,
// and anything else is preserved as raw JSON in Extras.
func (h *Hardware) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*h = Hardware{}
	for key, value := range raw {
		switch key {
		case "cpu":
			if err := json.Unmarshal(value, &h.CPU); err != nil {
				return err
			}
		case "memory":
			if err := json.Unmarshal(value, &h.Memory); err != nil {
				return err
			}
		case "video":
			if err := json.Unmarshal(value, &h.Video); err != nil {
				return err
			}
		case "memory_quota":
			if err := json.Unmarshal(value, &h.MemoryQuota); err != nil {
				return err
			}
		case "usb":
			if err := json.Unmarshal(value, &h.USB); err != nil {
				return err
			}
		default:
			if match := hardwareDeviceKey.FindStringSubmatch(key); match != nil {
				switch match[1] {
				case "hdd":
					var device Hdd
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if h.Hdds == nil {
						h.Hdds = map[string]Hdd{}
					}
					h.Hdds[key] = device
				case "cdrom":
					var device Cdrom
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if h.CdRoms == nil {
						h.CdRoms = map[string]Cdrom{}
					}
					h.CdRoms[key] = device
				case "net":
					var device Net
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if h.Nets == nil {
						h.Nets = map[string]Net{}
					}
					h.Nets[key] = device
				case "sound":
					var device Sound
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if h.Sounds == nil {
						h.Sounds = map[string]Sound{}
					}
					h.Sounds[key] = device
				case "fdd":
					var device Fdd
					if err := json.Unmarshal(value, &device); err != nil {
						return err
					}
					if h.Fdds == nil {
						h.Fdds = map[string]Fdd{}
					}
					h.Fdds[key] = device
				}
				continue
			}
			if h.Extras == nil {
				h.Extras = map[string]json.RawMessage{}
			}
			h.Extras[key] = append(json.RawMessage(nil), value...)
		}
	}

	return nil
}

// MarshalJSON re-emits the flat prlctl-shaped hardware object from the typed
// device maps, including the unknown keys captured in Extras. encoding/json
// sorts map keys, so the output is deterministic (golden tests depend on it).
func (h Hardware) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"cpu":          h.CPU,
		"memory":       h.Memory,
		"video":        h.Video,
		"memory_quota": h.MemoryQuota,
		"usb":          h.USB,
	}
	for key, device := range h.Hdds {
		out[key] = device
	}
	for key, device := range h.CdRoms {
		out[key] = device
	}
	for key, device := range h.Nets {
		out[key] = device
	}
	for key, device := range h.Sounds {
		out[key] = device
	}
	for key, device := range h.Fdds {
		out[key] = device
	}
	for key, raw := range h.Extras {
		if _, taken := out[key]; !taken {
			out[key] = raw
		}
	}

	return json.Marshal(out)
}

type Cdrom struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Image   string `json:"image"`
	State   string `json:"state,omitempty"`
}

type Hdd struct {
	Enabled       bool   `json:"enabled"`
	Port          string `json:"port"`
	Image         string `json:"image"`
	Type          string `json:"type"`
	Size          string `json:"size"`
	OnlineCompact string `json:"online-compact"`
}

type Fdd struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Image   string `json:"image"`
	State   string `json:"state,omitempty"`
}

type Net struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	MAC     string `json:"mac"`
	Card    string `json:"card"`
}

type Sound struct {
	Enabled bool   `json:"enabled"`
	Output  string `json:"output"`
	Mixer   string `json:"mixer"`
}

type CPU struct {
	Cpus    int64  `json:"cpus"`
	Auto    string `json:"auto"`
	VTX     bool   `json:"VT-x"`
	Hotplug bool   `json:"hotplug"`
	Accl    string `json:"accl"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
}

type Memory struct {
	Size    string `json:"size"`
	Auto    string `json:"auto"`
	Hotplug bool   `json:"hotplug"`
}

type MemoryQuota struct {
	Auto string `json:"auto"`
}

type Video struct {
	AdapterType           string `json:"adapter-type"`
	Size                  string `json:"size"`
	The3DAcceleration     string `json:"3d-acceleration"`
	VerticalSync          string `json:"vertical-sync"`
	HighResolution        string `json:"high-resolution"`
	HighResolutionInGuest string `json:"high-resolution-in-guest"`
	NativeScalingInGuest  string `json:"native-scaling-in-guest"`
	AutomaticVideoMemory  string `json:"automatic-video-memory"`
}

type MiscellaneousSharing struct {
	SharedClipboard string `json:"Shared clipboard"`
	SharedCloud     string `json:"Shared cloud"`
}

type Modality struct {
	OpacityPercentage  int64  `json:"Opacity (percentage)"`
	StayOnTop          string `json:"Stay on top"`
	ShowOnAllSpaces    string `json:"Show on all spaces "`
	CaptureMouseClicks string `json:"Capture mouse clicks"`
}

type MouseAndKeyboard struct {
	SmartMouseOptimizedForGames string `json:"Smart mouse optimized for games"`
	StickyMouse                 string `json:"Sticky mouse"`
	SmoothScrolling             string `json:"Smooth scrolling"`
	KeyboardOptimizationMode    string `json:"Keyboard optimization mode"`
}

type Optimization struct {
	FasterVirtualMachine     string `json:"Faster virtual machine"`
	HypervisorType           string `json:"Hypervisor type"`
	AdaptiveHypervisor       string `json:"Adaptive hypervisor"`
	DisabledWindowsLogo      string `json:"Disabled Windows logo"`
	AutoCompressVirtualDisks string `json:"Auto compress virtual disks"`
	NestedVirtualization     string `json:"Nested virtualization"`
	PMUVirtualization        string `json:"PMU virtualization"`
	LongerBatteryLife        string `json:"Longer battery life"`
	ShowBatteryStatus        string `json:"Show battery status"`
	ResourceQuota            string `json:"Resource quota"`
}

type PrintManagement struct {
	SynchronizeWithHostPrinters string `json:"Synchronize with host printers"`
	SynchronizeDefaultPrinter   string `json:"Synchronize default printer"`
	ShowHostPrinterUI           string `json:"Show host printer UI"`
}

type SMBIOSSettings struct {
	BIOSVersion        string `json:"BIOS Version"`
	SystemSerialNumber string `json:"System serial number"`
	BoardManufacturer  string `json:"Board Manufacturer"`
}

type Security struct {
	Encrypted                string `json:"Encrypted"`
	TPMEnabled               string `json:"TPM enabled"`
	TPMType                  string `json:"TPM type"`
	CustomPasswordProtection string `json:"Custom password protection"`
	ConfigurationIsLocked    string `json:"Configuration is locked"`
	Protected                string `json:"Protected"`
	Archived                 string `json:"Archived"`
	Packed                   string `json:"Packed"`
}

type SharedApplications struct {
	Enabled                      bool   `json:"enabled"`
	HostToGuestAppsSharing       string `json:"Host-to-guest apps sharing"`
	GuestToHostAppsSharing       string `json:"Guest-to-host apps sharing"`
	ShowGuestAppsFolderInDock    string `json:"Show guest apps folder in Dock"`
	ShowGuestNotifications       string `json:"Show guest notifications"`
	BounceDockIconWhenAppFlashes string `json:"Bounce dock icon when app flashes"`
}

type SmartMount struct {
	Enabled         bool   `json:"enabled"`
	RemovableDrives string `json:"Removable drives,omitempty"`
	CDDVDDrives     string `json:"CD/DVD drives,omitempty"`
	NetworkShares   string `json:"Network shares,omitempty"`
}

type StartupAndShutdown struct {
	Autostart      string `json:"Autostart"`
	AutostartDelay int64  `json:"Autostart delay"`
	Autostop       string `json:"Autostop"`
	StartupView    string `json:"Startup view"`
	OnShutdown     string `json:"On shutdown"`
	OnWindowClose  string `json:"On window close"`
	PauseIdle      string `json:"Pause idle"`
	UndoDisks      string `json:"Undo disks"`
}

type TimeSynchronization struct {
	Enabled                         bool   `json:"enabled"`
	SmartMode                       string `json:"Smart mode"`
	IntervalInSeconds               int64  `json:"Interval (in seconds)"`
	TimezoneSynchronizationDisabled string `json:"Timezone synchronization disabled"`
}

type TravelMode struct {
	EnterCondition string `json:"Enter condition"`
	EnterThreshold int64  `json:"Enter threshold"`
	QuitCondition  string `json:"Quit condition"`
}

type USBAndBluetooth struct {
	AutomaticSharingCameras    string `json:"Automatic sharing cameras"`
	AutomaticSharingBluetooth  string `json:"Automatic sharing bluetooth"`
	AutomaticSharingSmartCards string `json:"Automatic sharing smart cards"`
	AutomaticSharingGamepads   string `json:"Automatic sharing gamepads"`
	SupportUSB30               string `json:"Support USB 3.0"`
}

type NetworkInformation struct {
	Conditioned string                        `json:"Conditioned"`
	Inbound     NetworkInformationBound       `json:"Inbound"`
	Outbound    NetworkInformationBound       `json:"Outbound"`
	IPAddresses []NetworkInformationIPAddress `json:"ipAddresses"`
}

type NetworkInformationIPAddress struct {
	Type string `json:"type"`
	IP   string `json:"ip"`
}

type NetworkInformationBound struct {
	Bandwidth  string `json:"Bandwidth"`
	PacketLoss string `json:"Packet Loss"`
	Delay      string `json:"Delay"`
}
