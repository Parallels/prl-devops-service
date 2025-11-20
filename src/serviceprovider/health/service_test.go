package health

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
	"github.com/stretchr/testify/assert"
)

// MockBroadcaster for testing
type MockBroadcaster struct {
	LastMessage *models.EventMessage
}

func (m *MockBroadcaster) BroadcastMessage(msg *models.EventMessage) error {
	m.LastMessage = msg
	return nil
}

func (m *MockBroadcaster) RegisterHandler(eventType []constants.EventType, handler eventemitter.MessageHandler) {
	// No-op for testing
}

func TestHealthService_Handle_Ping(t *testing.T) {
	mock := &MockBroadcaster{}
	svc := NewHealthService(mock)
	ctx := basecontext.NewBaseContext()

	payload := []byte(`{"message": "ping"}`)
	msgID := "test-msg-1"

	svc.Handle(ctx, "client-1", constants.EventTypeHealth, payload, msgID)

	assert.NotNil(t, mock.LastMessage)
	assert.Equal(t, constants.EventTypeHealth, mock.LastMessage.Type)
	assert.Equal(t, "pong", mock.LastMessage.Message)
	assert.Equal(t, msgID, mock.LastMessage.RefID)
	assert.Equal(t, "client-1", mock.LastMessage.ClientID)
}

func TestHealthService_Handle_InvalidJSON(t *testing.T) {
	mock := &MockBroadcaster{}
	svc := NewHealthService(mock)
	ctx := basecontext.NewBaseContext()

	payload := []byte(`{invalid-json}`)
	msgID := "test-err-1"

	// Should log warning but not panic or crash
	svc.Handle(ctx, "client-1", constants.EventTypeHealth, payload, msgID)

	// No message should be broadcasted
	assert.Nil(t, mock.LastMessage)
}
