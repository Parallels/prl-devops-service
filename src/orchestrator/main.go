package orchestrator

import (
	"context"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/handlers"
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

	// Initialize WebSocket Manager and Handlers
	manager := NewHostWebSocketManager(s.ctx)
	pdfmHandler := handlers.NewPDfMEventHandler(manager)
	pdfmHandler.SetResourceUpdater(s)
	handlers.NewHostHealthHandler(manager)
	statsHandler := handlers.NewHostStatsHandler(manager)
	statsHandler.SetResourceUpdater(s)
	handlers.NewHostLogsHandler(manager)
	handlers.NewHostCatalogCacheEventHandler(manager, func(hostId string) {
		go globalOrchestratorService.RefreshHostCache(hostId)
	})

	// Initial refresh of connections
	if hosts, err := s.db.GetOrchestratorHosts(s.ctx, ""); err == nil {
		manager.RefreshConnections(hosts)
	}

	// Start periodic connection monitor (checks every 30 seconds)
	manager.StartConnectionMonitor(30 * time.Second)

	if waitForInit {
		s.ctx.LogInfof("[Orchestrator] Waiting for API to be initialized")
		<-restapi.Initialized
	}

	s.ctx.LogInfof("[Orchestrator] Starting Orchestrator Background Service")
	firstRun := true
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
				go s.processHostWaitingGroup(host, firstRun, &wg)
			}
			wg.Wait()

			if len(dtoOrchestratorHosts) > 0 {
				s.ctx.LogInfof("[Orchestrator] processed %v hosts", len(dtoOrchestratorHosts))
				s.ctx.LogInfof("[Orchestrator] Sleeping for %s seconds", s.refreshInterval)
			}

			firstRun = false
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

	manager := GetHostWebSocketManager()
	if manager != nil {
		manager.Shutdown()
	}

	s.ctx.LogInfof("[Orchestrator] Orchestrator Background Service Stopped")
}

func (s *OrchestratorService) Refresh() {
	dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
	if err != nil {
		return
	}

	for _, host := range dtoOrchestratorHosts {
		go s.processHost(host, true)
	}
}

func (s *OrchestratorService) processHostWaitingGroup(host models.OrchestratorHost, forceRefresh bool, wg *sync.WaitGroup) {
	defer wg.Done()

	select {
	case <-s.syncContext.Done():
		return
	default:
		s.processHost(host, forceRefresh)
	}
}

