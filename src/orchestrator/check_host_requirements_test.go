package orchestrator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	dbmodels "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/require"
)

func newCheckTestContext() basecontext.ApiContext {
	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()
	return ctx
}

// newCheckTestDatabase points the (singleton) JSON database at a scratch file.
// The database keeps background writers alive, so the scratch directory is
// removed best-effort instead of through t.TempDir's strict cleanup.
func newCheckTestDatabase(t *testing.T) *data.JsonDatabase {
	dir, err := os.MkdirTemp("", "devops-requirements-check-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	db := data.NewJsonDatabase(newCheckTestContext(), filepath.Join(dir, "requirements-check.json"))
	serviceprovider.NewMockProvider().JsonDatabase = db
	return db
}

func createCheckHost(t *testing.T, db *data.JsonDatabase, name string, resources *dbmodels.HostResources) *dbmodels.OrchestratorHost {
	ctx := newCheckTestContext()
	host, err := db.CreateOrchestratorHost(ctx, dbmodels.OrchestratorHost{
		Host:         name,
		Enabled:      true,
		State:        "healthy",
		Architecture: "arm64",
		Resources:    resources,
	})
	require.NoError(t, err)
	// Tests share the database singleton; leave no residue behind.
	t.Cleanup(func() { _ = db.DeleteOrchestratorHost(newCheckTestContext(), host.ID) })
	return host
}

func requirementsCheckByName(result models.MachineRequirementsHostResult, name string) (models.RequirementCheckResult, bool) {
	for _, check := range result.Checks {
		if check.Name == name {
			return check, true
		}
	}
	return models.RequirementCheckResult{}, false
}

func TestCheckHostRequirementsReportsStructuredShortage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.DiskSpaceAvailable{ParallelsHome: 100, PrlHomePath: "/Users/test/Parallels"})
	}))
	defer server.Close()

	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := dbmodels.OrchestratorHost{Host: server.URL, Enabled: true, State: "healthy", Architecture: "arm64",
		Resources: &dbmodels.HostResources{TotalAppleVms: MaxNumberAppleVms, TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 2, MemorySize: 1024}}}
	request := models.CreateVirtualMachineRequest{Architecture: "arm64"}
	specs := &models.CreateVirtualMachineSpecs{Type: "macvm", Cpu: "8", Memory: "8192", Size: 100 * 1024 * 1024}

	checks := svc.checkHostRequirements(host, request, specs)

	hostState, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameHostState)
	require.True(t, ok)
	require.True(t, hostState.Passed)

	cpu, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameCpu)
	require.True(t, ok)
	require.False(t, cpu.Passed)
	require.EqualValues(t, 8, *cpu.Required)
	require.EqualValues(t, 2, *cpu.Available)
	require.Equal(t, checkUnitLogicalCpus, cpu.Unit)

	memory, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameMemory)
	require.True(t, ok)
	require.False(t, memory.Passed)
	require.EqualValues(t, 8192, *memory.Required)
	require.EqualValues(t, 1024, *memory.Available)
	require.Equal(t, checkUnitMiB, memory.Unit)

	disk, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameDisk)
	require.True(t, ok)
	require.False(t, disk.Passed)
	require.EqualValues(t, 300, *disk.Required)
	require.EqualValues(t, 100, *disk.Available)

	inventory, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameMacosInventory)
	require.True(t, ok)
	require.True(t, inventory.Passed)

	capacity, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameMacosCapacity)
	require.True(t, ok)
	require.False(t, capacity.Passed)
	require.EqualValues(t, MaxNumberAppleVms, *capacity.Required)
	require.EqualValues(t, MaxNumberAppleVms, *capacity.Available)

	// The legacy aggregated message is preserved verbatim for the scheduling path.
	ok, failure := svc.validateHost(host, request, specs)
	require.False(t, ok)
	require.Equal(t, "Insufficient host resources: CPU: required 8 logical CPUs, available 2; "+
		"Memory: required 8192 MiB, available 1024 MiB; "+
		"VM disk (/Users/test/Parallels): required 300 MiB, available 100 MiB (estimated download/copy space); "+
		"Apple VM limit: current 2, pending 0, maximum 2", failure.Message)
}

