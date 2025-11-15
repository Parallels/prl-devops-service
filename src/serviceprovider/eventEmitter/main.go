package eventemitter

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/gorilla/websocket"
)

// hubCommand interface for all commands that modify hub state
type hubCommand interface {
	execute(*Hub)
}

// registerClientCmd registers a new client
type registerClientCmd struct {
	client *Client
}

// unregisterClientCmd unregisters a client
type unregisterClientCmd struct {
	clientID string
}

// unsubscribeCmd unsubscribes client from event types
type unsubscribeCmd struct {
	clientID   string
	userID     string
	eventTypes []constants.EventType
	response   chan unsubscribeResponse
}

type unsubscribeResponse struct {
	unsubscribed []string
	err          error
}

// getClientCmd retrieves client info
type getClientCmd struct {
	clientID string
	userID   string
	response chan getClientResponse
}

type getClientResponse struct {
	client *Client
	err    error
}

// updateClientStateCmd updates client state
type updateClientStateCmd struct {
	clientID    string
	updatePing  bool
	updatePong  bool
	updateAlive bool
	lastPingAt  time.Time
	lastPongAt  time.Time
	isAlive     bool
}

// broadcastCmd broadcasts a message
type broadcastCmd struct {
	message *models.EventMessage
}

// checkIPCmd checks if IP has active connection
type checkIPCmd struct {
	ip       string
	response chan bool
}

// shutdownCmd performs cleanup before hub shutdown
type shutdownCmd struct{}

// EventEmitter is the main service for managing WebSocket event broadcasting
type EventEmitter struct {
	ctx          basecontext.ApiContext
	hub          *Hub
	isRunning    int32 // atomic: 0 = not running, 1 = running
	startTime    time.Time
	messagesSent int64
}

// Hub manages client connections and message broadcasting
// All state modifications happen in the hub goroutine - no locks needed
type Hub struct {
	ctx           basecontext.ApiContext
	clients       map[string]*Client                      // Map of client ID to Client
	clientsByIP   map[string]string                       // Map of IP address to client ID
	subscriptions map[constants.EventType]map[string]bool // Map of event type to set of client IDs (type-safe)
	commands      chan hubCommand                         // Channel for all hub commands (never closed)
	shutdown      chan struct{}                           // Closed to signal shutdown started (never sent to)
	stopped       chan struct{}                           // Closed to signal hub has fully stopped
}

// Client represents a connected WebSocket client
// Immutable fields: ID, User, RemoteIP, ConnectedAt, Conn, Send, hub
// Mutable fields (LastPingAt, LastPongAt, IsAlive, Subscriptions) are modified via hub commands
type Client struct {
	ID            string
	User          *models.ApiUser
	hub           *Hub
	Conn          *websocket.Conn
	Send          chan *models.EventMessage
	Subscriptions []constants.EventType
	RemoteIP      string
	ConnectedAt   time.Time
	LastPingAt    time.Time
	LastPongAt    time.Time
	IsAlive       bool
}

var (
	globalEventEmitter *EventEmitter
	once               sync.Once
)

// NewEventEmitter creates a new EventEmitter instance (singleton)
func NewEventEmitter(ctx basecontext.ApiContext) *EventEmitter {
	once.Do(func() {
		globalEventEmitter = &EventEmitter{
			ctx:       ctx,
			startTime: time.Now(),
		}
	})
	return globalEventEmitter
}

// Get returns the global EventEmitter instance
func Get() *EventEmitter {
	return globalEventEmitter
}

// Initialize starts the event emitter service
func (e *EventEmitter) Initialize() *errors.Diagnostics {
	diag := errors.NewDiagnostics("EventEmitter.Initialize")
	defer diag.Complete()

	cfg := config.Get()

	// Only initialize in API or Orchestrator mode
	if !cfg.IsApi() && !cfg.IsOrchestrator() {
		diag.AddPathEntry("mode_check", "event_emitter")
		e.ctx.LogInfof("[EventEmitter] Not running in API or Orchestrator mode, skipping initialization")
		return diag
	}

	if atomic.LoadInt32(&e.isRunning) == 1 {
		diag.AddWarning("ALREADY_RUNNING", "Event emitter is already running", "event_emitter")
		e.ctx.LogWarnf("[EventEmitter] Already running, skipping initialization")
		return diag
	}

	diag.AddPathEntry("creating_hub", "event_emitter")
	e.ctx.LogInfof("[EventEmitter] Initializing Event Emitter service")

	// Create hub
	e.hub = &Hub{
		ctx:           e.ctx,
		clients:       make(map[string]*Client),
		clientsByIP:   make(map[string]string),
		subscriptions: make(map[constants.EventType]map[string]bool),
		commands:      make(chan hubCommand, 4096),
		shutdown:      make(chan struct{}),
		stopped:       make(chan struct{}),
	}

	// Start hub in background
	go e.hub.run()

	atomic.StoreInt32(&e.isRunning, 1)
	e.ctx.LogInfof("[EventEmitter] Event Emitter service initialized successfully")

	return diag
}

