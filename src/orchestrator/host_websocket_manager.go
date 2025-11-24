package orchestrator

import (
	"fmt"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
)

// HostWebSocketManager manages WebSocket connections to hosts
type HostWebSocketManager struct {
	ctx        basecontext.ApiContext
	clients    map[string]*HostWebSocketClient                              // hostID -> HostWebSocketClient
	handlers   map[constants.EventType]map[interfaces.HostEventHandler]bool // eventType -> handlers
	mu         sync.RWMutex
	handlersMu sync.RWMutex
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
	m.ctx.LogInfof("[HostWebSocketManager] Refreshing host connections")

	activeHostIDs := make(map[string]bool)

	// Connect new hosts or reconnect if needed
	for _, host := range hosts {
		if !host.Enabled {
			continue
		}
		activeHostIDs[host.ID] = true

		// We pass a copy of the host to avoid pointer issues in loop
		hostCopy := host
		if !m.IsConnected(host.ID) {
			m.ProbeAndConnect(hostCopy)
		}
	}

	// Disconnect hosts that are no longer in the list or disabled
	m.mu.Lock()
	for hostID := range m.clients {
		if !activeHostIDs[hostID] {
			go m.DisconnectHost(hostID) // Disconnect in background to avoid locking issues if DisconnectHost also locks
		}
	}
	m.mu.Unlock()
}

// Shutdown closes all WebSocket connections and cleans up resources
func (m *HostWebSocketManager) Shutdown() {
	m.ctx.LogInfof("[HostWebSocketManager] Shutting down...")
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
		m.ctx.LogInfof("[HostWebSocketManager] Probe failed for host %s, skipping WebSocket connection", host.Host)
	}
}
