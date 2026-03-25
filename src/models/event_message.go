package models

import (
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
)

// EventMessage represents an event that is sent to clients
type EventMessage struct {
	ID        string              `json:"id"`                  // Unique identifier for the event
	RefID     string              `json:"ref_id,omitempty"`    // Optional: Reference to a previous event ID (for replies)
	Type      constants.EventType `json:"event_type"`          // Type/routing key (e.g., pdfm, vm, host, system, global)
	Timestamp time.Time           `json:"timestamp"`           // When the event occurred
	Message   string              `json:"message"`             // Human-readable message
	Body      interface{}         `json:"body,omitempty"`      // Event-specific data (internal application data)
	ClientID  string              `json:"client_id,omitempty"` // Optional: Target specific client
}

// NewEventMessage creates a new event message with ID and timestamp
// Uses type-safe EventType to prevent arbitrary strings
func NewEventMessage(eventType constants.EventType, message string, body interface{}) *EventMessage {
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
	ID            string                `json:"id"`                  // Unique client identifier
	UserID        string                `json:"user_id"`             // User ID from authentication
	Username      string                `json:"username"`            // Username from authentication
	ConnectedAt   time.Time             `json:"connected_at"`        // Connection timestamp
	LastPingAt    time.Time             `json:"last_ping_at"`        // Last ping sent
	LastPongAt    time.Time             `json:"last_pong_at"`        // Last pong received
	Subscriptions []constants.EventType `json:"event_subscriptions"` // List of type subscriptions
	IsAlive       bool                  `json:"is_alive"`            // Connection health status
	QueueDepth    int                   `json:"queue_depth"`         // Number of pending outbound messages
}

// EventEmitterStats represents statistics about the event emitter
type EventEmitterStats struct {
	TotalClients       int                         `json:"total_clients"`
	TotalSubscriptions int                         `json:"total_subscriptions"`
	TypeStats          map[constants.EventType]int `json:"type_stats"`        // Number of subscribers per type
	Clients            []EventClientInfo           `json:"clients,omitempty"` // List of connected clients (admin only)
	MessagesSent       int64                       `json:"messages_sent"`     // Total messages sent since start
	StartTime          time.Time                   `json:"start_time"`        // When the emitter started
	Uptime             string                      `json:"uptime"`            // Human-readable uptime
}

type UnsubscribeRequest struct {
	ClientID   string   `json:"client_id"`             // Unique client identifier
	UserID     string   `json:"user_id"`               // User ID from authentication
	EventTypes []string `json:"event_types,omitempty"` // List of event types to unsubscribe from
}

type WebSocketConnectResponse struct {
	ClientID      string                `json:"client_id"`     // Unique client identifier
	Subscriptions []constants.EventType `json:"subscriptions"` // List of event types the client is subscribed to
}

type VmStateChange struct {
	PreviousState string `json:"previous_state"`
	CurrentState  string `json:"current_state"`
	VmID          string `json:"vm_id"`
}

type VmAdded struct {
	VmID  string      `json:"vm_id"`
	NewVm ParallelsVM `json:"new_vm"`
}
type VmRemoved struct {
	VmID string `json:"vm_id"`
}

type VmUpdated struct {
	VmID  string      `json:"vm_id"`
	NewVm ParallelsVM `json:"new_vm"`
}

type VmSnapshotsUpdated struct {
	VmID        string       `json:"vm_id"`
	VMSnapshots []VMSnapshot `json:"snapshots"`
}
type VmUptimeChanged struct {
	VmID   string `json:"vm_id"`
	Uptime string `json:"uptime"`
}

type HostHealthUpdate struct {
	HostID string `json:"host_id"`
	State  string `json:"state"`
}

type HostVmEvent struct {
	HostID string      `json:"host_id"`
	Event  interface{} `json:"event"` // VmStateChange, VmAdded, or VmRemoved
}

type HostStatsUpdate struct {
	HostID string      `json:"host_id"`
	Stats  interface{} `json:"stats"`
}

type HostLogsUpdate struct {
	HostID string      `json:"host_id"`
	Log    interface{} `json:"log"`
}

type ReverseProxyForwardEvent struct {
	ReverseProxyHostId string `json:"reverse_proxy_host_id,omitempty"`
	TargetVmId         string `json:"target_vm_id,omitempty"`
	TargetHost         string `json:"target_host,omitempty"`
	TargetPort         string `json:"target_port,omitempty"`
	Path               string `json:"path,omitempty"`
	TrafficType        string `json:"traffic_type"`
	InternalIpAddress  string `json:"internal_ip_address,omitempty"`
	Method             string `json:"method,omitempty"`
	SourceIp           string `json:"source_ip,omitempty"`
}

type ReverseProxyRouteUpdatedEvent struct {
	ReverseProxyHostId string `json:"reverse_proxy_host_id,omitempty"`
	TargetVmId         string `json:"target_vm_id,omitempty"`
	InternalIpAddress  string `json:"internal_ip_address,omitempty"`
}

type ReverseProxyRouteFailedEvent struct {
	ReverseProxyHostId string `json:"reverse_proxy_host_id,omitempty"`
	TargetVmId         string `json:"target_vm_id,omitempty"`
	InternalIpAddress  string `json:"internal_ip_address,omitempty"`
}

type CacheItemAddedEvent struct {
	CatalogId    string `json:"catalog_id"`
	Version      string `json:"version"`
	Architecture string `json:"architecture,omitempty"`
	CacheSize    int64  `json:"cache_size"`
	CacheType    string `json:"cache_type"`
	CachedDate   string `json:"cached_date,omitempty"`
}

type CacheItemRemovedEvent struct {
	CatalogId    string `json:"catalog_id"`
	Version      string `json:"version"`
	Architecture string `json:"architecture,omitempty"`
}

type MacVMsRunningNowEvent struct {
	MacVmsRunning []string `json:"mac_vms_running"`
}

// ── Auth event bodies ──────────────────────────────────────────────────────

type AuthUserEvent struct {
	UserID string `json:"user_id"`
}

type AuthRoleEvent struct {
	RoleID string `json:"role_id"`
}

type AuthClaimEvent struct {
	ClaimID string `json:"claim_id"`
}

type AuthRoleClaimEvent struct {
	RoleID  string `json:"role_id"`
	ClaimID string `json:"claim_id"`
}

type AuthUserRoleEvent struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

type AuthUserClaimEvent struct {
	UserID  string `json:"user_id"`
	ClaimID string `json:"claim_id"`
}

type HostAddedEvent struct {
	HostID      string `json:"host_id"`
	Host        string `json:"host"`
	Description string `json:"description,omitempty"`
}

type HostRemovedEvent struct {
	HostID string `json:"host_id"`
	Host   string `json:"host,omitempty"`
}

type HostDeployedEvent struct {
	HostID  string `json:"host_id"`
	Host    string `json:"host,omitempty"`
	Message string `json:"message,omitempty"`
}