func TestCheckHostRequirementsShortCircuitsStayVerbatim(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	request := models.CreateVirtualMachineRequest{Architecture: "arm64"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"}
	resources := &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}}

	blocked := dbmodels.OrchestratorHost{Enabled: false, State: "healthy", Architecture: "arm64", Resources: resources}
	ok, failure := svc.validateHost(blocked, request, specs)
	require.False(t, ok)
	require.Equal(t, "Host is not enabled", failure.Message)

	mismatched := dbmodels.OrchestratorHost{Enabled: true, State: "healthy", Architecture: "amd64", Resources: resources}
	ok, failure = svc.validateHost(mismatched, request, specs)
	require.False(t, ok)
	require.Equal(t, "Host does not have the same architecture", failure.Message)

	healthy := dbmodels.OrchestratorHost{Enabled: true, State: "healthy", Architecture: "arm64", Resources: resources}
	ok, failure = svc.validateHost(healthy, request, nil)
	require.False(t, ok)
	require.Equal(t, "VM resource requirements are unavailable", failure.Message)
}

func TestCheckMachineRequirementsFleetPassesWithSlimResponse(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	createCheckHost(t, db, "busy-host", &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 1, MemorySize: 512}})
	createCheckHost(t, db, "roomy-host", &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 16384}})

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		Requirements: &models.MachineRequirements{Cpu: 4, Memory: 4096, Os: "linux"},
	})
	require.Nil(t, apiError)
	require.True(t, response.Passed)
	require.Empty(t, response.Hosts)
	require.Equal(t, "pvm", response.Requirements.Type)
	require.Equal(t, 4, response.Requirements.Cpu)
	require.Equal(t, 4096, response.Requirements.Memory)

	// Success payload carries no host details at all.
	payload, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotContains(t, string(payload), `"hosts"`)
	require.Contains(t, string(payload), `"passed":true`)
}

func TestCheckMachineRequirementsFailureDetailsEveryHost(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	createCheckHost(t, db, "small-host", &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 2, MemorySize: 1024}})
	createCheckHost(t, db, "mac-host-at-limit", &dbmodels.HostResources{
		TotalAppleVms:  MaxNumberAppleVms,
		TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 4, MemorySize: 8192},
	})

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		Requirements: &models.MachineRequirements{Cpu: 8, Memory: 4096, Os: "macos"},
	})
	require.Nil(t, apiError)
	require.False(t, response.Passed)
	require.Equal(t, "No host meets the requirements", response.Message)
	require.Len(t, response.Hosts, 2)
	require.Equal(t, "macvm", response.Requirements.Type)

	for _, result := range response.Hosts {
		cpu, ok := requirementsCheckByName(result, checkNameCpu)
		require.True(t, ok, "host %s must carry a cpu check", result.Host)
		require.False(t, cpu.Passed)
		require.EqualValues(t, 8, *cpu.Required)

		capacity, ok := requirementsCheckByName(result, checkNameMacosCapacity)
		require.True(t, ok, "macos requirements must evaluate Apple VM capacity on host %s", result.Host)
		if result.Host == "mac-host-at-limit" {
			require.False(t, capacity.Passed)
			require.EqualValues(t, MaxNumberAppleVms, *capacity.Available)
		}
		// No architecture was requested: the check must not appear.
		_, ok = requirementsCheckByName(result, checkNameArchitecture)
		require.False(t, ok)
	}

	payload, err := json.Marshal(response)
	require.NoError(t, err)
	require.Contains(t, string(payload), `"hosts"`)
}

func TestCheckMachineRequirementsArchitectureMismatchIsReported(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	createCheckHost(t, db, "arm-host", &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}})

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		Architecture: "amd64",
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
	})
	require.Nil(t, apiError)
	require.False(t, response.Passed)
	require.Len(t, response.Hosts, 1)
	require.Len(t, response.Hosts[0].Checks, 1)
	require.Equal(t, checkNameArchitecture, response.Hosts[0].Checks[0].Name)
	require.Equal(t, "Host does not have the same architecture", response.Hosts[0].Checks[0].Message)
}

