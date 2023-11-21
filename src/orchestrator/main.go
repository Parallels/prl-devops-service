package orchestrator

import (
	"context"
	"sync"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/mappers"
	apimodels "github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"
)

var globalOrchestratorService *OrchestratorService

type OrchestratorService struct {
	ctx             basecontext.ApiContext
	timeout         time.Duration
	refreshInterval time.Duration
	syncContext     context.Context
	cancel          context.CancelFunc
	db              *data.JsonDatabase
}

func NewOrchestratorService(ctx basecontext.ApiContext) *OrchestratorService {
	if globalOrchestratorService == nil {
		globalOrchestratorService = &OrchestratorService{
			ctx:     ctx,
			timeout: 2 * time.Minute,
		}
		cfg := config.NewConfig()
		globalOrchestratorService.refreshInterval = time.Duration(cfg.GetOrchestratorPullFrequency()) * time.Second

	} else {
		globalOrchestratorService.ctx = ctx
	}

	return globalOrchestratorService
}

func (s *OrchestratorService) Start(waitForInit bool) error {
	s.syncContext, s.cancel = context.WithCancel(context.Background())

	dbService, err := serviceprovider.GetDatabaseService(s.ctx)
	if err != nil {
		return err
	}

	s.db = dbService

	if waitForInit {
		s.ctx.LogInfo("[Orchestrator] Waiting for API to be initialized")
		<-restapi.Initialized
	}

	s.ctx.LogInfo("[Orchestrator] Starting Orchestrator Background Service")
	for {
		select {
		case <-s.syncContext.Done():
			return nil
		default:
			var wg sync.WaitGroup
			dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
			if err != nil {
				return err
			}

			for _, host := range dtoOrchestratorHosts {
				wg.Add(1)
				go s.processHostWaitingGroup(host, &wg)
			}
			wg.Wait()

			s.ctx.LogInfo("[Orchestrator] Sleeping for %s seconds", s.refreshInterval)
			time.Sleep(s.refreshInterval)
		}
	}
}

func (s *OrchestratorService) Stop() {
	s.ctx.LogInfo("[Orchestrator] Stopping Orchestrator Background Service")
	s.cancel()
	s.syncContext.Done()
}

func (s *OrchestratorService) Refresh() error {
	dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
	if err != nil {
		return err
	}

	for _, host := range dtoOrchestratorHosts {
		go s.processHost(host)
	}

	return nil
}

func (s *OrchestratorService) processHostWaitingGroup(host models.OrchestratorHost, wg *sync.WaitGroup) {
	defer wg.Done()

	select {
	case <-s.syncContext.Done():
		return
	default:
		s.processHost(host)
	}
}

func (s *OrchestratorService) processHost(host models.OrchestratorHost) {
	s.ctx.LogInfo("[Orchestrator] Processing host %s", host.Host)

	host.HealthCheck = &apimodels.ApiHealthCheck{}
	if healthCheck, err := s.GetHostSystemHealthCheck(&host); err != nil {
		s.ctx.LogError("[Orchestrator] Error getting health check for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		s.persistHost(&host)
		return
	} else {
		host.SetHealthy()
		host.HealthCheck = healthCheck
	}

	// Updating the host resources
	hardwareInfo, err := s.GetHostHardwareInfo(&host)
	if err != nil {
		s.ctx.LogError("[Orchestrator] Error getting hardware info for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		s.persistHost(&host)
		return
	}

	if host.Resources == nil {
		host.Resources = &models.HostResources{}
	}

	dtoResources := mappers.MapHostResourcesFromSystemUsageResponse(*hardwareInfo)
	host.Resources = &dtoResources

	// Updating the Virtual Machines
	vms, err := s.GetHostVirtualMachinesInfo(&host)
	if err != nil {
		s.ctx.LogError("[Orchestrator] Error getting virtual machines for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		s.persistHost(&host)
		return
	}

	host.VirtualMachines = make([]models.VirtualMachine, 0)
	for _, vm := range vms {
		dtoVm := mappers.MapDtoVirtualMachineFromApi(vm)
		dtoVm.HostId = host.ID
		dtoVm.Host = host.GetHost()
		host.VirtualMachines = append(host.VirtualMachines, dtoVm)
	}

	s.persistHost(&host)
}

func (s *OrchestratorService) persistHost(host *models.OrchestratorHost) error {
	// persist the host
	s.db.Connect(s.ctx)
	if _, err := s.db.UpdateOrchestratorHost(s.ctx, host); err != nil {
		s.ctx.LogError("[Orchestrator] Error saving host %s: %v", host.Host, err.Error())
		return err
	}

	return nil
}

func (s *OrchestratorService) GetResources() error {

	return nil
}
