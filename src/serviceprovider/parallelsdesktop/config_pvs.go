package parallelsdesktop

type ParallelsDesktopConfigPvs struct {
	ParallelsVirtualMachine ConfigPvsParallelsVirtualMachine `json:"ParallelsVirtualMachine"`
}

type ConfigPvsParallelsVirtualMachine struct {
	AppVersion        string                  `json:"AppVersion"`
	ValidRC           int64                   `json:"ValidRc"`
	Identification    ConfigPvsIdentification `json:"Identification"`
	Security          ConfigPvsSecurity       `json:"Security"`
	Settings          ConfigPvsSettings       `json:"Settings"`
	Hardware          ConfigPvsHardware       `json:"Hardware"`
	InstalledSoftware int64                   `json:"InstalledSoftware"`
}

type ConfigPvsHardware struct {
	CPU               ConfigPvsCPU               `json:"Cpu"`
	Chipset           ConfigPvsChipset           `json:"Chipset"`
	Clock             ConfigPvsClock             `json:"Clock"`
	Memory            ConfigPvsMemory            `json:"Memory"`
	Video             ConfigPvsVideo             `json:"Video"`
	HibernateState    ConfigPvsHibernateState    `json:"HibernateState"`
	VirtIOSerial      ConfigPvsVirtIOSerial      `json:"VirtIOSerial"`
	CDROM             ConfigPvsCDROM             `json:"CdRom"`
	HDD               ConfigPvsHDD               `json:"Hdd"`
	NetworkAdapter    ConfigPvsNetworkAdapter    `json:"NetworkAdapter"`
	Sound             ConfigPvsHardwareSound     `json:"Sound"`
	USB               ConfigPvsUSB               `json:"USB"`
	USBConnectHistory ConfigPvsUSBConnectHistory `json:"UsbConnectHistory"`
	BTConnectHistory  string                     `json:"BTConnectHistory"`
	TPMChip           ConfigPvsTPMChip           `json:"TpmChip"`
}

type ConfigPvsCDROM struct {
	Index             int64  `json:"Index"`
	Enabled           int64  `json:"Enabled"`
	Connected         int64  `json:"Connected"`
	EmulatedType      int64  `json:"EmulatedType"`
	SystemName        string `json:"SystemName"`
	UserFriendlyName  string `json:"UserFriendlyName"`
	Remote            int64  `json:"Remote"`
	InterfaceType     int64  `json:"InterfaceType"`
	StackIndex        int64  `json:"StackIndex"`
	Passthrough       int64  `json:"Passthrough"`
	SubType           int64  `json:"SubType"`
	DeviceDescription string `json:"DeviceDescription"`
}

type ConfigPvsCPU struct {
	Number            int64 `json:"Number"`
	Mode              int64 `json:"Mode"`
	Type              int64 `json:"Type"`
	AutoCountEnabled  int64 `json:"AutoCountEnabled"`
	AccelerationLevel int64 `json:"AccelerationLevel"`
	EnableVTxSupport  int64 `json:"EnableVTxSupport"`
	EnableHotplug     int64 `json:"EnableHotplug"`
	VirtualizedHV     int64 `json:"VirtualizedHV"`
	VirtualizePMU     int64 `json:"VirtualizePMU"`
}

type ConfigPvsChipset struct {
	Type    int64 `json:"Type"`
	Version int64 `json:"Version"`
}

type ConfigPvsClock struct {
	TimeShift int64 `json:"TimeShift"`
}

type ConfigPvsHDD struct {
	UUID                string `json:"Uuid"`
	Index               int64  `json:"Index"`
	Enabled             int64  `json:"Enabled"`
	Connected           int64  `json:"Connected"`
	EmulatedType        int64  `json:"EmulatedType"`
	SystemName          string `json:"SystemName"`
	UserFriendlyName    string `json:"UserFriendlyName"`
	Remote              int64  `json:"Remote"`
	InterfaceType       int64  `json:"InterfaceType"`
	StackIndex          int64  `json:"StackIndex"`
	DiskType            int64  `json:"DiskType"`
	Size                int64  `json:"Size"`
	SizeOnDisk          int64  `json:"SizeOnDisk"`
	Passthrough         int64  `json:"Passthrough"`
	SubType             int64  `json:"SubType"`
	Splitted            int64  `json:"Splitted"`
	DiskVersion         int64  `json:"DiskVersion"`
	CompatLevel         string `json:"CompatLevel"`
	DeviceDescription   string `json:"DeviceDescription"`
	AutoCompressEnabled int64  `json:"AutoCompressEnabled"`
	OnlineCompactMode   int64  `json:"OnlineCompactMode"`
}