func TestCheckMachineRequirementsHostIdFilter(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	roomy := createCheckHost(t, db, "roomy-host", &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 16384}})

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		HostId:       roomy.ID,
		Architecture: "arm64",
		Requirements: &models.MachineRequirements{Cpu: 4, Memory: 4096},
	})
	require.Nil(t, apiError)
	require.True(t, response.Passed)
	require.Empty(t, response.Hosts)

	_, apiError = svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		HostId:       "missing-host",
		Architecture: "arm64",
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
	})
	require.NotNil(t, apiError)
	require.Equal(t, 404, apiError.Code)
}

func TestCheckMachineRequirementsCatalogModeResolvesManifest(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	_, err := db.CreateCatalogManifest(ctx, dbmodels.CatalogManifest{
		CatalogId:    "check-catalog",
		Name:         "check-catalog",
		Version:      "latest",
		Architecture: "arm64",
		Type:         "pvm",
		MinimumSpecRequirements: &dbmodels.MinimumSpecRequirement{
			Cpu: 4, Memory: 4096, Disk: 20480,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.DeleteCatalogManifest(ctx, "check-catalog") })
	createCheckHost(t, db, "roomy-host", &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 16384}})

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		CatalogId:    "check-catalog",
		Version:      "latest",
		Architecture: "arm64",
	})
	require.Nil(t, apiError)
	require.True(t, response.Passed)
	require.Equal(t, "pvm", response.Requirements.Type)
	require.Equal(t, 4, response.Requirements.Cpu)
	require.Equal(t, 4096, response.Requirements.Memory)
	require.Equal(t, 20480, response.Requirements.Disk)
}

func TestCheckMachineRequirementsCatalogLatestTagResolves(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	// Manifest stored with a literal version; latest only exists as a tag.
	_, err := db.CreateCatalogManifest(ctx, dbmodels.CatalogManifest{
		CatalogId:    "tagged-check-catalog",
		Name:         "tagged-check-catalog",
		Version:      "2.1.0",
		Architecture: "arm64",
		Type:         "pvm",
		MinimumSpecRequirements: &dbmodels.MinimumSpecRequirement{
			Cpu: 2, Memory: 2048, Disk: 10240,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.DeleteCatalogManifest(ctx, "tagged-check-catalog") })
	createCheckHost(t, db, "roomy-host", &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 16384}})

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		CatalogId:    "tagged-check-catalog",
		Version:      "latest",
		Architecture: "arm64",
	})
	require.Nil(t, apiError)
	require.True(t, response.Passed)
	require.Equal(t, 2, response.Requirements.Cpu)
	require.Equal(t, 2048, response.Requirements.Memory)
}

func TestCheckMachineRequirementsCatalogLookupFailureReturnsBadRequest(t *testing.T) {
	newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer server.Close()

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		CatalogId:    "macos",
		Version:      "latest",
		Architecture: "arm64",
		Connection:   server.URL,
	})
	require.Nil(t, response)
	require.NotNil(t, apiError)
	require.Equal(t, 400, apiError.Code)
	require.Contains(t, apiError.Message, "Cannot determine VM requirements")
}

func TestCheckMachineRequirementsWithoutHosts(t *testing.T) {
	newCheckTestDatabase(t)
	ctx := newCheckTestContext()

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
	})
	require.Nil(t, apiError)
	require.False(t, response.Passed)
	require.Equal(t, "No hosts registered", response.Message)
	require.Empty(t, response.Hosts)
}

func TestCheckMachineRequirementsOldHostDiskCheckIsSkipped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) }))
	defer server.Close()

	svc := &OrchestratorService{ctx: newCheckTestContext()}
	// Direct decision-factor evaluation with an unreachable disk endpoint.
	host := dbmodels.OrchestratorHost{Host: server.URL, Enabled: true, State: "healthy", Architecture: "arm64",
		Resources: &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}}}
	checks := svc.checkHostRequirements(host, models.CreateVirtualMachineRequest{Architecture: "arm64"},
		&models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048", Size: 1024 * 1024 * 1024})

	disk, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameDisk)
	require.True(t, ok)
	require.True(t, disk.Passed)
	require.Contains(t, disk.Message, "check skipped")
}

