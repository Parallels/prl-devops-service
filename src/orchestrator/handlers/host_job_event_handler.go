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

// HostJobEventHandler listens for job_manager events arriving from a connected
// orchestrator host over WebSocket and forwards them to the local UI event emitter.
// This allows the UI's existing job_manager subscription to receive job progress
// updates originating from operations running on a remote host.
type HostJobEventHandler struct {
	registrar interfaces.HostRegistrar
}

var (
	hostJobInstance *HostJobEventHandler
	hostJobOnce     sync.Once
)

func NewHostJobEventHandler(registrar interfaces.HostRegistrar) *HostJobEventHandler {
	hostJobOnce.Do(func() {
		hostJobInstance = &HostJobEventHandler{registrar: registrar}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypeJobManager}, hostJobInstance)
	})
	return hostJobInstance
}

func (h *HostJobEventHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	if eventType != constants.EventTypeJobManager {
		return
	}

	var event models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[HostJobEventHandler] Error unmarshalling job event from host %s: %v", hostID, err)
		return
	}

	emitter := serviceprovider.GetEventEmitter()
	if emitter == nil || !emitter.IsRunning() {
		return
	}

	// Forward the host job event to the local UI, tagging it with the originating host ID.
	// The UI's job_manager subscription will receive this alongside local job events.
	msg := models.NewEventMessage(constants.EventTypeJobManager, event.Message, models.HostJobEvent{
		HostID: hostID,
		Event:  event.Body,
	})
	go func() {
		if err := emitter.Broadcast(msg); err != nil {
			ctx.LogErrorf("[HostJobEventHandler] Failed to broadcast job event from host %s: %v", hostID, err)
		}
	}()
}
