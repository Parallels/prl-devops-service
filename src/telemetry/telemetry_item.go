package telemetry

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

type TelemetryItem struct {
	UserID     string
	HardwareID string
	Type       string
	Properties map[string]interface{}
	Options    map[string]interface{}
}

const (
	unknown_user    = "unknown_user"
	unknown_license = "unknown_license"
)

func NewTelemetryItem(ctx basecontext.ApiContext, eventType TelemetryEvent, properties, options map[string]interface{}) TelemetryItem {
	system := system.Get()
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
	item.Properties["os"] = system.GetOperatingSystem()
	if architecture, err := system.GetArchitecture(ctx); err == nil {
		item.Properties["architecture"] = architecture
	} else {
		item.Properties["architecture"] = unknown_user
	}

	if hid, err := system.GetUniqueId(ctx); err == nil {
		item.HardwareID = strings.ReplaceAll(hid, "\"", "")
		item.Properties["hardware_id"] = item.HardwareID
	} else {
		item.HardwareID = unknown_user
		item.Properties["hardware_id"] = item.HardwareID
	}

	if item.HardwareID != "" {
		hash := crypto.SHA256.New()
		hash.Write([]byte(item.HardwareID))
		hashedHardwareId := base64.StdEncoding.EncodeToString(hash.Sum(nil))
		item.HardwareID = hashedHardwareId
		item.Properties["hardware_id"] = hashedHardwareId
	}

	provider := serviceprovider.Get()
	key := provider.License
	if key == "" {
		key = "unknown_license"
	}

	if user, err := system.GetCurrentUser(ctx); err == nil {
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
