package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestHub_DrainCommands_EmptyQueue(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:         ctx,
		clientToHub: make(chan hubCommand, 256),
	}

	start := time.Now()
	hub.drainCommands()
	duration := time.Since(start)

	// Should return quickly when empty
	assert.Less(t, duration, 100*time.Millisecond)
}

func TestHub_DrainCommands_WithCommands(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clientToHub:   make(chan hubCommand, 256),
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	// Add a command
	cmd := &unregisterClientCmd{clientID: "test"}
	hub.clientToHub <- cmd

	hub.drainCommands()

	// Queue should be empty
	select {
	case <-hub.clientToHub:
		t.Fatal("Queue should be empty")
	default:
		// Expected
	}
}

func TestHub_Shutdown(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		clientsByIP:   make(map[string]string),
		subscriptions: make(map[constants.EventType]map[string]bool),
		shutdownChan:  make(chan struct{}),
	}

	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}
	hub.clients["test-client"] = client
	hub.clientsByIP["192.168.1.1"] = "test-client"
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{"test-client": true}

	hub.shutdown()

	assert.Empty(t, hub.clients)
	assert.Empty(t, hub.clientsByIP)
	assert.Empty(t, hub.subscriptions)

	// Channel should be closed
	_, ok := <-client.Send
	assert.False(t, ok)
}

func TestHub_BroadcastMessage_NoSubscribers(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	msg := models.NewEventMessage(constants.EventTypeVM, "test", nil)
	err := hub.broadcastMessage(msg)

	// Should not error when no subscribers
	assert.NoError(t, err)
}

// TestHub_RegisterClient_WithoutRemoteIP skipped - requires WebSocket connection (integration test)

func TestHub_UnsubscribeClientFromTypes_ClientNotFound(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	result := hub.unsubscribeClientFromTypes("nonexistent", "user1", []constants.EventType{
		constants.EventTypeVM,
	})

	assert.Empty(t, result)
}

func TestHub_UnsubscribeClientFromTypes_NotSubscribed(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	hub.clients["client1"] = &Client{ID: "client1"}
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{}

	result := hub.unsubscribeClientFromTypes("client1", "user1", []constants.EventType{
		constants.EventTypeHost, // Not subscribed to this
	})

	assert.Empty(t, result)
}

func TestEventEmitter_SendToType_NotRunning(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	emitter := &EventEmitter{
		ctx:       ctx,
		isRunning: 0,
	}

	err := emitter.SendToType(constants.EventTypeVM, "test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestEventEmitter_SendToClient_NotRunning(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	emitter := &EventEmitter{
		ctx:       ctx,
		isRunning: 0,
	}

	err := emitter.SendToClient("client1", constants.EventTypeVM, "test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestEventEmitter_SendToAll_NotRunning(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	emitter := &EventEmitter{
		ctx:       ctx,
		isRunning: 0,
	}

	err := emitter.SendToAll("test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestEventEmitter_SendToType_Running(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
		broadcast:     make(chan *models.EventMessage, 10),
	}

	emitter := &EventEmitter{
		ctx:       ctx,
		hub:       hub,
		isRunning: 1,
	}

	err := emitter.SendToType(constants.EventTypeVM, "test message", map[string]string{"key": "value"})

	assert.NoError(t, err)
}

func TestEventEmitter_SendToClient_Running(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
		broadcast:     make(chan *models.EventMessage, 10),
	}

	// Add client to hub
	client := &Client{
		ctx:  ctx,
		ID:   "client1",
		Send: make(chan *models.EventMessage, 10),
	}
	hub.clients["client1"] = client
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{"client1": true}

	emitter := &EventEmitter{
		ctx:       ctx,
		hub:       hub,
		isRunning: 1,
	}

	err := emitter.SendToClient("client1", constants.EventTypeVM, "test message", map[string]string{"key": "value"})

	assert.NoError(t, err)
}

func TestEventEmitter_SendToAll_Running(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
		broadcast:     make(chan *models.EventMessage, 10),
	}

	emitter := &EventEmitter{
		ctx:       ctx,
		hub:       hub,
		isRunning: 1,
	}

	err := emitter.SendToAll("test message", map[string]string{"key": "value"})

	assert.NoError(t, err)
}

func TestEventEmitter_BroadcastMessage_NotRunning(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	emitter := &EventEmitter{
		ctx:       ctx,
		isRunning: 0,
	}

	msg := models.NewEventMessage(constants.EventTypeVM, "test", nil)
	err := emitter.BroadcastMessage(msg)

	// Returns nil but logs warning
	assert.NoError(t, err)
}

func TestEventEmitter_BroadcastMessage_Running(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
		broadcast:     make(chan *models.EventMessage, 10),
	}

	emitter := &EventEmitter{
		ctx:       ctx,
		hub:       hub,
		isRunning: 1,
	}

	msg := models.NewEventMessage(constants.EventTypeVM, "test", map[string]string{"key": "value"})
	err := emitter.BroadcastMessage(msg)

	assert.NoError(t, err)
}

func TestEventEmitter_Shutdown_NilHub(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	emitter := &EventEmitter{
		ctx:       ctx,
		isRunning: 0,
		hub:       nil,
	}

	// Should not panic
	assert.NotPanics(t, func() {
		emitter.Shutdown()
	})
}

func TestHub_UnregisterClient_CleansUpEmptySubscriptions(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		clientsByIP:   make(map[string]string),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}

	hub.clients["test-client"] = client
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{"test-client": true}
	hub.subscriptions[constants.EventTypeHost] = map[string]bool{"test-client": true}

	hub.unregisterClient("test-client")

	// Empty subscription maps should be removed
	_, vmExists := hub.subscriptions[constants.EventTypeVM]
	assert.False(t, vmExists, "Empty VM subscription map should be removed")

	_, hostExists := hub.subscriptions[constants.EventTypeHost]
	assert.False(t, hostExists, "Empty HOST subscription map should be removed")
}

func TestHub_BroadcastMessage_MultipleSubscribers(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	client1 := &Client{
		ctx:  ctx,
		ID:   "client1",
		Send: make(chan *models.EventMessage, 10),
	}
	client2 := &Client{
		ctx:  ctx,
		ID:   "client2",
		Send: make(chan *models.EventMessage, 10),
	}
	client3 := &Client{
		ctx:  ctx,
		ID:   "client3",
		Send: make(chan *models.EventMessage, 10),
	}

	hub.clients["client1"] = client1
	hub.clients["client2"] = client2
	hub.clients["client3"] = client3

	hub.subscriptions[constants.EventTypeVM] = map[string]bool{
		"client1": true,
		"client2": true,
	}
	hub.subscriptions[constants.EventTypeHost] = map[string]bool{
		"client3": true,
	}

	// Broadcast VM message
	msg := models.NewEventMessage(constants.EventTypeVM, "test", nil)
	err := hub.broadcastMessage(msg)
	assert.NoError(t, err)

	// Only client1 and client2 should receive
	assert.Len(t, client1.Send, 1)
	assert.Len(t, client2.Send, 1)
	assert.Len(t, client3.Send, 0)
}
