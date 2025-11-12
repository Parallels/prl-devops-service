package eventemitter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, ctx basecontext.ApiContext) *models.ApiErrorResponse {
	clientIP := extractClientIP(r)
	ctx.LogInfof("WebSocket connection attempt from IP: %s", clientIP)

	if !isMultipleConnectionsPerIPAllowed() && clientIP != "" {
		if globalEventEmitter.hub.HasActiveConnectionFromIP(clientIP) {
			ctx.LogWarnf("Connection rejected: IP %s already has an active connection", clientIP)
			return &models.ApiErrorResponse{
				Message: fmt.Sprintf("IP address %s already has an active WebSocket connection", clientIP),
				Code:    http.StatusConflict,
			}
		}
	}

	subscriptions := parseSubscriptions(r, ctx)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctx.LogErrorf("Failed to upgrade WebSocket connection: %v", err)
		return &models.ApiErrorResponse{
			Message: "Failed to upgrade connection: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	usr := ctx.GetUser()
	if usr == nil {
		// for testing purposes only
		usr = &models.ApiUser{
			ID:       "anonymous",
			Username: "anonymous",
		}
		ctx.LogWarnf("WebSocket connection established without authenticated user")
	}
	// Create client
	client := &Client{
		ID:            uuid.NewString(),
		User:          usr,
		Hub:           globalEventEmitter.hub,
		Conn:          conn,
		Send:          make(chan *models.EventMessage, 1024),
		Subscriptions: subscriptions,
		RemoteIP:      clientIP,
		ConnectedAt:   time.Now(),
		LastPingAt:    time.Now(),
		LastPongAt:    time.Now(),
		IsAlive:       true,
	}

	ctx.LogInfof("WebSocket connection established for user %s (ID: %s) with subscriptions: %v",
		client.User.Username, client.ID, subscriptions)

	// Register client to hub
	globalEventEmitter.hub.register <- client

	// Start client goroutines
	go client.clientWriter()
	go client.clientReader()

	return nil
}

func parseSubscriptions(r *http.Request, ctx basecontext.ApiContext) []constants.EventType {
	typesParam := r.URL.Query().Get("event_types")

	if typesParam == "" {
		ctx.LogInfof("No event_types specified")
		return []constants.EventType{}
	}

	// Parse comma-separated types
	types := strings.Split(typesParam, ",")
	subscriptions := make([]constants.EventType, 0, len(types))
	invalidTypes := make([]string, 0)

	for _, t := range types {
		eventType := constants.EventType(strings.ToLower(strings.TrimSpace(t)))
		if !eventType.IsValid() {
			invalidTypes = append(invalidTypes, strings.TrimSpace(t))
			continue
		}
		subscriptions = append(subscriptions, eventType)
	}

	if len(invalidTypes) > 0 {
		allTypes := make([]string, 0, len(constants.GetAllEventTypes()))
		for _, et := range constants.GetAllEventTypes() {
			allTypes = append(allTypes, et.String())
		}
		ctx.LogInfof("Unknown event_type(s) requested: %s valid types are: %s", strings.Join(invalidTypes, ", "),
			strings.Join(allTypes, ", "))
	}

	if len(subscriptions) == 0 && len(invalidTypes) > 0 {
		ctx.LogWarnf("No valid event_type(s) provided in request")
	}
	return subscriptions
}

func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs, first one is the client)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP if there are multiple
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fall back to RemoteAddr (includes port, so strip it)
	if r.RemoteAddr != "" {
		// RemoteAddr is in format "IP:port", extract just the IP
		if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
			return r.RemoteAddr[:idx]
		}
		return r.RemoteAddr
	}
	return ""
}

func HandleUnsubscribe(r *http.Request, ctx basecontext.ApiContext) *models.ApiErrorResponse {

	var request models.UnsubscribeRequest
	if err := http_helper.MapRequestBody(r, &request); err != nil {
		ctx.LogWarnf("Invalid unsubscribe request body: %v", err)
		return &models.ApiErrorResponse{
			Message: "Invalid request body: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	// Lock the hub to safely access the clients map
	globalEventEmitter.hub.mu.RLock()
	client, ok := globalEventEmitter.hub.clients[request.ClientID]
	globalEventEmitter.hub.mu.RUnlock()

	if !ok {
		ctx.LogWarnf("No active WebSocket client found with ID: %s", request.ClientID)
		return &models.ApiErrorResponse{
			Message: "No active WebSocket client found with the provided ID",
			Code:    http.StatusNotFound,
		}
	}

	// Lock client to safely read User field
	client.mu.RLock()
	userID := client.User.ID
	client.mu.RUnlock()

	currentUser := ctx.GetUser()

	if userID != currentUser.ID {
		ctx.LogWarnf("Client ID: %s does not belong to authenticated user %s", request.ClientID, currentUser.ID)
		return &models.ApiErrorResponse{
			Message: "The specified client ID does not belong to the authenticated user",
			Code:    http.StatusUnauthorized,
		}
	}

	unsubscribed, err := client.unsubscribeToEvents(request.EventTypes)
	if err != nil {
		if len(unsubscribed) > 0 {
			ctx.LogWarnf("Partially unsubscribed from %v, but error occurred: %v", unsubscribed, err)
			return &models.ApiErrorResponse{
				Message: fmt.Sprintf("Partially successful: unsubscribed from %v, but %s", unsubscribed, err.Error()),
				Code:    http.StatusPartialContent,
			}
		}
		return &models.ApiErrorResponse{
			Message: "Failed to unsubscribe from event types: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	return nil
}
