package eventemitter

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/gorilla/websocket"
)

// EventEmitter is the main service for managing WebSocket event broadcasting
type EventEmitter struct {
	ctx       basecontext.ApiContext
	hub       *Hub
	isRunning int32 // atomic: 0 = not running, 1 = running
	startTime time.Time
}

// Hub manages client connections and message broadcasting
type Hub struct {
	ctx           basecontext.ApiContext
	clients       map[string]*Client                      // Map of client ID to Client
	clientsByIP   map[string]string                       // Map of IP address to client ID
	subscriptions map[constants.EventType]map[string]bool // Map of event type to set of client IDs (type-safe)
	broadcast     chan *models.EventMessage               // Channel for broadcasting messages
	clientToHub   chan hubCommand                         // Channel for commands from websocket clients
	shutdownChan  chan struct{}                           // Closed to signal shutdown started (never sent to)
	stopped       chan struct{}                           // Closed to signal hub has fully stopped
	mu            sync.RWMutex
}

// Client represents a connected WebSocket client
// Immutable fields: ID, User, RemoteIP, ConnectedAt, Conn, Send, hub
// Mutable fields (LastPingAt, LastPongAt, IsAlive, Subscriptions) are modified via hub commands
type Client struct {
	ctx         basecontext.ApiContext
	ID          string
	User        *models.ApiUser
	Conn        *websocket.Conn
	Send        chan *models.EventMessage
	ConnectedAt time.Time
	LastPingAt  time.Time
	LastPongAt  time.Time
	IsAlive     bool
	mu          sync.RWMutex
}

var (
	globalEventEmitter *EventEmitter
	once               sync.Once
)

// hubCommand interface for clients to send commands to the hub
type hubCommand interface {
	execute(*Hub)
}

// unregisterClientCmd unregister a client
type unregisterClientCmd struct {
	clientID string
}

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

func (e *EventEmitter) Initialize() *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("EventEmitter.Initialize")
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
		clientToHub:   make(chan hubCommand, 4096),
		broadcast:     make(chan *models.EventMessage, 4096),
		shutdownChan:  make(chan struct{}),
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

		e.hub.drainCommands()
		e.hub.shutdown()
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
		select {
		case <-h.shutdownChan:
			// Shutdown signal received
			h.ctx.LogInfof("[Hub] Shutdown signal received, exiting run loop")
			return
		case cmd, ok := <-h.clientToHub:
			if !ok {
				// Commands channel closed unexpectedly
				h.ctx.LogWarnf("[Hub] Commands channel closed")
				return
			}
			cmd.execute(h)
		case msg, ok := <-h.broadcast:
			if ok {
				h.broadcastMessage(msg)
			}
		}
	}
}

