package eventemitter

import (
	"encoding/json"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
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

		c.handleClientMessage(message)
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

// routingHeader is a lightweight struct for partial parsing
type routingHeader struct {
	Type constants.EventType `json:"event_type"`
	ID   string              `json:"id"`
}

func (c *Client) handleClientMessage(rawMsg []byte) {
	var header routingHeader
	if err := json.Unmarshal(rawMsg, &header); err != nil {
		c.ctx.LogWarnf("[Client %s] Failed to parse message header: %v", c.ID, err)
		msg := models.NewEventMessage(constants.EventTypeGlobal, c.ID, nil)
		msg.Message = "error"
		msg.Body = map[string]interface{}{
			"error": err.Error(),
		}
		c.Send <- msg
		return
	}

	if !constants.EventType.IsValid(header.Type) {
		c.ctx.LogWarnf("[Client %s] Received message with invalid type: %s", c.ID, header.Type)
		msg := models.NewEventMessage(constants.EventTypeGlobal, c.ID, nil)
		msg.Message = "error"
		msg.Body = map[string]interface{}{
			"error": "invalid message type " + header.Type.String(),
		}
		c.Send <- msg
		return
	}

	// Generate message ID if missing
	msgID := header.ID
	if msgID == "" {
		msgID = helpers.GenerateId()
	}

	// Create command and send to hub channel
	// We pass the original rawMsg directly - ZERO extra allocations/marshaling
	cmd := &RouteMessageCmd{
		ClientID: c.ID,
		Type:     header.Type,
		Payload:  rawMsg,
		MsgID:    msgID,
	}

	// Non-blocking send to avoid deadlocks
	select {
	case Get().hub.clientToHub <- cmd:
	default:
		c.ctx.LogWarnf("[Client %s] Hub command channel full, dropping message %s", c.ID, msgID)
	}
}
