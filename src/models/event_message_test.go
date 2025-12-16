package models

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

func TestNewEventMessage(t *testing.T) {
	message := "Test message"
	body := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	msg := NewEventMessage(constants.EventTypePDFM, message, body)

	assert.NotNil(t, msg)
	assert.NotEmpty(t, msg.ID, "ID should be generated")
	assert.Equal(t, constants.EventTypePDFM, msg.Type)
	assert.Equal(t, message, msg.Message)
	assert.Equal(t, body, msg.Body)
	assert.Empty(t, msg.ClientID, "ClientID should be empty by default")
	assert.WithinDuration(t, time.Now().UTC(), msg.Timestamp, 2*time.Second, "Timestamp should be current time")
}

func TestNewEventMessage_NilBody(t *testing.T) {
	msg := NewEventMessage("system", "Alert", nil)

	assert.NotNil(t, msg)
	assert.Nil(t, msg.Body)
	assert.Equal(t, constants.EventTypeSystem, msg.Type)
}

func TestNewEventMessage_EmptyMessage(t *testing.T) {
	msg := NewEventMessage(constants.EventTypeHealth, "", map[string]interface{}{})
	assert.NotNil(t, msg)
	assert.Empty(t, msg.Message)
	assert.NotEmpty(t, msg.ID)
	assert.NotEmpty(t, msg.Type)
}

func TestEventMessage_UniqueIDs(t *testing.T) {
	msg1 := NewEventMessage("test", "Message 1", nil)
	msg2 := NewEventMessage("test", "Message 2", nil)

	assert.NotEqual(t, msg1.ID, msg2.ID, "Each message should have a unique ID")
}

func TestEventMessage_TimestampUTC(t *testing.T) {
	msg := NewEventMessage("test", "Test", nil)

	assert.Equal(t, "UTC", msg.Timestamp.Location().String(), "Timestamp should be in UTC")
}

func TestEventClientInfo_Fields(t *testing.T) {
	now := time.Now()

	clientInfo := EventClientInfo{
		ID:            "client123",
		UserID:        "user456",
		Username:      "testuser",
		ConnectedAt:   now,
		LastPingAt:    now,
		LastPongAt:    now,
		Subscriptions: []constants.EventType{constants.EventTypePDFM, constants.EventTypeGlobal},
		IsAlive:       true,
	}

	assert.Equal(t, "client123", clientInfo.ID)
	assert.Equal(t, "user456", clientInfo.UserID)
	assert.Equal(t, "testuser", clientInfo.Username)
	assert.Equal(t, now, clientInfo.ConnectedAt)
	assert.Len(t, clientInfo.Subscriptions, 2)
	assert.True(t, clientInfo.IsAlive)
}

func TestEventEmitterStats_Fields(t *testing.T) {
	now := time.Now()

	stats := EventEmitterStats{
		TotalClients:       5,
		TotalSubscriptions: 12,
		TypeStats: map[constants.EventType]int{
			constants.EventTypePDFM:   2,
			constants.EventTypeGlobal: 5,
		},
		Clients:      []EventClientInfo{},
		MessagesSent: 100,
		StartTime:    now,
		Uptime:       "1h30m",
	}

	assert.Equal(t, 5, stats.TotalClients)
	assert.Equal(t, 12, stats.TotalSubscriptions)
	assert.Equal(t, 2, stats.TypeStats[constants.EventTypePDFM])
	assert.Equal(t, int64(100), stats.MessagesSent)
	assert.Equal(t, now, stats.StartTime)
}
