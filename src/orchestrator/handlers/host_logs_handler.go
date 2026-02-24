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

type HostLogsHandler struct {
	registrar interfaces.HostRegistrar
}

var (
	logsInstance *HostLogsHandler
	logsOnce     sync.Once
)

func NewHostLogsHandler(registrar interfaces.HostRegistrar) *HostLogsHandler {
	logsOnce.Do(func() {
		logsInstance = &HostLogsHandler{
			registrar: registrar,
		}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypeSystemLogs}, logsInstance)
	})
	return logsInstance
}

func (h *HostLogsHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	if eventType != constants.EventTypeSystemLogs {
		return
	}

	var event models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[HostLogsHandler] Error unmarshalling event message: %v", err)
		return
	}

	ctx.LogDebugf("[HostLogsHandler] Received logs from host %s", hostID)

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_LOGS_UPDATE", models.HostLogsUpdate{
			HostID: hostID,
			Log:    event.Body,
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				ctx.LogErrorf("[HostLogsHandler] Failed to broadcast event HOST_LOGS_UPDATE: %v", err)
			}
		}()
	}
}
