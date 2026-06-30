package models

import "time"

type Activity struct {
	BaseModel
	// Core activity information
	ActivityType  ActivityType  `json:"activity_type" gorm:"not null;type:text;index"`  // e.g., "user_login", "api_call", "system_event"
	ActivityLevel ActivityLevel `json:"activity_level" gorm:"not null;type:text;index"` // "info", "warning", "error", "critical"
	Message       string        `json:"message" gorm:"not null;type:text"`
	Service       string        `json:"service" gorm:"not null;type:text;index"` // e.g., "user_service", "certificate_service"
	Module        string        `json:"module" gorm:"not null;type:text;index"`  // e.g., "auth", "pipeline", "infrastructure"

	// Actor information
	ActorType ActorType `json:"actor_type" gorm:"not null;type:text;index"` // "user", "system", "api_key", "service"
	ActorID   string    `json:"actor_id" gorm:"type:text;index"`            // ID of the user, API key, or service
	ActorName string    `json:"actor_name" gorm:"type:text"`                // Human-readable name
	ActorIP   string    `json:"actor_ip" gorm:"type:text"`                  // IP address of the actor
	UserAgent string    `json:"user_agent" gorm:"type:text"`                // User agent string

	// Context information
	RequestID     string `json:"request_id" gorm:"type:text;index"`     // Request identifier for tracing
	CorrelationID string `json:"correlation_id" gorm:"type:text;index"` // Correlation ID for distributed tracing

	// Additional data
	Metadata JSONObject[map[string]interface{}] `json:"metadata" gorm:"type:text"` // Additional structured data
	Tags     StringSlice                        `json:"tags" gorm:"type:text"`     // Searchable tags

	// Timing information
	StartedAt   *time.Time `json:"started_at" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	CompletedAt *time.Time `json:"completed_at" gorm:"type:timestamp"` // For long-running activities
	DurationMs  int64      `json:"duration_ms" gorm:"type:bigint"`     // Duration in milliseconds

	// Outcome information
	Success      bool   `json:"success" gorm:"type:boolean;not null;default:true"`
	ErrorCode    string `json:"error_code" gorm:"type:text"`    // Error code if failed
	ErrorMessage string `json:"error_message" gorm:"type:text"` // Error message if failed
	StatusCode   int    `json:"status_code" gorm:"type:int"`    // HTTP status code for API calls

	// Security and compliance
	IsSensitive   bool `json:"is_sensitive" gorm:"type:boolean;not null;default:false"` // Whether this activity contains sensitive data
	RetentionDays int  `json:"retention_days" gorm:"type:int;not null;default:90"`      // How long to retain this record
}

// ActivitySummary represents aggregated activity data for reporting
type ActivitySummary struct {
	BaseModel
	SummaryType string    `json:"summary_type" gorm:"not null;type:text;index"` // "daily", "weekly", "monthly"
	SummaryDate time.Time `json:"summary_date" gorm:"type:date;not null;index"`
	Module      string    `json:"module" gorm:"not null;type:text;index"`
	Service     string    `json:"service" gorm:"not null;type:text;index"`

	// Aggregated counts
	TotalActivities int64 `json:"total_activities" gorm:"type:bigint;not null;default:0"`
	SuccessCount    int64 `json:"success_count" gorm:"type:bigint;not null;default:0"`
	ErrorCount      int64 `json:"error_count" gorm:"type:bigint;not null;default:0"`

	// Actor statistics
	UniqueActors int64                                `json:"unique_actors" gorm:"type:bigint;not null;default:0"`
	TopActors    JSONObject[[]map[string]interface{}] `json:"top_actors" gorm:"type:text"`

	// Performance metrics
	AvgDurationMs float64 `json:"avg_duration_ms" gorm:"type:float"`
	MaxDurationMs int64   `json:"max_duration_ms" gorm:"type:bigint"`
	MinDurationMs int64   `json:"min_duration_ms" gorm:"type:bigint"`

	// Activity type breakdown
	ActivityBreakdown JSONObject[map[string]int64] `json:"activity_breakdown" gorm:"type:text"`
}

// ActivityFilter represents filtering options for activity queries
type ActivityFilter struct {
	Module        []string   `json:"module"`
	Service       []string   `json:"service"`
	ActivityType  []string   `json:"activity_type"`
	ActivityLevel []string   `json:"activity_level"`
	ActorType     []string   `json:"actor_type"`
	ActorID       []string   `json:"actor_id"`
	TargetType    []string   `json:"target_type"`
	TargetID      []string   `json:"target_id"`
	TenantID      []string   `json:"tenant_id"`
	Success       *bool      `json:"success"`
	IsSensitive   *bool      `json:"is_sensitive"`
	Tags          []string   `json:"tags"`
	StartedAtFrom *time.Time `json:"started_at_from"`
	StartedAtTo   *time.Time `json:"started_at_to"`
	CreatedAtFrom *time.Time `json:"created_at_from"`
	CreatedAtTo   *time.Time `json:"created_at_to"`
}

func (Activity) TableName() string {
	return "activities"
}

func (ActivitySummary) TableName() string {
	return "activity_summaries"
}
