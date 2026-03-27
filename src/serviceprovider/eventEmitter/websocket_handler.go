package eventemitter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
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

func HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, ctx basecontext.ApiContext) *models.ApiErrorDiagnosticsResponse {
	clientIP := extractClientIP(r)
	ctx.LogInfof("WebSocket connection attempt from IP: %s", clientIP)
	wsHandleDiag := errors.NewDiagnostics("HandleWebSocketConnection")
	if !isMultipleConnectionsPerIPAllowed() && clientIP != "" {
		if Get().hub.HasActiveConnectionFromIP(clientIP) {
			ctx.LogWarnf("Connection rejected: IP %s already has an active connection", clientIP)
			wsHandleDiag.AddError(strconv.Itoa(http.StatusConflict), fmt.Sprintf("IP address %s already has an active WebSocket connection", clientIP), "HandleWebSocketConnection")
			resp := models.NewDiagnosticsWithCode(wsHandleDiag, http.StatusConflict)
			return &resp
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
			wsHandleDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "authentication required", "")
			resp := models.NewDiagnosticsWithCode(wsHandleDiag, http.StatusUnauthorized)
			return &resp
		}
	}

	typesParam := r.URL.Query().Get("event_types")
	subscriptions, _ := stringToEventTypes(strings.Split(typesParam, ","))

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctx.LogErrorf("Failed to upgrade WebSocket connection: %v", err)
		wsHandleDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Failed to upgrade connection: "+err.Error(), "")
		resp := models.NewDiagnosticsWithCode(wsHandleDiag, http.StatusBadRequest)
		return &resp
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

// HandleGetClients writes the list of connected clients (with queue depths) as JSON.
// returnApiErrorWithDiagnostics is a package-local helper that writes the
// HTTP status code and JSON-encodes the diagnostics response. It mirrors
// controllers.ReturnApiErrorWithDiagnostics without creating a cross-package
// import cycle.
func returnApiErrorWithDiagnostics(ctx basecontext.ApiContext, w http.ResponseWriter, err models.ApiErrorDiagnosticsResponse) {
	ctx.LogErrorf("Error: %v", err.Message)
	w.WriteHeader(err.Code)
	_ = json.NewEncoder(w).Encode(err)
}

func HandleGetClients(w http.ResponseWriter, r *http.Request, ctx basecontext.ApiContext) {
	clients := Get().GetClients()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(clients)
}

// HandleGetStats writes aggregate event emitter statistics as JSON.
func HandleGetStats(w http.ResponseWriter, r *http.Request, ctx basecontext.ApiContext) {
	stats := Get().GetStats()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stats)
}

func HandleUnsubscribe(w http.ResponseWriter, r *http.Request, ctx basecontext.ApiContext) {
	var request models.UnsubscribeRequest
	unsubscribeHandleDiag := errors.NewDiagnostics("HandleUnsubscribe")
	if err := http_helper.MapRequestBody(r, &request); err != nil {
		ctx.LogWarnf("Invalid unsubscribe request body: %v", err)
		unsubscribeHandleDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
		returnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(unsubscribeHandleDiag, http.StatusBadRequest))
		return
	}

	if len(request.EventTypes) == 0 {
		ctx.LogInfof("No event_types specified")
		unsubscribeHandleDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid event_types body parameter", "")
		returnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(unsubscribeHandleDiag, http.StatusBadRequest))
		return
	}

	eventTypesList, err := stringToEventTypes(request.EventTypes)
	if len(eventTypesList) <= 0 {
		ctx.LogWarnf("No valid event types to unsubscribe: %v", err)
		unsubscribeHandleDiag.AddError(strconv.Itoa(http.StatusBadRequest), "No valid event types to unsubscribe: "+err.Error(), "")
		returnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(unsubscribeHandleDiag, http.StatusBadRequest))
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
		unsubscribeHandleDiag.AddError(strconv.Itoa(http.StatusBadRequest), err.Error(), "")
		returnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(unsubscribeHandleDiag, http.StatusBadRequest))
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
