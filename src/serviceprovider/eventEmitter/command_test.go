package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetClientCmd_Execute_Success(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	// Get client with correct user ID
	respChan := make(chan getClientResponse, 1)
	hub.commands <- &getClientCmd{
		clientID: "test-client",
		userID:   "test-client",
		response: respChan,
	}

	resp := <-respChan
	require.NoError(t, resp.err)
	assert.NotNil(t, resp.client)
	assert.Equal(t, "test-client", resp.client.ID)
}

func TestGetClientCmd_Execute_ClientNotFound(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	respChan := make(chan getClientResponse, 1)
	hub.commands <- &getClientCmd{
		clientID: "nonexistent",
		userID:   "test",
		response: respChan,
	}

	resp := <-respChan
	assert.Error(t, resp.err)
	assert.Nil(t, resp.client)
	assert.Contains(t, resp.err.Error(), "not found")
}

func TestGetClientCmd_Execute_Unauthorized(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	// Try to get client with wrong user ID
	respChan := make(chan getClientResponse, 1)
	hub.commands <- &getClientCmd{
		clientID: "test-client",
		userID:   "wrong-user",
		response: respChan,
	}

	resp := <-respChan
	assert.Error(t, resp.err)
	assert.Nil(t, resp.client)
	assert.Contains(t, resp.err.Error(), "unauthorized")
}

func TestUpdateClientStateCmd_Execute_PingOnly(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	now := time.Now()
	hub.commands <- &updateClientStateCmd{
		clientID:   "test-client",
		updatePing: true,
		lastPingAt: now,
	}
	time.Sleep(50 * time.Millisecond)

	// Verify ping was updated
	assert.Equal(t, now.Unix(), client.LastPingAt.Unix())
}

func TestUpdateClientStateCmd_Execute_PongOnly(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	now := time.Now()
	hub.commands <- &updateClientStateCmd{
		clientID:   "test-client",
		updatePong: true,
		lastPongAt: now,
	}
	time.Sleep(50 * time.Millisecond)

	// Verify pong was updated
	assert.Equal(t, now.Unix(), client.LastPongAt.Unix())
}

func TestUpdateClientStateCmd_Execute_AliveOnly(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	client.IsAlive = false
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	hub.commands <- &updateClientStateCmd{
		clientID:    "test-client",
		updateAlive: true,
		isAlive:     true,
	}
	time.Sleep(50 * time.Millisecond)

	// Verify alive was updated
	assert.True(t, client.IsAlive)
}

func TestUpdateClientStateCmd_Execute_ClientNotFound(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	// Should not panic when client doesn't exist
	hub.commands <- &updateClientStateCmd{
		clientID:   "nonexistent",
		updatePing: true,
		lastPingAt: time.Now(),
	}
	time.Sleep(50 * time.Millisecond)
	// No assertion - just verify it doesn't crash
}

func TestUnsubscribeCmd_Execute_CannotUnsubscribeGlobal(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	respChan := make(chan unsubscribeResponse, 1)
	hub.commands <- &unsubscribeCmd{
		clientID:   "test-client",
		userID:     "test-client",
		eventTypes: []constants.EventType{constants.EventTypeGlobal},
		response:   respChan,
	}

	resp := <-respChan
	assert.Error(t, resp.err)
	assert.Contains(t, resp.err.Error(), "global")
	assert.Len(t, resp.unsubscribed, 0)
}

func TestUnsubscribeCmd_Execute_ClientNotFound(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	respChan := make(chan unsubscribeResponse, 1)
	hub.commands <- &unsubscribeCmd{
		clientID:   "nonexistent",
		userID:     "test",
		eventTypes: []constants.EventType{constants.EventTypeVM},
		response:   respChan,
	}

	resp := <-respChan
	assert.Error(t, resp.err)
	assert.Contains(t, resp.err.Error(), "not found")
}

func TestUnsubscribeCmd_Execute_Unauthorized(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeVM})
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	respChan := make(chan unsubscribeResponse, 1)
	hub.commands <- &unsubscribeCmd{
		clientID:   "test-client",
		userID:     "wrong-user",
		eventTypes: []constants.EventType{constants.EventTypeVM},
		response:   respChan,
	}

	resp := <-respChan
	assert.Error(t, resp.err)
	assert.Contains(t, resp.err.Error(), "unauthorized")
}

func TestShutdownCmd_Execute_ClosesAllClients(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	// Register multiple clients
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeHost})

	hub.commands <- &registerClientCmd{client: client1}
	hub.commands <- &registerClientCmd{client: client2}
	time.Sleep(50 * time.Millisecond)

	assert.Len(t, hub.clients, 2)

	// Execute shutdown command
	hub.commands <- &shutdownCmd{}
	time.Sleep(50 * time.Millisecond)

	// Verify all clients cleared
	assert.Len(t, hub.clients, 0)
	assert.Len(t, hub.clientsByIP, 0)
	assert.Len(t, hub.subscriptions, 0)

	// Verify channels are closed
	_, ok := <-client1.Send
	assert.False(t, ok, "client1 Send channel should be closed")
	_, ok = <-client2.Send
	assert.False(t, ok, "client2 Send channel should be closed")
}

func TestShutdownCmd_Execute_ClearsClientsByIP(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client.RemoteIP = "192.168.1.100"

	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	assert.Len(t, hub.clientsByIP, 1)

	hub.commands <- &shutdownCmd{}
	time.Sleep(50 * time.Millisecond)

	assert.Len(t, hub.clientsByIP, 0)
}

func TestCheckIPCmd_Execute_IPNotFound(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	respChan := make(chan bool, 1)
	hub.commands <- &checkIPCmd{
		ip:       "192.168.1.1",
		response: respChan,
	}

	result := <-respChan
	assert.False(t, result)
}

func TestCheckIPCmd_Execute_IPFound(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client.RemoteIP = "192.168.1.100"

	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	respChan := make(chan bool, 1)
	hub.commands <- &checkIPCmd{
		ip:       "192.168.1.100",
		response: respChan,
	}

	result := <-respChan
	assert.True(t, result)
}

func TestHasActiveConnectionFromIP_CommandTimeout(t *testing.T) {
	hub := createTestHub()
	// Don't start hub goroutine - test timeout behavior

	result := hub.HasActiveConnectionFromIP("192.168.1.1")
	assert.False(t, result, "Should return false on timeout")
}

func TestHasActiveConnectionFromIP_EmptyIP(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	result := hub.HasActiveConnectionFromIP("")
	assert.False(t, result)
}
