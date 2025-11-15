package eventemitter

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
)

// SendToType sends a message to all clients subscribed to a specific type
func (e *EventEmitter) SendToType(eventType constants.EventType, message string, body map[string]interface{}) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return errors.New("event emitter is not running")
	}

	msg := models.NewEventMessage(eventType, message, body)

	if !e.hub.trySendCommand(&broadcastCmd{message: msg}, 1*time.Second) {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message (shutdown or timeout)")
		return errors.New("failed to send message")
	}

	atomic.AddInt64(&e.messagesSent, 1)
	e.ctx.LogDebugf("[EventEmitter] Queued message %s for type: %s", msg.ID, eventType)
	return nil
}

// SendToClient sends a message to a specific client
func (e *EventEmitter) SendToClient(clientID string, eventType constants.EventType, message string, body map[string]interface{}) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return errors.New("event emitter is not running")
	}

	msg := models.NewEventMessage(eventType, message, body)
	msg.ClientID = clientID

	if !e.hub.trySendCommand(&broadcastCmd{message: msg}, 1*time.Second) {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message (shutdown or timeout)")
		return errors.New("failed to send message")
	}

	atomic.AddInt64(&e.messagesSent, 1)
	e.ctx.LogDebugf("[EventEmitter] Queued message %s for client: %s", msg.ID, clientID)
	return nil
}

// SendToAll sends a message to all connected clients
func (e *EventEmitter) SendToAll(message string, body map[string]interface{}) error {
	return e.SendToType(constants.EventTypeGlobal, message, body)
}

// BroadcastMessage sends a pre-constructed event message
func (e *EventEmitter) BroadcastMessage(msg *models.EventMessage) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return nil
	}

	if !e.hub.trySendCommand(&broadcastCmd{message: msg}, 1*time.Second) {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message (shutdown or timeout)")
		return errors.New("failed to send message")
	}

	atomic.AddInt64(&e.messagesSent, 1)
	e.ctx.LogDebugf("[EventEmitter] Queued pre-constructed message %s", msg.ID)
	return nil
}

// GetStats returns statistics about the event emitter
// Thread-safe via command pattern - retrieves snapshot from hub goroutine
func (e *EventEmitter) GetStats(includeClients bool) *models.EventEmitterStats {
	if atomic.LoadInt32(&e.isRunning) == 0 || e.hub == nil {
		return &models.EventEmitterStats{
			TotalClients:       0,
			TotalSubscriptions: 0,
			TypeStats:          make(map[constants.EventType]int),
			MessagesSent:       0,
			StartTime:          e.startTime,
			Uptime:             "0s",
		}
	}

	respChan := make(chan *models.EventEmitterStats, 1)
	cmd := &getStatsCmd{
		includeClients: includeClients,
		response:       respChan,
	}

	// Send command with timeout
	if !e.hub.trySendCommand(cmd, 1*time.Second) {
		e.ctx.LogWarnf("[EventEmitter] Cannot send getStats command (shutdown or timeout)")
		return &models.EventEmitterStats{
			TotalClients:       0,
			TotalSubscriptions: 0,
			TypeStats:          make(map[constants.EventType]int),
			MessagesSent:       atomic.LoadInt64(&e.messagesSent),
			StartTime:          e.startTime,
			Uptime:             e.getUptime(),
		}
	}

	// Wait for response with timeout
	select {
	case stats := <-respChan:
		// Add fields that are not in hub
		stats.MessagesSent = atomic.LoadInt64(&e.messagesSent)
		stats.StartTime = e.startTime
		stats.Uptime = e.getUptime()
		return stats
	case <-time.After(1 * time.Second):
		e.ctx.LogWarnf("[EventEmitter] Timeout waiting for getStats response")
		return &models.EventEmitterStats{
			TotalClients:       0,
			TotalSubscriptions: 0,
			TypeStats:          make(map[constants.EventType]int),
			MessagesSent:       atomic.LoadInt64(&e.messagesSent),
			StartTime:          e.startTime,
			Uptime:             e.getUptime(),
		}
	}
}

// getUptime returns human-readable uptime string
func (e *EventEmitter) getUptime() string {
	duration := e.startTime
	uptime := duration.String()
	return uptime
}
