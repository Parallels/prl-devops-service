package orchestrator

import (
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
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

	return host.Resources, nil
}
