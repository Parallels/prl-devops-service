package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

func TestHub_Run_CommandsChannelClosed(t *testing.T) {
	hub := createTestHub()

	// Close commands channel immediately
	hub.commands <- &shutdownCmd{}
	close(hub.shutdown)

	// run() should exit gracefully
	done := make(chan bool)
	go func() {
		hub.run()
		done <- true
	}()

	select {
	case <-done:
		// Success - run exited
	case <-time.After(1 * time.Second):
		t.Fatal("run() should have exited when commands channel closed")
	}
}

func TestHub_Run_DoneSignalReceived(t *testing.T) {
	hub := createTestHub()

	go hub.run()

	// Wait for hub to start
	time.Sleep(50 * time.Millisecond)

	// Signal shutdown
	hub.commands <- &shutdownCmd{}
	close(hub.shutdown)

	// Wait for hub to stop
	select {
	case <-hub.stopped:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Hub should have stopped")
	}
}

func TestHub_DrainCommands_EmptyQueue(t *testing.T) {
	hub := createTestHub()

	// drainCommands waits for quiescence period (100ms) to ensure no more commands
	start := time.Now()
	hub.drainCommands()
	duration := time.Since(start)

	// Should wait ~100ms for quiescence, allow 150ms for timing variance
	assert.GreaterOrEqual(t, duration, 95*time.Millisecond, "Should wait for quiescence period")
	assert.Less(t, duration, 150*time.Millisecond, "Should not wait too long")
}

func TestHub_DrainCommands_WithPendingCommands(t *testing.T) {
	hub := createTestHub()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})

	// Add some commands to queue
	hub.commands <- &registerClientCmd{client: client}
	hub.commands <- &unregisterClientCmd{clientID: client.ID}

	// Drain should process them
	hub.drainCommands()

	// Queue should be empty now
	select {
	case <-hub.commands:
		t.Fatal("Commands channel should be empty after drain")
	default:
		// Expected - channel is empty
	}
}

func TestHub_Run_MultipleCommandTypes(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}
		close(hub.shutdown)
		<-hub.stopped
	}()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})

	// Send different command types
	hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	hub.commands <- &updateClientStateCmd{
		clientID:   "client1",
		updatePing: true,
		lastPingAt: time.Now(),
	}
	time.Sleep(50 * time.Millisecond)

	respChan := make(chan bool, 1)
	hub.commands <- &checkIPCmd{
		ip:       client.RemoteIP,
		response: respChan,
	}

	// All commands should be processed
	select {
	case <-respChan:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Commands should have been processed")
	}
}

func TestHub_Run_HighCommandVolume(t *testing.T) {
	hub := createTestHub()
	go hub.run()
	defer func() {
		hub.commands <- &shutdownCmd{}
		close(hub.shutdown)
		<-hub.stopped
	}()

	// Send many commands rapidly
	for i := 0; i < 100; i++ {
		client := createTestClient("client"+string(rune(i)), "user", []constants.EventType{constants.EventTypeVM})
		hub.commands <- &registerClientCmd{client: client}
	}

	time.Sleep(200 * time.Millisecond)

	// All clients should be registered
	assert.Equal(t, 100, len(hub.clients))
}

func TestEventEmitter_Shutdown_WithPendingCommands(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	client := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})

	// Register client
	emitter.hub.commands <- &registerClientCmd{client: client}
	time.Sleep(50 * time.Millisecond)

	// Add pending commands
	emitter.hub.commands <- &updateClientStateCmd{
		clientID:   "client1",
		updatePing: true,
		lastPingAt: time.Now(),
	}

	// Shutdown should drain pending commands
	emitter.Shutdown()

	assert.False(t, emitter.IsRunning())
}

func TestEventEmitter_Shutdown_AlreadyStopped(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	// Shutdown once
	emitter.Shutdown()

	// Shutdown again - should not panic
	assert.NotPanics(t, func() {
		emitter.Shutdown()
	})
}

func TestEventEmitter_Initialize_NilHub(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	emitter.Shutdown()
	emitter.hub = nil

	// Shutdown with nil hub should not panic
	assert.NotPanics(t, func() {
		emitter.Shutdown()
	})
}