type ConfigPvsHibernateState struct {
	ConfigDirty       int64            `json:"ConfigDirty"`
	SMapType          int64            `json:"SMapType"`
	HardwareSignature int64            `json:"HardwareSignature"`
	ShutdownReason    int64            `json:"ShutdownReason"`
	LongReset         int64            `json:"LongReset"`
	ServerUUID        string           `json:"ServerUuid"`
	CPUFeatures       map[string]int64 `json:"CpuFeatures"`
}

type ConfigPvsMemory struct {
	RAM                  int64 `json:"RAM"`
	RAMAutoSizeEnabled   int64 `json:"RamAutoSizeEnabled"`
	EnableHotplug        int64 `json:"EnableHotplug"`
	HostMemQuotaMin      int64 `json:"HostMemQuotaMin"`
	HostMemQuotaMax      int64 `json:"HostMemQuotaMax"`
	HostMemQuotaPriority int64 `json:"HostMemQuotaPriority"`
	AutoQuota            int64 `json:"AutoQuota"`
	MaxBalloonSize       int64 `json:"MaxBalloonSize"`
	ExtendedMemoryLimits int64 `json:"ExtendedMemoryLimits"`
}

type ConfigPvsNetworkAdapter struct {
	Index                 int64              `json:"Index"`
	Enabled               int64              `json:"Enabled"`
	Connected             int64              `json:"Connected"`
	EmulatedType          int64              `json:"EmulatedType"`
	SystemName            string             `json:"SystemName"`
	UserFriendlyName      string             `json:"UserFriendlyName"`
	Remote                int64              `json:"Remote"`
	AdapterNumber         int64              `json:"AdapterNumber"`
	AdapterName           string             `json:"AdapterName"`
	MAC                   string             `json:"MAC"`
	VMNetUUID             string             `json:"VMNetUuid"`
	HostMAC               string             `json:"HostMAC"`
	HostInterfaceName     string             `json:"HostInterfaceName"`
	Router                int64              `json:"Router"`
	DHCPUseHostMAC        int64              `json:"DHCPUseHostMac"`
	ForceHostMACAddress   int64              `json:"ForceHostMacAddress"`
	AdapterType           int64              `json:"AdapterType"`
	StaticAddress         int64              `json:"StaticAddress"`
	PktFilter             ConfigPvsPktFilter `json:"PktFilter"`
	LinkRateLimit         map[string]int64   `json:"LinkRateLimit"`
	AutoApply             int64              `json:"AutoApply"`
	ConfigureWithDHCP     int64              `json:"ConfigureWithDhcp"`
	DefaultGateway        string             `json:"DefaultGateway"`
	ConfigureWithDHCPIPv6 int64              `json:"ConfigureWithDhcpIPv6"`
	DefaultGatewayIPv6    string             `json:"DefaultGatewayIPv6"`
	NetProfile            ConfigPvsProfile   `json:"NetProfile"`
	DeviceDescription     string             `json:"DeviceDescription"`
}

type ConfigPvsProfile struct {
	Type   int64 `json:"Type"`
	Custom int64 `json:"Custom"`
}

type ConfigPvsPktFilter struct {
	PreventPromisc  int64 `json:"PreventPromisc"`
	PreventMACSpoof int64 `json:"PreventMacSpoof"`
	PreventIPSpoof  int64 `json:"PreventIpSpoof"`
}

type ConfigPvsHardwareSound struct {
	Enabled           int64              `json:"Enabled"`
	Connected         int64              `json:"Connected"`
	BusType           int64              `json:"BusType"`
	EmulatedType      int64              `json:"EmulatedType"`
	AdvancedType      int64              `json:"AdvancedType"`
	HDAPatchApplied   int64              `json:"HDAPatchApplied"`
	SystemName        string             `json:"SystemName"`
	UserFriendlyName  string             `json:"UserFriendlyName"`
	Remote            int64              `json:"Remote"`
	VolumeSync        int64              `json:"VolumeSync"`
	Output            string             `json:"Output"`
	Mixer             string             `json:"Mixer"`
	Channel           int64              `json:"Channel"`
	Aec               int64              `json:"AEC"`
	SoundInputs       ConfigPvsSoundPuts `json:"SoundInputs"`
	SoundOutputs      ConfigPvsSoundPuts `json:"SoundOutputs"`
	DeviceDescription string             `json:"DeviceDescription"`
}

