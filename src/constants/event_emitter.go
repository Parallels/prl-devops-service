package constants

// Event Emitter Configuration
const (
	// WebSocket ping/pong configuration
	EVENT_EMITTER_PING_INTERVAL_SECONDS_ENV_VAR = "EVENT_EMITTER_PING_INTERVAL_SECONDS"
	EVENT_EMITTER_PONG_TIMEOUT_SECONDS_ENV_VAR  = "EVENT_EMITTER_PONG_TIMEOUT_SECONDS"

	// Ping interval: how often server sends ping to client
	DEFAULT_EVENT_EMITTER_PING_INTERVAL_SECONDS = 30
	// Pong timeout: how long to wait for pong response (should be > ping interval + network latency)
	DEFAULT_EVENT_EMITTER_PONG_TIMEOUT_SECONDS = 60
)

// EventType is a type-safe wrapper for event types
// This prevents arbitrary strings from being used as event types
type EventType string

// Event Message Types - predefined types for event routing
// Clients subscribe to these types and receive messages of that type
const (
	EventTypeGlobal EventType = "global" // Broadcasts to all subscribers
	EventTypeSystem EventType = "system" // System-level events
	EventTypeVM     EventType = "vm"     // Virtual machine events
	EventTypeHost   EventType = "host"   // Host-level events
	EventTypePDFM   EventType = "pdfm"   // PDFM-specific events
)

func (e EventType) String() string {
	return string(e)
}

// IsValid checks if the EventType is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventTypeGlobal, EventTypeSystem, EventTypeVM, EventTypeHost, EventTypePDFM:
		return true
	default:
		return false
	}
}

// GetAllEventTypes returns all valid EventType values
func GetAllEventTypes() []EventType {
	return []EventType{
		EventTypeGlobal,
		EventTypeSystem,
		EventTypeVM,
		EventTypeHost,
		EventTypePDFM,
	}
}
