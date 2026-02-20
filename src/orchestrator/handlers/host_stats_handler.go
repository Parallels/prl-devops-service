package handlers

import (
	"encoding/json"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type HostStatsHandler struct {
	registrar interfaces.HostRegistrar
}

var (
	statsInstance *HostStatsHandler
	statsOnce     sync.Once
)

func NewHostStatsHandler(registrar interfaces.HostRegistrar) *HostStatsHandler {
	statsOnce.Do(func() {
		statsInstance = &HostStatsHandler{
			registrar: registrar,
		}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypeStats}, statsInstance)
	})
	return statsInstance
}

func (h *HostStatsHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	if eventType != constants.EventTypeStats {
		return
	}

	var event models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[HostStatsHandler] Error unmarshalling event message: %v", err)
		return
	}

	ctx.LogDebugf("[HostStatsHandler] Received stats from host %s", hostID)

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_STATS_UPDATE", models.HostStatsUpdate{
			HostID: hostID,
			Stats:  event.Body,
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				ctx.LogErrorf("[HostStatsHandler] Failed to broadcast event HOST_STATS_UPDATE: %v", err)
			}
		}()
	}
}