type ConfigPvsSoundPuts struct {
	Sound ConfigPvsSoundInputsSound `json:"Sound"`
}

type ConfigPvsSoundInputsSound struct {
	Enabled           int64  `json:"Enabled"`
	Connected         int64  `json:"Connected"`
	BusType           int64  `json:"BusType"`
	EmulatedType      int64  `json:"EmulatedType"`
	AdvancedType      int64  `json:"AdvancedType"`
	HDAPatchApplied   int64  `json:"HDAPatchApplied"`
	SystemName        string `json:"SystemName"`
	UserFriendlyName  string `json:"UserFriendlyName"`
	Remote            int64  `json:"Remote"`
	VolumeSync        int64  `json:"VolumeSync"`
	Output            string `json:"Output"`
	Mixer             string `json:"Mixer"`
	Channel           int64  `json:"Channel"`
	Aec               int64  `json:"AEC"`
	SoundInputs       string `json:"SoundInputs"`
	SoundOutputs      string `json:"SoundOutputs"`
	DeviceDescription string `json:"DeviceDescription"`
}

type ConfigPvsTPMChip struct {
	Type   int64 `json:"Type"`
	Policy int64 `json:"Policy"`
}

type ConfigPvsUSB struct {
	Enabled           int64  `json:"Enabled"`
	Connected         int64  `json:"Connected"`
	EmulatedType      int64  `json:"EmulatedType"`
	SystemName        string `json:"SystemName"`
	UserFriendlyName  string `json:"UserFriendlyName"`
	Remote            int64  `json:"Remote"`
	AutoConnect       int64  `json:"AutoConnect"`
	ConnectReason     int64  `json:"ConnectReason"`
	DeviceDescription string `json:"DeviceDescription"`
	USBType           int64  `json:"UsbType"`
}

type ConfigPvsUSBConnectHistory struct {
	USBPort []ConfigPvsUSBPort `json:"USBPort"`
}

type ConfigPvsUSBPort struct {
	Location   int64  `json:"Location"`
	SystemName string `json:"SystemName"`
	Timestamp  int64  `json:"Timestamp"`
}

type ConfigPvsVideo struct {
	Enabled              int64                     `json:"Enabled"`
	Type                 int64                     `json:"Type"`
	VirtIOBusType        int64                     `json:"VirtIOBusType"`
	VideoMemorySize      int64                     `json:"VideoMemorySize"`
	EnableDirectXShaders int64                     `json:"EnableDirectXShaders"`
	ScreenResolutions    ConfigPvsArchivingOptions `json:"ScreenResolutions"`
	Enable3DAcceleration int64                     `json:"Enable3DAcceleration"`
	EnableVSync          int64                     `json:"EnableVSync"`
	MaxDisplays          int64                     `json:"MaxDisplays"`
	EnableHiResDrawing   int64                     `json:"EnableHiResDrawing"`
	UseHiResInGuest      int64                     `json:"UseHiResInGuest"`
	HostScaleFactor      int64                     `json:"HostScaleFactor"`
	NativeScalingInGuest int64                     `json:"NativeScalingInGuest"`
	ApertureOnlyCapable  int64                     `json:"ApertureOnlyCapable"`
}

type ConfigPvsArchivingOptions struct {
	Enabled int64 `json:"Enabled"`
}

type ConfigPvsVirtIOSerial struct {
	ToolgatePort int64 `json:"ToolgatePort"`
	LoopbackPort int64 `json:"LoopbackPort"`
}

type ConfigPvsIdentification struct {
	VMUUID                string `json:"VmUuid"`
	VMType                int64  `json:"VmType"`
	SourceVMUUID          string `json:"SourceVmUuid"`
	LinkedVMUUID          string `json:"LinkedVmUuid"`
	LinkedSnapshotUUID    string `json:"LinkedSnapshotUuid"`
	VMName                string `json:"VmName"`
	ServerUUID            string `json:"ServerUuid"`
	ServerUUIDAs          string `json:"ServerUuidAs"`
	LastServerUUID        string `json:"LastServerUuid"`
	ServerHost            string `json:"ServerHost"`
	VMHome                string `json:"VmHome"`
	VMFilesLocation       int64  `json:"VmFilesLocation"`
	VMLocationName        string `json:"VmLocationName"`
	VMCreationDate        string `json:"VmCreationDate"`
	VMUptimeStartDateTime string `json:"VmUptimeStartDateTime"`
	VMUptimeInSeconds     int64  `json:"VmUptimeInSeconds"`
}

