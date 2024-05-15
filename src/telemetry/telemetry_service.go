package telemetry

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/amplitude/analytics-go/amplitude"
	"github.com/amplitude/analytics-go/amplitude/types"
)

type TelemetryService struct {
	ctx             basecontext.ApiContext
	client          amplitude.Client
	EnableTelemetry bool
	CallBackChan    chan types.ExecuteResult
}

func (t *TelemetryService) TrackEvent(item TelemetryItem) {
	if !t.EnableTelemetry {
		t.ctx.LogDebugf("[Telemetry] Telemetry is disabled, ignoring event track")
		return
	}

	t.ctx.LogInfof("[Telemetry] Sending Amplitude Tracking event %s", item.Type)

	// Create a new event
	if len(item.UserID) < 5 {
		if item.HardwareID != "" {
			item.UserID = fmt.Sprintf("%s@%s", item.UserID, item.HardwareID)
		} else {
			item.UserID = fmt.Sprintf("%s@service", item.UserID)
		}
	}
	if len(item.HardwareID) < 5 {
		item.HardwareID = "service"
	}

	ev := amplitude.Event{
		UserID:          item.UserID,
		DeviceID:        item.HardwareID,
		EventType:       item.Type,
		EventProperties: item.Properties,
	}

	// Log the event
	t.client.Track(ev)
}

func (t *TelemetryService) Callback(result types.ExecuteResult) {
	if result.Code < 200 || result.Code >= 300 {
		t.ctx.LogErrorf("[Telemetry] Failed to send event to Amplitude: %v", result.Message)
		if result.Code == 401 || result.Code == 403 || result.Message == "Invalid API key" {
			t.ctx.LogErrorf("[Telemetry] Disabling telemetry as received invalid key")
			t.EnableTelemetry = false
		}
	} else {
		t.ctx.LogDebugf("[Telemetry] Event sent to Amplitude")
	}

	t.CallBackChan <- result
}

func (t *TelemetryService) Flush() {
	t.client.Flush()
}

func (t *TelemetryService) Close() {
	t.client.Shutdown()
}
