package handlers

import (
	"encoding/json"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type HostHealthHandler struct {
	registrar interfaces.HostRegistrar
}

var (
	healthInstance *HostHealthHandler
	healthOnce     sync.Once
)

func NewHostHealthHandler(registrar interfaces.HostRegistrar) *HostHealthHandler {
	healthOnce.Do(func() {
		healthInstance = &HostHealthHandler{
			registrar: registrar,
		}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypeHealth}, healthInstance)
	})
	return healthInstance
}

func (h *HostHealthHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	if eventType != constants.EventTypeHealth {
		return
	}

	var event models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[HostHealthHandler] Error unmarshalling event message: %v", err)
		return
	}

	if event.Message == "pong" {
		h.handlePong(ctx, hostID)
	}
}

func (h *HostHealthHandler) handlePong(ctx basecontext.ApiContext, hostID string) {
	ctx.LogDebugf("[HostHealthHandler] Received pong from host %s", hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[HostHealthHandler] Error getting database service: %v", err)
		return
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogErrorf("[HostHealthHandler] Error getting host %s from DB: %v", hostID, err)
		return
	}

	if host == nil {
		ctx.LogWarnf("[HostHealthHandler] Host %s not found in DB", hostID)
		return
	}

	host.UpdatedAt = helpers.GetUtcCurrentDateTime()
	stateChanged := false

	newState := ""
	if host.State != "healthy" {
		newState = "healthy"
		stateChanged = true
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_HEALTH_UPDATE", models.HostHealthUpdate{
				HostID: host.ID,
				State:  newState,
			})
			go func() {
				if err := emitter.Broadcast(msg); err != nil {
					ctx.LogErrorf("[HostHealthHandler] Failed to broadcast event %s: %v", "HOST_HEALTH_UPDATE", err)
				}
			}()
		}
	}

	// Use the lightweight timestamp update method instead of full UpdateOrchestratorHost
	if err := dbService.UpdateOrchestratorHostTimestamp(ctx, hostID, host.UpdatedAt, newState); err != nil {
		ctx.LogErrorf("[HostHealthHandler] Error updating host %s health timestamp in DB: %v", hostID, err)
	} else if stateChanged {
		ctx.LogDebugf("[HostHealthHandler] Host %s marked as healthy (pong received)", hostID)
	} else {
		ctx.LogDebugf("[HostHealthHandler] Host %s health timestamp updated (pong received)", hostID)
	}
}