type ConfigPvsSecurity struct {
	AccessControlList           string `json:"AccessControlList"`
	LockedOperationsList        string `json:"LockedOperationsList"`
	PasswordProtectedOperations string `json:"PasswordProtectedOperations"`
	LockDownHash                string `json:"LockDownHash"`
	Owner                       string `json:"Owner"`
	IsOwner                     int64  `json:"IsOwner"`
	AccessForOthers             int64  `json:"AccessForOthers"`
	LockedSign                  int64  `json:"LockedSign"`
}

type ConfigPvsSettings struct {
	General             ConfigPvsGeneral             `json:"General"`
	Startup             ConfigPvsStartup             `json:"Startup"`
	Shutdown            ConfigPvsShutdown            `json:"Shutdown"`
	SASProfile          ConfigPvsSASProfile          `json:"SasProfile"`
	Runtime             ConfigPvsRuntime             `json:"Runtime"`
	Schedule            ConfigPvsSchedule            `json:"Schedule"`
	Tools               ConfigPvsTools               `json:"Tools"`
	Autoprotect         ConfigPvsAutoprotect         `json:"Autoprotect"`
	AutoCompress        ConfigPvsAutoCompress        `json:"AutoCompress"`
	GlobalNetwork       ConfigPvsGlobalNetwork       `json:"GlobalNetwork"`
	VMEncryptionInfo    ConfigPvsVMEncryptionInfo    `json:"VmEncryptionInfo"`
	VMProtectionInfo    ConfigPvsVMProtectionInfo    `json:"VmProtectionInfo"`
	SharedCamera        ConfigPvsArchivingOptions    `json:"SharedCamera"`
	SharedCCID          ConfigPvsArchivingOptions    `json:"SharedCCID"`
	Keyboard            ConfigPvsKeyboard            `json:"Keyboard"`
	VirtualPrintersInfo ConfigPvsVirtualPrintersInfo `json:"VirtualPrintersInfo"`
	SharedBluetooth     ConfigPvsArchivingOptions    `json:"SharedBluetooth"`
	SharedGamepad       ConfigPvsArchivingOptions    `json:"SharedGamepad"`
	LockDown            ConfigPvsLockDown            `json:"LockDown"`
	USBController       ConfigPvsUSBController       `json:"UsbController"`
	OnlineCompact       ConfigPvsOnlineCompact       `json:"OnlineCompact"`
	TravelOptions       ConfigPvsTravelOptions       `json:"TravelOptions"`
	ArchivingOptions    ConfigPvsArchivingOptions    `json:"ArchivingOptions"`
	PackingOptions      ConfigPvsPackingOptions      `json:"PackingOptions"`
}

type ConfigPvsAutoCompress struct {
	Enabled            int64 `json:"Enabled"`
	Period             int64 `json:"Period"`
	FreeDiskSpaceRatio int64 `json:"FreeDiskSpaceRatio"`
}

type ConfigPvsAutoprotect struct {
	Enabled              int64 `json:"Enabled"`
	Period               int64 `json:"Period"`
	TotalSnapshots       int64 `json:"TotalSnapshots"`
	Schema               int64 `json:"Schema"`
	NotifyBeforeCreation int64 `json:"NotifyBeforeCreation"`
}

type ConfigPvsGeneral struct {
	OSType            int64                  `json:"OsType"`
	OSNumber          int64                  `json:"OsNumber"`
	SourceOSVersion   string                 `json:"SourceOsVersion"`
	VMDescription     string                 `json:"VmDescription"`
	IsTemplate        int64                  `json:"IsTemplate"`
	CustomProperty    string                 `json:"CustomProperty"`
	SwapDir           string                 `json:"SwapDir"`
	VMColor           int64                  `json:"VmColor"`
	Profile           ConfigPvsProfile       `json:"Profile"`
	CPURAMProfile     ConfigPvsCPURAMProfile `json:"CpuRamProfile"`
	AssetID           string                 `json:"AssetId"`
	SerialNumber      string                 `json:"SerialNumber"`
	SMBIOSBIOSVersion string                 `json:"SmbiosBiosVersion"`
	SMBIOSBoardID     string                 `json:"SmbiosBoardId"`
}

type ConfigPvsCPURAMProfile struct {
	Custom       int64                 `json:"Custom"`
	OtherCPU     int64                 `json:"OtherCpu"`
	OtherRAM     int64                 `json:"OtherRam"`
	CustomCPU    ConfigPvsCustomCPU    `json:"CustomCpu"`
	CustomMemory ConfigPvsCustomMemory `json:"CustomMemory"`
}