// Shutdown stops the event emitter service
func (e *EventEmitter) Shutdown() {
	if atomic.LoadInt32(&e.isRunning) == 0 {
		return
	}

	e.ctx.LogInfof("[EventEmitter] Shutting down Event Emitter service")

	if e.hub != nil {
		// Close shutdown channel to signal shutdown started
		// This prevents all new commands from being sent
		close(e.hub.shutdown)

		// Send shutdown command to close all clients
		select {
		case e.hub.commands <- &shutdownCmd{}:
		case <-time.After(2 * time.Second):
			e.ctx.LogWarnf("[EventEmitter] Timeout sending shutdown command")
		}

		// Wait for hub to finish draining and signal stopped
		<-e.hub.stopped
		e.ctx.LogInfof("[EventEmitter] Hub goroutine stopped")
	}

	atomic.StoreInt32(&e.isRunning, 0)
	e.ctx.LogInfof("[EventEmitter] Event Emitter service shut down successfully")
}

// IsRunning returns whether the event emitter is running
func (e *EventEmitter) IsRunning() bool {
	return atomic.LoadInt32(&e.isRunning) == 1
}

// run is the main loop for the hub, managing client registration and message broadcasting
func (h *Hub) run() {
	h.ctx.LogInfof("[Hub] Starting hub message routing")
	defer close(h.stopped) // Signal that hub has fully stopped

	for {
		cmd, ok := <-h.commands
		if !ok {
			// Commands channel closed unexpectedly
			h.ctx.LogWarnf("[Hub] Commands channel closed")
			return
		}

		// Check if this is a shutdown command
		if _, isShutdown := cmd.(*shutdownCmd); isShutdown {
			cmd.execute(h)
			h.ctx.LogInfof("[Hub] Shutdown command received, draining remaining commands")
			h.drainCommands()
			h.ctx.LogInfof("[Hub] Hub stopped gracefully")
			return
		}

		cmd.execute(h)
	}
}

// drainCommands processes remaining commands in the queue before shutdown
// Commands channel stays open (never closed) to avoid panics
// Shutdown channel is closed to prevent new commands from being queued
func (h *Hub) drainCommands() {
	drained := 0
	deadline := time.Now().Add(5 * time.Second) // Safety timeout

	for time.Now().Before(deadline) {
		select {
		case cmd := <-h.commands:
			cmd.execute(h)
			drained++
		case <-time.After(100 * time.Millisecond):
			// No more commands for 100ms, consider drain complete
			if drained > 0 {
				h.ctx.LogInfof("[Hub] Drained %d pending commands", drained)
			}
			return
		}
	}

	h.ctx.LogWarnf("[Hub] Drain timeout after %d commands, forcing shutdown", drained)
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	if client == nil {
		h.ctx.LogWarnf("[Hub] Attempted to register nil client")
		return
	}

	h.clients[client.ID] = client

	if client.RemoteIP != "" {
		h.clientsByIP[client.RemoteIP] = client.ID
	}

	// Always subscribe to global type automatically
	if h.subscriptions[constants.EventTypeGlobal] == nil {
		h.subscriptions[constants.EventTypeGlobal] = make(map[string]bool)
	}
	h.subscriptions[constants.EventTypeGlobal][client.ID] = true

	// Add global to client subscriptions if not already present
	if !slices.Contains(client.Subscriptions, constants.EventTypeGlobal) {
		client.Subscriptions = append(client.Subscriptions, constants.EventTypeGlobal)
	}

	// Register other subscriptions (skip invalid and global)
	for _, eventType := range client.Subscriptions {
		if eventType == constants.EventTypeGlobal {
			continue // Already registered above
		}
		// Skip invalid event types
		if !eventType.IsValid() {
			h.ctx.LogWarnf("[Hub] Client %s is subscribing to invalid event type %s, skipping", client.ID, eventType)
			continue
		}
		if h.subscriptions[eventType] == nil {
			h.subscriptions[eventType] = make(map[string]bool)
		}
		h.subscriptions[eventType][client.ID] = true
		h.ctx.LogDebugf("[Hub] Client %s subscribed to type: %s", client.ID, eventType)
	}

	h.ctx.LogInfof("[Hub] Registered client %s (user: %s) with %d subscriptions (global auto-subscribed)",
		client.ID, client.User.Username, len(client.Subscriptions))
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	if client == nil {
		h.ctx.LogWarnf("[Hub] Attempted to unregister nil client")
		return
	}

	if _, exists := h.clients[client.ID]; !exists {
		h.ctx.LogWarnf("[Hub] Attempted to unregister non-existent client %s", client.ID)
		return
	}

	h.ctx.LogInfof("[Hub] Unregistering client %s (user: %s)", client.ID, client.User.Username)

	// Remove from subscriptions
	for _, eventType := range client.Subscriptions {
		if h.subscriptions[eventType] != nil {
			delete(h.subscriptions[eventType], client.ID)
			if len(h.subscriptions[eventType]) == 0 {
				delete(h.subscriptions, eventType)
			}
		}
	}

	if client.RemoteIP != "" {
		delete(h.clientsByIP, client.RemoteIP)
	}

	// Close client connection
	delete(h.clients, client.ID)
	close(client.Send)
}

