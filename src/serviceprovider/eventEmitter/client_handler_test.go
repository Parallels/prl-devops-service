package eventemitter

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestClient_HandleClientMessage_ClientID(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client-123",
		Send: make(chan *models.EventMessage, 10),
	}

	msg := map[string]interface{}{
		"type": "client-id",
	}

	client.handleClientMessage(msg)

	// Check that client ID was sent
	select {
	case cidMsg := <-client.Send:
		assert.Equal(t, constants.EventTypeSystem, cidMsg.Type)
		assert.Equal(t, "test-client-123", cidMsg.Message)
	default:
		t.Fatal("Expected client-id message")
	}
}

func TestClient_HandleClientMessage_InvalidFormat(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}

	// Message without type field
	msg := map[string]interface{}{
		"data": "some data",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})

	// Should not send any message
	select {
	case <-client.Send:
		t.Fatal("Should not send message for invalid format")
	default:
		// Expected - no message
	}
}

func TestClient_HandleClientMessage_UnknownType(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}

	msg := map[string]interface{}{
		"type": "unknown-message-type",
	}

	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})
	// Should log debug message but not panic or send anything
}

func TestClient_HandleClientMessage_ClientID_ChannelFull(t *testing.T) {
	ctx := basecontext.NewBaseContext()

	// Create client with no buffer - channel will always be full
	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage),
	}

	msg := map[string]interface{}{
		"type": "client-id",
	}

	// Should not block (uses select default)
	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})
}