// drainCommands processes remaining commands in the queue before shutdown
// Commands channel stays open (never closed) to avoid panics
// Shutdown channel is closed to prevent new commands from being queued
func (h *Hub) drainCommands() {
	drained := 0
	deadline := time.Now().Add(30 * time.Millisecond) // Safety timeout

	for time.Now().Before(deadline) {
		select {
		case cmd := <-h.clientToHub:
			cmd.execute(h)
			drained++
		default:
			// No more commands, exit early
			if drained > 0 {
				h.ctx.LogInfof("[Hub] Drained %d pending commands", drained)
			}
			return
		}
	}
	if drained > 0 {
		h.ctx.LogInfof("[Hub] Drained %d pending commands", drained)
	}
	h.ctx.LogWarnf("[Hub] Drain timeout after %d commands, forcing shutdown", drained)
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client, subscriptions []constants.EventType, remoteIP string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client == nil || h.clients[client.ID] != nil {
		h.ctx.LogWarnf("[Hub] Attempted to register nil client or duplicate client ID")
		return false
	}

	h.clients[client.ID] = client
	if remoteIP != "" {
		h.clientsByIP[remoteIP] = client.ID
	}

	// Always subscribe to global type automatically
	if h.subscriptions[constants.EventTypeGlobal] == nil {
		h.subscriptions[constants.EventTypeGlobal] = make(map[string]bool)
	}
	h.subscriptions[constants.EventTypeGlobal][client.ID] = true

	// Register other subscriptions (skip invalid and global)
	for _, eventType := range subscriptions {
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
	rsp := models.WebSocketConnectResponse{
		ClientID:      client.ID,
		Subscriptions: subscriptions,
	}
	// Check if EventTypeGlobal is already in subscriptions
	hasGlobal := false
	for _, et := range subscriptions {
		if et == constants.EventTypeGlobal {
			hasGlobal = true
			break
		}
	}
	if hasGlobal {
		rsp.Subscriptions = subscriptions
	} else {
		rsp.Subscriptions = append(subscriptions, constants.EventTypeGlobal)
	}
	// Start client goroutines
	go client.clientWriter()
	go client.clientReader()

	client.Send <- models.NewEventMessage(constants.EventTypeGlobal, "WebSocket connection established subscribed to global by default", rsp)

	h.ctx.LogInfof("[Hub] Registered client %s (user: %s) with %d subscriptions %v + (global auto-subscribed)",
		client.ID, client.User.Username, len(rsp.Subscriptions), rsp.Subscriptions)
	return true
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[clientID]; !exists {
		h.ctx.LogWarnf("[Hub] Attempted to unregister non-existent client %s", clientID)
		return
	}

	h.ctx.LogInfof("[Hub] Unregistering client %s", clientID)

	// Get client reference before removal
	client := h.clients[clientID]

	// Remove client from subscriptions
	for eventType, subscribers := range h.subscriptions {
		if _, subscribed := subscribers[clientID]; subscribed {
			delete(subscribers, clientID)
			h.ctx.LogDebugf("[Hub] Client %s unsubscribed from type: %s", clientID, eventType)
			if len(subscribers) == 0 {
				delete(h.subscriptions, eventType)
			}
		}
	}

	// Remove from clientsByIP map
	for ip, cid := range h.clientsByIP {
		if cid == clientID {
			h.ctx.LogInfof("removed %s, %s", ip, cid)
			delete(h.clientsByIP, ip)
			break
		}
	}

	delete(h.clients, clientID)

	// Close the client's Send channel (hub owns this channel)
	// This is safe because client is removed from h.clients,
	// so no new messages will be sent to this channel
	if client != nil && client.Send != nil {
		close(client.Send)
	}
}

// trySendCommand safely sends a command to the hub, checking shutdown status
// Returns true if command was sent, false if hub is shutting down or timeout
func (h *Hub) trySendCommand(cmd hubCommand) bool {
	select {
	case <-h.shutdownChan:
		return false // Shutdown channel closed, don't send
	default:
	}

	select {
	case h.clientToHub <- cmd:
		return true
	case <-h.shutdownChan:
		h.ctx.LogWarnf("[Hub] Cannot send command, hub is shutting down")
		return false
	}
}

// HasActiveConnectionFromIP checks if there's already an active connection from the given IP
func (h *Hub) HasActiveConnectionFromIP(ip string) bool {
	if ip == "" {
		return false
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.clientsByIP[ip]
	return exists
}

// broadcastMessage sends a message to appropriate clients based on type and clientID
func (h *Hub) broadcastMessage(message *models.EventMessage) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if message == nil {
		h.ctx.LogWarnf("[Hub] Attempted to broadcast nil message")
		return errors.New("nil message")
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
				// Drop message - channel is full
			}
		} else {
			h.ctx.LogWarnf("[Hub] Target client %s not found for message %s",
				message.ClientID, message.ID)
			return fmt.Errorf("%s client not found", message.ClientID)
		}
		return nil
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
					// Drop message - channel is full
				}
			}
		}
	}
	return nil
}

func (h *Hub) unsubscribeClientFromTypes(clientID string, userId string, eventTypes []constants.EventType) ([]string, error) {

	h.mu.Lock()
	defer h.mu.Unlock()
	unsubscribed := []string{}
	err := error(nil)
	_, clientExists := h.clients[clientID]
	if !clientExists {
		h.ctx.LogWarnf("[Client %s] Attempted to unsubscribe but client does not exist", clientID)
		return unsubscribed, fmt.Errorf("client %s does not exist", clientID)
	}
	var globalAttempted bool
	for _, eventType := range eventTypes {
		if eventType == constants.EventTypeGlobal {
			globalAttempted = true
			continue
		}
		// Remove from hub subscriptions
		if h.subscriptions[eventType] != nil {
			delete(h.subscriptions[eventType], clientID)
			if len(h.subscriptions[eventType]) == 0 {
				delete(h.subscriptions, eventType)
				unsubscribed = append(unsubscribed, eventType.String())
			} else {
				unsubscribed = append(unsubscribed, eventType.String())
			}
		} else {
			h.ctx.LogWarnf("[Client %s] Not subscribed to event type %s, cannot unsubscribe", clientID, eventType)
			err = fmt.Errorf("not subscribed to event type %s", eventType)
		}
	}
	if len(unsubscribed) > 0 {
		h.ctx.LogInfof("[Client %s] Unsubscribed from event types: %v", clientID, unsubscribed)
	}

	if globalAttempted {
		h.ctx.LogWarnf("[Client %s] Cannot unsubscribe from global event type", clientID)
		err = fmt.Errorf("cannot unsubscribe from %s event type", constants.EventTypeGlobal)
	}
	return unsubscribed, err
}

func (c *unregisterClientCmd) execute(h *Hub) {
	h.unregisterClient(c.clientID)
}

func (h *Hub) shutdown() {
	h.ctx.LogInfof("[Hub] Shutting down, closing %d client connections", len(h.clients))
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all client connections
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
	close(h.shutdownChan)
}
