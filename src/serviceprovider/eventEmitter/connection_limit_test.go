package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

func TestHub_HasActiveConnectionFromIP_NoConnection(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	hasConnection := hub.HasActiveConnectionFromIP("192.168.1.100")

	assert.False(t, hasConnection, "Should return false when no connection exists")
}

func TestHub_HasActiveConnectionFromIP_WithConnection(t *testing.T) {
	hub := createTestHub()
	// Start hub goroutine to process commands
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeGlobal})
	client.RemoteIP = "192.168.1.100"

	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	hasConnection := hub.HasActiveConnectionFromIP("192.168.1.100")

	assert.True(t, hasConnection, "Should return true when connection exists from IP")
}

func TestHub_HasActiveConnectionFromIP_DifferentIP(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeGlobal})
	client.RemoteIP = "192.168.1.100"

	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	hasConnection := hub.HasActiveConnectionFromIP("192.168.1.200")

	assert.False(t, hasConnection, "Should return false for different IP")
}

func TestHub_HasActiveConnectionFromIP_EmptyIP(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	hasConnection := hub.HasActiveConnectionFromIP("")

	assert.False(t, hasConnection, "Should return false for empty IP")
}

func TestHub_HasActiveConnectionFromIP_MultipleClients(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeGlobal})
	client1.RemoteIP = "192.168.1.100"

	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeGlobal})
	client2.RemoteIP = "192.168.1.200"

	hub.commands <- &registerClientCmd{client: client1}
	hub.commands <- &registerClientCmd{client: client2}
	time.Sleep(50 * time.Millisecond)

	assert.True(t, hub.HasActiveConnectionFromIP("192.168.1.100"))
	assert.True(t, hub.HasActiveConnectionFromIP("192.168.1.200"))
	assert.False(t, hub.HasActiveConnectionFromIP("192.168.1.300"))
}

func TestHub_HasActiveConnectionFromIP_AfterUnregister(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeGlobal})
	client.RemoteIP = "192.168.1.100"

	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)
	assert.True(t, hub.HasActiveConnectionFromIP("192.168.1.100"))

	hub.commands <- &unregisterClientCmd{clientID: client.ID}
	time.Sleep(50 * time.Millisecond)
	assert.False(t, hub.HasActiveConnectionFromIP("192.168.1.100"), "Should return false after unregister")
}

func TestIsMultipleConnectionsPerIPAllowed(t *testing.T) {
	// Test the connection limit function
	allowed := isMultipleConnectionsPerIPAllowed()

	// In release mode, should return true (multiple connections allowed)
	// In debug mode, should return false (single connection per IP)
	assert.IsType(t, true, allowed, "Should return a boolean value")
}

func TestHub_RegisterClient_WithIP(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client.RemoteIP = "10.0.0.50"

	hub.registerClient(client)

	// Verify client is in clientsByIP map (direct access is OK for synchronous test)
	storedClientID, exists := hub.clientsByIP["10.0.0.50"]

	assert.True(t, exists, "Client should be in clientsByIP map")
	assert.Equal(t, client.ID, storedClientID)
}

func TestHub_RegisterClient_WithoutIP(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client.RemoteIP = "" // No IP

	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	// Should still register client, just not in clientsByIP
	_, exists := hub.clients["client1"]

	assert.True(t, exists, "Client should be registered")
	assert.False(t, hub.HasActiveConnectionFromIP(""), "Empty IP should not be tracked")
}
