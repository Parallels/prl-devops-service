package interfaces

// OrchestratorService mirrors orchestrator method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

type OrchestratorService interface {
	GetOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error)
	GetActiveOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error)
	GetOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) (*models.OrchestratorHost, error)
	CreateOrchestratorHost(ctx basecontext.ApiContext, host models.OrchestratorHost) (*models.OrchestratorHost, error)
	DeleteOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) error
	DeleteOrchestratorVirtualMachine(ctx basecontext.ApiContext, idOrHost string, vmIdOrName string) error
	UpdateOrchestratorHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error)
	UpdateOrchestratorHostDetails(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error)
	UpdateOrchestratorHostTimestamp(ctx basecontext.ApiContext, hostID string, timestamp string, state string) error
	UpdateOrchestratorHostWebsocketStatus(ctx basecontext.ApiContext, hostID string, hasWebsocketEvents bool) (bool, error)
	GetOrchestratorAvailableResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem
	GetOrchestratorTotalResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem
	GetOrchestratorInUseResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem
	GetOrchestratorReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem
	GetOrchestratorSystemReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem
	GetOrchestratorHostResources(ctx basecontext.ApiContext, hostId string) (*models.HostResources, error)
	GetOrchestratorVirtualMachines(ctx basecontext.ApiContext, filter string) ([]models.VirtualMachine, error)
	GetOrchestratorHostVirtualMachines(ctx basecontext.ApiContext, hostId string, filter string) ([]models.VirtualMachine, error)
	GetOrchestratorHostVirtualMachine(ctx basecontext.ApiContext, hostId string, machineId string) (*models.VirtualMachine, error)
	GetOrchestratorReverseProxyHosts(ctx basecontext.ApiContext, hostId string, filter string) ([]*models.ReverseProxyHost, error)
	GetOrchestratorReverseProxyHost(ctx basecontext.ApiContext, hostId string, rpIdOrName string) (*models.ReverseProxyHost, error)
	GetOrchestratorReverseProxyConfig(ctx basecontext.ApiContext, hostId string) (*models.ReverseProxy, error)
	UpsertOrchestratorHostVM(ctx basecontext.ApiContext, hostID string, vm models.VirtualMachine) error
	RemoveOrchestratorHostVM(ctx basecontext.ApiContext, hostID string, vmID string) error
	UpdateOrchestratorHostVMState(ctx basecontext.ApiContext, hostID string, vmID string, state string) error
	UpdateOrchestratorHostVMUptime(ctx basecontext.ApiContext, hostID string, vmID string, uptime string) error
	UpdateOrchestratorHostResources(ctx basecontext.ApiContext, hostID string, resources *models.HostResources) error
	UpdateOrchestratorHostCacheConfig(ctx basecontext.ApiContext, hostID string, cacheConfig *apimodels.CatalogCacheConfig) error
	ReplaceOrchestratorHostVMs(ctx basecontext.ApiContext, hostID string, vms []models.VirtualMachine) error
	GetHostVMSnapshots(ctx basecontext.ApiContext, hostId string) (*models.HostsVMSnapshotsRecord, error)
	SetHostVMSnapshots(ctx basecontext.ApiContext, hostId string, snapshots models.VMSnapshots) error
}
