package orchestrator

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	dataModels "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHostWebSocketManager() *HostWebSocketManager {
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	return &HostWebSocketManager{
		ctx:           ctx,
		clients:       make(map[string]*HostWebSocketClient),
		handlers:      make(map[constants.EventType]map[interfaces.HostEventHandler]bool),
		probeInFlight: make(map[string]bool),
		stopChan:      make(chan struct{}),
	}
}

func TestSyncConnections_SkipsProbeWhenClientAlreadyExists(t *testing.T) {
	manager := newTestHostWebSocketManager()
	manager.clients["host-a"] = &HostWebSocketClient{
		ctx:      manager.ctx,
		hostID:   "host-a",
		stopChan: make(chan struct{}),
	}

	probed := make(chan string, 2)
	manager.probeHost = func(host dataModels.OrchestratorHost) {
		probed <- host.ID
	}

	hosts := []dataModels.OrchestratorHost{
		{ID: "host-a", Host: "10.0.0.1", Enabled: true},
		{ID: "host-b", Host: "10.0.0.2", Enabled: true},
	}

	manager.syncConnections(hosts)

	select {
	case hostID := <-probed:
		assert.Equal(t, "host-b", hostID)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected a probe for the host without an existing client")
	}

	select {
	case hostID := <-probed:
		t.Fatalf("unexpected extra probe for host %s", hostID)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestConnectHost_ReplacesDisconnectedClient(t *testing.T) {
	manager := newTestHostWebSocketManager()
	oldClient := &HostWebSocketClient{
		ctx:      manager.ctx,
		hostID:   "host-a",
		stopChan: make(chan struct{}),
	}
	manager.clients["host-a"] = oldClient

	var startedClient *HostWebSocketClient
	manager.startClient = func(client *HostWebSocketClient, events []constants.EventType) {
		startedClient = client
	}

	host := &dataModels.OrchestratorHost{
		ID:      "host-a",
		Host:    "10.0.0.1",
		Enabled: true,
	}

	manager.ConnectHost(host, []constants.EventType{constants.EventTypeHealth})

	require.NotNil(t, startedClient)
	require.NotSame(t, oldClient, startedClient)
	assert.Same(t, startedClient, manager.clients["host-a"])

	select {
	case <-oldClient.stopChan:
	default:
		t.Fatal("expected previous disconnected client to be closed")
	}
}