// A host at its configured VM count limit is ineligible even though CPU, memory and
// disk are all free: that is the whole point of MAX_VMS_PER_HOST.
func TestCheckHostRequirementsVmCountLimitBlocksPlacement(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	request := models.CreateVirtualMachineRequest{Architecture: "arm64"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"}
	resources := &dbmodels.HostResources{
		MaxVmsPerHost:  3,
		TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 64, MemorySize: 131072},
		Total:          dbmodels.HostResourceItem{TotalVms: 3},
	}
	host := dbmodels.OrchestratorHost{Enabled: true, State: "healthy", Architecture: "arm64", Resources: resources}

	checks := svc.checkHostRequirements(host, request, specs)
	vmCount, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameVmCount)
	require.True(t, ok, "a host reporting a limit must carry a vm_count check")
	require.False(t, vmCount.Passed)
	require.Equal(t, checkUnitVms, vmCount.Unit)
	require.EqualValues(t, 3, *vmCount.Required)
	require.EqualValues(t, 3, *vmCount.Available)
	require.Equal(t, "VM limit: current 3, pending 0, maximum 3", vmCount.Message)

	// Creates already in flight consume the last slot.
	resources.Total.TotalVms = 2
	resources.TotalReserved.TotalVms = 1
	checks = svc.checkHostRequirements(host, request, specs)
	vmCount, ok = requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameVmCount)
	require.True(t, ok)
	require.False(t, vmCount.Passed)
	require.Equal(t, "VM limit: current 2, pending 1, maximum 3", vmCount.Message)

	// The shortage keeps the legacy aggregated message: no new error contract.
	ok, failure := svc.validateHost(host, request, specs)
	require.False(t, ok)
	require.Equal(t, "Insufficient host resources: VM limit: current 2, pending 1, maximum 3", failure.Message)
}

// "running" counts the VMs occupying a hypervisor slot, so a host full of stopped VMs
// still has room; "any" counts every registered VM.
func TestCheckHostRequirementsVmCountModeSelectsCountedSet(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	request := models.CreateVirtualMachineRequest{Architecture: "arm64"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"}
	host := dbmodels.OrchestratorHost{Enabled: true, State: "healthy", Architecture: "arm64",
		Resources: &dbmodels.HostResources{
			MaxVmsPerHost:     2,
			MaxVmsPerHostMode: "running",
			TotalAvailable:    dbmodels.HostResourceItem{LogicalCpuCount: 64, MemorySize: 131072},
			Total:             dbmodels.HostResourceItem{TotalVms: 40},
			TotalInUse:        dbmodels.HostResourceItem{TotalVms: 1},
		}}

	checks := svc.checkHostRequirements(host, request, specs)
	vmCount, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameVmCount)
	require.True(t, ok)
	require.True(t, vmCount.Passed, "stopped VMs are free capacity in running mode")
	require.Equal(t, "VM limit: current 1, pending 0, maximum 2", vmCount.Message)

	host.Resources.TotalInUse.TotalVms = 2
	checks = svc.checkHostRequirements(host, request, specs)
	vmCount, _ = requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameVmCount)
	require.False(t, vmCount.Passed)

	host.Resources.MaxVmsPerHostMode = ""
	checks = svc.checkHostRequirements(host, request, specs)
	vmCount, _ = requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameVmCount)
	require.False(t, vmCount.Passed, "the default any mode counts all 40 registered VMs")
	require.Equal(t, "VM limit: current 40, pending 0, maximum 2", vmCount.Message)
}

// Hosts that report no limit — unset, or an older host build without the field — must
// not gain a check, which is what keeps existing behaviour byte-identical.
func TestCheckHostRequirementsWithoutVmLimitOmitsCheck(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := dbmodels.OrchestratorHost{Enabled: true, State: "healthy", Architecture: "arm64",
		Resources: &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}}}

	checks := svc.checkHostRequirements(host, models.CreateVirtualMachineRequest{Architecture: "arm64"},
		&models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"})

	_, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameVmCount)
	require.False(t, ok)

	ok, failure := svc.validateHost(host, models.CreateVirtualMachineRequest{Architecture: "arm64"},
		&models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"})
	require.True(t, ok)
	require.Nil(t, failure)
}

