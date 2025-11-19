package eventemitter

import (
	"encoding/json"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

// clientReader reads messages from the WebSocket connection
func (c *Client) clientReader() {
	defer func() {
		c.ctx.LogInfof("[Client %s] Closing reader", c.ID)
		c.mu.Lock()
		if c.IsAlive {
			c.IsAlive = false
			c.Conn.Close()
			c.mu.Unlock()
			// Only unregister if we're the first goroutine to detect disconnection
			Get().hub.trySendCommand(&unregisterClientCmd{clientID: c.ID})
		} else {
			c.mu.Unlock()
		}
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.ctx.LogErrorf("[Client %s] WebSocket read error: %v", c.ID, err)
			}
			break
		}

		c.ctx.LogDebugf("[Client %s] Received message: %s", c.ID, string(message))

		// Parse and handle client messages if needed
		var clientMsg map[string]interface{}
		if err := json.Unmarshal(message, &clientMsg); err == nil {
			c.handleClientMessage(clientMsg)
		}
	}
}

// clientWriter writes messages from the hub to the WebSocket connection
func (c *Client) clientWriter() {
	defer func() {
		c.ctx.LogInfof("[Client %s] Closing writer", c.ID)
		c.mu.Lock()
		if c.IsAlive {
			c.IsAlive = false
			c.Conn.Close()
			c.mu.Unlock()
			// Only unregister if we're the first goroutine to detect disconnection
			Get().hub.trySendCommand(&unregisterClientCmd{clientID: c.ID})
		} else {
			c.mu.Unlock()
		}
	}()
	for message := range c.Send {
		c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.Conn.WriteJSON(message); err != nil {
			c.ctx.LogErrorf("[Client %s] WebSocket write error: %v", c.ID, err)
			return
		}
	}
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
	c.ctx.LogInfof("[Client %s] Send channel closed", c.ID)
}

func (c *Client) handleClientMessage(msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		c.ctx.LogWarnf("[Client %s] Received message without type", c.ID)
		return
	}

	switch msgType {
	case "client-id":
		cidMsg := models.NewEventMessage(constants.EventTypeSystem, c.ID, nil)
		cidMsg.ClientID = c.ID
		select {
		case c.Send <- cidMsg:
		default:
			c.ctx.LogWarnf("[Client %s] Failed to send client-id (channel full)", c.ID)
		}
	default:
		c.ctx.LogDebugf("[Client %s] Unknown message type: %s", c.ID, msgType)
	}
}
