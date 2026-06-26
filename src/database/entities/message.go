package entities

import "github.com/Parallels/prl-devops-service/database/common"

import (
	"time"

	"gorm.io/gorm"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusPending    MessageStatus = "pending"
	MessageStatusProcessing MessageStatus = "processing"
	MessageStatusCompleted  MessageStatus = "completed"
	MessageStatusFailed     MessageStatus = "failed"
	MessageStatusRetrying   MessageStatus = "retrying"
	MessageStatusAbandoned  MessageStatus = "abandoned"
)

// Message represents a message in the database
type Message struct {
	common.BaseModelWithTenant
	Type        string         `json:"type" gorm:"not null;index"`
	Priority    int            `json:"priority" gorm:"not null;default:1;index"`
	Payload     string         `json:"payload" gorm:"type:text"` // JSON string
	Status      MessageStatus  `json:"status" gorm:"not null;default:'pending';index"`
	RetryCount  int            `json:"retry_count" gorm:"not null;default:0"`
	MaxRetries  int            `json:"max_retries" gorm:"not null;default:3"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty" gorm:"index"`
	ProcessedAt *time.Time     `json:"processed_at,omitempty"`
	FailedAt    *time.Time     `json:"failed_at,omitempty"`
	Error       string         `json:"error,omitempty" gorm:"type:text"`
	WorkerName  string         `json:"worker_name,omitempty" gorm:"index"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// MessageStats represents statistics about messages
type MessageStats struct {
	TotalPending    int64 `json:"total_pending"`
	TotalProcessing int64 `json:"total_processing"`
	TotalCompleted  int64 `json:"total_completed"`
	TotalFailed     int64 `json:"total_failed"`
	TotalRetrying   int64 `json:"total_retrying"`
	TotalAbandoned  int64 `json:"total_abandoned"`
}

// TableName specifies the table name for Message
func (Message) TableName() string {
	return "messages"
}

// MessageEvent represents events that happen to messages
type MessageEvent struct {
	common.BaseModel
	MessageID  string        `json:"message_id" gorm:"not null;index"`
	Message    Message       `json:"message" gorm:"foreignKey:MessageID"`
	EventType  string        `json:"event_type" gorm:"not null"` // created, processing, completed, failed, retrying, abandoned
	Status     MessageStatus `json:"status" gorm:"not null"`
	WorkerName string        `json:"worker_name,omitempty"`
	Error      string        `json:"error,omitempty" gorm:"type:text"`
	Metadata   string        `json:"metadata,omitempty" gorm:"type:text"` // JSON string for additional data
	Timestamp  time.Time     `json:"timestamp" gorm:"not null"`
}

// TableName specifies the table name for MessageEvent
func (MessageEvent) TableName() string {
	return "message_events"
}

// Worker represents a registered worker
type Worker struct {
	common.BaseModel
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Version     string         `json:"version" gorm:"type:text"`
	Type        string         `json:"type" gorm:"not null"` // rabbitmq, interval, hybrid, database
	MessageType string         `json:"message_type" gorm:"not null;index"`
	Interval    *int64         `json:"interval,omitempty" gorm:"type:bigint"` // in seconds, for interval workers
	Enabled     bool           `json:"enabled" gorm:"not null;default:true"`
	IsRunning   bool           `json:"is_running" gorm:"not null;default:false"`
	LastSeen    *time.Time     `json:"last_seen,omitempty"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for Worker
func (Worker) TableName() string {
	return "workers"
}

// WorkerStats represents statistics about workers
type WorkerStats struct {
	WorkerID          string     `json:"worker_id"`
	WorkerName        string     `json:"worker_name"`
	MessagesProcessed int64      `json:"messages_processed"`
	MessagesFailed    int64      `json:"messages_failed"`
	LastProcessedAt   *time.Time `json:"last_processed_at,omitempty"`
	IsRunning         bool       `json:"is_running"`
}

// TableName specifies the table name for WorkerStats
func (WorkerStats) TableName() string {
	return "worker_stats"
}
