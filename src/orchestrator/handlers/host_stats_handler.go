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
	registrar       interfaces.HostRegistrar
	resourceUpdater ResourceUpdater
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

func (h *HostStatsHandler) SetResourceUpdater(updater ResourceUpdater) {
	h.resourceUpdater = updater
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
	if event.Message == "DISK_SPACE_CHANGED" {
		h.updateHostResources(ctx, hostID)
	}

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

func (h *HostStatsHandler) updateHostResources(ctx basecontext.ApiContext, hostID string) error {
	if h.resourceUpdater == nil {
		ctx.LogWarnf("[HostStatsHandler] [orchestrator] No resource updater configured - skipping host resource update")
		return nil
	}
	if err := h.resourceUpdater.UpdateHostResourcesForEvent(ctx, hostID); err != nil {
		ctx.LogErrorf("[HostStatsHandler] [orchestrator] Error updating host resources for host %s: %v", hostID, err)
		return err
	}
	return nil
}
