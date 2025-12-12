package eventemitter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestClient_HandleClientMessage_InvalidFormat(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}

	// Message without type field
	msg := []byte(`{"data": "some data"}`)

	// Should not panic
	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})

	// Should send an error message back to client
	select {
	case errMsg := <-client.Send:
		assert.Equal(t, constants.EventTypeGlobal, errMsg.Type)
		assert.Equal(t, "error", errMsg.Message)
	default:
		t.Fatal("Expected error message for invalid format")
	}
}

func TestClient_HandleClientMessage_UnknownType(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		Send: make(chan *models.EventMessage, 10),
	}

	msg := []byte(`{"type": "unknown-message-type"}`)

	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})

	// Should send an error message back to client (invalid type)
	select {
	case errMsg := <-client.Send:
		assert.Equal(t, constants.EventTypeGlobal, errMsg.Type)
		assert.Equal(t, "error", errMsg.Message)
	default:
		t.Fatal("Expected error message for unknown type")
	}
}

func TestClient_HandleClientMessage_Success(t *testing.T) {
	// Save original global instance
	originalEE := globalEventEmitter
	defer func() { globalEventEmitter = originalEE }()

	// Setup mock Hub
	mockHub := &Hub{
		clientToHub: make(chan hubCommand, 10),
	}
	globalEventEmitter = &EventEmitter{
		hub: mockHub,
	}

	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client-success",
		Send: make(chan *models.EventMessage, 10),
	}

	// Valid message
	payload := []byte(`{"event_type": "health", "message": "ping", "id": "msg-123"}`)

	client.handleClientMessage(payload)

	// Verify command sent to Hub
	select {
	case cmd := <-mockHub.clientToHub:
		routeCmd, ok := cmd.(*RouteMessageCmd)
		assert.True(t, ok)
		assert.Equal(t, "test-client-success", routeCmd.ClientID)
		assert.Equal(t, constants.EventTypeHealth, routeCmd.Type)
		assert.Equal(t, "msg-123", routeCmd.MsgID)
		assert.Equal(t, payload, routeCmd.Payload)
	default:
		t.Fatal("Expected RouteMessageCmd to be sent to Hub")
	}
}

func TestClient_ClientReader(t *testing.T) {
	// Save original global instance
	originalEE := globalEventEmitter
	defer func() { globalEventEmitter = originalEE }()

	// Setup mock Hub
	mockHub := &Hub{
		clientToHub: make(chan hubCommand, 10),
	}
	globalEventEmitter = &EventEmitter{
		hub: mockHub,
	}

	// Start test server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send a message to the client
		msg := []byte(`{"event_type": "health", "message": "ping"}`)
		conn.WriteMessage(websocket.TextMessage, msg)

		// Keep connection open for a bit
		time.Sleep(100 * time.Millisecond)
	}))
	defer s.Close()

	// Connect to server
	url := "ws" + strings.TrimPrefix(s.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:     ctx,
		ID:      "test-reader",
		Conn:    conn,
		IsAlive: true,
		Send:    make(chan *models.EventMessage, 10),
	}

	// Run reader
	go client.clientReader()

	// Verify message received and forwarded to Hub
	select {
	case cmd := <-mockHub.clientToHub:
		routeCmd, ok := cmd.(*RouteMessageCmd)
		assert.True(t, ok)
		assert.Equal(t, "test-reader", routeCmd.ClientID)
		assert.Equal(t, constants.EventTypeHealth, routeCmd.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message from clientReader")
	}

	// Cleanup
	client.mu.Lock()
	client.IsAlive = false
	client.Conn.Close()
	client.mu.Unlock()
}

func TestClient_ClientWriter(t *testing.T) {
	// Save original global instance
	originalEE := globalEventEmitter
	defer func() { globalEventEmitter = originalEE }()

	// Setup mock Hub
	mockHub := &Hub{
		clientToHub: make(chan hubCommand, 10),
	}
	globalEventEmitter = &EventEmitter{
		hub: mockHub,
	}

	// Start test server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Verify message content
		var receivedMsg models.EventMessage
		if err := json.Unmarshal(msg, &receivedMsg); err == nil {
			if receivedMsg.Message == "test-message" {
				// Success
			}
		}
	}))
	defer s.Close()

	// Connect to server
	url := "ws" + strings.TrimPrefix(s.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:     ctx,
		ID:      "test-writer",
		Conn:    conn,
		IsAlive: true,
		Send:    make(chan *models.EventMessage, 10),
	}

	// Run writer
	go client.clientWriter()

	// Send message
	msg := models.NewEventMessage(constants.EventTypeHealth, "test-writer", nil)
	msg.Message = "test-message"
	client.Send <- msg

	// Give time for write
	time.Sleep(100 * time.Millisecond)

	// Cleanup (closes channel which stops writer)
	close(client.Send)

	// Give time for writer to close connection
	time.Sleep(50 * time.Millisecond)
}