type ConfigPvsCustomCPU struct {
	Number int64 `json:"Number"`
	Auto   int64 `json:"Auto"`
}

type ConfigPvsCustomMemory struct {
	RAMSize int64 `json:"RamSize"`
	Auto    int64 `json:"Auto"`
}

type ConfigPvsGlobalNetwork struct {
	HostName           string `json:"HostName"`
	DefaultGateway     string `json:"DefaultGateway"`
	DefaultGatewayIPv6 string `json:"DefaultGatewayIPv6"`
	AutoApplyIPOnly    int64  `json:"AutoApplyIpOnly"`
}

type ConfigPvsKeyboard struct {
	HardwareLayout int64 `json:"HardwareLayout"`
}

type ConfigPvsLockDown struct {
	Hash string `json:"Hash"`
}

type ConfigPvsOnlineCompact struct {
	Mode int64 `json:"Mode"`
}

type ConfigPvsPackingOptions struct {
	Direction int64 `json:"Direction"`
	Progress  int64 `json:"Progress"`
}

type ConfigPvsRuntime struct {
	ForegroundPriority           int64                `json:"ForegroundPriority"`
	BackgroundPriority           int64                `json:"BackgroundPriority"`
	DiskCachePolicy              int64                `json:"DiskCachePolicy"`
	CloseAppOnShutdown           int64                `json:"CloseAppOnShutdown"`
	ActionOnStop                 int64                `json:"ActionOnStop"`
	DockIcon                     int64                `json:"DockIcon"`
	OSResolutionInFullScreen     int64                `json:"OsResolutionInFullScreen"`
	FullScreen                   ConfigPvsFullScreen  `json:"FullScreen"`
	UndoDisks                    int64                `json:"UndoDisks"`
	SafeMode                     int64                `json:"SafeMode"`
	SystemFlags                  string               `json:"SystemFlags"`
	DisableAPIC                  int64                `json:"DisableAPIC"`
	OptimizePowerConsumptionMode int64                `json:"OptimizePowerConsumptionMode"`
	ShowBatteryStatus            int64                `json:"ShowBatteryStatus"`
	Enabled                      int64                `json:"Enabled"`
	EnableAdaptiveHypervisor     int64                `json:"EnableAdaptiveHypervisor"`
	UseSMBIOSData                int64                `json:"UseSMBiosData"`
	DisableSpeaker               int64                `json:"DisableSpeaker"`
	HideBIOSOnStartEnabled       int64                `json:"HideBiosOnStartEnabled"`
	UseDefaultAnswers            int64                `json:"UseDefaultAnswers"`
	CompactHDDMask               int64                `json:"CompactHddMask"`
	CompactMode                  int64                `json:"CompactMode"`
	DisableWin7Logo              int64                `json:"DisableWin7Logo"`
	OptimizeModifiers            int64                `json:"OptimizeModifiers"`
	StickyMouse                  int64                `json:"StickyMouse"`
	PauseOnDeactivation          int64                `json:"PauseOnDeactivation"`
	FeaturesMask                 int64                `json:"FEATURES_MASK"`
	EXTFeaturesMask              int64                `json:"EXT_FEATURES_MASK"`
	EXT80000001_EcxMask          int64                `json:"EXT_80000001_ECX_MASK"`
	EXT80000001_EdxMask          int64                `json:"EXT_80000001_EDX_MASK"`
	EXT80000007_EdxMask          int64                `json:"EXT_80000007_EDX_MASK"`
	EXT80000008_Eax              int64                `json:"EXT_80000008_EAX"`
	EXT00000007_EbxMask          int64                `json:"EXT_00000007_EBX_MASK"`
	EXT00000007_EdxMask          int64                `json:"EXT_00000007_EDX_MASK"`
	EXT0000000DEaxMask           int64                `json:"EXT_0000000D_EAX_MASK"`
	EXT00000006_EaxMask          int64                `json:"EXT_00000006_EAX_MASK"`
	CPUFeaturesMaskValid         int64                `json:"CpuFeaturesMaskValid"`
	UnattendedInstallLocale      string               `json:"UnattendedInstallLocale"`
	UnattendedInstallEdition     string               `json:"UnattendedInstallEdition"`
	UnattendedEvaluationVersion  int64                `json:"UnattendedEvaluationVersion"`
	HostRetinaEnabled            int64                `json:"HostRetinaEnabled"`
	DebugServer                  ConfigPvsDebugServer `json:"DebugServer"`
	HypervisorType               int64                `json:"HypervisorType"`
	ResourceQuota                int64                `json:"ResourceQuota"`
	RestoreImage                 string               `json:"RestoreImage"`
	NetworkBusType               int64                `json:"NetworkBusType"`
}

