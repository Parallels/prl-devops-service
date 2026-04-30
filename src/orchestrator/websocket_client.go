package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
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
	wsWriteDeadline   = 10 * time.Second
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
	pingStop    chan struct{}
	pingTicker  *time.Ticker
	pongWait    time.Duration
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
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		if resp != nil {
			body := readHandshakeBody(resp)
			if body != "" {
				c.ctx.LogWarnf("[HostWebSocketClient] Connection to host %s failed with status %d: %s", c.hostName, resp.StatusCode, body)
			} else {
				c.ctx.LogWarnf("[HostWebSocketClient] Connection to host %s failed with status %d", c.hostName, resp.StatusCode)
			}
		}
		return err
	}

	c.conn = conn
	c.setConnected(true)
	c.ctx.LogInfof("[HostWebSocketClient] Connected to host %s", c.hostName)

	// Compute pong wait from config (3× ping interval gives us headroom for 2 missed pings)
	cfg := config.Get()
	pingInterval := time.Duration(cfg.OrchestratorPullFrequency()) * time.Second
	c.pongWait = pingInterval * 3

	// Set pong handler: extend read deadline and dispatch a synthetic pong event so that
	// HostHealthHandler.handlePong updates UpdatedAt in the DB. This is the ONLY place
	// that should update UpdatedAt — all other direct writes have been removed.
	c.conn.SetPongHandler(func(_ string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
		manager := GetHostWebSocketManager()
		if manager != nil {
			manager.DispatchMessage(c.hostID, constants.EventTypeHealth, []byte(`{"message":"pong"}`))
		}
		return nil
	})

	// Set an initial read deadline; the pong handler will keep extending it.
	_ = c.conn.SetReadDeadline(time.Now().Add(c.pongWait))

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
				c.ctx.LogErrorf("[Orchestrator] [WS Read] Error unmarshalling message from host %s: %v", c.hostName, err)
				continue
			}

			// Log all incoming events for debugging — helps identify which types reach the orchestrator
			// Exclude noisy stat and log events to keep logs focused on job tracking
			if event.Type != constants.EventTypeStats && event.Type != constants.EventTypeSystemLogs {
				c.ctx.LogDebugf("[Orchestrator] [WS Received] host=%s type=%q message=%q id=%s", c.hostName, event.Type, event.Message, event.ID)
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
	c.conn.SetWriteDeadline(time.Now().Add(wsWriteDeadline))
	return c.conn.WriteJSON(message)
}

func (c *HostWebSocketClient) SendPing() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	c.ctx.LogDebugf("[HostWebSocketClient] Sending ping to host %s", c.hostID)
	return c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(wsWriteDeadline))
}

func (c *HostWebSocketClient) startPingRoutine() {
	c.mu.Lock()
	// Stop the previous ping goroutine by closing its dedicated stop channel.
	// This prevents goroutine accumulation across reconnects where old goroutines
	// would otherwise continue competing for the new ticker.
	if c.pingStop != nil {
		close(c.pingStop)
	}
	c.pingStop = make(chan struct{})
	pingStop := c.pingStop // capture by value so the goroutine owns this reference

	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
	cfg := config.Get()
	c.pingTicker = time.NewTicker(time.Duration(cfg.OrchestratorPullFrequency()) * time.Second)
	ticker := c.pingTicker // capture by value to avoid the goroutine reading c.pingTicker after replacement
	c.mu.Unlock()

	go func() {
		// Send initial ping immediately to confirm connection status
		if err := c.SendPing(); err != nil {
			c.ctx.LogDebugf("[HostWebSocketClient] Failed to send initial ping to host %s: %v", c.hostName, err)
		}

		for {
			select {
			case <-c.stopChan:
				return
			case <-pingStop:
				return
			case <-ticker.C:
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

	// Update DB status
	if dbService, err := serviceprovider.GetDatabaseService(c.ctx); err == nil {
		updated, _ := dbService.UpdateOrchestratorHostWebsocketStatus(c.ctx, c.hostID, false)
		if updated {
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
			}
		}
	} else {
		c.ctx.LogWarnf("[HostWebSocketClient] EventEmitter not available to broadcast disconnection for host %s", c.hostName)
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

	conn, resp, err := dialer.Dial(u.String(), header)
	if err != nil {
		if resp != nil {
			body := readHandshakeBody(resp)
			if body != "" {
				c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: connection error: %v (status: %d, body: %s)", err, resp.StatusCode, body)
			} else {
				c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: connection error: %v (status: %d)", err, resp.StatusCode)
			}
		} else {
			c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: connection error: %v", err)
		}
		return false
	}
	defer conn.Close()

	// Use a protocol-level ping; the host auto-responds with a pong frame.
	pongCh := make(chan struct{}, 1)
	conn.SetPongHandler(func(_ string) error {
		select {
		case pongCh <- struct{}{}:
		default:
		}
		return nil
	})

	// Drive reads in a goroutine so the pong handler fires
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(wsWriteDeadline)); err != nil {
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: write error: %v", err)
		return false
	}

	select {
	case <-pongCh:
		c.ctx.LogInfof("[HostWebSocketClient] Probe successful for host %s", c.hostName)
		return true
	case <-time.After(3 * time.Second):
		c.ctx.LogWarnf("[HostWebSocketClient] Probe failed: no pong received from host %s", c.hostName)
		return false
	}
}

func readHandshakeBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil {
		return ""
	}

	return helpers.CleanOutputString(string(body))
}
