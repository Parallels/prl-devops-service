package entities

// Activity related types
type ActivityType string
type ActivityLevel string

const (
	ActivityTypeCreate ActivityType  = "create"
	ActivityTypeUpdate ActivityType  = "update"
	ActivityTypeDelete ActivityType  = "delete"
	ActivityTypeRead   ActivityType  = "read"
	
	ActivityLevelInfo    ActivityLevel = "info"
	ActivityLevelWarning ActivityLevel = "warning"
	ActivityLevelError   ActivityLevel = "error"
)

// Actor related types
type ActorType string

const (
	ActorTypeUser   ActorType = "user"
	ActorTypeSystem ActorType = "system"
	ActorTypeApiKey ActorType = "api_key"
)

// Access and Security levels
type AccessLevel string
type SecurityLevel string

const (
	AccessLevelRead  AccessLevel = "read"
	AccessLevelWrite AccessLevel = "write"
	AccessLevelAdmin AccessLevel = "admin"
	
	SecurityLevelLow    SecurityLevel = "low"
	SecurityLevelMedium SecurityLevel = "medium"
	SecurityLevelHigh   SecurityLevel = "high"
)

// Record status
type RecordStatus string

const (
	RecordStatusActive   RecordStatus = "active"
	RecordStatusInactive RecordStatus = "inactive"
	RecordStatusDeleted  RecordStatus = "deleted"
)
