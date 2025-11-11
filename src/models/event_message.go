package models

import (
	"time"

	"github.com/Parallels/prl-devops-service/helpers"
)

// EventMessage represents an event that is sent to clients
type EventMessage struct {
	ID        string                 `json:"id"`                  // Unique identifier for the event
	Type      string                 `json:"type"`                // Type/routing key (e.g., pdfm, hi, bye, vm, host, system, global)
	Timestamp time.Time              `json:"timestamp"`           // When the event occurred
	Message   string                 `json:"message"`             // Human-readable message
	Body      map[string]interface{} `json:"body,omitempty"`      // Event-specific data (internal application data)
	ClientID  string                 `json:"client_id,omitempty"` // Optional: Target specific client
}

// NewEventMessage creates a new event message with ID and timestamp
func NewEventMessage(eventType, message string, body map[string]interface{}) *EventMessage {
	return &EventMessage{
		ID:        helpers.GenerateId(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Message:   message,
		Body:      body,
	}
}

// EventClientInfo represents information about a connected WebSocket client
type EventClientInfo struct {
	ID            string    `json:"id"`            // Unique client identifier
	UserID        string    `json:"user_id"`       // User ID from authentication
	Username      string    `json:"username"`      // Username from authentication
	ConnectedAt   time.Time `json:"connected_at"`  // Connection timestamp
	LastPingAt    time.Time `json:"last_ping_at"`  // Last ping sent
	LastPongAt    time.Time `json:"last_pong_at"`  // Last pong received
	Subscriptions []string  `json:"subscriptions"` // List of type subscriptions
	IsAlive       bool      `json:"is_alive"`      // Connection health status
}

// EventEmitterStats represents statistics about the event emitter
type EventEmitterStats struct {
	TotalClients       int               `json:"total_clients"`
	TotalSubscriptions int               `json:"total_subscriptions"`
	TypeStats          map[string]int    `json:"type_stats"`        // Number of subscribers per type
	Clients            []EventClientInfo `json:"clients,omitempty"` // List of connected clients (admin only)
	MessagesSent       int64             `json:"messages_sent"`     // Total messages sent since start
	StartTime          time.Time         `json:"start_time"`        // When the emitter started
	Uptime             string            `json:"uptime"`            // Human-readable uptime
}
