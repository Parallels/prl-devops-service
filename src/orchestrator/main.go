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
	"github.com/Parallels/prl-devops-service/mappers"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/handlers"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/telemetry"
)

const stalenessMultiplier = 3

var globalOrchestratorService *OrchestratorService

type OrchestratorService struct {
	ctx                 basecontext.ApiContext
	timeout             time.Duration
	healthCheckTimeout  time.Duration
	refreshInterval     time.Duration
	fullRefreshInterval time.Duration
	syncContext         context.Context
	cancel              context.CancelFunc
	db                  *data.JsonDatabase
	hwQueue             *hardwareUpdateQueue
}

func NewOrchestratorService(ctx basecontext.ApiContext) *OrchestratorService {
	if globalOrchestratorService == nil {
		globalOrchestratorService = &OrchestratorService{
			ctx:                ctx,
			timeout:            5 * time.Minute,
			healthCheckTimeout: 3 * time.Second,
		}
		cfg := config.Get()
		pullFreq := time.Duration(cfg.OrchestratorPullFrequency()) * time.Second
		globalOrchestratorService.refreshInterval = pullFreq
		// Full refresh (VMs + snapshots + hardware) runs every 10× the health-check interval.
		globalOrchestratorService.fullRefreshInterval = pullFreq * 10
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
	handlers.NewHostHealthHandler(manager)
	statsHandler := handlers.NewHostStatsHandler(manager)
	statsHandler.SetResourceUpdater(s)
	handlers.NewHostLogsHandler(manager)
	handlers.NewHostCatalogCacheEventHandler(manager, func(hostId string) {
		go globalOrchestratorService.RefreshHostCache(hostId)
	})
	rpHandler := handlers.NewHostReverseProxyEventHandler(manager)
	rpHandler.SetReverseProxyUpdater(s)
	handlers.NewHostJobEventHandler(manager)

	// Initialize per-host hardware update queue and wire it to the PDFM handler.
	s.hwQueue = newHardwareUpdateQueue(s)
	pdfmHandler.SetHardwareEnqueuer(s.hwQueue)
	s.hwQueue.Start(s.ctx)

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

	// Startup: full data load for every host (hardware, VMs, snapshots, cache, reverse proxy).
	if startupHosts, err := s.db.GetOrchestratorHosts(s.ctx, ""); err == nil && len(startupHosts) > 0 {
		var wg sync.WaitGroup
		for _, host := range startupHosts {
			wg.Add(1)
			go func(h models.OrchestratorHost) {
				defer wg.Done()
				select {
				case <-s.syncContext.Done():
				default:
					s.fullRefreshHost(h, true)
				}
			}(host)
		}
		wg.Wait()
		s.ctx.LogInfof("[Orchestrator] Startup full refresh complete for %d hosts", len(startupHosts))
	}

	// Background: periodic full refresh (self-healing) on a longer interval.
	go s.runFullRefreshLoop()

	// Background: lightweight health check on every refreshInterval tick.
	for {
		select {
		case <-s.syncContext.Done():
			return
		default:
			time.Sleep(s.refreshInterval)

			dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
			if err != nil {
				continue
			}

			var wg sync.WaitGroup
			for _, host := range dtoOrchestratorHosts {
				wg.Add(1)
				go s.processHostWaitingGroup(host, false, &wg)
			}
			wg.Wait()
		}
	}
}

// runFullRefreshLoop runs a full data refresh for all hosts every fullRefreshInterval.
// This is the self-healing mechanism that re-syncs VMs, snapshots, hardware, and cache
// in case any WebSocket events were missed.
func (s *OrchestratorService) runFullRefreshLoop() {
	ticker := time.NewTicker(s.fullRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.syncContext.Done():
			return
		case <-ticker.C:
			dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
			if err != nil {
				continue
			}
			s.ctx.LogInfof("[Orchestrator] Periodic full refresh for %d hosts", len(dtoOrchestratorHosts))
			for _, host := range dtoOrchestratorHosts {
				go s.fullRefreshHost(host, true)
			}
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

	if s.hwQueue != nil {
		s.hwQueue.Stop()
	}

	manager := GetHostWebSocketManager()
	if manager != nil {
		manager.Shutdown()
	}

	s.ctx.LogInfof("[Orchestrator] Orchestrator Background Service Stopped")
}

// Refresh forces a full data sync for all hosts. Called when noCache=true is passed to
// the read APIs (e.g. GetVirtualMachines). Delegates to fullRefreshHost so that VMs,
// snapshots, hardware, and cache are all refreshed in one pass.
func (s *OrchestratorService) Refresh() {
	dtoOrchestratorHosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
	if err != nil {
		return
	}

	for _, host := range dtoOrchestratorHosts {
		go s.fullRefreshHost(host, false)
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

// processHost is the lightweight health-check tick. It only verifies WebSocket freshness
// and, when stale, performs an HTTP health probe. Heavy data work (VMs, snapshots,
// hardware, cache) lives in fullRefreshHost which runs at startup and on a longer interval.
func (s *OrchestratorService) processHost(host models.OrchestratorHost, forceRefresh bool) {
	if !host.Enabled {
		return
	}

	manager := GetHostWebSocketManager()
	websocketPingFailed := false

	if manager != nil && manager.IsConnected(host.ID) && host.State == "healthy" {
		lastUpdated, err := time.Parse(time.RFC3339Nano, host.UpdatedAt)
		stalenessThreshold := s.refreshInterval * stalenessMultiplier

		if err == nil && time.Since(lastUpdated) < stalenessThreshold && !forceRefresh {
			s.ctx.LogDebugf("[Orchestrator] Host %s is connected and fresh (last updated: %s). Skipping health check.", host.Host, host.UpdatedAt)
			if !host.HasWebsocketEvents {
				host.HasWebsocketEvents = true
				_, _ = s.db.UpdateOrchestratorHostWebsocketStatus(s.ctx, host.ID, true)
			}
			return
		}

		// Host is connected but pong is late — fall back to HTTP for this tick.
		// Do NOT clear HasWebsocketEvents: the connection is still alive; staleness
		// only means we haven't received a pong recently. The flag is cleared only
		// when the connection actually drops (notifyDisconnection / DisconnectHost).
		s.ctx.LogWarnf("[Orchestrator] Host %s is connected but stale (last updated: %s). Falling back to HTTP health check.", host.Host, host.UpdatedAt)
		websocketPingFailed = true
		// Defensive: if the flag was incorrectly cleared while the connection is
		// still alive (e.g. due to a previous bug or a concurrent race), restore it
		// now so the UI reflects the actual connection state.
		if !host.HasWebsocketEvents {
			_, _ = s.db.UpdateOrchestratorHostWebsocketStatus(s.ctx, host.ID, true)
		}
	}

	s.ctx.LogDebugf("[Orchestrator] Health checking host %s", host.Host)

	host.HealthCheck = &apimodels.ApiHealthCheck{}
	healthCheck, err := s.GetHostSystemHealthCheck(&host)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] Error getting health check for host %s: %v", host.Host, err.Error())
		host.SetUnhealthy(err.Error())
		_ = s.persistHost(&host)
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_HEALTH_UPDATE", apimodels.HostHealthUpdate{
				HostID: host.ID,
				State:  host.State,
			})
			go func() { _ = emitter.Broadcast(msg) }()
		}
		return
	}

	host.SetHealthy()
	host.HealthCheck = healthCheck

	if websocketPingFailed {
		s.ctx.LogWarnf("[Orchestrator] Host %s has degraded WebSocket connection, using HTTP fallback", host.Host)
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_WEBSOCKET_DEGRADED", apimodels.HostHealthUpdate{
				HostID: host.ID,
				State:  host.State,
			})
			go func() { _ = emitter.Broadcast(msg) }()
		}
	}

	_ = s.persistHost(&host)
	host.HealthCheck = nil

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_HEALTH_UPDATE", apimodels.HostHealthUpdate{
			HostID: host.ID,
			State:  host.State,
		})
		go func() { _ = emitter.Broadcast(msg) }()
	}
}

