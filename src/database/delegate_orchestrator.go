package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/orchestrator delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) DeleteOrchestratorVirtualMachine(ctx basecontext.ApiContext, idOrHost string, vmIdOrName string) error {
	return g.legacy.DeleteOrchestratorVirtualMachine(ctx, idOrHost, vmIdOrName)
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

func (g *GormDatabase) ReplaceOrchestratorHostVMs(ctx basecontext.ApiContext, hostID string, vms []models.VirtualMachine) error {
	return g.legacy.ReplaceOrchestratorHostVMs(ctx, hostID, vms)
}

func (g *GormDatabase) GetHostVMSnapshots(ctx basecontext.ApiContext, hostId string) (*models.HostsVMSnapshotsRecord, error) {
	return g.legacy.GetHostVMSnapshots(ctx, hostId)
}

func (g *GormDatabase) SetHostVMSnapshots(ctx basecontext.ApiContext, hostId string, snapshots models.VMSnapshots) error {
	return g.legacy.SetHostVMSnapshots(ctx, hostId, snapshots)
}
