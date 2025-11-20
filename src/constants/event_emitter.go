package constants

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
	EventTypeHealth EventType = "health" // Health check events
)

func (e EventType) String() string {
	return string(e)
}

// IsValid checks if the EventType is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventTypeGlobal, EventTypeSystem, EventTypeVM, EventTypeHost, EventTypePDFM, EventTypeHealth:
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
		EventTypeHealth,
	}
}