type ConfigPvsDebugServer struct {
	Port  int64 `json:"Port"`
	State int64 `json:"State"`
}

type ConfigPvsFullScreen struct {
	UseAllDisplays        int64   `json:"UseAllDisplays"`
	UseActiveCorners      int64   `json:"UseActiveCorners"`
	UseNativeFullScreen   int64   `json:"UseNativeFullScreen"`
	CornerAction          []int64 `json:"CornerAction"`
	ScaleViewMode         int64   `json:"ScaleViewMode"`
	EnableGammaControl    int64   `json:"EnableGammaControl"`
	OptimiseForGames      int64   `json:"OptimiseForGames"`
	ActivateSpacesOnClick int64   `json:"ActivateSpacesOnClick"`
}

type ConfigPvsSASProfile struct {
	Custom int64 `json:"Custom"`
}

type ConfigPvsSchedule struct {
	SchedBasis       int64  `json:"SchedBasis"`
	SchedGranularity int64  `json:"SchedGranularity"`
	SchedDayOfWeek   int64  `json:"SchedDayOfWeek"`
	SchedDayOfMonth  int64  `json:"SchedDayOfMonth"`
	SchedDay         int64  `json:"SchedDay"`
	SchedWeek        int64  `json:"SchedWeek"`
	SchedMonth       int64  `json:"SchedMonth"`
	SchedStartDate   string `json:"SchedStartDate"`
	SchedStartTime   string `json:"SchedStartTime"`
	SchedStopDate    string `json:"SchedStopDate"`
	SchedStopTime    string `json:"SchedStopTime"`
}

type ConfigPvsShutdown struct {
	AutoStop         int64 `json:"AutoStop"`
	OnVMWindowClose  int64 `json:"OnVmWindowClose"`
	WindowOnShutdown int64 `json:"WindowOnShutdown"`
	ReclaimDiskSpace int64 `json:"ReclaimDiskSpace"`
}

type ConfigPvsStartup struct {
	AutoStart                int64                 `json:"AutoStart"`
	AutoStartDelay           int64                 `json:"AutoStartDelay"`
	VMStartLoginMode         int64                 `json:"VmStartLoginMode"`
	WindowMode               int64                 `json:"WindowMode"`
	StartInDetachedWindow    int64                 `json:"StartInDetachedWindow"`
	BootingOrder             ConfigPvsBootingOrder `json:"BootingOrder"`
	AllowSelectBootDevice    int64                 `json:"AllowSelectBootDevice"`
	BIOS                     ConfigPvsBIOS         `json:"Bios"`
	ExternalDeviceSystemName string                `json:"ExternalDeviceSystemName"`
}

type ConfigPvsBIOS struct {
	EFIEnabled    int64 `json:"EfiEnabled"`
	EFISecureBoot int64 `json:"EfiSecureBoot"`
}

type ConfigPvsBootingOrder struct {
	BootDevice []ConfigPvsBootDevice `json:"BootDevice"`
}

type ConfigPvsBootDevice struct {
	Index         int64 `json:"Index"`
	Type          int64 `json:"Type"`
	BootingNumber int64 `json:"BootingNumber"`
	InUse         int64 `json:"InUse"`
}

