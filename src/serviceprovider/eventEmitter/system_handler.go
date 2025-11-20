package eventemitter

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
)

// Broadcaster defines the capability to send messages
type Broadcaster interface {
	BroadcastMessage(msg *models.EventMessage) error
}

// SystemManager defines the interface required by SystemHandler
type SystemManager interface {
	Broadcaster
	Registrar
}

// SystemHandler handles system-level messages like client-id requests
type SystemHandler struct {
	broadcaster Broadcaster
}

var (
	sysInstance *SystemHandler
	sysOnce     sync.Once
)

func NewSystemHandler(manager SystemManager) *SystemHandler {
	sysOnce.Do(func() {
		sysInstance = &SystemHandler{broadcaster: manager}
		manager.RegisterHandler([]constants.EventType{constants.EventTypeSystem}, sysInstance)
	})
	return sysInstance
}

// Handle processes system messages
func (h *SystemHandler) Handle(ctx basecontext.ApiContext, clientID string, eventType constants.EventType, payload []byte, msgID string) {
	if eventType == constants.EventTypeSystem {
		// Parse payload to check for specific action/message
		var request struct {
			Message string `json:"message"` // e.g., "client-id"
		}
		if err := json.Unmarshal(payload, &request); err != nil {
			msg := models.NewEventMessage(constants.EventTypeSystem, clientID, nil)
			msg.ClientID = clientID
			msg.RefID = msgID
			msg.Message = "error"
			msg.Body = map[string]interface{}{
				"error": err.Error(),
			}
			h.broadcaster.BroadcastMessage(msg)
			return
		}

		if strings.ToLower(request.Message) == "client-id" {
			cidMsg := models.NewEventMessage(constants.EventTypeSystem, clientID, nil)
			cidMsg.ClientID = clientID
			cidMsg.RefID = msgID
			cidMsg.Message = "client-id"
			cidMsg.Body = map[string]interface{}{
				"client-id": clientID,
			}
			h.broadcaster.BroadcastMessage(cidMsg)
		}
	}
}
