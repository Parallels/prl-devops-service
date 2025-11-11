package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

func TestGracefulShutdown(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Register a client
	client := createTestClient("test1", "testuser", []string{constants.EVENT_TYPE_PDFM})
	emitter.hub.register <- client
	time.Sleep(50 * time.Millisecond)

	// Verify client is registered
	stats := emitter.GetStats(false)
	assert.Equal(t, 1, stats.TotalClients)

	// Shutdown
	emitter.Shutdown()

	// Verify it's not running
	assert.False(t, emitter.IsRunning())

	// Try to send a message - should not panic
	err := emitter.SendToType(constants.EVENT_TYPE_PDFM, "Test", nil)
	assert.Error(t, err) // Should log warning, and error
}

func TestShutdownWithMultipleClients(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Register multiple clients
	client1 := createTestClient("client1", "user1", []string{constants.EVENT_TYPE_PDFM})
	client2 := createTestClient("client2", "user2", []string{constants.EVENT_TYPE_VM})
	client3 := createTestClient("client3", "user3", []string{constants.EVENT_TYPE_HOST})

	emitter.hub.register <- client1
	emitter.hub.register <- client2
	emitter.hub.register <- client3
	time.Sleep(50 * time.Millisecond)

	stats := emitter.GetStats(false)
	assert.Equal(t, 3, stats.TotalClients)

	// Shutdown should close all client channels gracefully
	emitter.Shutdown()

	assert.False(t, emitter.IsRunning())
}

func TestShutdownIdempotency(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// First shutdown
	emitter.Shutdown()
	assert.False(t, emitter.IsRunning())

	// Second shutdown should be safe (no panic)
	emitter.Shutdown()
	assert.False(t, emitter.IsRunning())
}
