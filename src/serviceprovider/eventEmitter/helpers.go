package eventemitter

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
)

// SendToType sends a message to all clients subscribed to a specific type
func (e *EventEmitter) SendToType(eventType constants.EventType, message string, body interface{}) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return errors.New("event emitter is not running")
	}

	msg := models.NewEventMessage(eventType, message, body)

	e.hub.broadcastMessage(msg)
	return nil
}

// SendToClient sends a message to a specific client
func (e *EventEmitter) SendToClient(clientID string, eventType constants.EventType, message string, body interface{}) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return errors.New("event emitter is not running")
	}

	msg := models.NewEventMessage(eventType, message, body)
	msg.ClientID = clientID

	return e.hub.broadcastMessage(msg)
}

// SendToAll sends a message to all connected clients
func (e *EventEmitter) SendToAll(message string, body interface{}) error {
	return e.SendToType(constants.EventTypeGlobal, message, body)
}

// BroadcastMessage sends a pre-constructed event message
func (e *EventEmitter) BroadcastMessage(msg *models.EventMessage) error {
	if !e.IsRunning() {
		e.ctx.LogWarnf("[EventEmitter] Cannot send message, service is not running")
		return nil
	}
	return e.hub.broadcastMessage(msg)
}

func stringToEventTypes(eventTypesString []string) ([]constants.EventType, error) {

	if len(eventTypesString) == 0 {
		return []constants.EventType{}, fmt.Errorf("no event types provided")
	}

	subscriptions := make([]constants.EventType, 0, len(eventTypesString))
	invalidTypes := make([]string, 0)

	for _, t := range eventTypesString {
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
		return subscriptions, fmt.Errorf("invalid event type(s): %s. Valid types are: %s", strings.Join(invalidTypes, ", "), strings.Join(allTypes, ", "))
	}

	if len(subscriptions) == 0 && len(invalidTypes) > 0 {
		return subscriptions, fmt.Errorf("no valid event types provided")
	}
	return subscriptions, nil
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