// fullRefreshHost performs a complete data sync for one host: hardware info, all VMs,
// snapshots for each VM, cache items, and optionally reverse proxy config.
// It is called at startup, by the periodic full-refresh loop, and by Refresh().
// Snapshots are fetched directly from the host (no per-VM health probe).
func (s *OrchestratorService) fullRefreshHost(host models.OrchestratorHost, loadReverseProxy bool) {
	if !host.Enabled {
		return
	}

	s.ctx.LogInfof("[Orchestrator] Full refresh: host %s", host.Host)

	hardwareInfo, err := s.GetHostHardwareInfo(&host)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] Full refresh: hardware info error for host %s: %v", host.Host, err)
		return
	}
	// Hardware info fetch succeeded — the host is reachable, so mark it healthy
	// before proceeding. This ensures CallGetHostReverseProxyConfig (which guards
	// on host.State == HealthyState) can run during the same refresh pass.
	host.State = HealthyState
	s.updateHostWithHardwareInfo(&host, hardwareInfo)

	vms, err := s.GetHostVirtualMachinesInfo(&host)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] Full refresh: VMs error for host %s: %v", host.Host, err)
		return
	}

	host.VirtualMachines = make([]models.VirtualMachine, 0, len(vms))
	totalAppleVms := 0
	for _, vm := range vms {
		dtoVm := mappers.MapDtoVirtualMachineFromApi(vm)
		dtoVm.HostId = host.ID
		dtoVm.HostName = getHostName(host)
		dtoVm.Host = host.GetHost()
		dtoVm.HostUrl = host.GetHostUrl()
		host.VirtualMachines = append(host.VirtualMachines, dtoVm)
		if vm.Type == "APPLE_VZ_VM" {
			totalAppleVms++
		}
	}
	host.Resources.TotalAppleVms = int64(totalAppleVms)

	if loadReverseProxy {
		host.ReverseProxyHosts = make([]*models.ReverseProxyHost, 0)
		if rpConfig, err := s.CallGetHostReverseProxyConfig(&host); err == nil && rpConfig != nil {
			host.ReverseProxy = &models.ReverseProxy{
				ID:      rpConfig.ID,
				Host:    rpConfig.Host,
				Port:    rpConfig.Port,
				HostID:  host.ID,
				Enabled: rpConfig.Enabled,
			}
			if rpHosts, err := s.CallGetHostReverseProxyHosts(&host); err == nil && rpHosts != nil {
				host.ReverseProxyHosts = rpHosts
			}
		}
	}

	if cacheList, err := s.CallGetHostCatalogCache(&host); err == nil && cacheList != nil {
		host.CacheItems = make([]apimodels.HostCatalogCacheItem, 0, len(cacheList.Manifests))
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
		s.ctx.LogWarnf("[Orchestrator] Full refresh: cache error for host %s: %v", host.Host, err)
	}

	s.ctx.LogInfof("[Orchestrator] Host %s: %v CPU, %v MB RAM, %v VMs (%v MacVMs)",
		host.Host, host.Resources.Total.LogicalCpuCount, host.Resources.Total.MemorySize,
		len(host.VirtualMachines), host.Resources.TotalAppleVms)

	// Fetch snapshots directly — no health probe per VM.
	for _, vm := range host.VirtualMachines {
		snapshots, err := s.callGetVMSnapshotsFromHost(&host, vm.ID)
		if err != nil {
			s.ctx.LogErrorf("[Orchestrator] Full refresh: snapshots error for VM %s on host %s: %v", vm.ID, host.Host, err)
			continue
		}
		var dbVMSnapshots []models.VMSnapshot
		if snapshots != nil {
			dbVMSnapshots = mappers.VMSnapshotsApiToDto(snapshots.Snapshots)
		}
		s.db.SetHostVMSnapshots(s.ctx, host.ID, models.VMSnapshots{
			VMId:       vm.ID,
			VMSnapshot: dbVMSnapshots,
		})
	}

	// Atomically replace the VM list in the DB. This is the only place that bulk-replaces
	// VMs; PDFM event handlers use targeted per-VM atomic methods.
	if err := s.db.ReplaceOrchestratorHostVMs(s.ctx, host.ID, host.VirtualMachines); err != nil {
		s.ctx.LogErrorf("[Orchestrator] Full refresh: failed to replace VMs for host %s: %v", host.Host, err)
	}

	// Persist health/resources/config — VMs are managed separately above.
	host.VirtualMachines = nil
	_ = s.persistHost(&host)

	host.HealthCheck = nil
	host.Resources = nil
	host.ReverseProxy = nil
	host.ReverseProxyHosts = nil
	host.CacheItems = nil
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
			Host:    hardwareInfo.ReverseProxy.Host,
			Port:    hardwareInfo.ReverseProxy.Port,
			Enabled: hardwareInfo.ReverseProxy.Enabled,
		}
		host.ReverseProxyHosts = make([]*models.ReverseProxyHost, 0)
		for _, rpHost := range hardwareInfo.ReverseProxy.Hosts {
			dtoHost := mappers.ApiReverseProxyHostToDto(rpHost)
			host.ReverseProxyHosts = append(host.ReverseProxyHosts, &dtoHost)
		}
	}
}

