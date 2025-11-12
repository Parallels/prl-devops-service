package eventemitter

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	apiModels "github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestParseSubscriptions_AllTypes(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)

	subscriptions := parseSubscriptions(req, ctx)

	// When no event_types specified, should return empty slice
	assert.Equal(t, []constants.EventType{}, subscriptions)
}

func TestParseSubscriptions_SpecificTypes(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=vm,host", nil)

	subscriptions := parseSubscriptions(req, ctx)

	assert.Len(t, subscriptions, 2)
	assert.Contains(t, subscriptions, constants.EventTypeVM)
	assert.Contains(t, subscriptions, constants.EventTypeHost)
}

func TestParseSubscriptions_CaseInsensitive(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=Vm,HoSt,SyStEm", nil)

	subscriptions := parseSubscriptions(req, ctx)

	assert.Len(t, subscriptions, 3)
	assert.Contains(t, subscriptions, constants.EventTypeVM)
	assert.Contains(t, subscriptions, constants.EventTypeHost)
	assert.Contains(t, subscriptions, constants.EventTypeSystem)
}

func TestParseSubscriptions_CaseInsensitiveAllTypes(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	// Test all event types with various cases including PDfM which has mixed case in constant
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=GLOBAL,VM,HOST,SYSTEM,PDFM", nil)

	subscriptions := parseSubscriptions(req, ctx)

	// Should get 5 types, all in lowercase
	assert.Len(t, subscriptions, 5)
	assert.Contains(t, subscriptions, constants.EventTypeGlobal)
	assert.Contains(t, subscriptions, constants.EventTypeVM)
	assert.Contains(t, subscriptions, constants.EventTypeHost)
	assert.Contains(t, subscriptions, constants.EventTypeSystem)
	assert.Contains(t, subscriptions, constants.EventTypePDFM)
}

func TestParseSubscriptions_CaseInsensitiveDuplicates(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	// Test that duplicates in different cases are all accepted (not deduplicated in parseSubscriptions)
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=vm,VM,Vm", nil)

	subscriptions := parseSubscriptions(req, ctx)

	// parseSubscriptions doesn't deduplicate, so we get 3 entries
	assert.Len(t, subscriptions, 3)
	// All should be lowercase
	assert.Equal(t, constants.EventTypeVM, subscriptions[0])
	assert.Equal(t, constants.EventTypeVM, subscriptions[1])
	assert.Equal(t, constants.EventTypeVM, subscriptions[2])
}

func TestParseSubscriptions_WithWhitespace(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=%20%20vm%20%20,%20host%20%20,%20%20system%20%20", nil)

	subscriptions := parseSubscriptions(req, ctx)

	assert.Len(t, subscriptions, 3)
	assert.Contains(t, subscriptions, constants.EventTypeVM)
	assert.Contains(t, subscriptions, constants.EventTypeHost)
	assert.Contains(t, subscriptions, constants.EventTypeSystem)
}

func TestParseSubscriptions_InvalidType(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=vm,INVALID_TYPE", nil)

	subscriptions := parseSubscriptions(req, ctx)

	// When there's a mix of valid and invalid types, valid ones are kept
	assert.Len(t, subscriptions, 1)
	assert.Contains(t, subscriptions, constants.EventTypeVM)
}

func TestParseSubscriptions_OnlyInvalidTypes(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=INVALID1,INVALID2", nil)

	subscriptions := parseSubscriptions(req, ctx)
	assert.Len(t, subscriptions, 0)
}

func TestParseSubscriptions_EmptyString(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	req := httptest.NewRequest("GET", "/ws/subscribe?event_types=", nil)

	subscriptions := parseSubscriptions(req, ctx)
	assert.Len(t, subscriptions, 0)
}

func TestClient_HandleClientMessage_Ping(t *testing.T) {
	hub := createTestHub()
	client := &Client{
		ID:   "test-client",
		Hub:  hub,
		Send: make(chan *apiModels.EventMessage, 10),
	}

	msg := map[string]interface{}{
		"type": "invalid_type",
	}

	client.handleClientMessage(msg)

	select {
	case <-client.Send:
		t.Fatal("Should not send message for unsupported message type")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message sent
	}
}