func (s *OrchestratorService) processHost(host models.OrchestratorHost, forceRefresh bool) {
	// Check if host is connected via WebSocket
	manager := GetHostWebSocketManager()
	websocketPingFailed := false

	if manager != nil && manager.IsConnected(host.ID) && host.State == "healthy" {
		// Check for staleness
		// If the host hasn't updated its status (via pong) in a while, we should verify health via HTTP
		lastUpdated, err := time.Parse(time.RFC3339Nano, host.UpdatedAt)
		stalenessThreshold := s.refreshInterval * stalenessMultipler

		if err == nil && time.Since(lastUpdated) < stalenessThreshold && !forceRefresh {
			s.ctx.LogDebugf("[Orchestrator] Host %s is connected and fresh (last updated: %s). Skipping HTTP health check.", host.Host, host.UpdatedAt)
			// Ping successful (implied by freshness), skip HTTP health check
			if !host.HasWebsocketEvents {
				host.HasWebsocketEvents = true
				_, _ = s.db.UpdateOrchestratorHostWebsocketStatus(s.ctx, host.ID, true)
			}
			return
		} else {
			s.ctx.LogWarnf("[Orchestrator] Host %s is connected but stale (last updated: %s). Falling back to HTTP health check.", host.Host, host.UpdatedAt)
			websocketPingFailed = true
			if host.HasWebsocketEvents {
				host.HasWebsocketEvents = false
				updated, _ := s.db.UpdateOrchestratorHostWebsocketStatus(s.ctx, host.ID, false)
				if updated {
					if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
						msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_WEBSOCKET_DISCONNECTED", apimodels.HostHealthUpdate{
							HostID: host.ID,
							State:  "websocket_disconnected",
						})
						go func() {
							if err := emitter.Broadcast(msg); err != nil {
								s.ctx.LogErrorf("[Orchestrator] Failed to broadcast event %s: %v", "HOST_WEBSOCKET_DISCONNECTED", err)
							} else {
								s.ctx.LogInfof("[Orchestrator] Broadcasted HOST_WEBSOCKET_DISCONNECTED event for host %s (detected staleness in processHost)", host.Host)
							}
						}()
					}
				}
			}
		}
	}

	s.ctx.LogInfof("[Orchestrator] Processing host %s", host.Host)

	host.HealthCheck = &apimodels.ApiHealthCheck{}
	if healthCheck, err := s.GetHostSystemHealthCheck(&host); err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error getting health check for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		_ = s.persistHost(&host)

		// Broadcast host unhealthy event
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_HEALTH_UPDATE", apimodels.HostHealthUpdate{
				HostID: host.ID,
				State:  host.State,
			})
			go func() {
				if err := emitter.Broadcast(msg); err != nil {
					s.ctx.LogErrorf("[Orchestrator] Failed to broadcast HOST_HEALTH_UPDATE event: %v", err)
				}
			}()
		}
		return
	} else {
		s.ctx.LogInfof("[Orchestrator] host %s is alive and well: %s", host.Host, healthCheck.Message)
		host.SetHealthy()
		host.HealthCheck = healthCheck

		// If WebSocket ping failed but HTTP health check succeeded, broadcast degraded WebSocket event
		if websocketPingFailed {
			s.ctx.LogWarnf("[Orchestrator] Host %s has degraded WebSocket connection, using HTTP fallback", host.Host)
			if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
				msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_WEBSOCKET_DEGRADED", apimodels.HostHealthUpdate{
					HostID: host.ID,
					State:  host.State,
				})
				go func() {
					if err := emitter.Broadcast(msg); err != nil {
						s.ctx.LogErrorf("[Orchestrator] Failed to broadcast HOST_WEBSOCKET_DEGRADED event: %v", err)
					}
				}()
			}
		}
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

	// Update host with hardware information using common function
	s.updateHostWithHardwareInfo(&host, hardwareInfo)

	// Updating the Virtual Machines
	vms, err := s.GetHostVirtualMachinesInfo(&host)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error getting virtual machines for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_HEALTH_UPDATE", apimodels.HostHealthUpdate{
				HostID: host.ID,
				State:  host.State,
			})
			go func() {
				if err := emitter.Broadcast(msg); err != nil {
					s.ctx.LogErrorf("[Orchestrator] Failed to broadcast event %s: %v", "HOST_HEALTH_UPDATE", err)
				}
			}()
		}
		_ = s.persistHost(&host)
		return
	}

	host.VirtualMachines = make([]models.VirtualMachine, 0)
	for _, vm := range vms {
		dtoVm := mappers.MapDtoVirtualMachineFromApi(vm)
		dtoVm.HostId = host.ID
		dtoVm.HostName = getHostName(host)
		dtoVm.Host = host.GetHost()
		dtoVm.HostUrl = host.GetHostUrl()
		host.VirtualMachines = append(host.VirtualMachines, dtoVm)
	}

	totalAppleVms := 0
	for _, vm := range host.VirtualMachines {
		if vm.Type == "APPLE_VZ_VM" {
			totalAppleVms++
		}
	}

	host.ReverseProxyHosts = make([]*models.ReverseProxyHost, 0)
	// Updating the reverse proxy hosts
	rpConfig, err := s.CallGetHostReverseProxyConfig(&host)
	if err == nil && rpConfig != nil {
		host.ReverseProxy = &models.ReverseProxy{
			ID:      rpConfig.ID,
			Host:    rpConfig.Host,
			Port:    rpConfig.Port,
			HostID:  host.ID,
			Enabled: rpConfig.Enabled,
		}
		// Getting all of the hosts in the reverse proxy config
		hosts, err := s.CallGetHostReverseProxyHosts(&host)
		if err == nil && hosts != nil {
			for _, rpHost := range hosts {
				host.ReverseProxyHosts = append(host.ReverseProxyHosts, rpHost)
			}
		}
	}

	host.Resources.TotalAppleVms = int64(totalAppleVms)
	host.UpdatedAt = helpers.GetUtcCurrentDateTime()

	// Getting Cache Items
	if cacheList, err := s.CallGetHostCatalogCache(&host); err == nil && cacheList != nil {
		host.CacheItems = make([]apimodels.HostCatalogCacheItem, 0)
		for _, manifest := range cacheList.Manifests {
			host.CacheItems = append(host.CacheItems, apimodels.HostCatalogCacheItem{
				CatalogId:    manifest.CatalogId,
				Version:      manifest.Version,
				Architecture: manifest.Architecture,
				CacheSize:    manifest.CacheSize,
				CacheType:    manifest.CacheType,
				CachedDate:   manifest.CacheDate,
			})
		}
	} else {
		s.ctx.LogWarnf("[Orchestrator] Defaulting or error getting cache items for host %s: %v", host.Host, err)
	}

	s.ctx.LogInfof("[Orchestrator] Host %s has %v CPU Cores and %v Mb of RAM, contains %v VMs of which %v are MacVMs", host.Host, host.Resources.Total.LogicalCpuCount, host.Resources.Total.MemorySize, len(host.VirtualMachines), host.Resources.TotalAppleVms)

	for _, vm := range host.VirtualMachines {
		listVMSnapshotResponse, err := s.GetHostVirtualMachineSnapshotsWithAPI(s.ctx, host.ID, vm.ID, false)
		if err != nil {
			s.ctx.LogErrorf("[Orchestrator] Error getting snapshots for VM %s: %v", vm.ID, err.Error())
		}
		var dbVMSnapshots []models.VMSnapshot
		if listVMSnapshotResponse != nil {
			dbVMSnapshots = mappers.VMSnapshotsApiToDto(listVMSnapshotResponse.Snapshots)
		}
		s.db.SetHostVMSnapshots(s.ctx, host.ID, models.VMSnapshots{
			VMId:       vm.ID,
			VMSnapshot: dbVMSnapshots,
		})
	}
	_ = s.persistHost(&host)

	// Free up memory
	host.HealthCheck = nil
	host.Resources = nil
	host.VirtualMachines = nil
	host.ReverseProxy = nil
	host.ReverseProxyHosts = nil
	host.CacheItems = nil

	// Emit host health update event
	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_HEALTH_UPDATE", apimodels.HostHealthUpdate{
			HostID: host.ID,
			State:  host.State,
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				s.ctx.LogErrorf("[Orchestrator] Failed to broadcast event %s: %v", "HOST_HEALTH_UPDATE", err)
			}
		}()
	}
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
		// Another process (e.g., ping/pong handler) updated the timestamp more recently
		// Use the newer timestamp but still save our health check data
		s.ctx.LogDebugf("[Orchestrator] Host %s was updated by another process, using newer timestamp", host.Host)
		hostToSave.UpdatedAt = oldHost.UpdatedAt
	}

	s.ctx.LogDebugf("[Orchestrator] Saving host %s", host.Host)
	if _, err := s.db.UpdateOrchestratorHost(s.ctx, &hostToSave); err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error saving host %s: %v", host.Host, err.Error())
		return err
	}

	// Free up memory
	hostToSave.HealthCheck = nil
	hostToSave.Resources = nil
	hostToSave.VirtualMachines = nil
	hostToSave.ReverseProxy = nil
	hostToSave.ReverseProxyHosts = nil
	hostToSave.CacheItems = nil

	return nil
}

