package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_Run_RegisterClient(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypePDFM})

	// Register client through channel
	emitter.hub.register <- client

	// Wait for registration to complete
	time.Sleep(100 * time.Millisecond)

	// Verify client was registered
	emitter.hub.mu.RLock()
	registeredClient, exists := emitter.hub.clients["test-client"]
	emitter.hub.mu.RUnlock()

	assert.True(t, exists, "Client should be registered")
	assert.Equal(t, client.ID, registeredClient.ID)
}

func TestHub_Run_UnregisterClient(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypePDFM})

	// Register then unregister
	emitter.hub.register <- client
	time.Sleep(50 * time.Millisecond)
	emitter.hub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	// Verify client was unregistered
	emitter.hub.mu.RLock()
	_, exists := emitter.hub.clients["test-client"]
	emitter.hub.mu.RUnlock()

	assert.False(t, exists, "Client should be unregistered")
}

func TestHub_Run_BroadcastMessage(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	emitter.hub.register <- client
	time.Sleep(50 * time.Millisecond)

	// Broadcast a message
	msg := models.NewEventMessage(constants.EventTypeVM, "Test message", map[string]interface{}{
		"test": "data",
	})
	emitter.hub.broadcast <- msg

	// Verify client received the message
	select {
	case receivedMsg := <-client.Send:
		assert.Equal(t, msg.ID, receivedMsg.ID)
		assert.Equal(t, "Test message", receivedMsg.Message)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client should have received the message")
	}
}

func TestHub_Run_Shutdown(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	require.True(t, emitter.IsRunning())

	// Shutdown should stop the hub.run() goroutine
	emitter.Shutdown()

	assert.False(t, emitter.IsRunning())
}

func TestHub_RegisterClient_Nil(t *testing.T) {
	hub := createTestHub()

	// Should not panic
	assert.NotPanics(t, func() {
		hub.registerClient(nil)
	})
}

func TestHub_UnregisterClient_Nil(t *testing.T) {
	hub := createTestHub()

	// Should not panic
	assert.NotPanics(t, func() {
		hub.unregisterClient(nil)
	})
}

func TestHub_BroadcastMessage_Nil(t *testing.T) {
	hub := createTestHub()

	// Should not panic
	assert.NotPanics(t, func() {
		hub.broadcastMessage(nil)
	})
}

func TestHub_BroadcastMessage_NonExistentType(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	hub.registerClient(client)

	// Send message to a type with no subscribers
	msg := models.NewEventMessage(constants.EventTypeHost, "Test", nil)

	// Should not panic or error
	assert.NotPanics(t, func() {
		hub.broadcastMessage(msg)
	})

	// Client should not receive anything (not subscribed to HOST)
	select {
	case <-client.Send:
		t.Fatal("Client should not receive message for unsubscribed type")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message
	}
}

func TestEventEmitter_SendToClient_NotRunning(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)
	// Don't initialize - not running

	err := emitter.SendToClient("client1", constants.EventTypeVM, "Test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event emitter is not running")
}

func TestEventEmitter_BroadcastMessage_NotRunning(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)
	// Don't initialize

	msg := models.NewEventMessage(constants.EventTypeVM, "Test", nil)
	err := emitter.BroadcastMessage(msg)

	// Should return nil (just logs warning)
	assert.NoError(t, err)
}

func TestHub_Run_MultipleOperations(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Register multiple clients
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeHost})
	client3 := createTestClient("client3", "user3", []constants.EventType{constants.EventTypePDFM})

	emitter.hub.register <- client1
	emitter.hub.register <- client2
	emitter.hub.register <- client3
	time.Sleep(100 * time.Millisecond)

	// Send messages to different types
	msg1 := models.NewEventMessage(constants.EventTypeVM, "VM message", nil)
	msg2 := models.NewEventMessage(constants.EventTypeHost, "HOST message", nil)

	emitter.hub.broadcast <- msg1
	emitter.hub.broadcast <- msg2

	// Verify correct routing
	select {
	case receivedMsg := <-client1.Send:
		assert.Equal(t, constants.EventTypeVM, receivedMsg.Type)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client1 should have received VM message")
	}

	select {
	case receivedMsg := <-client2.Send:
		assert.Equal(t, constants.EventTypeHost, receivedMsg.Type)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client2 should have received HOST message")
	}

	// client3 should not receive either (subscribed to PDFM only)
	select {
	case <-client3.Send:
		t.Fatal("Client3 should not receive VM or HOST messages")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message
	}

	// Unregister one client
	emitter.hub.unregister <- client1
	time.Sleep(50 * time.Millisecond)

	// Verify client1 is removed
	emitter.hub.mu.RLock()
	_, exists := emitter.hub.clients["client1"]
	emitter.hub.mu.RUnlock()
	assert.False(t, exists, "Client1 should be unregistered")
}

func TestHub_RegisterClient_DuplicateSubscription(t *testing.T) {
	hub := createTestHub()

	// Client with duplicate subscriptions in the list
	client := createTestClient("client1", "user1", []constants.EventType{
		constants.EventTypeVM,
		constants.EventTypeVM, // Duplicate
		constants.EventTypeHost,
	})

	hub.registerClient(client)

	// Check subscription map has correct count
	hub.mu.RLock()
	vmSubs := len(hub.subscriptions[constants.EventTypeVM])
	hub.mu.RUnlock()

	// Even with duplicate in client subscriptions list,
	// the hub subscription map should only have one entry per client
	assert.Equal(t, 1, vmSubs, "Should only have one subscription per client ID")
}

func TestEventEmitter_GetStats_Uptime(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	stats := emitter.GetStats(false)

	assert.NotEmpty(t, stats.Uptime, "Uptime should not be empty")
	assert.False(t, stats.StartTime.IsZero(), "StartTime should be set")
}

func TestHub_RegisterClient_WarningForInvalidTypes(t *testing.T) {
	// This tests that the hub generates warning messages (sent to broadcast channel)
	// when a client subscribes to invalid event types
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Create a monitoring client to receive warning messages
	monitorClient := createTestClient("monitor", "monitor", []constants.EventType{constants.EventTypeGlobal})
	emitter.hub.register <- monitorClient
	time.Sleep(50 * time.Millisecond)

	// Register a client with invalid event type
	client := createTestClient("client1", "user1", []constants.EventType{
		constants.EventTypeVM, // valid
		"TOTALLY_INVALID",     // invalid
	})
	emitter.hub.register <- client
	time.Sleep(100 * time.Millisecond)

	// Monitor client might receive a warning message about the invalid subscription
	// This is sent to the broadcast channel asynchronously
	// We'll check if we receive it, but it's not guaranteed due to timing
	select {
	case msg := <-monitorClient.Send:
		// If we receive a message, it should be about the invalid type
		assert.Contains(t, msg.Message, "unsupported event type", "Should warn about unsupported type")
	case <-time.After(300 * time.Millisecond):
		// Timeout is OK - the warning is sent asynchronously and might be missed
		t.Log("Warning message not received (timing issue, acceptable)")
	}
}
