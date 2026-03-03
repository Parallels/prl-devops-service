package handlers

import (
	"encoding/json"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	global_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type HostReverseProxyEventHandler struct {
	registrar interfaces.HostRegistrar
}

var (
	rpInstance *HostReverseProxyEventHandler
	rpOnce     sync.Once
)

func NewHostReverseProxyEventHandler(registrar interfaces.HostRegistrar) *HostReverseProxyEventHandler {
	rpOnce.Do(func() {
		rpInstance = &HostReverseProxyEventHandler{
			registrar: registrar,
		}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypeReverseProxy}, rpInstance)
	})
	return rpInstance
}

func (h *HostReverseProxyEventHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	if eventType != constants.EventTypeReverseProxy {
		return
	}

	var event global_models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[HostReverseProxyEventHandler] Error unmarshalling event message: %v", err)
		return
	}

	ctx.LogDebugf("[HostReverseProxyEventHandler] Received reverse proxy event from host %s: %s", hostID, event.Message)

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := global_models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_REVERSE_PROXY_EVENT", map[string]interface{}{
			"host_id": hostID,
			"event":   event,
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				ctx.LogErrorf("[HostReverseProxyEventHandler] Failed to broadcast event HOST_REVERSE_PROXY_EVENT: %v", err)
			}
		}()
	}
}
