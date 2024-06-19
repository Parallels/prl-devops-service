package telemetry

import (
	"crypto"
	"encoding/base64"
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
	unknown_user    = "unknown_user"
	unknown_license = "unknown_license"
)

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

	// Adding default properties
	item.Properties["os"] = sys.GetOperatingSystem()
	if architecture, err := sys.GetArchitecture(ctx); err == nil {
		item.Properties["architecture"] = architecture
	} else {
		item.Properties["architecture"] = unknown_user
	}

	if hid, err := sys.GetUniqueId(ctx); err == nil {
		item.DeviceId = strings.ReplaceAll(hid, "\"", "")
		item.Properties["hardware_id"] = item.DeviceId
	} else {
		item.DeviceId = unknown_user
		item.Properties["hardware_id"] = item.DeviceId
	}

	item.Properties["version"] = system.VersionSvc.String()

	config := config.Get()
	if config != nil {
		item.Properties["call_source"] = config.Source()
	}

	if item.DeviceId != "" {
		hash := crypto.SHA256.New()
		hash.Write([]byte(item.DeviceId))
		hashedHardwareId := base64.StdEncoding.EncodeToString(hash.Sum(nil))
		item.DeviceId = hashedHardwareId
		item.Properties["hardware_id"] = hashedHardwareId
	}

	key := "unknown_license"
	provider := serviceprovider.Get()
	if provider != nil {
		key = provider.License
	}

	if user, err := sys.GetCurrentUser(ctx); err == nil {
		hash := crypto.SHA256.New()
		hash.Write([]byte(user))
		hashedUser := base64.StdEncoding.EncodeToString(hash.Sum(nil))
		item.UserID = hashedUser
	} else {
		item.UserID = unknown_user
	}

	userId := fmt.Sprintf("%s@%s", item.UserID, key)
	hash := crypto.SHA256.New()
	hash.Write([]byte(userId))
	hashedUserId := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	if len(hashedUserId) > 10 {
		item.Properties["user_id"] = hashedUserId
	}

	return item
}
