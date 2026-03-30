package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
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

	// Synchronously populate VMs when the host module is enabled.
	// This ensures VMs are in the DB before RegisterHost returns, avoiding
	// the window where the host appears healthy but has no VMs while the
	// async Refresh goroutine hasn't completed yet.
	// Per-VM atomic upserts are used so concurrent PDFM WebSocket events
	// (VM_ADDED, VM_UPDATED) can safely overlap without data loss.
	if hasModule(hw.EnabledModules, "host") {
		vms, err := s.GetHostVirtualMachinesInfo(dbHost)
		if err != nil {
			ctx.LogWarnf("[Orchestrator] RegisterHost: could not fetch VMs for host %s (will retry on next refresh): %v", dbHost.Host, err)
		} else {
			synced := 0
			for _, vm := range vms {
				dtoVm := mappers.MapDtoVirtualMachineFromApi(vm)
				dtoVm.HostId = dbHost.ID
				dtoVm.HostName = getHostName(*dbHost)
				dtoVm.Host = dbHost.GetHost()
				dtoVm.HostUrl = dbHost.GetHostUrl()
				if err := dbService.UpsertOrchestratorHostVM(ctx, dbHost.ID, dtoVm); err != nil {
					ctx.LogWarnf("[Orchestrator] RegisterHost: failed to upsert VM %s on host %s: %v", vm.ID, dbHost.Host, err)
					continue
				}
				synced++
			}
			ctx.LogInfof("[Orchestrator] RegisterHost: synced %d VMs for host %s", synced, dbHost.Host)
		}
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
