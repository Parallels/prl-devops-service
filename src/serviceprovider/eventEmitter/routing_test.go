package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestHub_BroadcastMessage_ToType(t *testing.T) {
	hub := createTestHub()
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypePDFM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM})

	hub.registerClient(client1)
	hub.registerClient(client2)

	// Create and broadcast a PDFM message
	message := models.NewEventMessage(constants.EventTypePDFM, "Test PDFM message", map[string]interface{}{
		"test": "data",
	})

	go hub.broadcastMessage(message)

	// Client1 should receive it (subscribed to PDFM)
	select {
	case msg := <-client1.Send:
		assert.Equal(t, message.ID, msg.ID, "Client1 should receive the message")
		assert.Equal(t, constants.EventTypePDFM, msg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client1 should have received the message")
	}

	// Client2 should NOT receive it (not subscribed to PDFM, only to VM)
	// But wait, client2 should receive it because of global auto-subscription!
	// Actually no - PDFM messages only go to PDFM subscribers. Global subscribers get global messages.
	select {
	case <-client2.Send:
		t.Fatal("Client2 should NOT receive PDFM message (only subscribed to VM)")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message
	}
}

func TestHub_BroadcastMessage_ToGlobal(t *testing.T) {
	hub := createTestHub()
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypePDFM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM})

	hub.registerClient(client1)
	hub.registerClient(client2)

	// Create and broadcast a GLOBAL message
	message := models.NewEventMessage(constants.EventTypeGlobal, "Global broadcast", map[string]interface{}{
		"broadcast": "all",
	})

	go hub.broadcastMessage(message)

	// Both clients should receive it (both auto-subscribed to global)
	select {
	case msg := <-client1.Send:
		assert.Equal(t, message.ID, msg.ID)
		assert.Equal(t, constants.EventTypeGlobal, msg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client1 should have received the global message")
	}

	select {
	case msg := <-client2.Send:
		assert.Equal(t, message.ID, msg.ID)
		assert.Equal(t, constants.EventTypeGlobal, msg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client2 should have received the global message")
	}
}

func TestHub_BroadcastMessage_ToSpecificClient(t *testing.T) {
	hub := createTestHub()
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypePDFM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypePDFM})

	hub.registerClient(client1)
	hub.registerClient(client2)

	// Create a message targeting specific client
	message := models.NewEventMessage(constants.EventTypePDFM, "Direct message", map[string]interface{}{
		"direct": true,
	})
	message.ClientID = "client1"

	go hub.broadcastMessage(message)

	// Only client1 should receive it
	select {
	case msg := <-client1.Send:
		assert.Equal(t, message.ID, msg.ID)
		assert.Equal(t, "client1", msg.ClientID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client1 should have received the direct message")
	}

	// Client2 should NOT receive it
	select {
	case <-client2.Send:
		t.Fatal("Client2 should NOT receive direct message to client1")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message
	}
}

func TestHub_BroadcastMessage_ClientNotFound(t *testing.T) {
	hub := createTestHub()

	// Create a message targeting non-existent client
	message := models.NewEventMessage(constants.EventTypePDFM, "Message to nobody", nil)
	message.ClientID = "nonexistent"

	// Should not panic
	assert.NotPanics(t, func() {
		hub.broadcastMessage(message)
	})
}

func TestHub_BroadcastMessage_ChannelFull(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypePDFM})

	// Fill up the channel (capacity is 10)
	for i := 0; i < 10; i++ {
		msg := models.NewEventMessage(constants.EventTypePDFM, "Fill message", nil)
		client.Send <- msg
	}

	hub.registerClient(client)

	// Try to send one more message - should be dropped
	message := models.NewEventMessage(constants.EventTypePDFM, "Overflow message", nil)

	// Should not panic or block
	done := make(chan bool)
	go func() {
		hub.broadcastMessage(message)
		done <- true
	}()

	select {
	case <-done:
		// Success - message was dropped without blocking
	case <-time.After(200 * time.Millisecond):
		t.Fatal("broadcastMessage should not block when channel is full")
	}
}

func TestHub_BroadcastMessage_NoSubscribers(t *testing.T) {
	hub := createTestHub()

	// No clients registered
	message := models.NewEventMessage(constants.EventTypePDFM, "Message to nobody", nil)

	// Should not panic
	assert.NotPanics(t, func() {
		hub.broadcastMessage(message)
	})
}

func TestHub_BroadcastMessage_MultipleSubscribersSameType(t *testing.T) {
	hub := createTestHub()
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM})
	client3 := createTestClient("client3", "user3", []constants.EventType{constants.EventTypeVM})

	hub.registerClient(client1)
	hub.registerClient(client2)
	hub.registerClient(client3)

	// Broadcast VM message
	message := models.NewEventMessage(constants.EventTypeVM, "VM event", map[string]interface{}{
		"vm_id": "123",
	})

	go hub.broadcastMessage(message)

	// All 3 clients should receive it
	receivedCount := 0
	timeout := time.After(200 * time.Millisecond)

receiveLoop:
	for i := 0; i < 3; i++ {
		select {
		case <-client1.Send:
			receivedCount++
		case <-client2.Send:
			receivedCount++
		case <-client3.Send:
			receivedCount++
		case <-timeout:
			break receiveLoop
		}
	}

	assert.Equal(t, 3, receivedCount, "All 3 clients should receive the message")
}
