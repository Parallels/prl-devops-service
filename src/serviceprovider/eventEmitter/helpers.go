package eventemitter

import (
	"errors"
	"sync/atomic"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
)

// SendToType sends a message to all clients subscribed to a specific type
func (e *EventEmitter) SendToType(eventType string, message string, body map[string]interface{}) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return errors.New("event emitter is not running")
	}

	msg := models.NewEventMessage(eventType, message, body)

	e.hub.broadcast <- msg
	atomic.AddInt64(&e.messagesSent, 1)

	e.ctx.LogDebugf("[EventEmitter] Queued message %s for type: %s", msg.ID, eventType)
	return nil
}

// SendToClient sends a message to a specific client
func (e *EventEmitter) SendToClient(clientID string, eventType string, message string, body map[string]interface{}) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return errors.New("event emitter is not running")
	}

	msg := models.NewEventMessage(eventType, message, body)
	msg.ClientID = clientID

	e.hub.broadcast <- msg
	atomic.AddInt64(&e.messagesSent, 1)

	e.ctx.LogDebugf("[EventEmitter] Queued message %s for client: %s", msg.ID, clientID)
	return nil
}

// SendToAll sends a message to all connected clients
func (e *EventEmitter) SendToAll(message string, body map[string]interface{}) error {
	return e.SendToType(constants.EVENT_TYPE_GLOBAL, message, body)
}

// BroadcastMessage sends a pre-constructed event message
func (e *EventEmitter) BroadcastMessage(msg *models.EventMessage) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return nil
	}

	e.hub.broadcast <- msg
	atomic.AddInt64(&e.messagesSent, 1)

	e.ctx.LogDebugf("[EventEmitter] Queued pre-constructed message %s", msg.ID)
	return nil
}

// GetStats returns statistics about the event emitter
func (e *EventEmitter) GetStats(includeClients bool) *models.EventEmitterStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.isRunning || e.hub == nil {
		return &models.EventEmitterStats{
			TotalClients:       0,
			TotalSubscriptions: 0,
			TypeStats:          make(map[string]int),
			MessagesSent:       0,
			StartTime:          e.startTime,
			Uptime:             "0s",
		}
	}

	e.hub.mu.RLock()
	defer e.hub.mu.RUnlock()

	stats := &models.EventEmitterStats{
		TotalClients:       len(e.hub.clients),
		TotalSubscriptions: 0,
		TypeStats:          make(map[string]int),
		MessagesSent:       atomic.LoadInt64(&e.messagesSent),
		StartTime:          e.startTime,
		Uptime:             e.getUptime(),
	}

	// Count subscriptions per type
	for eventType, subscribers := range e.hub.subscriptions {
		count := len(subscribers)
		stats.TypeStats[eventType] = count
		stats.TotalSubscriptions += count
	}

	// Include client details if requested (admin only)
	if includeClients {
		stats.Clients = make([]models.EventClientInfo, 0, len(e.hub.clients))
		for _, client := range e.hub.clients {
			client.mu.RLock()
			clientInfo := models.EventClientInfo{
				ID:            client.ID,
				UserID:        client.UserID,
				Username:      client.Username,
				ConnectedAt:   client.ConnectedAt,
				LastPingAt:    client.LastPingAt,
				LastPongAt:    client.LastPongAt,
				Subscriptions: client.Subscriptions,
				IsAlive:       client.IsAlive,
			}
			client.mu.RUnlock()
			stats.Clients = append(stats.Clients, clientInfo)
		}
	}

	return stats
}

// getUptime returns human-readable uptime string
func (e *EventEmitter) getUptime() string {
	duration := e.startTime
	uptime := duration.String()
	return uptime
}