type ConfigPvsTools struct {
	ToolsVersion          string                         `json:"ToolsVersion"`
	BusType               int64                          `json:"BusType"`
	IsolatedVM            int64                          `json:"IsolatedVm"`
	NonAdminToolsUpgrade  int64                          `json:"NonAdminToolsUpgrade"`
	LockGuestOnSuspend    int64                          `json:"LockGuestOnSuspend"`
	SyncVMHostname        int64                          `json:"SyncVmHostname"`
	SyncSSHIDS            int64                          `json:"SyncSshIds"`
	Coherence             map[string]int64               `json:"Coherence"`
	SharedFolders         ConfigPvsSharedFolders         `json:"SharedFolders"`
	SharedProfile         ConfigPvsSharedProfile         `json:"SharedProfile"`
	SharedApplications    ConfigPvsSharedApplications    `json:"SharedApplications"`
	AutoUpdate            ConfigPvsArchivingOptions      `json:"AutoUpdate"`
	ClipboardSync         ConfigPvsClipboardSync         `json:"ClipboardSync"`
	DragAndDrop           ConfigPvsArchivingOptions      `json:"DragAndDrop"`
	KeyboardLayoutSync    ConfigPvsArchivingOptions      `json:"KeyboardLayoutSync"`
	MouseSync             ConfigPvsArchivingOptions      `json:"MouseSync"`
	MouseVtdSync          ConfigPvsArchivingOptions      `json:"MouseVtdSync"`
	SmartMouse            ConfigPvsArchivingOptions      `json:"SmartMouse"`
	SmoothScrolling       ConfigPvsArchivingOptions      `json:"SmoothScrolling"`
	TimeSync              ConfigPvsTimeSync              `json:"TimeSync"`
	TisDatabase           ConfigPvsTisDatabase           `json:"TisDatabase"`
	Modality              ConfigPvsModality              `json:"Modality"`
	SharedVolumes         ConfigPvsSharedVolumes         `json:"SharedVolumes"`
	Gestures              ConfigPvsGestures              `json:"Gestures"`
	RemoteControl         ConfigPvsArchivingOptions      `json:"RemoteControl"`
	LocationProvider      ConfigPvsArchivingOptions      `json:"LocationProvider"`
	AutoSyncOSType        ConfigPvsArchivingOptions      `json:"AutoSyncOSType"`
	WinMaintenance        ConfigPvsWinMaintenance        `json:"WinMaintenance"`
	DevelopOptions        ConfigPvsDevelopOptions        `json:"DevelopOptions"`
	DiskSpaceOptimization ConfigPvsDiskSpaceOptimization `json:"DiskSpaceOptimization"`
	RosettaLinux          ConfigPvsArchivingOptions      `json:"RosettaLinux"`
}

type ConfigPvsClipboardSync struct {
	Enabled                int64 `json:"Enabled"`
	PreserveTextFormatting int64 `json:"PreserveTextFormatting"`
}

type ConfigPvsDevelopOptions struct {
	ShowInMenu int64 `json:"ShowInMenu"`
}

type ConfigPvsDiskSpaceOptimization struct {
	SyncFreeSpaceFromHost int64 `json:"SyncFreeSpaceFromHost"`
}

type ConfigPvsGestures struct {
	Enabled        int64 `json:"Enabled"`
	OneFingerSwipe int64 `json:"OneFingerSwipe"`
}

type ConfigPvsModality struct {
	Opacity                float64 `json:"Opacity"`
	StayOnTop              int64   `json:"StayOnTop"`
	CaptureMouseClicks     int64   `json:"CaptureMouseClicks"`
	UseWhenAppInBackground int64   `json:"UseWhenAppInBackground"`
	ShowOnAllSpaces        int64   `json:"ShowOnAllSpaces"`
}

type ConfigPvsSharedApplications struct {
	FromWinToMAC                        int64                    `json:"FromWinToMac"`
	FromMACToWin                        int64                    `json:"FromMacToWin"`
	SmartSelect                         int64                    `json:"SmartSelect"`
	AppInDock                           int64                    `json:"AppInDock"`
	ShowWindowsAppInDock                int64                    `json:"ShowWindowsAppInDock"`
	ShowGuestNotifications              int64                    `json:"ShowGuestNotifications"`
	BounceDockIconWhenAppFlashes        int64                    `json:"BounceDockIconWhenAppFlashes"`
	WebApplications                     ConfigPvsWebApplications `json:"WebApplications"`
	IconGroupingEnabled                 int64                    `json:"IconGroupingEnabled"`
	DisableRecentDocs                   int64                    `json:"DisableRecentDocs"`
	StoreInternetPasswordsInOSXKeychain int64                    `json:"StoreInternetPasswordsInOSXKeychain"`
}

type ConfigPvsWebApplications struct {
	WebBrowser   int64 `json:"WebBrowser"`
	EmailClient  int64 `json:"EmailClient"`
	FTPClient    int64 `json:"FtpClient"`
	Newsgroups   int64 `json:"Newsgroups"`
	RSS          int64 `json:"Rss"`
	RemoteAccess int64 `json:"RemoteAccess"`
}

type ConfigPvsSharedFolders struct {
	HostSharing  map[string]int64      `json:"HostSharing"`
	GuestSharing ConfigPvsGuestSharing `json:"GuestSharing"`
}

type ConfigPvsGuestSharing struct {
	Enabled                int64 `json:"Enabled"`
	AutoMount              int64 `json:"AutoMount"`
	AutoMountNetworkDrives int64 `json:"AutoMountNetworkDrives"`
	EnableSpotlight        int64 `json:"EnableSpotlight"`
	AutoMountCloudDrives   int64 `json:"AutoMountCloudDrives"`
	ShareRemovableDrives   int64 `json:"ShareRemovableDrives"`
	PortNumber             int64 `json:"PortNumber"`
	AllowExec              int64 `json:"AllowExec"`
}

