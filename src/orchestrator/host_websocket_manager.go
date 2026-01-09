package orchestrator

import (
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	event "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// HostWebSocketManager manages WebSocket connections to hosts
type HostWebSocketManager struct {
	ctx           basecontext.ApiContext
	clients       map[string]*HostWebSocketClient                              // hostID -> HostWebSocketClient
	handlers      map[constants.EventType]map[interfaces.HostEventHandler]bool // eventType -> handlers
	mu            sync.RWMutex
	handlersMu    sync.RWMutex
	stopChan      chan struct{}
	refreshTicker *time.Ticker
}

var (
	managerInstance *HostWebSocketManager
	managerOnce     sync.Once
)

// NewHostWebSocketManager creates or returns the singleton instance
func NewHostWebSocketManager(ctx basecontext.ApiContext) *HostWebSocketManager {
	managerOnce.Do(func() {
		managerInstance = &HostWebSocketManager{
			ctx:      ctx,
			clients:  make(map[string]*HostWebSocketClient),
			handlers: make(map[constants.EventType]map[interfaces.HostEventHandler]bool),
			stopChan: make(chan struct{}),
		}
	})
	return managerInstance
}

// GetHostWebSocketManager returns the singleton instance
func GetHostWebSocketManager() *HostWebSocketManager {
	return managerInstance
}

// RegisterHandler registers a handler for specific event types
func (m *HostWebSocketManager) RegisterHandler(eventTypes []constants.EventType, handler interfaces.HostEventHandler) {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()

	for _, eventType := range eventTypes {
		if m.handlers[eventType] == nil {
			m.handlers[eventType] = make(map[interfaces.HostEventHandler]bool)
		}
		m.handlers[eventType][handler] = true
	}
}

// DispatchMessage dispatches a message to registered handlers
func (m *HostWebSocketManager) DispatchMessage(hostID string, eventType constants.EventType, payload []byte) {
	m.handlersMu.RLock()
	handlers, exists := m.handlers[eventType]
	m.handlersMu.RUnlock()

	if !exists {
		m.ctx.LogDebugf("[HostWebSocketManager] No handlers registered for event type: %s", eventType)
		return
	}
	for handler := range handlers {
		go handler.Handle(m.ctx, hostID, eventType, payload)
	}
}

// ConnectHost initiates a WebSocket connection to the host
func (m *HostWebSocketManager) ConnectHost(host *models.OrchestratorHost, events []constants.EventType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[host.ID]; exists {
		m.ctx.LogInfof("[HostWebSocketManager] Host %s already connected", host.Host)
		return
	}

	client := NewHostWebSocketClient(m.ctx, host, m)
	m.clients[host.ID] = client
	go client.Connect(events)
}

// DisconnectHost closes the WebSocket connection to the host
func (m *HostWebSocketManager) DisconnectHost(hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.clients[hostID]; exists {
		client.Close()
		delete(m.clients, hostID)
		m.ctx.LogInfof("[HostWebSocketManager] Disconnected host %s", hostID)

		// Update DB status
		dbService, err := serviceprovider.GetDatabaseService(m.ctx)
		if err == nil {
			updated, _ := dbService.UpdateOrchestratorHostWebsocketStatus(m.ctx, hostID, false)
			if updated {
				if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
					msg := event.NewEventMessage(constants.EventTypeOrchestrator, "HOST_WEBSOCKET_DISCONNECTED",
						event.HostHealthUpdate{
							HostID: hostID,
							State:  "websocket_disconnected",
						})
					go func() {
						if err := emitter.Broadcast(msg); err != nil {
							m.ctx.LogErrorf("[HostWebSocketManager] Failed to broadcast HOST_WEBSOCKET_DISCONNECTED event: %v", err)
						} else {
							m.ctx.LogInfof("[HostWebSocketManager] Broadcasted HOST_WEBSOCKET_DISCONNECTED event for host %s", hostID)
						}
					}()
				}
			}
		}
	}
}

// IsConnected checks if a host has an active connection
func (m *HostWebSocketManager) IsConnected(hostID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if client, exists := m.clients[hostID]; exists {
		return client.IsConnected()
	}
	return false
}

