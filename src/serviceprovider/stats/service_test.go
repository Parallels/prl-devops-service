package stats

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

type MockBroadcaster struct {
	messages  []*models.EventMessage
	msgChan   chan *models.EventMessage
	isRunning bool
}

func (m *MockBroadcaster) BroadcastMessage(msg *models.EventMessage) error {
	m.messages = append(m.messages, msg)
	if m.msgChan != nil {
		m.msgChan <- msg
	}
	return nil
}

func (m *MockBroadcaster) IsRunning() bool {
	return m.isRunning
}

func TestStatsService_Run(t *testing.T) {
	// Setup
	mockBroadcaster := &MockBroadcaster{
		msgChan:   make(chan *models.EventMessage, 10),
		isRunning: true,
	}

	// Direct initialization to bypass singleton for testing isolation
	service := &StatsService{
		broadcaster: mockBroadcaster,
		stopChan:    make(chan struct{}),
	}

	ctx := basecontext.NewBaseContext()

	// Execute
	// Run with a short interval
	go service.Run(ctx, 50*time.Millisecond)

	// Verify
	select {
	case msg := <-mockBroadcaster.msgChan:
		assert.Equal(t, constants.EventTypeStats, msg.Type)
		assert.Equal(t, "System Stats", msg.Message)

		// Assert payload
		assert.IsType(t, StatsMessage{}, msg.Body)
		statsBody := msg.Body.(StatsMessage) // Type assertion works because we are in the same package

		assert.Greater(t, statsBody.Memory, uint64(0), "Memory should be greater than 0")
		assert.GreaterOrEqual(t, statsBody.CpuUserTime, float64(0), "CpuUserTime should be >= 0")
		assert.Greater(t, statsBody.Goroutines, 0, "Goroutines should be > 0")

	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for stats message")
	}

	service.Stop()
	// Allow some time for stop to propagate
	time.Sleep(10 * time.Millisecond)
	assert.False(t, service.isRunning)
}
