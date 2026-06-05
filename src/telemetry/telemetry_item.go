package telemetry

import (
	"crypto"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

type TelemetryItem struct {
	UserID     string
	DeviceId   string
	Type       string
	Properties map[string]interface{}
	Options    map[string]interface{}
}

const (
	unknownUser    = "unknown_user"
	unknownLicense = "unknown_license"
)

// sha256Hex returns the lowercase hex digest of the SHA256 hash of data.
func sha256Hex(data string) string {
	hash := crypto.SHA256.New()
	hash.Write([]byte(data))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func NewTelemetryItem(ctx basecontext.ApiContext, eventType TelemetryEvent, properties, options map[string]interface{}) TelemetryItem {
	sys := system.Get()
	item := TelemetryItem{
		Type:       string(eventType),
		Properties: properties,
		Options:    options,
	}
	if item.Properties == nil {
		item.Properties = make(map[string]interface{})
	}
	if item.Options == nil {
		item.Options = make(map[string]interface{})
	}

	// --- Raw unique ID (machine UUID) ---
	var rawId string
	if hid, err := sys.GetUniqueId(ctx); err == nil {
		rawId = strings.ReplaceAll(hid, "\"", "")
	} else {
		rawId = unknownUser
	}

	// --- hardware_id: raw machine UUID (no encoding, not PII) ---
	item.DeviceId = rawId
	item.Properties["hardware_id"] = rawId

	// --- Default properties ---
	item.Properties["os"] = sys.GetOperatingSystem()
	if arch, err := sys.GetArchitecture(ctx); err == nil {
		item.Properties["architecture"] = arch
	}

	item.Properties["version"] = system.VersionSvc.String()

	cfg := config.Get()
	if cfg != nil {
		item.Properties["call_source"] = cfg.Source()

		// --- enabled_modules: comma-separated list of active modules ---
		modules := cfg.GetEnabledModules()
		if len(modules) > 0 {
			item.Properties["enabled_modules"] = strings.Join(modules, ",")
		}
	}

	// --- user_id: SHA256("<sha256(raw_id)>@<hardware_id>") ---
	uidHash := sha256Hex(rawId)
	userInput := fmt.Sprintf("%s@%s", uidHash, rawId)
	hash := crypto.SHA256.New()
	hash.Write([]byte(userInput))
	hashedUserId := fmt.Sprintf("%x", hash.Sum(nil))

	item.UserID = hashedUserId
	if len(hashedUserId) > 10 {
		item.Properties["user_id"] = hashedUserId
	}

	// --- Parallels Desktop license properties (raw, not PII) ---
	provider := serviceprovider.Get()
	if provider != nil && provider.ParallelsDesktopService != nil && provider.ParallelsDesktopService.Installed() {
		if license, err := provider.ParallelsDesktopService.GetCachedLicense(); err == nil && license != nil {
			hashedLicenseSerial := sha256Hex(license.Serial)
			item.Properties["pd_edition"] = license.Edition
			item.Properties["pd_serial"] = hashedLicenseSerial
			item.Properties["pd_uuid"] = license.UUID
			item.Properties["pd_is_trial"] = license.IsTrial
		}
	}

	return item
}
