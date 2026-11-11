package orchestrator

import (
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"

	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostVirtualMachines(ctx basecontext.ApiContext, hostId string, filter string, noCache bool) ([]*models.VirtualMachine, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	if noCache {
		// Targeted live fetch of this host only (no full-orchestrator refresh):
		// ask the host for live prlctl data and merge it into the DB.
		s.refetchHostVmsLive(ctx, dbService, hostId)
	}

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	vms, err := dbService.GetOrchestratorHostVirtualMachines(ctx, hostId, "")
	if err != nil {
		return nil, err
	}

	result := make([]*models.VirtualMachine, 0)

	// Updating Host State and Name for each VM
	for _, vm := range vms {
	hostLoop:
		for _, host := range hosts {
			if strings.EqualFold(vm.HostId, host.ID) {
				vm.HostState = host.State
				vm.HostName = getHostName(host)
				break hostLoop
			}
		}

		result = append(result, &vm)
	}

	return result, nil
}

func (s *OrchestratorService) GetHostVirtualMachine(ctx basecontext.ApiContext, hostId string, vmId string, noCache bool) (*models.VirtualMachine, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	if noCache {
		// Targeted live fetch of this single VM (not the whole fleet).
		s.refetchHostVmLive(ctx, dbService, hostId, vmId)
	}

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	vm, err := dbService.GetOrchestratorHostVirtualMachine(ctx, hostId, vmId)
	if err != nil {
		return nil, err
	}

	// Updating Host State and Name for each VM
	for _, host := range hosts {
		if vm.HostId == host.ID {
			vm.HostState = host.State
			vm.HostName = getHostName(host)
			break
		}
	}

	return vm, nil
}

func (s *OrchestratorService) GetHostAppleVirtualMachines(ctx basecontext.ApiContext, hostId string, vmId string) ([]*models.VirtualMachine, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	appleVms := make([]*models.VirtualMachine, 0)

	vms, err := dbService.GetOrchestratorHostVirtualMachines(ctx, hostId, "")
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Type == "APPLE_VZ_VM" {
			appleVms = append(appleVms, &vm)
		}
	}

	// Updating Host State for each VM
	for _, host := range hosts {
		for _, vm := range appleVms {
			if vm.HostId == host.ID {
				vm.HostState = host.State
				break
			}
		}
	}

	return appleVms, nil
}

func (s *OrchestratorService) GetHostVirtualMachinesInfo(host *models.OrchestratorHost) ([]api_models.ParallelsVM, error) {
	return s.callGetHostVms(host, false)
}

// GetHostVirtualMachinesInfoLive fetches the host's VMs bypassing the host-side
// cache (GET /v1/machines?noCache=true) — live data straight from prlctl.
func (s *OrchestratorService) GetHostVirtualMachinesInfoLive(host *models.OrchestratorHost) ([]api_models.ParallelsVM, error) {
	return s.callGetHostVms(host, true)
}

func (s *OrchestratorService) callGetHostVms(host *models.OrchestratorHost, live bool) ([]api_models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	path := "/v1/machines"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}
	if live {
		query := url.Query()
		query.Set("noCache", "true")
		url.RawQuery = query.Encode()
	}

	var response []api_models.ParallelsVM
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting virtual machines for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return response, nil
}

// CallGetHostVirtualMachineInfo fetches a single VM from the host by ID via HTTP.
// Used to sync VM config after a state-change event where only partial data is received.
func (s *OrchestratorService) CallGetHostVirtualMachineInfo(host *models.OrchestratorHost, vmID string) (*api_models.ParallelsVM, error) {
	return s.callGetHostVm(host, vmID, false)
}

// CallGetHostVirtualMachineInfoLive fetches a single VM from the host by ID,
// bypassing the host-side cache (GET /v1/machines/{id}?noCache=true).
func (s *OrchestratorService) CallGetHostVirtualMachineInfoLive(host *models.OrchestratorHost, vmID string) (*api_models.ParallelsVM, error) {
	return s.callGetHostVm(host, vmID, true)
}

func (s *OrchestratorService) callGetHostVm(host *models.OrchestratorHost, vmID string, live bool) (*api_models.ParallelsVM, error) {
	httpClient := s.getApiClient(*host)
	path := fmt.Sprintf("/v1/machines/%s", vmID)
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}
	if live {
		query := url.Query()
		query.Set("noCache", "true")
		url.RawQuery = query.Encode()
	}

	var response api_models.ParallelsVM
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting virtual machine %s for host %s: %v", vmID, host.Host, apiResponse.StatusCode)
	}

	return &response, nil
}

// refetchHostVmsLive live-fetches all VMs of one host (targeted — other hosts are
// untouched), merges preserved IPs/statuses and upserts each VM into the DB.
// Best-effort: on a fetch error it logs and the caller serves the last DB state.
func (s *OrchestratorService) refetchHostVmsLive(ctx basecontext.ApiContext, dbService interfaces.DatabaseService, hostID string) {
	dbHost, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogWarnf("[Orchestrator] Live VM refetch: host %s not found: %v", hostID, err)
		return
	}
	vms, err := s.GetHostVirtualMachinesInfoLive(dbHost)
	if err != nil {
		ctx.LogWarnf("[Orchestrator] Live VM refetch failed for host %s: %v", dbHost.Host, err)
		return
	}
	for _, vm := range vms {
		s.upsertLiveVm(ctx, dbService, dbHost, vm)
	}
	ctx.LogDebugf("[Orchestrator] Live VM refetch complete for host %s: %d VMs", dbHost.Host, len(vms))
}

// refetchHostVmLive live-fetches a single VM of one host and upserts it into the DB.
func (s *OrchestratorService) refetchHostVmLive(ctx basecontext.ApiContext, dbService interfaces.DatabaseService, hostID, vmID string) {
	dbHost, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogWarnf("[Orchestrator] Live VM refetch: host %s not found: %v", hostID, err)
		return
	}
	vm, err := s.CallGetHostVirtualMachineInfoLive(dbHost, vmID)
	if err != nil || vm == nil {
		ctx.LogWarnf("[Orchestrator] Live VM refetch failed for VM %s on host %s: %v", vmID, dbHost.Host, err)
		return
	}
	s.upsertLiveVm(ctx, dbService, dbHost, *vm)
}

// upsertLiveVm maps a live host-side VM to the orchestrator DB model and upserts
// it, keeping any IP/status that only the stored record knows about.
func (s *OrchestratorService) upsertLiveVm(ctx basecontext.ApiContext, dbService interfaces.DatabaseService, host *models.OrchestratorHost, vm api_models.ParallelsVM) {
	dtoVm := mappers.MapDtoVirtualMachineFromApi(vm)
	dtoVm.HostId = host.ID
	dtoVm.HostName = getHostName(*host)
	dtoVm.Host = host.GetHost()
	dtoVm.HostUrl = host.GetHostUrl()

	if stored, err := dbService.GetOrchestratorHostVirtualMachine(ctx, host.ID, vm.ID); err == nil && stored != nil {
		dtoVm = mergePreservedIp(*stored, dtoVm)
	}

	if err := dbService.UpsertOrchestratorHostVM(ctx, host.ID, dtoVm); err != nil {
		ctx.LogErrorf("[Orchestrator] Live VM refetch: error upserting VM %s on host %s: %v", vm.ID, host.Host, err)
	}
}