func (s *OrchestratorService) UpdateHostResourcesForEvent(ctx basecontext.ApiContext, hostID string) error {
	// Use GetDatabaseHost to avoid a live /health/probe HTTP call on every stats event.
	host, err := s.GetDatabaseHost(ctx, hostID)
	if err != nil {
		return err
	}

	hardwareInfo, err := s.GetHostHardwareInfo(host)
	if err != nil {
		ctx.LogErrorf("[Orchestrator] Error getting hardware info for host %s: %v", hostID, err)
		return err
	}

	s.updateHostWithHardwareInfo(host, hardwareInfo)

	if err := s.persistHost(host); err != nil {
		ctx.LogErrorf("[Orchestrator] Error persisting host %s: %v", hostID, err)
		return err
	}

	ctx.LogInfof("[Orchestrator] Updated host %s resources after stats event", hostID)
	return nil
}

// UpdateHostVMForEvent fetches a single VM from the host and updates its record in the DB.
// Called after VM_STATE_CHANGED so that any config differences (RAM, CPU, IP, etc.)
// that arrived with or just before the state transition are captured without
// pulling the entire hardware info or all VMs from the host.
//
// IMPORTANT: the HTTP call to the host is made first (slow path), and only THEN
// do we re-read the host from the DB. This avoids overwriting concurrent DB changes
// (e.g. a VM_ADDED written by handleVmAdded while the HTTP call was in flight).
func (s *OrchestratorService) UpdateHostVMForEvent(ctx basecontext.ApiContext, hostID string, vmID string) error {
	// We need the host's connection info to make the HTTP call, but we must not hold
	// a stale snapshot of VirtualMachines across the slow network round-trip.
	// Read a minimal host just for the URL/auth, make the call, then re-read fresh.
	connHost, err := s.GetDatabaseHost(ctx, hostID)
	if err != nil {
		return err
	}
	if connHost == nil {
		return nil
	}

	// Slow path: fetch the VM from the host. Other events (VM_ADDED, VM_REMOVED, etc.)
	// may update the DB while we wait here — that is expected and correct.
	vm, err := s.CallGetHostVirtualMachineInfo(connHost, vmID)
	if err != nil {
		ctx.LogErrorf("[Orchestrator] Error fetching VM %s from host %s: %v", vmID, hostID, err)
		return err
	}

	// Re-read the host from DB after the HTTP call so we have the latest VirtualMachines
	// list — any VM_ADDED / VM_REMOVED events that fired during the HTTP call will
	// already be reflected here, and we won't overwrite them.
	host, err := s.GetDatabaseHost(ctx, hostID)
	if err != nil {
		return err
	}
	if host == nil {
		return nil
	}

	dtoVm := mappers.MapDtoVirtualMachineFromApi(*vm)
	dtoVm.HostId = host.ID
	dtoVm.HostName = getHostName(*host)
	dtoVm.Host = host.GetHost()
	dtoVm.HostUrl = host.GetHostUrl()

	vmIndex := -1
	for i, v := range host.VirtualMachines {
		if v.ID == vmID {
			vmIndex = i
			break
		}
	}
	if vmIndex != -1 {
		host.VirtualMachines[vmIndex] = dtoVm
	} else {
		host.VirtualMachines = append(host.VirtualMachines, dtoVm)
	}

	// Use persistHost so we don't overwrite a fresher UpdatedAt that handlePong may have just written.
	if err := s.persistHost(host); err != nil {
		ctx.LogErrorf("[Orchestrator] Error persisting VM %s update for host %s: %v", vmID, hostID, err)
		return err
	}

	ctx.LogInfof("[Orchestrator] Synced VM %s on host %s after state change", vmID, hostID)
	return nil
}

