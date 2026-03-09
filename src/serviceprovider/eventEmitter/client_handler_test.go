package eventemitter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestClient_HandleClientMessage_InvalidFormat(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		done: make(chan struct{}),
	}

	// Message without type field
	msg := []byte(`{"data": "some data"}`)

	// Should not panic
	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})

	// Should enqueue an error message
	client.pendingMu.Lock()
	pending := client.pending
	client.pendingMu.Unlock()
	if len(pending) == 0 {
		t.Fatal("Expected error message for invalid format")
	}
	errMsg := pending[0]
	assert.Equal(t, constants.EventTypeGlobal, errMsg.Type)
	assert.Equal(t, "error", errMsg.Message)
}

func TestClient_HandleClientMessage_UnknownType(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:  ctx,
		ID:   "test-client",
		done: make(chan struct{}),
	}

	msg := []byte(`{"type": "unknown-message-type"}`)

	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})

	// Should enqueue an error message (invalid type)
	client.pendingMu.Lock()
	pending := client.pending
	client.pendingMu.Unlock()
	if len(pending) == 0 {
		t.Fatal("Expected error message for unknown type")
	}
	errMsg := pending[0]
	assert.Equal(t, constants.EventTypeGlobal, errMsg.Type)
	assert.Equal(t, "error", errMsg.Message)
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
		done: make(chan struct{}),
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
		done:    make(chan struct{}),
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

func TestClient_ClientQueueWorker_StopsOnDone(t *testing.T) {
	// Save original global instance
	originalEE := globalEventEmitter
	defer func() { globalEventEmitter = originalEE }()

	mockHub := &Hub{
		clientToHub:  make(chan hubCommand, 10),
		shutdownChan: make(chan struct{}),
	}
	globalEventEmitter = &EventEmitter{hub: mockHub}

	// Start test server that stays open long enough
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(2 * time.Second)
	}))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)

	ctx := basecontext.NewBaseContext()
	client := &Client{
		ctx:     ctx,
		ID:      "test-queue-worker",
		Conn:    conn,
		IsAlive: true,
		done:    make(chan struct{}),
	}

	workerDone := make(chan struct{})
	go func() {
		client.clientQueueWorker()
		close(workerDone)
	}()

	// Signal worker to stop via done channel
	client.closeOnce.Do(func() { close(client.done) })

	select {
	case <-workerDone:
		// Worker stopped as expected
	case <-time.After(2 * time.Second):
		t.Fatal("clientQueueWorker did not stop after done channel closed")
	}
}