func TestClient_HandleClientMessage_UnknownType(t *testing.T) {
	hub := createTestHub()
	client := &Client{
		ID:   "test-client",
		Hub:  hub,
		Send: make(chan *apiModels.EventMessage, 10),
	}

	msg := map[string]interface{}{
		"type": "unknown_type",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		client.handleClientMessage(msg)
	})
}

func TestClient_UnsubscribeToEvents(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeGlobal, constants.EventTypeVM, constants.EventTypeHost})
	client.Hub = hub

	hub.registerClient(client)

	types := []string{constants.EventTypeVM.String()}
	client.unsubscribeToEvents(types)

	// Check subscription was removed
	assert.NotContains(t, client.Subscriptions, constants.EventTypeVM)
	assert.Contains(t, client.Subscriptions, constants.EventTypeHost)
	assert.Contains(t, client.Subscriptions, constants.EventTypeGlobal)

	// Check hub subscriptions
	if subs, exists := hub.subscriptions[constants.EventTypeVM]; exists {
		assert.False(t, subs["test-client"])
	}
}

func TestClient_UnsubscribeToEvents_InvalidType(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeGlobal})
	client.Hub = hub

	hub.registerClient(client)
	initialSubCount := len(client.Subscriptions)

	types := []string{"INVALID_TYPE"}
	client.unsubscribeToEvents(types)

	// Invalid type should be ignored
	assert.Len(t, client.Subscriptions, initialSubCount)
}

func TestClient_UnsubscribeToEvents_CaseInsensitive(t *testing.T) {
	hub := createTestHub()
	client := createTestClient("test-client", "testuser", []constants.EventType{constants.EventTypeGlobal, constants.EventTypeVM, constants.EventTypeHost})
	client.Hub = hub

	hub.registerClient(client)

	// Unsubscribe with mixed case
	types := []string{"VM", "HoSt"}
	client.unsubscribeToEvents(types)

	// Check subscriptions were removed (comparison should be case-insensitive)
	assert.NotContains(t, client.Subscriptions, "vm")
	assert.NotContains(t, client.Subscriptions, "host")
	assert.Contains(t, client.Subscriptions, constants.EventTypeGlobal)
}

func TestExtractClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1, 172.16.0.1")

	ip := extractClientIP(req)

	// Should return the first IP in the X-Forwarded-For header
	assert.Equal(t, "192.168.1.100", ip)
}

func TestExtractClientIP_XForwardedForSingle(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.42")

	ip := extractClientIP(req)

	assert.Equal(t, "203.0.113.42", ip)
}

func TestExtractClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.Header.Set("X-Real-IP", "198.51.100.50")

	ip := extractClientIP(req)

	assert.Equal(t, "198.51.100.50", ip)
}

func TestExtractClientIP_XForwardedForTakesPrecedence(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	req.Header.Set("X-Real-IP", "198.51.100.50")

	ip := extractClientIP(req)

	// X-Forwarded-For should take precedence
	assert.Equal(t, "192.168.1.100", ip)
}

func TestExtractClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.RemoteAddr = "203.0.113.195:54321"

	ip := extractClientIP(req)

	// Should strip the port and return just the IP
	assert.Equal(t, "203.0.113.195", ip)
}

func TestExtractClientIP_RemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.RemoteAddr = "203.0.113.195"

	ip := extractClientIP(req)

	assert.Equal(t, "203.0.113.195", ip)
}

func TestExtractClientIP_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.RemoteAddr = ""

	ip := extractClientIP(req)

	assert.Equal(t, "", ip)
}

func TestExtractClientIP_WithWhitespace(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/subscribe", nil)
	req.Header.Set("X-Forwarded-For", "  192.168.1.100  , 10.0.0.1")

	ip := extractClientIP(req)

	// Should trim whitespace
	assert.Equal(t, "192.168.1.100", ip)
}
