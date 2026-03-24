package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

// validateHostAndVM is a shared helper function to validate host and VM for snapshot operations
func (s *OrchestratorService) validateHostAndVM(ctx basecontext.ApiContext, hostId string, vmId string, noCache bool) (*data_models.OrchestratorHost, *data_models.VirtualMachine, error) {
	if noCache {
		ctx.LogDebugf("[Orchestrator] No cache set, refreshing all hosts...")
		s.Refresh()
	}

	vm, err := s.GetVirtualMachine(ctx, vmId, false)
	if err != nil {
		return nil, nil, err
	}
	if vm == nil {
		return nil, nil, errors.NewWithCodef(404, "Virtual machine %s not found", vmId)
	}

	// Use GetDatabaseHost to avoid a live /health/probe HTTP call per snapshot operation.
	host, err := s.GetDatabaseHost(ctx, hostId)
	if err != nil {
		return nil, nil, err
	}
	if host == nil {
		return nil, nil, errors.NewWithCodef(404, "Host %s not found", hostId)
	}

	if !host.Enabled {
		return nil, nil, errors.NewWithCodef(400, "Host %s is disabled", host.Host)
	}
	if host.State != "healthy" {
		return nil, nil, errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	return host, vm, nil
}

// GetHostVirtualMachineSnapshotsWithAPI lists all snapshots for a virtual machine on an orchestrator host
func (s *OrchestratorService) GetHostVirtualMachineSnapshotsWithAPI(ctx basecontext.ApiContext, hostId string, vmId string, noCache bool) (*apimodels.ListVMSnapshotResponse, error) {
	host, vm, err := s.validateHostAndVM(ctx, hostId, vmId, noCache)
	if err != nil {
		return nil, err
	}

	return s.callGetVMSnapshotsFromHost(host, vm.ID)
}

// callGetVMSnapshotsFromHost fetches snapshots for a single VM directly from the host
// without performing a health probe. Used by fullRefreshHost to avoid an extra HTTP
// round-trip per VM during the snapshot sync loop.
func (s *OrchestratorService) callGetVMSnapshotsFromHost(host *data_models.OrchestratorHost, vmID string) (*apimodels.ListVMSnapshotResponse, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(2 * time.Minute)
	path := "/machines/" + vmID + "/snapshots"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response apimodels.ListVMSnapshotResponse
	_, err = httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetHostVirtualMachineSnapshots lists all snapshots for a virtual machine on an orchestrator host
func (s *OrchestratorService) GetHostVirtualMachineSnapshots(ctx basecontext.ApiContext, hostId string, vmId string) (*apimodels.ListVMSnapshotResponse, error) {
	hostSnapshots, err := s.db.GetHostVMSnapshots(ctx, hostId)
	if err != nil {
		return nil, err
	}
	var response apimodels.ListVMSnapshotResponse
	mapper := mappers.VMSnapshotsDtoToApi(hostSnapshots.VMSnapshots[vmId])
	response.Snapshots = mapper

	return &response, nil
}

// CreateHostVirtualMachineSnapshot creates a new snapshot for a virtual machine on an orchestrator host
func (s *OrchestratorService) CreateHostVirtualMachineSnapshot(ctx basecontext.ApiContext, hostId string, vmId string, request apimodels.CreateVMSnapshotRequest, noCache bool) (*apimodels.CreateVMSnapshotResponse, error) {
	host, vm, err := s.validateHostAndVM(ctx, hostId, vmId, noCache)
	if err != nil {
		return nil, err
	}

	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(10 * time.Minute) // creating snapshots can take a while
	path := "/machines/" + vm.ID + "/snapshots"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response apimodels.CreateVMSnapshotResponse
	_, err = httpClient.Post(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteAllHostVirtualMachineSnapshots deletes all snapshots for a virtual machine on an orchestrator host
func (s *OrchestratorService) DeleteAllHostVirtualMachineSnapshots(ctx basecontext.ApiContext, hostId string, vmId string, noCache bool) error {
	host, vm, err := s.validateHostAndVM(ctx, hostId, vmId, noCache)
	if err != nil {
		return err
	}

	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(10 * time.Minute) // deleting snapshots can take a while
	path := "/machines/" + vm.ID + "/snapshots"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	_, err = httpClient.Delete(url.String(), nil)
	if err != nil {
		return err
	}

	return nil
}

// DeleteHostVirtualMachineSnapshot deletes a specific snapshot for a virtual machine on an orchestrator host
func (s *OrchestratorService) DeleteHostVirtualMachineSnapshot(ctx basecontext.ApiContext, hostId string, vmId string, snapshotId string, noCache bool) error {
	host, vm, err := s.validateHostAndVM(ctx, hostId, vmId, noCache)
	if err != nil {
		return err
	}

	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(10 * time.Minute) // deleting snapshots can take a while
	path := "/machines/" + vm.ID + "/snapshots/" + snapshotId
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	_, err = httpClient.Delete(url.String(), nil)
	if err != nil {
		return err
	}

	return nil
}

// RevertHostVirtualMachineSnapshot reverts a virtual machine to a specific snapshot on an orchestrator host
func (s *OrchestratorService) RevertHostVirtualMachineSnapshot(ctx basecontext.ApiContext, hostId string, vmId string, snapshotId string, request apimodels.RevertVMSnapshotRequest, noCache bool) error {
	host, vm, err := s.validateHostAndVM(ctx, hostId, vmId, noCache)
	if err != nil {
		return err
	}

	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(10 * time.Minute) // reverting snapshots can take a while
	path := "/machines/" + vm.ID + "/snapshots/" + snapshotId + "/revert"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	_, err = httpClient.Post(url.String(), request, nil)
	if err != nil {
		return err
	}

	return nil
}
