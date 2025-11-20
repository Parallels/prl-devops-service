package health

import (
	"encoding/json"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
)

// Broadcaster defines the capability required by HealthService to send messages.
// This decouples the service from the concrete EventEmitter implementation.
type Broadcaster interface {
	BroadcastMessage(msg *models.EventMessage) error
}

// ServiceManager defines the interface required by HealthService
type ServiceManager interface {
	eventemitter.Broadcaster
	eventemitter.Registrar
}

// HealthService handles health-related events
type HealthService struct {
	broadcaster eventemitter.Broadcaster
}

var (
	instance *HealthService
	once     sync.Once
)

// NewHealthService returns the singleton instance of HealthService
func NewHealthService(manager ServiceManager) *HealthService {
	once.Do(func() {
		instance = &HealthService{
			broadcaster: manager,
		}
		manager.RegisterHandler([]constants.EventType{constants.EventTypeHealth}, instance)
	})
	return instance
}

// Handle processes incoming messages.
// Implements eventemitter.MessageHandler interface.
func (s *HealthService) Handle(ctx basecontext.ApiContext, clientID string, eventType constants.EventType, payload []byte, msgID string) {
	if eventType == constants.EventTypeHealth {
		// Parse payload to check for specific action/message
		var request struct {
			Message string `json:"message"` // e.g., "ping"
		}
		if err := json.Unmarshal(payload, &request); err != nil {
			ctx.LogWarnf("[HealthService] Failed to parse payload: %v", err)
			return
		}

		if request.Message == "ping" {
			// Create reply
			reply := models.NewEventMessage(constants.EventTypeHealth, "pong", nil)
			reply.ClientID = clientID
			reply.RefID = msgID // Correlate with the request

			// Broadcast using the interface
			_ = s.broadcaster.BroadcastMessage(reply)
		}
	}
}
