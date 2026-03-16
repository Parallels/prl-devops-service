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
	writeWait         = 10 * time.Second
	pingInterval      = 30 * time.Second
	pongWait          = 45 * time.Second
	queueTickInterval = 100 * time.Millisecond
	// maxClientQueue is the per-client queue cap. Oldest messages are evicted
	// once this limit is reached (tail-drop prevention).
	maxClientQueue = 2048
)

// enqueue appends a message to the client's pending queue.
// If the queue is at capacity the oldest message is evicted to make room,
// keeping the queue bounded without silently dropping the new message.
func (c *Client) enqueue(msg *models.EventMessage) {
	c.pendingMu.Lock()
	if len(c.pending) >= maxClientQueue {
		c.ctx.LogWarnf("[Client %s] Queue full (%d), evicting oldest message", c.ID, maxClientQueue)
		c.pending = c.pending[1:]
	}
	c.pending = append(c.pending, msg)
	c.pendingMu.Unlock()
}

// clientQueueWorker drains the per-client pending queue every queueTickInterval
// and writes each message to the WebSocket. It also sends periodic pings and
// disconnects the client if a pong is not received within pongWait.
func (c *Client) clientQueueWorker() {
	ticker := time.NewTicker(queueTickInterval)
	pingTicker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	defer pingTicker.Stop()
	defer func() {
		c.ctx.LogInfof("[Client %s] Closing queue worker", c.ID)
		c.mu.Lock()
		if c.IsAlive {
			c.IsAlive = false
			c.Conn.Close()
			c.mu.Unlock()
			Get().hub.trySendCommand(&unregisterClientCmd{clientID: c.ID})
		} else {
			c.mu.Unlock()
		}
	}()

	// Configure pong handler — updates LastPongAt on every pong received.
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.LastPongAt = time.Now()
		c.mu.Unlock()
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.done:
			return
		case <-pingTicker.C:
			c.mu.Lock()
			c.LastPingAt = time.Now()
			c.mu.Unlock()
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.ctx.LogWarnf("[Client %s] Ping failed: %v", c.ID, err)
				return
			}
		case <-ticker.C:
			if err := c.flushQueue(); err != nil {
				c.ctx.LogErrorf("[Client %s] Write error, disconnecting: %v", c.ID, err)
				return
			}
		}
	}
}

// flushQueue drains all pending messages and writes them to the WebSocket.
// Returns an error if any write fails.
func (c *Client) flushQueue() error {
	c.pendingMu.Lock()
	if len(c.pending) == 0 {
		c.pendingMu.Unlock()
		return nil
	}
	msgs := c.pending
	c.pending = make([]*models.EventMessage, 0, 64)
	c.pendingMu.Unlock()

	for _, msg := range msgs {
		c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.Conn.WriteJSON(msg); err != nil {
			return err
		}
	}
	return nil
}

// clientReader reads inbound messages from the WebSocket connection.
func (c *Client) clientReader() {
	defer func() {
		c.ctx.LogInfof("[Client %s] Closing reader", c.ID)
		c.mu.Lock()
		if c.IsAlive {
			c.IsAlive = false
			c.Conn.Close()
			c.mu.Unlock()
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

// routingHeader is a lightweight struct for partial parsing.
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
		msg.Body = map[string]interface{}{"error": err.Error()}
		c.enqueue(msg)
		return
	}

	if !constants.EventType.IsValid(header.Type) {
		c.ctx.LogWarnf("[Client %s] Received message with invalid type: %s", c.ID, header.Type)
		msg := models.NewEventMessage(constants.EventTypeGlobal, c.ID, nil)
		msg.Message = "error"
		msg.Body = map[string]interface{}{"error": "invalid message type " + header.Type.String()}
		c.enqueue(msg)
		return
	}

	msgID := header.ID
	if msgID == "" {
		msgID = helpers.GenerateId()
	}

	cmd := &RouteMessageCmd{
		ClientID: c.ID,
		Type:     header.Type,
		Payload:  rawMsg,
		MsgID:    msgID,
	}

	select {
	case Get().hub.clientToHub <- cmd:
	default:
		c.ctx.LogWarnf("[Client %s] Hub command channel full, dropping message %s", c.ID, msgID)
	}
}
