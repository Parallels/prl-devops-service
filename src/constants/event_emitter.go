package constants

// EventType is a type-safe wrapper for event types
// This prevents arbitrary strings from being used as event types
type EventType string

// Event Message Types - predefined types for event routing
// Clients subscribe to these types and receive messages of that type
const (
	EventTypeGlobal       EventType = "global"       // Broadcasts to all subscribers
	EventTypePDFM         EventType = "pdfm"         // PDFM-specific events
	EventTypeSystem       EventType = "system"       // System-level events
	EventTypeHealth       EventType = "health"       // Health check events
	EventTypeOrchestrator EventType = "orchestrator" // Orchestrator events
	EventTypeStats        EventType = "stats"        // Statistics events
)

func (e EventType) String() string {
	return string(e)
}

// IsValid checks if the EventType is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventTypeGlobal, EventTypePDFM, EventTypeSystem, EventTypeHealth, EventTypeOrchestrator, EventTypeStats:
		return true
	default:
		return false
	}
}

// GetAllEventTypes returns all valid EventType values
func GetAllEventTypes() []EventType {
	return []EventType{
		EventTypeGlobal,
		EventTypePDFM,
		EventTypeHealth,
		EventTypeOrchestrator,
		EventTypeStats,
	}
}