// trySendCommand safely sends a command to the hub, checking shutdown status
// Returns true if command was sent, false if hub is shutting down or timeout
func (h *Hub) trySendCommand(cmd hubCommand, timeout time.Duration) bool {
	select {
	case <-h.shutdown:
		return false // Shutdown channel closed, don't send
	default:
	}

	select {
	case h.commands <- cmd:
		return true
	case <-h.shutdown:
		return false // Shutdown started during send
	case <-time.After(timeout):
		return false
	}
}

// HasActiveConnectionFromIP checks if there's already an active connection from the given IP
func (h *Hub) HasActiveConnectionFromIP(ip string) bool {
	if ip == "" {
		return false
	}

	// Use command for thread-safe access
	respChan := make(chan bool, 1)
	cmd := &checkIPCmd{ip: ip, response: respChan}

	if !h.trySendCommand(cmd, 1*time.Second) {
		h.ctx.LogWarnf("[Hub] Cannot check IP connection (shutdown or timeout)")
		return false
	}

	select {
	case result := <-respChan:
		return result
	case <-time.After(1 * time.Second):
		h.ctx.LogWarnf("[Hub] Timeout waiting for IP check response")
		return false
	}
}

// broadcastMessage sends a message to appropriate clients based on type and clientID
func (h *Hub) broadcastMessage(message *models.EventMessage) {
	if message == nil {
		h.ctx.LogWarnf("[Hub] Attempted to broadcast nil message")
		return
	}

	// If message targets a specific client
	if message.ClientID != "" {
		if client, exists := h.clients[message.ClientID]; exists {
			select {
			case client.Send <- message:
				h.ctx.LogDebugf("[Hub] Sent message %s to client %s", message.ID, client.ID)
			default:
				h.ctx.LogWarnf("[Hub] Client %s send channel is full, dropping message %s",
					client.ID, message.ID)
			}
		} else {
			h.ctx.LogWarnf("[Hub] Target client %s not found for message %s",
				message.ClientID, message.ID)
		}
		return
	}

	if subscribers, exists := h.subscriptions[message.Type]; exists {
		for clientID := range subscribers {
			if client, exists := h.clients[clientID]; exists {
				select {
				case client.Send <- message:
					h.ctx.LogDebugf("[Hub] Sent message %s to client %s (type: %s)",
						message.ID, client.ID, message.Type)
				default:
					h.ctx.LogWarnf("[Hub] Client %s send channel is full, dropping message %s",
						client.ID, message.ID)
				}
			}
		}
	}
}

// Command implementations
func (c *registerClientCmd) execute(h *Hub) {
	h.registerClient(c.client)
}

func (c *unregisterClientCmd) execute(h *Hub) {
	if client, exists := h.clients[c.clientID]; exists {
		h.unregisterClient(client)
	}
}

