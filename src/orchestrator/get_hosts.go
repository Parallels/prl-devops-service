package orchestrator

import (
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/placement"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHosts(ctx basecontext.ApiContext, filter string) ([]*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}
	dtoOrchestratorHosts, err := dbService.GetOrchestratorHosts(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]*models.OrchestratorHost, 0)

	if len(dtoOrchestratorHosts) == 0 {
		return result, nil
	}

	var wg sync.WaitGroup
	mutex := sync.Mutex{}
	wg.Add(len(dtoOrchestratorHosts))
	for _, host := range dtoOrchestratorHosts {
		starTime := time.Now()
		go func(host models.OrchestratorHost) {
			defer wg.Done()
			ctx.LogDebugf("[Orchestrator] Processing Host: %v", host.Host)
			if host.Enabled {
				host.State = s.GetHostHealthCheckState(&host)
				ctx.LogDebugf("[Orchestrator] Host %v state: %v", host.Host, host.State)
			}

			mutex.Lock()
			result = append(result, &host)
			mutex.Unlock()
			ctx.LogDebugf("[Orchestrator] Processing Host: %v - Time: %v", host.Host, time.Since(starTime))
		}(host)
	}

	wg.Wait()
	return result, nil
}

func (s *OrchestratorService) GetHost(ctx basecontext.ApiContext, idOrName string) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}
	host, err := dbService.GetOrchestratorHost(ctx, idOrName)
	if err != nil {
		return nil, err
	}

	if host.Enabled {
		host.State = s.GetHostHealthCheckState(host)
	}

	return host, nil
}

func (s *OrchestratorService) GetDatabaseHost(ctx basecontext.ApiContext, idOrName string) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}
	host, err := dbService.GetOrchestratorHost(ctx, idOrName)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func (s *OrchestratorService) GetHostResources(ctx basecontext.ApiContext, idOrName string) (*models.HostResources, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}
	host, err := dbService.GetOrchestratorHost(ctx, idOrName)
	if err != nil {
		return nil, err
	}
	if host.Resources == nil {
		return nil, nil
	}

	// The host publishes what it is holding for itself; the ledger holds what
	// placement promised and the host has not applied yet. This endpoint reports
	// the figure placement will actually subtract — the larger of the two per
	// dimension, the same combine the scheduling decision makes rather than a
	// sum of two overlapping views — so what an operator reads here is what a
	// create is measured against, and a reservation is visible before the VM it
	// pays for exists anywhere.
	//
	// A copy, deliberately: the cached host must keep reporting the host's own
	// numbers, or the next scheduling decision would read its own ledger back to
	// itself as if the host had said it.
	resources := *host.Resources
	service := placement.Get()
	if service.Enabled() {
		pending := placement.CombinePending(service.PendingForHost(ctx, host.ID), hostReportedPending(&resources))
		resources.PendingCpu = pending.CPU
		resources.PendingMemory = pending.Memory
	}

	return &resources, nil
}
