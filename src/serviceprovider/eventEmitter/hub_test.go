package eventemitter

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestHub() *Hub {
	ctx := basecontext.NewBaseContext()
	return &Hub{
		ctx:           ctx,
		clients:       make(map[string]*Client),
		clientsByIP:   make(map[string]*Client),
		subscriptions: make(map[constants.EventType]map[string]bool),
		broadcast:     make(chan *models.EventMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

func createTestClient(id string, username string, subscriptions []constants.EventType) *Client {
	return &Client{
		ID: id,
		User: &models.ApiUser{
			ID:       id,
			Username: username,
			Email:    username + "@test.com",
			Roles:    []string{},
			Claims:   []string{},
		},
		Hub:           nil,
		Conn:          nil, // Not needed for these tests
		Send:          make(chan *models.EventMessage, 10),
		Subscriptions: subscriptions,
		ConnectedAt:   time.Now(),
		LastPingAt:    time.Now(),
		LastPongAt:    time.Now(),
		IsAlive:       true,
	}
}

func TestHub_RegisterClient(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("client1", "testuser", []constants.EventType{constants.EventTypePDFM})

	hub.registerClient(client)

	// Check client is registered
	assert.Len(t, hub.clients, 1, "Should have one client")
	assert.Contains(t, hub.clients, "client1", "Client should be in clients map")

	// Check subscriptions
	assert.Contains(t, client.Subscriptions, constants.EventTypePDFM, "Should have PDFM subscription")
	assert.Contains(t, client.Subscriptions, constants.EventTypeGlobal, "Should have auto-added GLOBAL subscription")

	// Check hub subscriptions map
	assert.True(t, hub.subscriptions[constants.EventTypePDFM]["client1"], "Should be subscribed to PDFM")
	assert.True(t, hub.subscriptions[constants.EventTypeGlobal]["client1"], "Should be subscribed to GLOBAL")
}

func TestHub_RegisterClient_AutoGlobalSubscription(t *testing.T) {
	hub := createTestHub()

	// Client without global subscription
	client := createTestClient("client1", "testuser", []constants.EventType{constants.EventTypeVM})

	hub.registerClient(client)

	// Global should be auto-added
	assert.Contains(t, client.Subscriptions, constants.EventTypeGlobal, "Global should be auto-added")
	assert.True(t, hub.subscriptions[constants.EventTypeGlobal]["client1"], "Should be subscribed to global")
	assert.Len(t, client.Subscriptions, 2, "Should have 2 subscriptions (vm + global)")
}

func TestHub_RegisterClient_AlreadyHasGlobal(t *testing.T) {
	hub := createTestHub()

	// Client with global already in subscriptions
	client := createTestClient("client1", "testuser", []constants.EventType{constants.EventTypeGlobal, constants.EventTypeHost})

	hub.registerClient(client)

	// Should not duplicate global
	globalCount := 0
	for _, sub := range client.Subscriptions {
		if sub == constants.EventTypeGlobal {
			globalCount++
		}
	}
	assert.Equal(t, 1, globalCount, "Global should appear only once")
	assert.Len(t, client.Subscriptions, 2, "Should have 2 subscriptions")
}

func TestHub_RegisterClient_MultipleSubscriptions(t *testing.T) {
	hub := createTestHub()

	client := createTestClient("client1", "testuser", []constants.EventType{
		constants.EventTypeVM,
		constants.EventTypeHost,
		constants.EventTypeSystem,
	})

	hub.registerClient(client)

	// Check all subscriptions are registered
	assert.True(t, hub.subscriptions[constants.EventTypeVM]["client1"])
	assert.True(t, hub.subscriptions[constants.EventTypeHost]["client1"])
	assert.True(t, hub.subscriptions[constants.EventTypeSystem]["client1"])
	assert.True(t, hub.subscriptions[constants.EventTypeGlobal]["client1"])
	assert.Len(t, client.Subscriptions, 4, "Should have 4 subscriptions (3 + global)")
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("client1", "testuser", []constants.EventType{constants.EventTypePDFM})

	hub.registerClient(client)
	require.Len(t, hub.clients, 1)

	hub.unregisterClient(client)

	// Check client is removed
	assert.Len(t, hub.clients, 0, "Should have no clients")
	assert.NotContains(t, hub.clients, "client1", "Client should not be in clients map")

	// Check subscriptions are cleaned up
	if pdfmSubs, exists := hub.subscriptions[constants.EventTypePDFM]; exists {
		assert.NotContains(t, pdfmSubs, "client1", "Client should not be in PDFM subscriptions")
	}
	if globalSubs, exists := hub.subscriptions[constants.EventTypeGlobal]; exists {
		assert.NotContains(t, globalSubs, "client1", "Client should not be in GLOBAL subscriptions")
	}
}

func TestHub_UnregisterClient_NonExistent(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("client1", "testuser", []constants.EventType{constants.EventTypePDFM})

	// Try to unregister a client that was never registered
	// Should not panic
	assert.NotPanics(t, func() {
		hub.unregisterClient(client)
	})
}

func TestHub_UnregisterClient_CleansUpEmptySubscriptions(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("client1", "testuser", []constants.EventType{constants.EventTypePDFM})

	hub.registerClient(client)
	hub.unregisterClient(client)

	// Empty subscription maps should be removed
	_, pdfmExists := hub.subscriptions[constants.EventTypePDFM]
	assert.False(t, pdfmExists, "Empty PDFM subscription map should be removed")
}

func TestHub_MultipleClients(t *testing.T) {
	hub := createTestHub()
	client1 := createTestClient("client1", "user1", []constants.EventType{constants.EventTypeVM})
	client2 := createTestClient("client2", "user2", []constants.EventType{constants.EventTypeVM})
	client3 := createTestClient("client3", "user3", []constants.EventType{constants.EventTypeHost})

	hub.registerClient(client1)
	hub.registerClient(client2)
	hub.registerClient(client3)

	// Check all clients registered
	assert.Len(t, hub.clients, 3, "Should have 3 clients")

	// Check VM subscriptions
	assert.Len(t, hub.subscriptions[constants.EventTypeVM], 2, "VM should have 2 subscribers")
	assert.True(t, hub.subscriptions[constants.EventTypeVM]["client1"])
	assert.True(t, hub.subscriptions[constants.EventTypeVM]["client2"])

	// Check HOST subscriptions
	assert.Len(t, hub.subscriptions[constants.EventTypeHost], 1, "HOST should have 1 subscriber")
	assert.True(t, hub.subscriptions[constants.EventTypeHost]["client3"])

	// Check GLOBAL subscriptions (all 3 should be subscribed)
	assert.Len(t, hub.subscriptions[constants.EventTypeGlobal], 3, "GLOBAL should have 3 subscribers")
}

func TestHub_RegisterClient_InvalidEventTypes(t *testing.T) {
	hub := createTestHub()

	// Client with mix of valid and invalid event types
	client := createTestClient("client1", "testuser", []constants.EventType{
		constants.EventTypeVM,   // valid
		"INVALID_TYPE_1",        // invalid
		constants.EventTypeHost, // valid
		"ANOTHER_INVALID",       // invalid
	})

	hub.registerClient(client)

	// Check client was still registered
	assert.Contains(t, hub.clients, "client1", "Client should be registered despite invalid types")

	// Check only valid subscriptions were added
	assert.True(t, hub.subscriptions[constants.EventTypeVM]["client1"], "Should be subscribed to VM")
	assert.True(t, hub.subscriptions[constants.EventTypeHost]["client1"], "Should be subscribed to HOST")
	assert.True(t, hub.subscriptions[constants.EventTypeGlobal]["client1"], "Should be subscribed to GLOBAL (auto)")

	// Check invalid types were NOT added to subscriptions
	_, invalidExists := hub.subscriptions["INVALID_TYPE_1"]
	assert.False(t, invalidExists, "Invalid type should not exist in subscriptions")
	_, anotherInvalidExists := hub.subscriptions["ANOTHER_INVALID"]
	assert.False(t, anotherInvalidExists, "Another invalid type should not exist in subscriptions")

	// Note: Error messages ARE sent to broadcast channel, but we don't test that here
	// since hub.run() would consume them. The important thing is invalid types are skipped.
}

func TestHub_RegisterClient_OnlyInvalidEventTypes(t *testing.T) {
	hub := createTestHub()

	// Client with only invalid event types
	client := createTestClient("client1", "testuser", []constants.EventType{
		"TOTALLY_INVALID",
		"ALSO_INVALID",
	})

	hub.registerClient(client)

	// Client should still be registered (only has global subscription)
	assert.Contains(t, hub.clients, "client1", "Client should be registered")

	// Should only have global subscription
	assert.True(t, hub.subscriptions[constants.EventTypeGlobal]["client1"], "Should have auto-subscribed GLOBAL")
	assert.Len(t, hub.subscriptions, 1, "Should only have GLOBAL subscription type")

	// Invalid types should not exist
	_, exists := hub.subscriptions["TOTALLY_INVALID"]
	assert.False(t, exists, "Invalid type should not exist")
}
