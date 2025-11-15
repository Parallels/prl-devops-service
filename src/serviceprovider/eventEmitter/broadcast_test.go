package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestBroadcastMessage_EmptySubscribers(t *testing.T) {
	hub := createTestHub()

	msg := models.NewEventMessage(constants.EventTypeVM, "Test", nil)
	hub.broadcastMessage(msg)

	// Should not panic with no subscribers
	assert.Len(t, hub.clients, 0)
}

func TestBroadcastMessage_TargetClientNotFound(t *testing.T) {
	hub := createTestHub()

	msg := models.NewEventMessage(constants.EventTypeVM, "Test", nil)
	msg.ClientID = "nonexistent"

	// Should not panic
	assert.NotPanics(t, func() {
		hub.broadcastMessage(msg)
	})
}

func TestBroadcastMessage_TargetSpecificClient(t *testing.T) {
	hub := createTestHub()

	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM})

	hub.registerClient(client1)
	hub.registerClient(client2)

	msg := models.NewEventMessage(constants.EventTypeVM, "Test", nil)
	msg.ClientID = "client1"

	hub.broadcastMessage(msg)

	// Only client1 should receive
	select {
	case receivedMsg := <-client1.Send:
		assert.Equal(t, msg.ID, receivedMsg.ID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client1 should have received message")
	}

	// client2 should not receive
	select {
	case <-client2.Send:
		t.Fatal("client2 should not have received message")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message
	}
}

func TestBroadcastMessage_MultipleSubscribers(t *testing.T) {
	hub := createTestHub()

	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM})
	client3 := createTestClient("client3", "user3", []constants.EventType{constants.EventTypeHost})

	hub.registerClient(client1)
	hub.registerClient(client2)
	hub.registerClient(client3)

	msg := models.NewEventMessage(constants.EventTypeVM, "Test", nil)

	hub.broadcastMessage(msg)

	// Both VM subscribers should receive
	select {
	case <-client1.Send:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client1 should have received message")
	}

	select {
	case <-client2.Send:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client2 should have received message")
	}

	// Host subscriber should not receive VM message
	select {
	case <-client3.Send:
		t.Fatal("client3 should not have received VM message")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message
	}
}

func TestBroadcastMessage_ChannelFull(t *testing.T) {
	hub := createTestHub()

	// Create client with small buffer
	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client.Send = make(chan *models.EventMessage, 1) // Small buffer

	hub.registerClient(client)

	// Fill the channel
	msg1 := models.NewEventMessage(constants.EventTypeVM, "Test1", nil)
	hub.broadcastMessage(msg1)

	// Try to send another when full - should drop
	msg2 := models.NewEventMessage(constants.EventTypeVM, "Test2", nil)
	hub.broadcastMessage(msg2)

	// Drain and verify only first message received
	<-client.Send

	// Second should have been dropped (logged as warning)
	select {
	case <-client.Send:
		t.Fatal("Second message should have been dropped")
	case <-time.After(50 * time.Millisecond):
		// Expected - message was dropped
	}
}

func TestBroadcastMessage_GlobalSubscribers(t *testing.T) {
	hub := createTestHub()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeGlobal})
	hub.registerClient(client)

	msg := models.NewEventMessage(constants.EventTypeGlobal, "Test", nil)
	hub.broadcastMessage(msg)

	select {
	case receivedMsg := <-client.Send:
		assert.Equal(t, msg.ID, receivedMsg.ID)
		assert.Equal(t, constants.EventTypeGlobal, receivedMsg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Global subscriber should have received message")
	}
}

func TestBroadcastCmd_Execute(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}; close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	msg := models.NewEventMessage(constants.EventTypeVM, "Test", nil)
	hub.commands <- &broadcastCmd{message: msg}

	select {
	case receivedMsg := <-client.Send:
		assert.Equal(t, msg.ID, receivedMsg.ID)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client should have received broadcast message")
	}
}

func TestBroadcastMessage_NoSubscribersForType(t *testing.T) {
	hub := createTestHub()

	// Register client with VM subscription
	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	hub.registerClient(client)

	// Broadcast to HOST type (no subscribers)
	msg := models.NewEventMessage(constants.EventTypeHost, "Test", nil)
	hub.broadcastMessage(msg)

	// Client should not receive anything
	select {
	case <-client.Send:
		t.Fatal("Client should not receive message for unsubscribed type")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message
	}
}
