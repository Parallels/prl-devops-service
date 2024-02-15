package telemetry

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
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
	unknown = "unknown"
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
		item.Properties["architecture"] = unknown
	}

	if hid, err := system.GetUniqueId(ctx); err == nil {
		item.HardwareID = strings.ReplaceAll(hid, "\"", "")
		item.Properties["hardware_id"] = item.HardwareID
	} else {
		item.HardwareID = unknown
		item.Properties["hardware_id"] = item.HardwareID
	}

	if user, err := system.GetCurrentUser(ctx); err == nil {
		item.UserID = user
		item.Properties["user_id"] = item.UserID
	} else {
		item.UserID = unknown
		item.Properties["user_id"] = unknown
	}

	return item
}
