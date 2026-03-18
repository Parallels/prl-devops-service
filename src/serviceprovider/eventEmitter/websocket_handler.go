package eventemitter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, ctx basecontext.ApiContext) *models.ApiErrorResponse {
	clientIP := extractClientIP(r)
	ctx.LogInfof("WebSocket connection attempt from IP: %s", clientIP)

	if !isMultipleConnectionsPerIPAllowed() && clientIP != "" {
		if Get().hub.HasActiveConnectionFromIP(clientIP) {
			ctx.LogWarnf("Connection rejected: IP %s already has an active connection", clientIP)
			return &models.ApiErrorResponse{
				Message: fmt.Sprintf("IP address %s already has an active WebSocket connection", clientIP),
				Code:    http.StatusConflict,
			}
		}
	}
	// Validate that the request has an authenticated user before upgrading the
	// connection.  Checking here (before Upgrade) lets us return a proper HTTP
	// 401 response instead of a WebSocket close frame, which some clients
	// misinterpret as a generic error.
	usr := ctx.GetUser()
	if usr == nil {
		// API key auth (IsMicroService=true) does not populate the user.
		// Allow the connection with a synthetic identity so microservices can
		// subscribe to events.
		if authCtx := ctx.GetAuthorizationContext(); authCtx != nil && authCtx.IsMicroService {
			keyName := authCtx.ApiKeyName
			if keyName == "" {
				keyName = "api-key-client"
			}
			usr = &models.ApiUser{
				Username: keyName,
				Name:     keyName,
			}
		} else {
			ctx.LogErrorf("WebSocket connection attempt without authenticated user")
			return &models.ApiErrorResponse{
				Message: "authentication required",
				Code:    http.StatusUnauthorized,
			}
		}
	}

	typesParam := r.URL.Query().Get("event_types")
	subscriptions, _ := stringToEventTypes(strings.Split(typesParam, ","))

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctx.LogErrorf("Failed to upgrade WebSocket connection: %v", err)
		return &models.ApiErrorResponse{
			Message: "Failed to upgrade connection: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
	}
	// Create client
	client := &Client{
		ctx:         ctx,
		ID:          uuid.NewString(),
		User:        usr,
		Conn:        conn,
		done:        make(chan struct{}),
		ConnectedAt: time.Now(),
		LastPingAt:  time.Now(),
		LastPongAt:  time.Now(),
		IsAlive:     true,
	}

	if !Get().hub.registerClient(client, subscriptions, clientIP) {
		ctx.LogWarnf("Failed to register client (shutdown or timeout)")
		// Connection is already hijacked; send a WebSocket close frame instead of
		// writing to the HTTP response writer (which would cause a 'hijacked
		// connection' runtime warning).
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "service is shutting down"))
		conn.Close()
		return nil
	}

	return nil
}

func HandleUnsubscribe(w http.ResponseWriter, r *http.Request, ctx basecontext.ApiContext) {
	var request models.UnsubscribeRequest
	if err := http_helper.MapRequestBody(r, &request); err != nil {
		ctx.LogWarnf("Invalid unsubscribe request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(models.ApiErrorResponse{
			Message: "Invalid request body: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	if len(request.EventTypes) == 0 {
		ctx.LogInfof("No event_types specified")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(models.ApiErrorResponse{
			Message: "Invalid event_types body parameter",
			Code:    http.StatusBadRequest,
		})
	}

	eventTypesList, err := stringToEventTypes(request.EventTypes)
	if len(eventTypesList) <= 0 {
		ctx.LogWarnf("No valid event types to unsubscribe: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(models.ApiErrorResponse{
			Message: "No valid event types to unsubscribe: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	username := ""
	if u := ctx.GetUser(); u != nil {
		username = u.Username
	} else if authCtx := ctx.GetAuthorizationContext(); authCtx != nil && authCtx.IsMicroService {
		username = authCtx.ApiKeyName
		if username == "" {
			username = "api-key-client"
		}
	}
	unsubscribed, err := Get().hub.unsubscribeClientFromTypes(request.ClientID, username, eventTypesList)

	if len(unsubscribed) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(models.ApiErrorResponse{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err != nil && len(unsubscribed) > 0 {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(models.ApiErrorResponse{
			Message: err.Error() + " unsubscribed from: " + strings.Join(unsubscribed, ", "),
			Code:    http.StatusOK,
		})
		return
	}
	ctx.LogInfof("Client %s unsubscribed from event types: %v", request.ClientID, unsubscribed)
}