// Placement across a fleet must skip the host that is full by VM count and pick the
// one that has room, without a CPU or memory shortage anywhere.
func TestCheckMachineRequirementsSkipsHostAtVmCountLimit(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	full := createCheckHost(t, db, "full-host", &dbmodels.HostResources{
		MaxVmsPerHost:  2,
		TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 16, MemorySize: 32768},
		Total:          dbmodels.HostResourceItem{TotalVms: 2},
	})
	createCheckHost(t, db, "roomy-host", &dbmodels.HostResources{
		MaxVmsPerHost:  20,
		TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 16, MemorySize: 32768},
		Total:          dbmodels.HostResourceItem{TotalVms: 2},
	})
	require.NotNil(t, full)

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
	})
	require.Nil(t, apiError)
	require.True(t, response.Passed, "the host below its VM limit must carry the placement")

	// With the fleet reduced to the full host the pre-flight reports why.
	fullHost, err := db.GetOrchestratorHost(ctx, full.ID)
	require.NoError(t, err)
	fullHost.Enabled = true
	response, apiError = svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		HostId:       fullHost.ID,
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
	})
	require.Nil(t, apiError)
	require.False(t, response.Passed)
	require.Len(t, response.Hosts, 1)
	vmCount, ok := requirementsCheckByName(response.Hosts[0], checkNameVmCount)
	require.True(t, ok)
	require.False(t, vmCount.Passed)
}

// createCheckHostWithTags registers a host carrying the given placement tags.
func createCheckHostWithTags(t *testing.T, db *data.JsonDatabase, name string, tags []string, resources *dbmodels.HostResources) *dbmodels.OrchestratorHost {
	ctx := newCheckTestContext()
	host, err := db.CreateOrchestratorHost(ctx, dbmodels.OrchestratorHost{
		Host:         name,
		Enabled:      true,
		State:        "healthy",
		Architecture: "arm64",
		Tags:         tags,
		Resources:    resources,
	})
	require.NoError(t, err)
	// Tests share the database singleton; leave no residue behind.
	t.Cleanup(func() { _ = db.DeleteOrchestratorHost(newCheckTestContext(), host.ID) })
	return host
}

// target_tags is a hard AND-constraint evaluated before any other decision
// factor: only a host carrying every tag — here the one that is otherwise
// capacity-eligible as well — keeps the fleet check passing.
func TestCheckMachineRequirementsFleetHonorTargetTags(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	rich := &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 16, MemorySize: 32768}}
	createCheckHostWithTags(t, db, "test-host", []string{"pool=test"}, rich)
	createCheckHostWithTags(t, db, "build-host", []string{"pool=build", "OS=Sonoma"}, rich)

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
		TargetTags:   []string{"POOL=BUILD", "os=sonoma"}, // case-insensitive AND match
	})
	require.Nil(t, apiError)
	require.True(t, response.Passed, "the host carrying both tags must carry the placement")
	require.Empty(t, response.Hosts, "a passing verdict stays slim")
}

// With no tag match anywhere the verdict fails and every evaluated host
// reports the host_tags check naming exactly what it was missing.
func TestCheckMachineRequirementsFleetWithoutTagMatchReportsHostTags(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	rich := &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 16, MemorySize: 32768}}
	build := createCheckHostWithTags(t, db, "build-host", []string{"pool=build"}, rich)
	test := createCheckHostWithTags(t, db, "test-host", nil, rich)

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
		TargetTags:   []string{"pool=gold", "os=sonoma"},
	})
	require.Nil(t, apiError)
	require.False(t, response.Passed)
	require.Equal(t, "No host meets the requirements", response.Message)
	require.Len(t, response.Hosts, 2)

	byId := map[string]models.MachineRequirementsHostResult{}
	for _, h := range response.Hosts {
		byId[h.HostId] = h
	}
	buildResult, ok := byId[build.ID]
	require.True(t, ok)
	require.Len(t, buildResult.Checks, 1, "a tag short-circuit reports only the tag failure")
	require.Equal(t, checkNameHostTags, buildResult.Checks[0].Name)
	require.False(t, buildResult.Checks[0].Passed)
	require.Equal(t, "Host is missing required tags: pool=gold, os=sonoma", buildResult.Checks[0].Message)

	testResult, ok := byId[test.ID]
	require.True(t, ok)
	require.Equal(t, "Host is missing required tags: pool=gold, os=sonoma", testResult.Checks[0].Message)
}

