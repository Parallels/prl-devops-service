package orchestrator

import (
	"context"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/telemetry"
)

var globalOrchestratorService *OrchestratorService

type OrchestratorService struct {
	ctx                basecontext.ApiContext
	timeout            time.Duration
	healthCheckTimeout time.Duration
	refreshInterval    time.Duration
	syncContext        context.Context
	cancel             context.CancelFunc
	db                 *data.JsonDatabase
}

func NewOrchestratorService(ctx basecontext.ApiContext) *OrchestratorService {
	if globalOrchestratorService == nil {
		globalOrchestratorService = &OrchestratorService{
			ctx:                ctx,
			timeout:            5 * time.Minute,
			healthCheckTimeout: 3 * time.Second,
		}
		cfg := config.Get()
		globalOrchestratorService.refreshInterval = time.Duration(cfg.OrchestratorPullFrequency()) * time.Second
	} else {
		globalOrchestratorService.ctx = ctx
	}
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil
	}

	globalOrchestratorService.db = dbService

	return globalOrchestratorService
}

func (s *OrchestratorService) Start(waitForInit bool) {
	ts := telemetry.Get()
	ts.TrackEvent(telemetry.NewTelemetryItem(s.ctx, telemetry.EventStartOrchestrator, nil, nil))
	s.syncContext, s.cancel = context.WithCancel(context.Background())

	dbService, err := serviceprovider.GetDatabaseService(s.ctx)
	if err != nil {
		return
	}

	s.db = dbService

	if waitForInit {
		s.ctx.LogInfof("[Orchestrator] Waiting for API to be initialized")
		<-restapi.Initialized
	}

	s.ctx.LogInfof("[Orchestrator] Starting Orchestrator Background Service")
	for {
		select {
		case <-s.syncContext.Done():
			return
		default:
			var wg sync.WaitGroup
			dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
			if err != nil {
				return
			}

			for _, host := range dtoOrchestratorHosts {
				wg.Add(1)
				go s.processHostWaitingGroup(host, &wg)
			}
			wg.Wait()

			if len(dtoOrchestratorHosts) > 0 {
				s.ctx.LogInfof("[Orchestrator] processed %v hosts", len(dtoOrchestratorHosts))
				s.ctx.LogInfof("[Orchestrator] Sleeping for %s seconds", s.refreshInterval)
			}

			time.Sleep(s.refreshInterval)
		}
	}
}

func (s *OrchestratorService) Stop() {
	s.ctx.LogInfof("[Orchestrator] Stopping Orchestrator Background Service")
	sp := serviceprovider.Get()
	if sp != nil {
		db := sp.JsonDatabase
		if db != nil {
			ctx := basecontext.NewRootBaseContext()
			ctx.LogInfof("[Orchestrator] Saving database")
			if err := db.SaveNow(ctx); err != nil {
				ctx.LogErrorf("[Core] Error saving database: %v", err)
			} else {
				ctx.LogInfof("[Orchestrator] Database saved")
			}
		}
	}
	s.cancel()
	s.syncContext.Done()

	s.ctx.LogInfof("[Orchestrator] Orchestrator Background Service Stopped")
}

