package eventemitter

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

// clientReader pumps messages from the WebSocket connection to the hub
func (c *Client) clientReader() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	cfg := config.Get()
	pongWait := cfg.EventEmitterPongTimeout()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.mu.Lock()
		c.LastPongAt = time.Now()
		c.IsAlive = true
		c.mu.Unlock()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.ctx.LogErrorf("[Client %s] WebSocket read error: %v", c.ID, err)
			}
			break
		}

		c.Hub.ctx.LogDebugf("[Client %s] Received message: %s", c.ID, string(message))

		// Parse and handle client messages if needed
		var clientMsg map[string]interface{}
		if err := json.Unmarshal(message, &clientMsg); err == nil {
			c.handleClientMessage(clientMsg)
		}
	}
}

// clientWriter pumps messages from the hub to the WebSocket connection
func (c *Client) clientWriter() {
	cfg := config.Get()
	pingPeriod := cfg.EventEmitterPingInterval()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send JSON message
			if err := c.Conn.WriteJSON(message); err != nil {
				c.Hub.ctx.LogErrorf("[Client %s] WebSocket write error: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			c.mu.Lock()
			c.LastPingAt = time.Now()
			c.mu.Unlock()
		}
	}
}

func (c *Client) handleClientMessage(msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		c.Hub.ctx.LogWarnf("[Client %s] Received message without type", c.ID)
		return
	}

	switch msgType {
	case "client-id":
		cidMsg := models.NewEventMessage(constants.EventTypeSystem, c.ID, nil)
		cidMsg.ClientID = c.ID
		select {
		case c.Send <- cidMsg:
		default:
			c.Hub.ctx.LogWarnf("[Client %s] Failed to send client-id (channel full)", c.ID)
		}
	default:
		c.Hub.ctx.LogDebugf("[Client %s] Unknown message type: %s", c.ID, msgType)
	}
}

func (c *Client) unsubscribeToEvents(types []string) ([]string, error) {
	c.Hub.mu.Lock()
	c.mu.Lock()
	defer c.Hub.mu.Unlock()
	defer c.mu.Unlock()

	var globalAttempted bool
	unsubscribed := make([]string, 0)

	for _, typeStr := range types {
		// Convert to EventType (case-insensitive)
		eventType := constants.EventType(strings.ToLower(strings.TrimSpace(typeStr)))

		if !eventType.IsValid() {
			c.Hub.ctx.LogWarnf("[Client %s] Attempted to unsubscribe from unknown event type: %s", c.ID, typeStr)
			continue
		}

		if eventType == constants.EventTypeGlobal {
			c.Hub.ctx.LogWarnf("[Client %s] Cannot unsubscribe from global event type", c.ID)
			globalAttempted = true
			continue
		}

		// Remove from hub subscriptions
		if c.Hub.subscriptions[eventType] != nil {
			delete(c.Hub.subscriptions[eventType], c.ID)
			if len(c.Hub.subscriptions[eventType]) == 0 {
				delete(c.Hub.subscriptions, eventType)
			}
		}

		// Remove from client's subscription list
		newSubs := make([]constants.EventType, 0, len(c.Subscriptions))
		for _, sub := range c.Subscriptions {
			if sub != eventType {
				newSubs = append(newSubs, sub)
			}
		}
		c.Subscriptions = newSubs
		unsubscribed = append(unsubscribed, eventType.String())
	}

	if len(unsubscribed) > 0 {
		c.Hub.ctx.LogInfof("[Client %s] Unsubscribed from event types: %v", c.ID, unsubscribed)
	}

	if globalAttempted {
		return unsubscribed, errors.New("cannot unsubscribe from global event type")
	}
	return unsubscribed, nil
}