// An explicit host that lacks a required tag fails with the tag breakdown —
// naming a host does not bypass the placement constraint.
func TestCheckMachineRequirementsExplicitHostTagMismatch(t *testing.T) {
	db := newCheckTestDatabase(t)
	ctx := newCheckTestContext()
	rich := &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 16, MemorySize: 32768}}
	host := createCheckHostWithTags(t, db, "test-host", []string{"pool=test"}, rich)

	svc := &OrchestratorService{ctx: ctx}
	response, apiError := svc.CheckMachineRequirements(ctx, models.CheckMachineRequirementsRequest{
		HostId:       host.ID,
		Requirements: &models.MachineRequirements{Cpu: 2, Memory: 2048},
		TargetTags:   []string{"pool=build"},
	})
	require.Nil(t, apiError)
	require.False(t, response.Passed)
	require.Len(t, response.Hosts, 1)
	require.Equal(t, "Host is missing required tags: pool=build", response.Hosts[0].Checks[0].Message)
}

// The hard tag check runs before the disk-space probe and short-circuits: a
// mismatched host never reaches the network.
func TestCheckHostRequirementsHostTagsShortCircuitBeforeDiskProbe(t *testing.T) {
	probes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probes++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.DiskSpaceAvailable{ParallelsHome: 100, PrlHomePath: "/Users/test/Parallels"})
	}))
	defer server.Close()

	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := dbmodels.OrchestratorHost{Host: server.URL, Enabled: true, State: "healthy", Architecture: "arm64",
		Tags:      []string{"pool=test"},
		Resources: &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}}}
	request := models.CreateVirtualMachineRequest{Architecture: "arm64", TargetTags: []string{"pool=build", "os=sonoma"}}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048", Size: 1024}

	checks := svc.checkHostRequirements(host, request, specs)
	require.Len(t, checks, 1)
	require.Equal(t, checkNameHostTags, checks[0].Name)
	require.False(t, checks[0].Passed)
	require.Zero(t, probes, "tag-mismatched hosts are rejected before the disk probe")

	// The scheduling path reports the blocking failure verbatim, not wrapped
	// in the aggregated "Insufficient host resources" message.
	ok, failure := svc.validateHost(host, request, specs)
	require.False(t, ok)
	require.NotNil(t, failure)
	require.Equal(t, "Host is missing required tags: pool=build, os=sonoma", failure.Message)
}

// A host carrying every requested tag (case-insensitively) gains a passing
// host_tags check and stays eligible for the remaining decision factors.
func TestCheckHostRequirementsHostTagsPassReported(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := dbmodels.OrchestratorHost{Host: "tagged.local", Enabled: true, State: "healthy", Architecture: "arm64",
		Tags:      []string{"pool=build", "OS=Sonoma ", "lab"},
		Resources: &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}}}
	request := models.CreateVirtualMachineRequest{Architecture: "arm64", TargetTags: []string{"POOL=BUILD", "os=sonoma"}}

	checks := svc.checkHostRequirements(host, request, &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"})

	require.Equal(t, checkNameHostTags, checks[0].Name, "the tag check leads the breakdown")
	require.True(t, checks[0].Passed)
	require.Contains(t, checks[0].Message, "all 2 required tags")

	ok, failure := svc.validateHost(host, request, &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"})
	require.True(t, ok)
	require.Nil(t, failure)
}

// Without target_tags the check list must not gain a host_tags entry — the
// response shape stays byte-identical for requests that don't use the feature.
func TestCheckHostRequirementsWithoutTargetTagsOmitsHostTagsCheck(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := dbmodels.OrchestratorHost{Host: "plain.local", Enabled: true, State: "healthy", Architecture: "arm64",
		Tags:      []string{"pool=build"},
		Resources: &dbmodels.HostResources{TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 8, MemorySize: 8192}}}

	checks := svc.checkHostRequirements(host, models.CreateVirtualMachineRequest{Architecture: "arm64"},
		&models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "2048"})

	_, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, checkNameHostTags)
	require.False(t, ok)
}