// UpdateHostReverseProxyForEvent implements handlers.ReverseProxyUpdater.
// It fetches the current reverse proxy config and hosts for a single host via HTTP
// and persists the result to the DB. Called by HostReverseProxyEventHandler when
// a reverse_proxy event arrives over WebSocket.
func (s *OrchestratorService) UpdateHostReverseProxyForEvent(ctx basecontext.ApiContext, hostID string) error {
	host, err := s.GetDatabaseHost(ctx, hostID)
	if err != nil {
		ctx.LogErrorf("[Orchestrator] Error getting host %s for reverse proxy update: %v", hostID, err)
		return err
	}
	if host == nil {
		ctx.LogWarnf("[Orchestrator] Host %s not found for reverse proxy update", hostID)
		return nil
	}

	rpConfig, err := s.CallGetHostReverseProxyConfig(host)
	if err == nil && rpConfig != nil {
		host.ReverseProxy = &models.ReverseProxy{
			ID:      rpConfig.ID,
			Host:    rpConfig.Host,
			Port:    rpConfig.Port,
			HostID:  host.ID,
			Enabled: rpConfig.Enabled,
		}
		rpHosts, err := s.CallGetHostReverseProxyHosts(host)
		if err == nil && rpHosts != nil {
			host.ReverseProxyHosts = rpHosts
		}
	}

	if err := s.persistHost(host); err != nil {
		ctx.LogErrorf("[Orchestrator] Error persisting reverse proxy update for host %s: %v", hostID, err)
		return err
	}

	ctx.LogInfof("[Orchestrator] Updated reverse proxy config for host %s", hostID)
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
