package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/gorilla/websocket"
)

const (
	reconnectInterval = 5 * time.Second
	maxReconnectDelay = 1 * time.Minute
)

type HostWebSocketClient struct {
	ctx         basecontext.ApiContext
	hostID      string
	hostName    string
	hostPort    string
	hostSchema  string
	hostAuth    *models.OrchestratorHostAuthentication
	conn        *websocket.Conn
	isConnected bool
	mu          sync.RWMutex
	stopChan    chan struct{}
	pingTicker  *time.Ticker
}

func NewHostWebSocketClient(ctx basecontext.ApiContext, host *models.OrchestratorHost, manager *HostWebSocketManager) *HostWebSocketClient {
	return &HostWebSocketClient{
		ctx:        ctx,
		hostID:     host.ID,
		hostName:   host.Host,
		hostPort:   host.Port,
		hostSchema: host.Schema,
		hostAuth:   host.Authentication,
		stopChan:   make(chan struct{}),
	}
}

func (c *HostWebSocketClient) Connect(events []constants.EventType) {
	backoff := reconnectInterval

	for {
		select {
		case <-c.stopChan:
			return
		default:
			if err := c.establishConnection(events); err != nil {
				c.ctx.LogErrorf("[HostWebSocketClient] Failed to connect to host %s: %v. Retrying in %v", c.hostName, err, backoff)
				time.Sleep(backoff)
				if backoff < maxReconnectDelay {
					backoff *= 2
				}
			} else {
				backoff = reconnectInterval
				c.startPingRoutine()
				c.readLoop()

				// Check if we are stopping
				select {
				case <-c.stopChan:
					return
				default:
				}

				// If readLoop returns and we are not stopping, it means connection was closed or error occurred
				c.setConnected(false)
				c.ctx.LogWarnf("[HostWebSocketClient] Connection to host %s lost. Reconnecting...", c.hostName)
			}
		}
	}
}

func (c *HostWebSocketClient) establishConnection(events []constants.EventType) error {
	path := "/api/v1/ws/subscribe"
	c.ctx.LogInfof("[HostWebSocketClient] Establishing connection to host %s", c.hostName)
	scheme := c.hostSchema
	if scheme == "" {
		scheme = "http"
	}

	hostStr := c.hostName
	if c.hostPort != "" {
		hostStr = fmt.Sprintf("%s:%s", c.hostName, c.hostPort)
	}

	baseUrl := fmt.Sprintf("%s://%s", scheme, hostStr)

	u, err := helpers.JoinUrl([]string{baseUrl, path})
	if err != nil {
		return err
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	q := u.Query()

	eventStrings := make([]string, len(events))
	for i, event := range events {
		eventStrings[i] = string(event)
	}
	q.Set("event_types", strings.Join(eventStrings, ","))
	u.RawQuery = q.Encode()

	// Use shared helper to get authentication header
	host := models.OrchestratorHost{
		Host:           c.hostName,
		Port:           c.hostPort,
		Schema:         c.hostSchema,
		Authentication: c.hostAuth,
	}
	header, err := getAuthHeaderForWebSocket(c.ctx, host)
	if err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Failed to get auth header: %v", err)
		header = http.Header{} // Use empty header on error
	}

	c.ctx.LogInfof("[HostWebSocketClient] Connecting to %s", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return err
	}

	c.conn = conn
	c.setConnected(true)
	c.ctx.LogInfof("[HostWebSocketClient] Connected to host %s", c.hostName)

	c.broadcastConnectionEvent()

	return nil
}

func (c *HostWebSocketClient) readLoop() {
	defer func() {
		c.conn.Close()
		c.notifyDisconnection()
	}()

	for {
		select {
		case <-c.stopChan:
			c.ctx.LogDebugf("[HostWebSocketClient] ReadLoop stopping for host %s (stop signal)", c.hostName)
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				// Check if we are stopping
				select {
				case <-c.stopChan:
					c.ctx.LogDebugf("[HostWebSocketClient] ReadLoop stopping for host %s (stop signal)", c.hostName)
					return
				default:
					c.ctx.LogErrorf("[HostWebSocketClient] Error reading message from host %s: %v", c.hostName, err)
				}
				return
			}

			var event api_models.EventMessage
			if err := json.Unmarshal(message, &event); err != nil {
				c.ctx.LogErrorf("[HostWebSocketClient] Error unmarshalling message from host %s: %v", c.hostName, err)
				continue
			}

			// Dispatch message to manager
			manager := GetHostWebSocketManager()
			if manager != nil {
				manager.DispatchMessage(c.hostID, event.Type, message)
			}
		}
	}
}

func (c *HostWebSocketClient) Send(message interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WriteJSON(message)
}

func (c *HostWebSocketClient) SendPing() error {
	pingMsg := map[string]string{
		"event_type": string(constants.EventTypeHealth),
		"message":    "ping",
	}
	c.ctx.LogInfof("[HostWebSocketClient] Sending ping to host %s", c.hostID)
	return c.Send(pingMsg)
}

