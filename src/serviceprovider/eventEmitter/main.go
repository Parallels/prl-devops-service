package eventemitter

import (
	"slices"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/gorilla/websocket"
)

// EventEmitter is the main service for managing WebSocket event broadcasting
type EventEmitter struct {
	ctx          basecontext.ApiContext
	hub          *Hub
	isRunning    bool
	startTime    time.Time
	messagesSent int64
	mu           sync.RWMutex
}

// Hub manages client connections and message broadcasting
type Hub struct {
	ctx           basecontext.ApiContext
	clients       map[string]*Client         // Map of client ID to Client
	subscriptions map[string]map[string]bool // Map of event type to set of client IDs
	broadcast     chan *models.EventMessage  // Channel for broadcasting messages
	register      chan *Client               // Channel for registering new clients
	unregister    chan *Client               // Channel for unregistering clients
	done          chan struct{}              // Channel to signal hub shutdown
	mu            sync.RWMutex
}

// Client represents a connected WebSocket client
type Client struct {
	ID            string
	UserID        string
	Username      string
	Hub           *Hub
	Conn          *websocket.Conn
	Send          chan *models.EventMessage
	Subscriptions []string
	ConnectedAt   time.Time
	LastPingAt    time.Time
	LastPongAt    time.Time
	IsAlive       bool
	mu            sync.RWMutex
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

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
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
		subscriptions: make(map[string]map[string]bool),
		broadcast:     make(chan *models.EventMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		done:          make(chan struct{}),
	}

	// Start hub in background
	go e.hub.run()

	e.isRunning = true
	e.ctx.LogInfof("[EventEmitter] Event Emitter service initialized successfully")

	return diag
}

// Shutdown stops the event emitter service
func (e *EventEmitter) Shutdown() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning {
		return
	}

	e.ctx.LogInfof("[EventEmitter] Shutting down Event Emitter service")

	// Signal hub to stop before closing channels
	if e.hub != nil {
		// Signal hub goroutine to stop
		close(e.hub.done)

		// Give hub a moment to process the shutdown signal
		time.Sleep(50 * time.Millisecond)

		e.hub.mu.Lock()
		// Close all client connections
		for _, client := range e.hub.clients {
			close(client.Send)
			if client.Conn != nil {
				client.Conn.Close()
			}
		}
		e.hub.mu.Unlock()

		// Close channels after hub has stopped
		close(e.hub.broadcast)
		close(e.hub.register)
		close(e.hub.unregister)
	}

	e.isRunning = false
	e.ctx.LogInfof("[EventEmitter] Event Emitter service shut down successfully")
}

// IsRunning returns whether the event emitter is running
func (e *EventEmitter) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isRunning
}

// run is the main loop for the hub, managing client registration and message broadcasting
func (h *Hub) run() {
	h.ctx.LogInfof("[Hub] Starting hub message routing")

	for {
		select {
		case <-h.done:
			h.ctx.LogInfof("[Hub] Hub shutdown signal received, stopping message routing")
			return

		case client, ok := <-h.register:
			if !ok {
				h.ctx.LogWarnf("[Hub] Register channel closed")
				return
			}
			h.registerClient(client)

		case client, ok := <-h.unregister:
			if !ok {
				h.ctx.LogWarnf("[Hub] Unregister channel closed")
				return
			}
			h.unregisterClient(client)

		case message, ok := <-h.broadcast:
			if !ok {
				h.ctx.LogWarnf("[Hub] Broadcast channel closed")
				return
			}
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	if client == nil {
		h.ctx.LogWarnf("[Hub] Attempted to register nil client")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	// Always subscribe to global type automatically
	if h.subscriptions[constants.EVENT_TYPE_GLOBAL] == nil {
		h.subscriptions[constants.EVENT_TYPE_GLOBAL] = make(map[string]bool)
	}
	h.subscriptions[constants.EVENT_TYPE_GLOBAL][client.ID] = true

	// Add global to client subscriptions if not already present
	if !slices.Contains(client.Subscriptions, constants.EVENT_TYPE_GLOBAL) {
		client.Subscriptions = append(client.Subscriptions, constants.EVENT_TYPE_GLOBAL)
	}

	for _, eventType := range client.Subscriptions {
		if !slices.Contains(constants.AllEventTypes, eventType) {
			h.ctx.LogWarnf("[Hub] Client %s subscribed to unknown event type: %s", client.ID, eventType)

			msg := models.NewEventMessage(constants.EVENT_TYPE_GLOBAL, "Client subscribed to unsupported event type: "+eventType, nil)
			msg.ClientID = client.ID
			// Send async to avoid blocking registration
			go func(m *models.EventMessage) {
				select {
				case h.broadcast <- m:
				case <-time.After(100 * time.Millisecond):
					h.ctx.LogWarnf("[Hub] Failed to send validation error message (timeout)")
				}
			}(msg)
			continue
		}
	}

	h.ctx.LogInfof("[Hub] Registered client %s (user: %s) with %d subscriptions (global auto-subscribed)",
		client.ID, client.Username, len(client.Subscriptions))

	// Register other subscriptions (skip invalid and global)
	for _, eventType := range client.Subscriptions {
		if eventType == constants.EVENT_TYPE_GLOBAL {
			continue // Already registered above
		}
		// Skip invalid event types
		if !slices.Contains(constants.AllEventTypes, eventType) {
			continue
		}
		if h.subscriptions[eventType] == nil {
			h.subscriptions[eventType] = make(map[string]bool)
		}
		h.subscriptions[eventType][client.ID] = true
		h.ctx.LogDebugf("[Hub] Client %s subscribed to type: %s", client.ID, eventType)
	}
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	if client == nil {
		h.ctx.LogWarnf("[Hub] Attempted to unregister nil client")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.ID]; !exists {
		h.ctx.LogWarnf("[Hub] Attempted to unregister non-existent client %s", client.ID)
		return
	}

	h.ctx.LogInfof("[Hub] Unregistering client %s (user: %s)", client.ID, client.Username)

	// Remove from subscriptions
	for _, eventType := range client.Subscriptions {
		if h.subscriptions[eventType] != nil {
			delete(h.subscriptions[eventType], client.ID)
			if len(h.subscriptions[eventType]) == 0 {
				delete(h.subscriptions, eventType)
			}
		}
	}

	// Close client connection
	delete(h.clients, client.ID)
	close(client.Send)
}

// broadcastMessage sends a message to appropriate clients based on type and clientID
func (h *Hub) broadcastMessage(message *models.EventMessage) {
	if message == nil {
		h.ctx.LogWarnf("[Hub] Attempted to broadcast nil message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

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

	// Broadcast to type subscribers
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
