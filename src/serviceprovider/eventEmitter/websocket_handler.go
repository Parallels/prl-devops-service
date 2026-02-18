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

	usr := ctx.GetUser()
	if usr == nil {
		ctx.LogErrorf("WebSocket connection without authenticated user")
		return &models.ApiErrorResponse{
			Message: "Authentication required",
			Code:    http.StatusUnauthorized,
		}
	}
	// Create client
	client := &Client{
		ctx:         ctx,
		ID:          uuid.NewString(),
		User:        usr,
		Conn:        conn,
		Send:        make(chan *models.EventMessage, 1024),
		ConnectedAt: time.Now(),
		LastPingAt:  time.Now(),
		LastPongAt:  time.Now(),
		IsAlive:     true,
	}

	if !Get().hub.registerClient(client, subscriptions, clientIP) {
		ctx.LogWarnf("Failed to register client (shutdown or timeout)")
		conn.Close()
		return &models.ApiErrorResponse{
			Message: "Service is shutting down",
			Code:    http.StatusServiceUnavailable,
		}
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

	unsubscribed, err := Get().hub.unsubscribeClientFromTypes(request.ClientID, ctx.GetUser().Username, eventTypesList)

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
