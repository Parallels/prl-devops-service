package constants

// Event Emitter Configuration
const (
	// WebSocket ping/pong configuration
	EVENT_EMITTER_PING_INTERVAL_SECONDS_ENV_VAR = "EVENT_EMITTER_PING_INTERVAL_SECONDS"
	EVENT_EMITTER_PONG_TIMEOUT_SECONDS_ENV_VAR  = "EVENT_EMITTER_PONG_TIMEOUT_SECONDS"

	// Default values
	DEFAULT_EVENT_EMITTER_PING_INTERVAL_SECONDS = 30
	DEFAULT_EVENT_EMITTER_PONG_TIMEOUT_SECONDS  = 10
)

// Event Message Types - predefined types for event routing
// Clients subscribe to these types and receive messages of that type
const (
	EVENT_TYPE_GLOBAL = "global" // Broadcasts to all subscribers
	EVENT_TYPE_SYSTEM = "system" // System-level events
	EVENT_TYPE_VM     = "vm"     // Virtual machine events
	EVENT_TYPE_HOST   = "host"   // Host-level events
	EVENT_TYPE_PDFM   = "pdfm"   // PDFM-specific events
)

// AllEventTypes returns all valid event types for subscription
var AllEventTypes = []string{
	EVENT_TYPE_GLOBAL,
	EVENT_TYPE_SYSTEM,
	EVENT_TYPE_VM,
	EVENT_TYPE_HOST,
	EVENT_TYPE_PDFM,
}
