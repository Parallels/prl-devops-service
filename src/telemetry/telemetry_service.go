package telemetry

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/amplitude/analytics-go/amplitude"
)

type TelemetryService struct {
	ctx             basecontext.ApiContext
	client          amplitude.Client
	EnableTelemetry bool
}

func (t *TelemetryService) TrackEvent(item TelemetryItem) error {
	if !t.EnableTelemetry {
		t.ctx.LogDebugf("[Telemetry] Telemetry is disabled, ignoring event track")
		return nil
	}

	// If the context is nil, return an error
	if t.ctx == nil {
		return fmt.Errorf("context is nil")
	}
	t.ctx.LogInfof("[Telemetry] Sending Amplitude Tracking")

	// Create a new event
	ev := amplitude.Event{
		UserID:          item.UserID,
		DeviceID:        item.HardwareID,
		EventType:       item.Type,
		EventProperties: item.Properties,
	}

	// Log the event
	t.client.Track(ev)

	return nil
}
