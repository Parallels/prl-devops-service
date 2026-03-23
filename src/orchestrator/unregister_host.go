package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) UnregisterHost(ctx basecontext.ApiContext, hostId string) error {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return err
	}

	existingHost, _ := dbService.GetOrchestratorHost(ctx, hostId)

	err = dbService.DeleteOrchestratorHost(ctx, hostId)
	if err != nil {
		return err
	}

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		event := apimodels.HostRemovedEvent{HostID: hostId}
		if existingHost != nil {
			event.Host = existingHost.Host
		}
		msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_REMOVED", event)
		if err := emitter.Broadcast(msg); err != nil {
			ctx.LogErrorf("[Orchestrator] Failed to broadcast HOST_REMOVED for host %s: %v", hostId, err)
		}
	}

	s.Refresh()
	return nil
}