func (s *OrchestratorService) updateHostWithHardwareInfo(host *models.OrchestratorHost, hardwareInfo *apimodels.SystemUsageResponse) {
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
	host.IsLogStreamingEnabled = hardwareInfo.IsLogStreamingEnabled
	host.EnabledModules = hardwareInfo.EnabledModules
	host.CacheConfig = hardwareInfo.CacheConfig
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
}

func (s *OrchestratorService) UpdateHostResourcesForEvent(ctx basecontext.ApiContext, hostID string) error {
	host, err := s.GetHost(ctx, hostID)
	if err != nil {
		return err
	}

	// Get hardware info using existing orchestrator service method
	hardwareInfo, err := s.GetHostHardwareInfo(host)
	if err != nil {
		ctx.LogErrorf("[Orchestrator] Error getting hardware info for host %s: %v", hostID, err)
		return err
	}

	// Update host with hardware information
	s.updateHostWithHardwareInfo(host, hardwareInfo)

	if err := s.persistHost(host); err != nil {
		ctx.LogErrorf("[Orchestrator] Error persisting host %s: %v", hostID, err)
		return err
	}

	ctx.LogInfof("[Orchestrator] Updated host %s resources after VM event", hostID)
	return nil
}

func (s *OrchestratorService) RefreshHostCache(hostId string) {
	if host, err := s.db.GetOrchestratorHost(s.ctx, hostId); err == nil && host != nil {
		if cacheList, err := s.CallGetHostCatalogCache(host); err == nil && cacheList != nil {
			host.CacheItems = make([]apimodels.HostCatalogCacheItem, 0)
			for _, manifest := range cacheList.Manifests {
				host.CacheItems = append(host.CacheItems, apimodels.HostCatalogCacheItem{
					CatalogId:    manifest.CatalogId,
					Version:      manifest.Version,
					Architecture: manifest.Architecture,
					CacheSize:    manifest.CacheSize,
					CacheType:    manifest.CacheType,
					CachedDate:   manifest.CacheDate,
				})
			}
			host.UpdatedAt = helpers.GetUtcCurrentDateTime()
			_ = s.persistHost(host)
		} else {
			s.ctx.LogWarnf("[Orchestrator] Error doing real-time cache refresh for host %s: %v", host.Host, err)
		}
	}
}

func (s *OrchestratorService) SetHealthCheckTimeout(timeout time.Duration) {
	s.healthCheckTimeout = timeout
}

func getHostName(host models.OrchestratorHost) string {
	if host.Description != "" {
		return host.Description
	}
	return host.Host
}
