package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	dbmodels "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/require"
)

func TestValidateHostReportsAllResourceShortages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.DiskSpaceAvailable{ParallelsHome: 100, PrlHomePath: "/Users/test/Parallels"})
	}))
	defer server.Close()
	svc := &OrchestratorService{ctx: basecontext.NewRootBaseContext()}
	host := dbmodels.OrchestratorHost{Host: server.URL, Enabled: true, State: "healthy", Architecture: "arm64",
		Resources: &dbmodels.HostResources{TotalAppleVms: MaxNumberAppleVms, TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 2, MemorySize: 1024}}}
	request := models.CreateVirtualMachineRequest{Architecture: "arm64"}
	ok, failure := svc.validateHost(host, request, &models.CreateVirtualMachineSpecs{Type: "macvm", Cpu: "8", Memory: "8192", Size: 100 * 1024 * 1024})
	require.False(t, ok)
	require.NotNil(t, failure)
	for _, detail := range []string{"CPU: required 8 logical CPUs, available 2", "Memory: required 8192 MiB, available 1024 MiB", "VM disk (/Users/test/Parallels): required 300 MiB, available 100 MiB", "Apple VM limit:"} {
		require.Contains(t, failure.Message, detail)
	}
	// Exact capacity must remain eligible.
	ok, failure = svc.validateHost(host, request, &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "1024"})
	require.True(t, ok)
	require.Nil(t, failure)
}

func TestHostSelectionPreservesResourceFailuresInJobs(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()
	db := data.NewJsonDatabase(ctx, filepath.Join(t.TempDir(), "resource-failures.json"))
	serviceprovider.NewMockProvider().JsonDatabase = db
	manager := jobs.New(ctx)
	svc := &OrchestratorService{ctx: ctx}
	for _, name := range []string{"resource-host-a", "resource-host-b"} {
		host, err := db.CreateOrchestratorHost(ctx, dbmodels.OrchestratorHost{Host: name, State: "healthy", Architecture: "arm64", Resources: &dbmodels.HostResources{}})
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.DeleteOrchestratorHost(ctx, host.ID) })
	}
	for _, async := range []bool{false, true} {
		t.Run(fmt.Sprintf("async=%v", async), func(t *testing.T) {
			job, err := manager.CreateNewJob("test-user", "orchestrator", "create", "Creating VM")
			require.NoError(t, err)
			var failure *models.ApiErrorResponse
			if async {
				_, failure = svc.DispatchCreateVirtualMachine(ctx, job.ID, models.CreateVirtualMachineRequest{Architecture: "arm64"})
			} else {
				_, failure = svc.CreateVirtualMachine(ctx, job.ID, models.CreateVirtualMachineRequest{Architecture: "arm64"})
			}
			require.NotNil(t, failure)
			for _, detail := range []string{"resource-host-a", "resource-host-b", "CPU: required 2 logical CPUs, available 0", "Memory: required 2048 MiB, available 0 MiB"} {
				require.Contains(t, failure.Message, detail)
			}
			// Controllers pass the returned message to MarkJobError; polling and
			// event delivery both use this same API mapper.
			require.NoError(t, manager.MarkJobError(job.ID, fmt.Errorf("%s", failure.Message)))
			stored, err := db.GetJob(ctx, job.ID)
			require.NoError(t, err)
			response := mappers.MapJobToApiJob(*stored)
			require.Equal(t, constants.JobStateFailed, response.State)
			require.Equal(t, failure.Message, response.Error)
			require.Equal(t, failure.Message, response.Message)
		})
	}
}

func TestCatalogLookupFailureDoesNotDefaultToNonMacVM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer server.Close()
	svc := &OrchestratorService{ctx: basecontext.NewRootBaseContext()}
	specs, err := svc.getSpecsFromRequest(models.CreateVirtualMachineRequest{Architecture: "arm64", CatalogManifest: &models.CreateCatalogVirtualMachineRequest{Connection: server.URL, CatalogId: "macos", Version: "latest"}})
	require.Error(t, err)
	require.Nil(t, specs)
	require.Contains(t, err.Error(), "Cannot determine VM requirements")
}

func TestValidateHostRejectsPendingMacStartsAndIncompleteInventory(t *testing.T) {
	svc := &OrchestratorService{ctx: basecontext.NewRootBaseContext()}
	complete := false
	host := dbmodels.OrchestratorHost{Enabled: true, State: "healthy", Architecture: "arm64", Resources: &dbmodels.HostResources{TotalAppleVms: 1, MacVMInventoryComplete: &complete, TotalReserved: dbmodels.HostResourceItem{TotalAppleVms: 1}, TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}}}
	ok, failure := svc.validateHost(host, models.CreateVirtualMachineRequest{Architecture: "arm64"}, &models.CreateVirtualMachineSpecs{Type: "macvm", Cpu: "2", Memory: "2048"})
	require.False(t, ok)
	require.Contains(t, failure.Message, "inventory is incomplete")
	require.Contains(t, failure.Message, "pending 1")
}
