package eventemitter

import (
	"sync"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestSendToType(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Register a test client
	client := createTestClient("test1", "testuser", []constants.EventType{constants.EventTypePDFM})
	emitter.hub.register <- client

	// Give it time to register
	time.Sleep(50 * time.Millisecond)

	// Send message
	err := emitter.SendToType(constants.EventTypePDFM, "Test message", map[string]interface{}{
		"key": "value",
	})

	assert.NoError(t, err)

	// Check message was queued
	select {
	case msg := <-client.Send:
		assert.Equal(t, constants.EventTypePDFM, msg.Type)
		assert.Equal(t, "Test message", msg.Message)
		if body, ok := msg.Body.(map[string]interface{}); ok {
			assert.Equal(t, "value", body["key"])
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client should have received the message")
	}

	// Check message counter incremented
	assert.Greater(t, emitter.messagesSent, int64(0), "Message counter should increment")
}

func TestSendToType_NotRunning(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)

	// Don't initialize - not running
	err := emitter.SendToType(constants.EventTypePDFM, "Test", nil)

	// Should return error but should warn in logs
	assert.Error(t, err)
}

func TestSendToClient(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Register a test client
	client := createTestClient("client1", "testuser", []constants.EventType{constants.EventTypePDFM})
	emitter.hub.register <- client

	time.Sleep(50 * time.Millisecond)

	// Send message to specific client
	err := emitter.SendToClient("client1", constants.EventTypePDFM, "Direct message", map[string]interface{}{
		"direct": true,
	})

	assert.NoError(t, err)

	select {
	case msg := <-client.Send:
		assert.Equal(t, "client1", msg.ClientID)
		assert.Equal(t, "Direct message", msg.Message)
		if body, ok := msg.Body.(map[string]interface{}); ok {
			assert.True(t, body["direct"].(bool))
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client should have received the direct message")
	}
}

func TestSendToAll(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Register multiple clients with different subscriptions
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypePDFM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM})

	emitter.hub.register <- client1
	emitter.hub.register <- client2

	time.Sleep(50 * time.Millisecond)

	// Send to all
	err := emitter.SendToAll("Broadcast to all", map[string]interface{}{
		"broadcast": true,
	})

	assert.NoError(t, err)

	// Both should receive (via global subscription)
	receivedCount := 0
	timeout := time.After(200 * time.Millisecond)

receiveLoop:
	for i := 0; i < 2; i++ {
		select {
		case msg := <-client1.Send:
			assert.Equal(t, constants.EventTypeGlobal, msg.Type)
			receivedCount++
		case msg := <-client2.Send:
			assert.Equal(t, constants.EventTypeGlobal, msg.Type)
			receivedCount++
		case <-timeout:
			break receiveLoop
		}
	}

	assert.Equal(t, 2, receivedCount, "Both clients should receive global message")
}

func TestBroadcastMessage(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	client := createTestClient("test1", "testuser", []constants.EventType{constants.EventTypeSystem})
	emitter.hub.register <- client

	time.Sleep(50 * time.Millisecond)

	// Create and broadcast pre-constructed message
	message := models.NewEventMessage(constants.EventTypeSystem, "System alert", map[string]interface{}{
		"level": "warning",
	})

	err := emitter.BroadcastMessage(message)
	assert.NoError(t, err)

	select {
	case msg := <-client.Send:
		assert.Equal(t, message.ID, msg.ID)
		assert.Equal(t, constants.EventTypeSystem, msg.Type)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client should have received the broadcast message")
	}
}

func TestGetStats_NoClients(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	stats := emitter.GetStats(false)

	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalClients)
	assert.Equal(t, 0, stats.TotalSubscriptions)
	assert.Empty(t, stats.Clients, "Should not include client details")
}

func TestGetStats_WithClients(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Register clients
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypePDFM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM, constants.EventTypeHost})

	emitter.hub.register <- client1
	emitter.hub.register <- client2

	time.Sleep(50 * time.Millisecond)

	stats := emitter.GetStats(false)

	assert.Equal(t, 2, stats.TotalClients)
	// Each client gets global auto-subscribed
	// client1: pdfm + global = 2
	// client2: vm + host + global = 3
	// Total subscriptions per type:
	// pdfm: 1, vm: 1, host: 1, global: 2
	assert.Equal(t, 5, stats.TotalSubscriptions)
	assert.Equal(t, 1, stats.TypeStats[constants.EventTypePDFM])
	assert.Equal(t, 1, stats.TypeStats[constants.EventTypeVM])
	assert.Equal(t, 1, stats.TypeStats[constants.EventTypeHost])
	assert.Equal(t, 2, stats.TypeStats[constants.EventTypeGlobal])
}

func TestGetStats_IncludeClients(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypePDFM})
	emitter.hub.register <- client

	time.Sleep(50 * time.Millisecond)

	stats := emitter.GetStats(true)

	assert.Equal(t, 1, stats.TotalClients)
	assert.Len(t, stats.Clients, 1, "Should include client details")
	assert.Equal(t, "client1", stats.Clients[0].ID)
	assert.Equal(t, "user1", stats.Clients[0].Username)
	assert.Contains(t, stats.Clients[0].Subscriptions, constants.EventTypePDFM)
	assert.Contains(t, stats.Clients[0].Subscriptions, constants.EventTypeGlobal)
}

func TestGetStats_NotRunning(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)

	// Don't initialize
	stats := emitter.GetStats(false)

	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalClients)
	assert.Equal(t, int64(0), stats.MessagesSent)
}

func TestMessageCounter(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	client := createTestClient("test1", "testuser", []constants.EventType{constants.EventTypePDFM})
	emitter.hub.register <- client

	time.Sleep(50 * time.Millisecond)

	// Send multiple messages
	for i := 0; i < 5; i++ {
		emitter.SendToType(constants.EventTypePDFM, "Test", nil)
	}

	stats := emitter.GetStats(false)
	assert.Equal(t, int64(5), stats.MessagesSent)
}
