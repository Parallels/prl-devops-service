package constants

// EventType is a type-safe wrapper for event types
// This prevents arbitrary strings from being used as event types
// Clients subscribe to these types and receive messages of that type
type EventType string

const (
	EventTypeGlobal       EventType = "global"        // Broadcasts to all subscribers
	EventTypePDFM         EventType = "pdfm"          // PDFM-specific events
	EventTypeSystem       EventType = "system"        // System-level events
	EventTypeSystemLogs   EventType = "system_logs"   // System logs events
	EventTypeHealth       EventType = "health"        // Health check events
	EventTypeOrchestrator EventType = "orchestrator"  // Orchestrator events
	EventTypeStats        EventType = "stats"         // Statistics events
	EventTypeReverseProxy EventType = "reverse_proxy" // Reverse Proxy events
	EventTypeCatalogCache EventType = "catalog_cache" // Catalog cache events
	EventTypeJobManager   EventType = "job_manager"   // Job Manager events
	EventTypeAuth         EventType = "auth"          // Auth events (users, roles, claims)
)

// Auth event message constants
const (
	EventAuthUserAdded   = "USER_ADDED"
	EventAuthUserUpdated = "USER_UPDATED"
	EventAuthUserRemoved = "USER_REMOVED"

	EventAuthRoleAdded   = "ROLE_ADDED"
	EventAuthRoleRemoved = "ROLE_REMOVED"

	EventAuthRoleClaimAdded   = "ROLE_CLAIM_ADDED"
	EventAuthRoleClaimRemoved = "ROLE_CLAIM_REMOVED"

	EventAuthUserRoleAdded   = "USER_ROLE_ADDED"
	EventAuthUserRoleRemoved = "USER_ROLE_REMOVED"

	EventAuthUserClaimAdded   = "USER_CLAIM_ADDED"
	EventAuthUserClaimRemoved = "USER_CLAIM_REMOVED"

	EventAuthClaimAdded   = "CLAIM_ADDED"
	EventAuthClaimRemoved = "CLAIM_REMOVED"
)

func (e EventType) String() string {
	return string(e)
}

// IsValid checks if the EventType is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventTypeGlobal, EventTypePDFM, EventTypeSystem, EventTypeSystemLogs, EventTypeHealth, EventTypeOrchestrator, EventTypeStats, EventTypeReverseProxy, EventTypeCatalogCache, EventTypeJobManager, EventTypeAuth:
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
		EventTypeSystem,
		EventTypeSystemLogs,
		EventTypeHealth,
		EventTypeOrchestrator,
		EventTypeStats,
		EventTypeReverseProxy,
		EventTypeCatalogCache,
		EventTypeJobManager,
		EventTypeAuth,
	}
}
