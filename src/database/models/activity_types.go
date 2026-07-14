package models

type ActivityType string

const (
	ActivityTypeUserLogin     ActivityType = "user_login"
	ActivityTypeUserLogout    ActivityType = "user_logout"
	ActivityTypeUserCreated   ActivityType = "user_created"
	ActivityTypeUserUpdated   ActivityType = "user_updated"
	ActivityTypeUserDeleted   ActivityType = "user_deleted"
	ActivityTypeAPICall       ActivityType = "api_call"
	ActivityTypeSystemEvent   ActivityType = "system_event"
	ActivityTypeConfigChanged ActivityType = "config_changed"
	ActivityTypeVMCreated     ActivityType = "vm_created"
	ActivityTypeVMDeleted     ActivityType = "vm_deleted"
	ActivityTypeVMStarted     ActivityType = "vm_started"
	ActivityTypeVMStopped     ActivityType = "vm_stopped"
)

type ActivityLevel string

const (
	ActivityLevelInfo     ActivityLevel = "info"
	ActivityLevelWarning  ActivityLevel = "warning"
	ActivityLevelError    ActivityLevel = "error"
	ActivityLevelCritical ActivityLevel = "critical"
	ActivityLevelDebug    ActivityLevel = "debug"
)

type ActorType string

const (
	ActorTypeUser    ActorType = "user"
	ActorTypeSystem  ActorType = "system"
	ActorTypeAPIKey  ActorType = "api_key"
	ActorTypeService ActorType = "service"
)

type JSONObject[T any] struct {
	Data T
}

type StringSlice []string

type RecordStatus string

const (
	RecordStatusActive   RecordStatus = "active"
	RecordStatusInactive RecordStatus = "inactive"
	RecordStatusDeleted  RecordStatus = "deleted"
)

type SecurityLevel string

const (
	SecurityLevelPublic       SecurityLevel = "public"
	SecurityLevelInternal     SecurityLevel = "internal"
	SecurityLevelConfidential SecurityLevel = "confidential"
	SecurityLevelSecret       SecurityLevel = "secret"
)
