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

// ReverseProxyUpdater is implemented by OrchestratorService and called when a
// reverse_proxy event arrives so the DB is kept current without a full refresh.
type ReverseProxyUpdater interface {
	UpdateHostReverseProxyForEvent(ctx basecontext.ApiContext, hostID string) error
}

type HostReverseProxyEventHandler struct {
	registrar           interfaces.HostRegistrar
	reverseProxyUpdater ReverseProxyUpdater
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

// SetReverseProxyUpdater injects the updater dependency (mirrors PDfMEventHandler.SetResourceUpdater).
func (h *HostReverseProxyEventHandler) SetReverseProxyUpdater(updater ReverseProxyUpdater) {
	h.reverseProxyUpdater = updater
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

	// Refresh the DB with the latest reverse proxy state for this host.
	// Run asynchronously so the WebSocket read loop is not blocked.
	if h.reverseProxyUpdater != nil {
		go func() {
			if err := h.reverseProxyUpdater.UpdateHostReverseProxyForEvent(ctx, hostID); err != nil {
				ctx.LogErrorf("[HostReverseProxyEventHandler] Error updating reverse proxy config for host %s: %v", hostID, err)
			} else {
				ctx.LogInfof("[HostReverseProxyEventHandler] Updated reverse proxy config for host %s", hostID)
			}
		}()
	}

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := global_models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_REVERSE_PROXY_UPDATED", map[string]interface{}{
			"host_id": hostID,
			"event":   event,
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				ctx.LogErrorf("[HostReverseProxyEventHandler] Failed to broadcast event HOST_REVERSE_PROXY_UPDATED: %v", err)
			}
		}()
	}
}