type ConfigPvsSharedProfile struct {
	Enabled      int64 `json:"Enabled"`
	UseDesktop   int64 `json:"UseDesktop"`
	UseDocuments int64 `json:"UseDocuments"`
	UsePictures  int64 `json:"UsePictures"`
	UseMusic     int64 `json:"UseMusic"`
	UseMovies    int64 `json:"UseMovies"`
	UseDownloads int64 `json:"UseDownloads"`
	UseTrashBin  int64 `json:"UseTrashBin"`
}

type ConfigPvsSharedVolumes struct {
	Enabled             int64 `json:"Enabled"`
	UseExternalDisks    int64 `json:"UseExternalDisks"`
	UseDVDs             int64 `json:"UseDVDs"`
	UseConnectedServers int64 `json:"UseConnectedServers"`
	UseInversedDisks    int64 `json:"UseInversedDisks"`
}

type ConfigPvsTimeSync struct {
	Enabled              int64 `json:"Enabled"`
	SyncInterval         int64 `json:"SyncInterval"`
	KeepTimeDiff         int64 `json:"KeepTimeDiff"`
	SyncHostToGuest      int64 `json:"SyncHostToGuest"`
	SyncTimezoneDisabled int64 `json:"SyncTimezoneDisabled"`
}

type ConfigPvsTisDatabase struct {
	Data string `json:"Data"`
}

type ConfigPvsWinMaintenance struct {
	Enabled      int64  `json:"Enabled"`
	ScheduleDay  int64  `json:"ScheduleDay"`
	ScheduleTime string `json:"ScheduleTime"`
}

type ConfigPvsTravelOptions struct {
	Enabled      int64              `json:"Enabled"`
	Condition    ConfigPvsCondition `json:"Condition"`
	SavedOptions string             `json:"SavedOptions"`
}

type ConfigPvsCondition struct {
	Enter                 int64 `json:"Enter"`
	EnterBetteryThreshold int64 `json:"EnterBetteryThreshold"`
	Quit                  int64 `json:"Quit"`
}

type ConfigPvsUSBController struct {
	BusType         int64                    `json:"BusType"`
	UhcEnabled      int64                    `json:"UhcEnabled"`
	EhcEnabled      int64                    `json:"EhcEnabled"`
	XhcEnabled      int64                    `json:"XhcEnabled"`
	ExternalDevices ConfigPvsExternalDevices `json:"ExternalDevices"`
}

type ConfigPvsExternalDevices struct {
	Disks           int64 `json:"Disks"`
	HumanInterfaces int64 `json:"HumanInterfaces"`
	Communication   int64 `json:"Communication"`
	Audio           int64 `json:"Audio"`
	Video           int64 `json:"Video"`
	SmartCards      int64 `json:"SmartCards"`
	Printers        int64 `json:"Printers"`
	SmartPhones     int64 `json:"SmartPhones"`
	Other           int64 `json:"Other"`
}

type ConfigPvsVMEncryptionInfo struct {
	Enabled  int64  `json:"Enabled"`
	PluginID string `json:"PluginId"`
	Hash1    string `json:"Hash1"`
	Hash2    string `json:"Hash2"`
	Salt     string `json:"Salt"`
}

type ConfigPvsVMProtectionInfo struct {
	Enabled        int64                   `json:"Enabled"`
	Hash1          string                  `json:"Hash1"`
	Hash2          string                  `json:"Hash2"`
	Hash3          string                  `json:"Hash3"`
	Salt           string                  `json:"Salt"`
	ExpirationInfo ConfigPvsExpirationInfo `json:"ExpirationInfo"`
}

type ConfigPvsExpirationInfo struct {
	Enabled                  int64  `json:"Enabled"`
	ExpirationDate           string `json:"ExpirationDate"`
	TrustedTimeServerURL     string `json:"TrustedTimeServerUrl"`
	Note                     string `json:"Note"`
	TimeCheckIntervalSeconds int64  `json:"TimeCheckIntervalSeconds"`
	OfflineTimeToLiveSeconds int64  `json:"OfflineTimeToLiveSeconds"`
}

type ConfigPvsVirtualPrintersInfo struct {
	UseHostPrinters    int64 `json:"UseHostPrinters"`
	SyncDefaultPrinter int64 `json:"SyncDefaultPrinter"`
	ShowHostPrinterUI  int64 `json:"ShowHostPrinterUI"`
}
