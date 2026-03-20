package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) RegisterHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hw, err := s.GetHostHardwareInfo(host)
	if err != nil {
		return nil, err
	}
	s.updateHostWithHardwareInfo(host, hw)

	host.Enabled = true
	dbHost, err := dbService.CreateOrchestratorHost(ctx, *host)
	if err != nil {
		return nil, err
	}

	s.Refresh()
	dbService.SaveNow(ctx)

	manager := GetHostWebSocketManager()
	if manager != nil {
		manager.ProbeAndConnect(*dbHost)
	}

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_ADDED", apimodels.HostAddedEvent{
			HostID:      dbHost.ID,
			Host:        dbHost.Host,
			Description: dbHost.Description,
		})
		if err := emitter.Broadcast(msg); err != nil {
			ctx.LogErrorf("[Orchestrator] Failed to broadcast HOST_ADDED for host %s: %v", dbHost.ID, err)
		}
	}

	return dbHost, nil
}
