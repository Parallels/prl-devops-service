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
		// Send unregister command if hub is still running
		c.hub.trySendCommand(&unregisterClientCmd{clientID: c.ID}, 1*time.Second)
		c.Conn.Close()
	}()

	cfg := config.Get()
	pongWait := cfg.EventEmitterPongTimeout()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		// Update state if hub is still running
		c.hub.trySendCommand(&updateClientStateCmd{
			clientID:    c.ID,
			updatePong:  true,
			updateAlive: true,
			lastPongAt:  time.Now(),
			isAlive:     true,
		}, 100*time.Millisecond)
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.ctx.LogErrorf("[Client %s] WebSocket read error: %v", c.ID, err)
			}
			break
		}

		c.hub.ctx.LogDebugf("[Client %s] Received message: %s", c.ID, string(message))

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
				c.hub.ctx.LogErrorf("[Client %s] WebSocket write error: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			// Update ping state if hub is still running
			c.hub.trySendCommand(&updateClientStateCmd{
				clientID:   c.ID,
				updatePing: true,
				lastPingAt: time.Now(),
			}, 100*time.Millisecond)
		}
	}
}

func (c *Client) handleClientMessage(msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		c.hub.ctx.LogWarnf("[Client %s] Received message without type", c.ID)
		return
	}

	switch msgType {
	case "client-id":
		cidMsg := models.NewEventMessage(constants.EventTypeSystem, c.ID, nil)
		cidMsg.ClientID = c.ID
		select {
		case c.Send <- cidMsg:
		default:
			c.hub.ctx.LogWarnf("[Client %s] Failed to send client-id (channel full)", c.ID)
		}
	default:
		c.hub.ctx.LogDebugf("[Client %s] Unknown message type: %s", c.ID, msgType)
	}
}

func (c *Client) unsubscribeToEvents(types []string, userID string) ([]string, error) {
	// Convert to event types
	eventTypes := make([]constants.EventType, 0, len(types))
	for _, typeStr := range types {
		eventType := constants.EventType(strings.ToLower(strings.TrimSpace(typeStr)))
		if !eventType.IsValid() {
			c.hub.ctx.LogWarnf("[Client %s] Attempted to unsubscribe from unknown event type: %s", c.ID, typeStr)
			continue
		}
		eventTypes = append(eventTypes, eventType)
	}

	// Send command to hub
	respChan := make(chan unsubscribeResponse, 1)
	cmd := &unsubscribeCmd{
		clientID:   c.ID,
		userID:     userID,
		eventTypes: eventTypes,
		response:   respChan,
	}

	if !c.hub.trySendCommand(cmd, 1*time.Second) {
		return nil, errors.New("cannot unsubscribe (shutdown or timeout)")
	}

	// Wait for response
	resp := <-respChan
	return resp.unsubscribed, resp.err
}