func (c *unsubscribeCmd) execute(h *Hub) {
	client, clientExists := h.clients[c.clientID]
	if !clientExists {
		c.response <- unsubscribeResponse{
			unsubscribed: []string{},
			err:          errors.New("client not found"),
		}
		return
	}

	// Security: Verify ownership
	if client.User.ID != c.userID {
		h.ctx.LogWarnf("[Hub] User %s attempted to unsubscribe client %s owned by %s",
			c.userID, c.clientID, client.User.ID)
		c.response <- unsubscribeResponse{
			unsubscribed: []string{},
			err:          errors.New("unauthorized"),
		}
		return
	}

	unsubscribed := make([]string, 0)
	var globalAttempted bool

	for _, eventType := range c.eventTypes {
		if eventType == constants.EventTypeGlobal {
			h.ctx.LogWarnf("[Client %s] Cannot unsubscribe from global event type", c.clientID)
			globalAttempted = true
			continue
		}

		// Remove from hub subscriptions
		if h.subscriptions[eventType] != nil {
			delete(h.subscriptions[eventType], c.clientID)
			if len(h.subscriptions[eventType]) == 0 {
				delete(h.subscriptions, eventType)
			}
		}

		// Remove from client's subscription list
		newSubs := make([]constants.EventType, 0, len(client.Subscriptions))
		for _, sub := range client.Subscriptions {
			if sub != eventType {
				newSubs = append(newSubs, sub)
			}
		}
		client.Subscriptions = newSubs
		unsubscribed = append(unsubscribed, eventType.String())
	}

	if len(unsubscribed) > 0 {
		h.ctx.LogInfof("[Client %s] Unsubscribed from event types: %v", c.clientID, unsubscribed)
	}

	var respErr error
	if globalAttempted {
		respErr = errors.New("cannot unsubscribe from global event type")
	}

	c.response <- unsubscribeResponse{
		unsubscribed: unsubscribed,
		err:          respErr,
	}
}

func (c *getClientCmd) execute(h *Hub) {
	client, exists := h.clients[c.clientID]
	if !exists {
		c.response <- getClientResponse{
			client: nil,
			err:    errors.New("client not found"),
		}
		return
	}

	// Verify ownership
	if client.User.ID != c.userID {
		c.response <- getClientResponse{
			client: nil,
			err:    errors.New("unauthorized"),
		}
		return
	}

	c.response <- getClientResponse{
		client: client,
		err:    nil,
	}
}

func (c *updateClientStateCmd) execute(h *Hub) {
	client, exists := h.clients[c.clientID]
	if !exists {
		return
	}

	if c.updatePing {
		client.LastPingAt = c.lastPingAt
	}
	if c.updatePong {
		client.LastPongAt = c.lastPongAt
	}
	if c.updateAlive {
		client.IsAlive = c.isAlive
	}
}

func (c *broadcastCmd) execute(h *Hub) {
	h.broadcastMessage(c.message)
}

func (c *checkIPCmd) execute(h *Hub) {
	_, exists := h.clientsByIP[c.ip]
	c.response <- exists
}

// getStatsCmd retrieves hub statistics
type getStatsCmd struct {
	includeClients bool
	response       chan *models.EventEmitterStats
}

func (cmd *getStatsCmd) execute(h *Hub) {
	stats := &models.EventEmitterStats{
		TotalClients:       len(h.clients),
		TotalSubscriptions: 0,
		TypeStats:          make(map[constants.EventType]int),
	}

	// Count subscriptions per type
	for eventType, subscribers := range h.subscriptions {
		count := len(subscribers)
		stats.TypeStats[eventType] = count
		stats.TotalSubscriptions += count
	}

	// Include client details if requested
	if cmd.includeClients {
		stats.Clients = make([]models.EventClientInfo, 0, len(h.clients))
		for _, client := range h.clients {
			clientInfo := models.EventClientInfo{
				ID:            client.ID,
				UserID:        client.User.ID,
				Username:      client.User.Username,
				ConnectedAt:   client.ConnectedAt,
				LastPingAt:    client.LastPingAt,
				LastPongAt:    client.LastPongAt,
				Subscriptions: client.Subscriptions,
				IsAlive:       client.IsAlive,
			}
			stats.Clients = append(stats.Clients, clientInfo)
		}
	}

	cmd.response <- stats
}

func (c *shutdownCmd) execute(h *Hub) {
	h.ctx.LogInfof("[Hub] Shutting down, closing %d client connections", len(h.clients))
	for _, client := range h.clients {
		close(client.Send)
		if client.Conn != nil {
			client.Conn.Close()
		}
	}
	// Clear all maps
	h.clients = make(map[string]*Client)
	h.clientsByIP = make(map[string]string)
	h.subscriptions = make(map[constants.EventType]map[string]bool)
	h.ctx.LogInfof("[Hub] All clients disconnected")
}