func (s *OrchestratorService) Refresh() {
	dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
	if err != nil {
		return
	}

	for _, host := range dtoOrchestratorHosts {
		go s.processHost(host)
	}
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
	s.ctx.LogInfof("[Orchestrator] Processing host %s", host.Host)

	host.HealthCheck = &apimodels.ApiHealthCheck{}
	if healthCheck, err := s.GetHostSystemHealthCheck(&host); err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error getting health check for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		_ = s.persistHost(&host)
		return
	} else {
		s.ctx.LogInfof("[Orchestrator] host %s is alive and well: %s", host.Host, healthCheck.Message)
		host.SetHealthy()
		host.HealthCheck = healthCheck
	}

	s.ctx.LogInfof("[Orchestrator] Getting hardware info for host %s", host.Host)
	// Updating the host resources
	hardwareInfo, err := s.GetHostHardwareInfo(&host)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error getting hardware info for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		_ = s.persistHost(&host)
		return
	}

	if host.Resources == nil {
		host.Resources = &models.HostResources{}
	}

	dtoResources := mappers.MapHostResourcesFromSystemUsageResponse(*hardwareInfo)
	host.DevOpsVersion = hardwareInfo.DevOpsVersion
	host.OsName = hardwareInfo.OsName
	host.OsVersion = hardwareInfo.OsVersion
	host.ExternalIpAddress = hardwareInfo.ExternalIpAddress
	host.Resources = &dtoResources
	host.Architecture = hardwareInfo.CpuType
	host.CpuModel = hardwareInfo.CpuBrand
	host.ParallelsDesktopVersion = hardwareInfo.ParallelsDesktopVersion
	host.ParallelsDesktopLicensed = hardwareInfo.ParallelsDesktopLicensed
	host.IsReverseProxyEnabled = hardwareInfo.IsReverseProxyEnabled
	if hardwareInfo.ReverseProxy != nil {
		host.ReverseProxy = &models.ReverseProxy{
			Host: hardwareInfo.ReverseProxy.Host,
			Port: hardwareInfo.ReverseProxy.Port,
		}
		host.ReverseProxyHosts = make([]*models.ReverseProxyHost, 0)
		for _, rpHost := range hardwareInfo.ReverseProxy.Hosts {
			dtoHost := mappers.ApiReverseProxyHostToDto(rpHost)
			host.ReverseProxyHosts = append(host.ReverseProxyHosts, &dtoHost)
		}
	}

	// Updating the Virtual Machines
	vms, err := s.GetHostVirtualMachinesInfo(&host)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error getting virtual machines for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		_ = s.persistHost(&host)
		return
	}

	host.VirtualMachines = make([]models.VirtualMachine, 0)
	for _, vm := range vms {
		dtoVm := mappers.MapDtoVirtualMachineFromApi(vm)
		dtoVm.HostId = host.ID
		dtoVm.Host = host.GetHost()
		host.VirtualMachines = append(host.VirtualMachines, dtoVm)
	}

	totalAppleVms := 0
	for _, vm := range host.VirtualMachines {
		if vm.Type == "APPLE_VZ_VM" {
			totalAppleVms++
		}
	}

	host.Resources.TotalAppleVms = int64(totalAppleVms)
	host.UpdatedAt = helpers.GetUtcCurrentDateTime()

	s.ctx.LogInfof("[Orchestrator] Host %s has %v CPU Cores and %v Mb of RAM, contains %v VMs of which %v are MacVMs", host.Host, host.Resources.Total.LogicalCpuCount, host.Resources.Total.MemorySize, len(host.VirtualMachines), host.Resources.TotalAppleVms)

	_ = s.persistHost(&host)

	// Free up memory
	host.HealthCheck = nil
	host.Resources = nil
	host.VirtualMachines = nil
	host.ReverseProxy = nil
	host.ReverseProxyHosts = nil
}

func (s *OrchestratorService) persistHost(host *models.OrchestratorHost) error {
	// persist the host
	_ = s.db.Connect(s.ctx)
	// trying to fix the concurrency issues
	hostToSave := *host
	oldHost, err := s.db.GetOrchestratorHost(s.ctx, host.ID)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error getting host %s: %v", host.Host, err.Error())
		return err
	}
	s.ctx.LogDebugf("[Orchestrator] oldHost: %v, updated at %s and new one %s updated at %v", oldHost.ID, oldHost.UpdatedAt, host.ID, host.UpdatedAt)
	if oldHost.UpdatedAt > host.UpdatedAt {
		s.ctx.LogDebugf("[Orchestrator] Host %s was updated by another process, skipping", host.Host)
	} else {
		s.ctx.LogDebugf("[Orchestrator] Saving host %s", host.Host)
		if _, err := s.db.UpdateOrchestratorHost(s.ctx, &hostToSave); err != nil {
			s.ctx.LogErrorf("[Orchestrator] Error saving host %s: %v", host.Host, err.Error())
			return err
		}
	}

	s.ctx.LogDebugf("[Orchestrator] Host %s saved, freeing up memory", host.Host)
	// Free up memory
	hostToSave.HealthCheck = nil
	hostToSave.Resources = nil
	hostToSave.VirtualMachines = nil

	return nil
}

func (s *OrchestratorService) SetHealthCheckTimeout(timeout time.Duration) {
	s.healthCheckTimeout = timeout
}
