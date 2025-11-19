package eventemitter

import (
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestNewEventEmitter_Singleton(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	emitter1 := NewEventEmitter(ctx)
	emitter2 := NewEventEmitter(ctx)

	assert.Same(t, emitter1, emitter2, "Should return same instance")
	assert.NotNil(t, emitter1.ctx)
}

func TestGet(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	NewEventEmitter(ctx)

	retrieved := Get()
	assert.NotNil(t, retrieved)
	assert.Same(t, globalEventEmitter, retrieved)
}

func TestEventEmitter_Initialize_NotAPIOrOrchestrator(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	t.Setenv(constants.MODE_ENV_VAR, "cli")

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)
	diag := emitter.Initialize()

	assert.NotNil(t, diag)
	assert.False(t, emitter.IsRunning())
	assert.Nil(t, emitter.hub)
}

func TestEventEmitter_Initialize_AlreadyRunning(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	t.Setenv(constants.MODE_ENV_VAR, "api")

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)
	emitter.Initialize()
	defer emitter.Shutdown()

	// Try to initialize again
	diag2 := emitter.Initialize()
	assert.True(t, diag2.HasWarnings())
}

func TestEventEmitter_Shutdown_NotRunning(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)

	// Should not panic
	assert.NotPanics(t, func() {
		emitter.Shutdown()
	})
}

func TestEventEmitter_IsRunning(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)

	assert.False(t, emitter.IsRunning())
}

// TestHub_RegisterClient_Success skipped - requires WebSocket connection (integration test)

// TestHub_RegisterClient_NilClient skipped - requires WebSocket connection (integration test)

// TestHub_RegisterClient_DuplicateID skipped - requires WebSocket connection (integration test)

// TestHub_RegisterClient_InvalidEventType skipped - requires WebSocket connection (integration test)

func TestHub_UnregisterClient_Success(t *testing.T) {
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
	hub.clientsByIP["192.168.1.1"] = "test-client"
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{"test-client": true}

	hub.unregisterClient("test-client")

	assert.NotContains(t, hub.clients, "test-client")
	assert.NotContains(t, hub.clientsByIP, "192.168.1.1")
	assert.Empty(t, hub.subscriptions[constants.EventTypeVM])
}

func TestHub_UnregisterClient_NonExistent(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	// Should not panic
	assert.NotPanics(t, func() {
		hub.unregisterClient("nonexistent")
	})
}

func TestHub_BroadcastMessage_ToSpecificClient(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}
	hub.clients["test-client"] = client

	msg := models.NewEventMessage(constants.EventTypeVM, "test", nil)
	msg.ClientID = "test-client"

	err := hub.broadcastMessage(msg)
	assert.NoError(t, err)

	select {
	case received := <-client.Send:
		assert.Equal(t, msg.ID, received.ID)
	default:
		t.Fatal("Client should have received message")
	}
}

func TestHub_BroadcastMessage_ClientNotFound(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	msg := models.NewEventMessage(constants.EventTypeVM, "test", nil)
	msg.ClientID = "nonexistent"

	err := hub.broadcastMessage(msg)
	assert.Error(t, err)
}

func TestHub_BroadcastMessage_ToSubscribers(t *testing.T) {
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

	hub.clients["client1"] = client1
	hub.clients["client2"] = client2
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{
		"client1": true,
		"client2": true,
	}

	msg := models.NewEventMessage(constants.EventTypeVM, "test", nil)
	err := hub.broadcastMessage(msg)
	assert.NoError(t, err)

	// Both should receive
	select {
	case <-client1.Send:
	default:
		t.Fatal("Client1 should have received message")
	}
	select {
	case <-client2.Send:
	default:
		t.Fatal("Client2 should have received message")
	}
}

func TestHub_BroadcastMessage_NilMessage(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	err := hub.broadcastMessage(nil)
	assert.Error(t, err)
}

func TestHub_BroadcastMessage_ChannelFull(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 1),
	}
	hub.clients["test-client"] = client
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{"test-client": true}

	// Fill channel
	client.Send <- models.NewEventMessage(constants.EventTypeVM, "fill", nil)

	// Try to send another - should drop
	msg := models.NewEventMessage(constants.EventTypeVM, "overflow", nil)
	err := hub.broadcastMessage(msg)
	assert.NoError(t, err) // No error, just drops
}

func TestHub_HasActiveConnectionFromIP(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:         ctx,
		clientsByIP: make(map[string]string),
	}

	hub.clientsByIP["192.168.1.1"] = "client1"

	assert.True(t, hub.HasActiveConnectionFromIP("192.168.1.1"))
	assert.False(t, hub.HasActiveConnectionFromIP("192.168.1.2"))
	assert.False(t, hub.HasActiveConnectionFromIP(""))
}

func TestHub_UnsubscribeClientFromTypes(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	hub.clients["client1"] = &Client{ID: "client1"}
	hub.subscriptions[constants.EventTypeVM] = map[string]bool{"client1": true}
	hub.subscriptions[constants.EventTypeHost] = map[string]bool{"client1": true}
	hub.subscriptions[constants.EventTypeGlobal] = map[string]bool{"client1": true}

	result, err := hub.unsubscribeClientFromTypes("client1", "user1", []constants.EventType{
		constants.EventTypeVM,
		constants.EventTypeHost,
	})

	assert.Len(t, result, 2)
	assert.NoError(t, err)
	assert.Contains(t, result, "vm")
	assert.Contains(t, result, "host")
	assert.Empty(t, hub.subscriptions[constants.EventTypeVM])
	assert.Empty(t, hub.subscriptions[constants.EventTypeHost])
	// Global should remain
	assert.Len(t, hub.subscriptions[constants.EventTypeGlobal], 1)
}

func TestHub_UnsubscribeClientFromTypes_CannotUnsubscribeGlobal(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	hub.clients["client1"] = &Client{ID: "client1"}
	hub.subscriptions[constants.EventTypeGlobal] = map[string]bool{"client1": true}

	result, err := hub.unsubscribeClientFromTypes("client1", "user1", []constants.EventType{
		constants.EventTypeGlobal,
	})

	assert.Empty(t, result)
	assert.Error(t, err)
	// Global should still be there
	assert.Len(t, hub.subscriptions[constants.EventTypeGlobal], 1)
}

func TestHub_TrySendCommand_Success(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:          ctx,
		clientToHub:  make(chan hubCommand, 10),
		shutdownChan: make(chan struct{}),
	}

	cmd := &unregisterClientCmd{clientID: "test"}
	result := hub.trySendCommand(cmd)

	assert.True(t, result)
	select {
	case receivedCmd := <-hub.clientToHub:
		assert.Equal(t, cmd, receivedCmd)
	default:
		t.Fatal("Command should be in channel")
	}
}

func TestHub_TrySendCommand_ShutdownInProgress(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:          ctx,
		clientToHub:  make(chan hubCommand, 10),
		shutdownChan: make(chan struct{}),
	}

	close(hub.shutdownChan)

	cmd := &unregisterClientCmd{clientID: "test"}
	result := hub.trySendCommand(cmd)

	assert.False(t, result)
}

func TestUnregisterClientCmd_Execute(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	hub := &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
	}

	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}
	hub.clients["test-client"] = client

	cmd := &unregisterClientCmd{clientID: "test-client"}
	cmd.execute(hub)

	assert.NotContains(t, hub.clients, "test-client")
}