// SendPing sends a ping message to the host
func (m *HostWebSocketManager) SendPing(hostID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if client, exists := m.clients[hostID]; exists {
		return client.SendPing()
	}
	return fmt.Errorf("host %s not connected", hostID)
}

// RefreshConnections synchronizes active connections with the provided list of hosts
func (m *HostWebSocketManager) RefreshConnections(hosts []models.OrchestratorHost) {
	m.ctx.LogDebugf("[HostWebSocketManager] Refreshing host connections - found %d hosts", len(hosts))

	m.syncConnections(hosts)
}

// Shutdown closes all WebSocket connections and cleans up resources
func (m *HostWebSocketManager) Shutdown() {
	m.ctx.LogInfof("[HostWebSocketManager] Shutting down...")

	// Stop the connection monitor
	if m.stopChan != nil {
		close(m.stopChan)
	}
	if m.refreshTicker != nil {
		m.refreshTicker.Stop()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, client := range m.clients {
		client.Close()
		delete(m.clients, id)
	}
	m.ctx.LogInfof("[HostWebSocketManager] Shutdown complete")
}

func (m *HostWebSocketManager) ProbeAndConnect(host models.OrchestratorHost) {
	// Create a temporary client to probe
	client := NewHostWebSocketClient(m.ctx, &host, nil) // No manager needed for probe
	if client.Probe() {
		m.handlersMu.RLock()
		eventTypes := make([]constants.EventType, 0, len(m.handlers))
		for eventType := range m.handlers {
			eventTypes = append(eventTypes, eventType)
		}
		m.handlersMu.RUnlock()
		m.ConnectHost(&host, eventTypes)
	} else {
		m.ctx.LogDebugf("[HostWebSocketManager] Probe failed for host %s, skipping WebSocket connection", host.Host)
	}
}

// StartConnectionMonitor starts a background goroutine that periodically checks
// for disconnected hosts and attempts to reconnect them
func (m *HostWebSocketManager) StartConnectionMonitor(checkInterval time.Duration) {
	m.refreshTicker = time.NewTicker(checkInterval)

	go func() {
		m.ctx.LogInfof("[HostWebSocketManager] Starting connection monitor (interval: %v)", checkInterval)
		for {
			select {
			case <-m.stopChan:
				m.ctx.LogInfof("[HostWebSocketManager] Connection monitor stopped")
				return
			case <-m.refreshTicker.C:
				m.checkAndReconnectHosts()
			}
		}
	}()
}

// checkAndReconnectHosts checks all enabled hosts and attempts to reconnect disconnected ones
func (m *HostWebSocketManager) checkAndReconnectHosts() {
	// Get database service
	dbService, err := serviceprovider.GetDatabaseService(m.ctx)
	if err != nil {
		m.ctx.LogErrorf("[HostWebSocketManager] Error getting database service: %v", err)
		return
	}

	// Get all hosts from database
	hosts, err := dbService.GetOrchestratorHosts(m.ctx, "")
	if err != nil {
		m.ctx.LogErrorf("[HostWebSocketManager] Error getting hosts: %v", err)
		return
	}

	m.syncConnections(hosts)
}

func (m *HostWebSocketManager) syncConnections(hosts []models.OrchestratorHost) {
	activeHostIDs := make(map[string]bool)

	// Check each enabled host and attempt to connect if needed
	for _, host := range hosts {
		if !host.Enabled {
			continue
		}

		activeHostIDs[host.ID] = true

		// If host is not connected, attempt to connect
		if !m.IsConnected(host.ID) {
			m.ctx.LogDebugf("[HostWebSocketManager] Host %s is not connected, attempting reconnection", host.Host)
			hostCopy := host
			go m.ProbeAndConnect(hostCopy)
		}
	}

	// Disconnect hosts that are no longer enabled or in the database
	m.mu.Lock()
	for hostID := range m.clients {
		if !activeHostIDs[hostID] {
			m.ctx.LogDebugf("[HostWebSocketManager] Host %s is no longer enabled or in database, disconnecting", hostID)
			go m.DisconnectHost(hostID)
		}
	}
	m.mu.Unlock()
}
