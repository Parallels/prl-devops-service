package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/orchestrator delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

func (g *GormDatabase) GetOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	return g.legacy.GetOrchestratorHosts(ctx, filter)
}

func (g *GormDatabase) GetActiveOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	return g.legacy.GetActiveOrchestratorHosts(ctx, filter)
}

func (g *GormDatabase) GetOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) (*models.OrchestratorHost, error) {
	return g.legacy.GetOrchestratorHost(ctx, idOrHost)
}

func (g *GormDatabase) CreateOrchestratorHost(ctx basecontext.ApiContext, host models.OrchestratorHost) (*models.OrchestratorHost, error) {
	return g.legacy.CreateOrchestratorHost(ctx, host)
}

func (g *GormDatabase) DeleteOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) error {
	return g.legacy.DeleteOrchestratorHost(ctx, idOrHost)
}

func (g *GormDatabase) DeleteOrchestratorVirtualMachine(ctx basecontext.ApiContext, idOrHost string, vmIdOrName string) error {
	return g.legacy.DeleteOrchestratorVirtualMachine(ctx, idOrHost, vmIdOrName)
}

func (g *GormDatabase) UpdateOrchestratorHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	return g.legacy.UpdateOrchestratorHost(ctx, host)
}

func (g *GormDatabase) UpdateOrchestratorHostDetails(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	return g.legacy.UpdateOrchestratorHostDetails(ctx, host)
}

func (g *GormDatabase) UpdateOrchestratorHostTimestamp(ctx basecontext.ApiContext, hostID string, timestamp string, state string) error {
	return g.legacy.UpdateOrchestratorHostTimestamp(ctx, hostID, timestamp, state)
}

func (g *GormDatabase) UpdateOrchestratorHostWebsocketStatus(ctx basecontext.ApiContext, hostID string, hasWebsocketEvents bool) (bool, error) {
	return g.legacy.UpdateOrchestratorHostWebsocketStatus(ctx, hostID, hasWebsocketEvents)
}

func (g *GormDatabase) GetOrchestratorAvailableResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.legacy.GetOrchestratorAvailableResources(ctx)
}

func (g *GormDatabase) GetOrchestratorTotalResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.legacy.GetOrchestratorTotalResources(ctx)
}

func (g *GormDatabase) GetOrchestratorInUseResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.legacy.GetOrchestratorInUseResources(ctx)
}

func (g *GormDatabase) GetOrchestratorReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.legacy.GetOrchestratorReservedResources(ctx)
}

func (g *GormDatabase) GetOrchestratorSystemReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.legacy.GetOrchestratorSystemReservedResources(ctx)
}

func (g *GormDatabase) GetOrchestratorHostResources(ctx basecontext.ApiContext, hostId string) (*models.HostResources, error) {
	return g.legacy.GetOrchestratorHostResources(ctx, hostId)
}

func (g *GormDatabase) GetOrchestratorVirtualMachines(ctx basecontext.ApiContext, filter string) ([]models.VirtualMachine, error) {
	return g.legacy.GetOrchestratorVirtualMachines(ctx, filter)
}

func (g *GormDatabase) GetOrchestratorHostVirtualMachines(ctx basecontext.ApiContext, hostId string, filter string) ([]models.VirtualMachine, error) {
	return g.legacy.GetOrchestratorHostVirtualMachines(ctx, hostId, filter)
}

func (g *GormDatabase) GetOrchestratorHostVirtualMachine(ctx basecontext.ApiContext, hostId string, machineId string) (*models.VirtualMachine, error) {
	return g.legacy.GetOrchestratorHostVirtualMachine(ctx, hostId, machineId)
}

func (g *GormDatabase) GetOrchestratorReverseProxyHosts(ctx basecontext.ApiContext, hostId string, filter string) ([]*models.ReverseProxyHost, error) {
	return g.legacy.GetOrchestratorReverseProxyHosts(ctx, hostId, filter)
}

func (g *GormDatabase) GetOrchestratorReverseProxyHost(ctx basecontext.ApiContext, hostId string, rpIdOrName string) (*models.ReverseProxyHost, error) {
	return g.legacy.GetOrchestratorReverseProxyHost(ctx, hostId, rpIdOrName)
}

func (g *GormDatabase) GetOrchestratorReverseProxyConfig(ctx basecontext.ApiContext, hostId string) (*models.ReverseProxy, error) {
	return g.legacy.GetOrchestratorReverseProxyConfig(ctx, hostId)
}

func (g *GormDatabase) UpsertOrchestratorHostVM(ctx basecontext.ApiContext, hostID string, vm models.VirtualMachine) error {
	return g.legacy.UpsertOrchestratorHostVM(ctx, hostID, vm)
}

func (g *GormDatabase) RemoveOrchestratorHostVM(ctx basecontext.ApiContext, hostID string, vmID string) error {
	return g.legacy.RemoveOrchestratorHostVM(ctx, hostID, vmID)
}

func (g *GormDatabase) UpdateOrchestratorHostVMState(ctx basecontext.ApiContext, hostID string, vmID string, state string) error {
	return g.legacy.UpdateOrchestratorHostVMState(ctx, hostID, vmID, state)
}

func (g *GormDatabase) UpdateOrchestratorHostVMUptime(ctx basecontext.ApiContext, hostID string, vmID string, uptime string) error {
	return g.legacy.UpdateOrchestratorHostVMUptime(ctx, hostID, vmID, uptime)
}

func (g *GormDatabase) UpdateOrchestratorHostResources(ctx basecontext.ApiContext, hostID string, resources *models.HostResources) error {
	return g.legacy.UpdateOrchestratorHostResources(ctx, hostID, resources)
}

func (g *GormDatabase) UpdateOrchestratorHostCacheConfig(ctx basecontext.ApiContext, hostID string, cacheConfig *apimodels.CatalogCacheConfig) error {
	return g.legacy.UpdateOrchestratorHostCacheConfig(ctx, hostID, cacheConfig)
}

func (g *GormDatabase) ReplaceOrchestratorHostVMs(ctx basecontext.ApiContext, hostID string, vms []models.VirtualMachine) error {
	return g.legacy.ReplaceOrchestratorHostVMs(ctx, hostID, vms)
}

func (g *GormDatabase) GetHostVMSnapshots(ctx basecontext.ApiContext, hostId string) (*models.HostsVMSnapshotsRecord, error) {
	return g.legacy.GetHostVMSnapshots(ctx, hostId)
}

func (g *GormDatabase) SetHostVMSnapshots(ctx basecontext.ApiContext, hostId string, snapshots models.VMSnapshots) error {
	return g.legacy.SetHostVMSnapshots(ctx, hostId, snapshots)
}
