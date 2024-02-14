package telemetry

import (
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

func NewTelemetryItem(ctx basecontext.ApiContext, eventType string, properties, options map[string]interface{}) TelemetryItem {
	system := system.Get()
	item := TelemetryItem{
		Type:       eventType,
		Properties: properties,
		Options:    options,
	}
	if item.Properties == nil {
		item.Properties = make(map[string]interface{})
	}
	if item.Options == nil {
		item.Options = make(map[string]interface{})
	}

	if hid, err := system.GetUniqueId(ctx); err == nil {
		item.HardwareID = hid
	} else {
		item.HardwareID = "unknown"
	}

	if user, err := system.GetCurrentUser(ctx); err == nil {
		item.UserID = user
	}

	return item
}
