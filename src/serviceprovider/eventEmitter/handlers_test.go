package eventemitter

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
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

func (m *MockBroadcaster) RegisterHandler(eventType []constants.EventType, handler MessageHandler) {
	// No-op for testing
}

func TestSystemHandler_Handle_ClientID(t *testing.T) {
	mock := &MockBroadcaster{}
	svc := NewSystemHandler(mock)
	svc.broadcaster = mock

	ctx := basecontext.NewBaseContext()
	payload := []byte(`{"message": "client-id"}`)
	msgID := "test-sys-1"

	svc.Handle(ctx, "client-1", constants.EventTypeSystem, payload, msgID)

	assert.NotNil(t, mock.LastMessage)
	assert.Equal(t, constants.EventTypeSystem, mock.LastMessage.Type)
	assert.Equal(t, "client-id", mock.LastMessage.Message)
	assert.Equal(t, msgID, mock.LastMessage.RefID)
	assert.Equal(t, "client-1", mock.LastMessage.ClientID)

	body, ok := mock.LastMessage.Body.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "client-1", body["client-id"])
}

func TestSystemHandler_Handle_InvalidMessage(t *testing.T) {
	mock := &MockBroadcaster{}
	svc := NewSystemHandler(mock)
	svc.broadcaster = mock
	ctx := basecontext.NewBaseContext()

	// Invalid JSON
	payload := []byte(`{invalid-json}`)
	msgID := "test-err-1"

	svc.Handle(ctx, "client-1", constants.EventTypeSystem, payload, msgID)

	assert.NotNil(t, mock.LastMessage)
	assert.Equal(t, constants.EventTypeSystem, mock.LastMessage.Type)
	assert.Equal(t, "error", mock.LastMessage.Message)
	assert.Equal(t, msgID, mock.LastMessage.RefID)
}