func (c *HostWebSocketClient) startPingRoutine() {
	c.mu.Lock()
	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
	cfg := config.Get()
	c.pingTicker = time.NewTicker(time.Duration(cfg.OrchestratorPullFrequency()) * time.Second)
	c.mu.Unlock()

	go func() {
		for {
			select {
			case <-c.stopChan:
				return
			case <-c.pingTicker.C:
				if err := c.SendPing(); err != nil {
					c.ctx.LogDebugf("[HostWebSocketClient] Failed to send periodic ping to host %s: %v", c.hostName, err)
				}
			}
		}
	}()
}

func (c *HostWebSocketClient) Close() {
	close(c.stopChan)
	c.mu.Lock()
	if c.pingTicker != nil {
		c.pingTicker.Stop()
		c.pingTicker = nil
	}
	c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.setConnected(false)
}

func (c *HostWebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

func (c *HostWebSocketClient) setConnected(connected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isConnected = connected
}

func (c *HostWebSocketClient) notifyDisconnection() {
	c.ctx.LogWarnf("[HostWebSocketClient] Host %s WebSocket disconnection detected - current state: connected=%v", c.hostName, c.IsConnected())

	// Always broadcast disconnection event if we're not intentionally stopping
	select {
	case <-c.stopChan:
		c.ctx.LogDebugf("[HostWebSocketClient] Host %s disconnection was intentional (stop signal), skipping event", c.hostName)
		return
	default:
		// This is an unexpected disconnection
	}

	c.setConnected(false)

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := api_models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_WEBSOCKET_DISCONNECTED", api_models.HostHealthUpdate{
			HostID: c.hostID,
			State:  "websocket_disconnected",
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				c.ctx.LogErrorf("[HostWebSocketClient] Failed to broadcast HOST_WEBSOCKET_DISCONNECTED event: %v", err)
			} else {
				c.ctx.LogInfof("[HostWebSocketClient] Broadcasted HOST_WEBSOCKET_DISCONNECTED event for host %s", c.hostName)
			}
		}()
	} else {
		c.ctx.LogWarnf("[HostWebSocketClient] EventEmitter not available to broadcast disconnection for host %s", c.hostName)
	}
}

func (c *HostWebSocketClient) broadcastConnectionEvent() {
	c.ctx.LogInfof("[HostWebSocketClient] Host %s WebSocket connection established", c.hostName)

	// Broadcast WebSocket connection event
	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := api_models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_WEBSOCKET_CONNECTED", api_models.HostHealthUpdate{
			HostID: c.hostID,
			State:  "websocket_connected",
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				c.ctx.LogErrorf("[HostWebSocketClient] Failed to broadcast HOST_WEBSOCKET_CONNECTED event: %v", err)
			} else {
				c.ctx.LogInfof("[HostWebSocketClient] Broadcasted HOST_WEBSOCKET_CONNECTED event for host %s", c.hostName)
			}
		}()
	} else {
		c.ctx.LogWarnf("[HostWebSocketClient] EventEmitter not available to broadcast connection for host %s", c.hostName)
	}
}

func (c *HostWebSocketClient) Probe() bool {
	// Use a short timeout for probing
	c.ctx.LogInfof("[HostWebSocketClient] Probing host %s for WebSocket support...", c.hostName)

	path := "/api/v1/ws/subscribe"
	scheme := c.hostSchema
	if scheme == "" {
		scheme = "http"
	}
	hostStr := c.hostName
	if c.hostPort != "" {
		hostStr = fmt.Sprintf("%s:%s", c.hostName, c.hostPort)
	}
	baseUrl := fmt.Sprintf("%s://%s", scheme, hostStr)
	hostUrl, err := helpers.JoinUrl([]string{baseUrl, path})
	if err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: invalid URL: %v", err)
		return false
	}

	u, err := url.Parse(hostUrl.String())
	if err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: invalid URL parse: %v", err)
		return false
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	// Add query params
	q := u.Query()
	q.Set("event_types", string(constants.EventTypeHealth))
	u.RawQuery = q.Encode()

	// Use shared helper to get authentication header
	host := models.OrchestratorHost{
		Host:           c.hostName,
		Port:           c.hostPort,
		Schema:         c.hostSchema,
		Authentication: c.hostAuth,
	}
	header, err := getAuthHeaderForWebSocket(c.ctx, host)
	if err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: auth error: %v", err)
		return false
	}

	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = 2 * time.Second

	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: connection error: %v", err)
		return false
	}
	defer conn.Close()

	// Send Ping
	pingMsg := map[string]string{
		"type":    string(constants.EventTypeHealth),
		"message": "ping",
	}
	if err := conn.WriteJSON(pingMsg); err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: write error: %v", err)
		return false
	}

	// Wait for Pong (Read message)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn.ReadMessage()
	if err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: read error (timeout or closed): %v", err)
		return false
	}

	// We received a message, assuming it's a pong or at least the connection works
	c.ctx.LogInfof("[HostWebSocketClient] Probe successful for host %s", c.hostName)
	return true
}
