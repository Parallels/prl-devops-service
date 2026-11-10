package orchestrator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// Names of the individual decision-factor checks reported by the pre-flight endpoint.
const (
	checkNameHostState      = "host_state"
	checkNameArchitecture   = "architecture"
	checkNameVmRequirements = "vm_requirements"
	checkNameCpu            = "cpu"
	checkNameMemory         = "memory"
	checkNameDisk           = "disk"
	checkNameMacosInventory = "macos_inventory"
	checkNameMacosCapacity  = "macos_capacity"
)

// unit constants shared with the pre-flight response payload.
const (
	checkUnitLogicalCpus = "logical_cpus"
	checkUnitMiB         = "MiB"
	checkUnitAppleVms    = "apple_vms"
)

// checkHostRequirements evaluates the scheduling decision factor for a single host and
// returns the outcome of every individual check, with required-vs-available detail.
//
// The evaluation short-circuits on blocking conditions (host state, architecture,
// missing specs) exactly like the original validateHost, returning a single failing
// check so validateHost can preserve its legacy single-message error responses.
func (s *OrchestratorService) checkHostRequirements(host data_models.OrchestratorHost, request models.CreateVirtualMachineRequest, specs *models.CreateVirtualMachineSpecs) []models.RequirementCheckResult {
	// Host state: enabled, healthy and reporting resources (legacy short-circuit order).
	if !host.Enabled {
		return []models.RequirementCheckResult{failedCheck(checkNameHostState, "Host is not enabled")}
	}
	if host.State != HealthyState {
		return []models.RequirementCheckResult{failedCheck(checkNameHostState, "Host is not healthy")}
	}
	if host.Resources == nil {
		return []models.RequirementCheckResult{failedCheck(checkNameHostState, "Host does not have resources information")}
	}
	stateCheck := models.RequirementCheckResult{
		Name:    checkNameHostState,
		Passed:  true,
		Message: "Host is enabled and healthy",
	}

	// Architecture, evaluated exactly like the legacy scheduler. Pre-flight callers
	// that impose no architecture constraint are matched per host by the caller so
	// this check passes trivially, and dropped from the reported result.
	if !strings.EqualFold(host.Architecture, request.Architecture) {
		return []models.RequirementCheckResult{failedCheck(checkNameArchitecture, "Host does not have the same architecture")}
	}
	archCheck := models.RequirementCheckResult{
		Name:    checkNameArchitecture,
		Passed:  true,
		Message: fmt.Sprintf("Host architecture %s matches requested architecture %s", host.Architecture, request.Architecture),
	}

	if specs == nil {
		return []models.RequirementCheckResult{failedCheck(checkNameVmRequirements, "VM resource requirements are unavailable")}
	}

	checks := []models.RequirementCheckResult{stateCheck, archCheck}

	// We will trust that the host has the reserved cpus setup correctly
	// otherwise we would potentially go above the reserved cpus
	availableCpus := host.Resources.TotalAvailable.LogicalCpuCount
	availableMemory := host.Resources.TotalAvailable.MemorySize

	requiredCpus := specs.GetCpuCount()
	cpuCheck := models.RequirementCheckResult{
		Name:      checkNameCpu,
		Passed:    availableCpus >= requiredCpus,
		Required:  models.NewRequirementCheckValue(float64(requiredCpus)),
		Available: models.NewRequirementCheckValue(float64(availableCpus)),
		Unit:      checkUnitLogicalCpus,
		Message:   fmt.Sprintf("CPU: required %d logical CPUs, available %d", requiredCpus, availableCpus),
	}
	checks = append(checks, cpuCheck)

	requiredMemory := specs.GetMemorySize()
	memoryCheck := models.RequirementCheckResult{
		Name:      checkNameMemory,
		Passed:    availableMemory >= requiredMemory,
		Required:  models.NewRequirementCheckValue(requiredMemory),
		Available: models.NewRequirementCheckValue(availableMemory),
		Unit:      checkUnitMiB,
		Message:   fmt.Sprintf("Memory: required %.0f MiB, available %.0f MiB", requiredMemory, availableMemory),
	}
	checks = append(checks, memoryCheck)

	if specs.Size > 0 {
		diskSpace, diskErr := s.getHostDiskSpace(s.ctx, host, request.Owner)
		if diskErr != nil {
			// Host may be running an older version that doesn't expose the disk-space
			// endpoint. Log a warning and skip the check so the host remains eligible.
			s.ctx.LogWarnf("[Orchestrator] Could not get disk space info for host %s (may be older version, skipping check): %v", host.Host, diskErr)
			checks = append(checks, models.RequirementCheckResult{
				Name:    checkNameDisk,
				Passed:  true,
				Message: fmt.Sprintf("Disk space information unavailable (host may be running an older version), check skipped: %v", diskErr),
			})
		} else {
			cacheFolder := ""
			if host.CacheConfig != nil {
				cacheFolder = host.CacheConfig.Folder
			}
			// Same volume: download + copy to cache + create = 3× size.
			// Different volumes: only 2× size needed across the two volumes.
			var requiredSpace int64
			if isSameVolume(diskSpace.PrlHomePath, cacheFolder) {
				requiredSpace = 3 * (specs.Size / 1024.0 / 1024.0) // Convert from bytes to MB
			} else {
				requiredSpace = 2 * (specs.Size / 1024.0 / 1024.0) // Convert from bytes to MB
			}
			diskCheck := models.RequirementCheckResult{
				Name:      checkNameDisk,
				Passed:    diskSpace.ParallelsHome >= requiredSpace,
				Required:  models.NewRequirementCheckValue(float64(requiredSpace)),
				Available: models.NewRequirementCheckValue(float64(diskSpace.ParallelsHome)),
				Unit:      checkUnitMiB,
				Message:   fmt.Sprintf("VM disk (%s): required %d MiB, available %d MiB (estimated download/copy space)", diskSpace.PrlHomePath, requiredSpace, diskSpace.ParallelsHome),
			}
			checks = append(checks, diskCheck)
		}
	}

	if strings.EqualFold(specs.Type, "macvm") {
		active, pending := host.Resources.TotalAppleVms, host.Resources.TotalReserved.TotalAppleVms
		s.ctx.LogInfof("[MacVMCapacity] Scheduling host=%s active=%d pending=%d maximum=%d", host.ID, active, pending, MaxNumberAppleVms)

		inventoryOk := host.Resources.MacVMInventoryComplete == nil || *host.Resources.MacVMInventoryComplete
		inventoryCheck := models.RequirementCheckResult{
			Name:   checkNameMacosInventory,
			Passed: inventoryOk,
		}
		if inventoryOk {
			inventoryCheck.Message = "macOS VM inventory is complete"
		} else {
			inventoryCheck.Message = "macOS VM inventory is incomplete; run the host service as root to enumerate all users"
		}
		checks = append(checks, inventoryCheck)

		capacityOk := active+pending < MaxNumberAppleVms
		checks = append(checks, models.RequirementCheckResult{
			Name:      checkNameMacosCapacity,
			Passed:    capacityOk,
			Required:  models.NewRequirementCheckValue(MaxNumberAppleVms),
			Available: models.NewRequirementCheckValue(float64(active + pending)),
			Unit:      checkUnitAppleVms,
			Message:   fmt.Sprintf("Apple VM limit: current %d, pending %d, maximum %d", active, pending, MaxNumberAppleVms),
		})
	}

	return checks
}

// CheckMachineRequirements runs the scheduling decision factor against registered hosts
// for a catalog image or an explicit set of requirements, and reports what is missing.
// When any host passes every check the response is slim; otherwise every evaluated host
// is returned with its full per-check breakdown.
func (s *OrchestratorService) CheckMachineRequirements(ctx basecontext.ApiContext, request models.CheckMachineRequirementsRequest) (*models.CheckMachineRequirementsResponse, *models.ApiErrorResponse) {
	specs, specsError := s.getCheckRequirementsSpecs(request)
	if specsError != nil {
		return nil, specsError
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, &models.ApiErrorResponse{Message: "There was an error getting the database", Code: 500}
	}

	var hosts []data_models.OrchestratorHost
	if strings.TrimSpace(request.HostId) != "" {
		host, hostErr := dbService.GetOrchestratorHost(ctx, request.HostId)
		if hostErr != nil {
			apiError := models.NewFromError(hostErr)
			return nil, &apiError
		}
		hosts = []data_models.OrchestratorHost{*host}
	} else {
		allHosts, hostsErr := dbService.GetOrchestratorHosts(ctx, "")
		if hostsErr != nil {
			return nil, &models.ApiErrorResponse{Message: "There was an error getting the hosts from the database", Code: 500}
		}
		hosts = allHosts
	}

	response := &models.CheckMachineRequirementsResponse{Requirements: resolvedRequirements(specs, request)}

	archUnspecified := strings.TrimSpace(request.Architecture) == ""
	failedHosts := make([]models.MachineRequirementsHostResult, 0, len(hosts))
	for _, host := range hosts {
		vmRequest := models.CreateVirtualMachineRequest{Architecture: request.Architecture, Owner: request.Owner}
		if archUnspecified {
			// No architecture constraint requested: match the host so the decision
			// factor keeps its semantics, and drop the trivial pass from the report.
			vmRequest.Architecture = host.Architecture
		}
		checks := s.checkHostRequirements(host, vmRequest, specs)
		if archUnspecified {
			checks = dropRequirementCheck(checks, checkNameArchitecture)
		}
		result := models.MachineRequirementsHostResult{
			HostId: host.ID,
			Host:   host.GetHost(),
			Passed: true,
			Checks: checks,
		}
		for _, check := range checks {
			if !check.Passed {
				result.Passed = false
				break
			}
		}
		// First fully passing host settles the verdict: the answer is slim and the
		// remaining hosts (and their disk probes) are skipped.
		if result.Passed {
			response.Passed = true
			return response, nil
		}
		failedHosts = append(failedHosts, result)
	}

	if len(hosts) == 0 {
		response.Message = "No hosts registered"
		return response, nil
	}
	response.Message = "No host meets the requirements"
	response.Hosts = failedHosts
	return response, nil
}

// getCheckRequirementsSpecs resolves the specs to evaluate from either the catalog
// manifest (same resolution path as VM creation) or the explicit request requirements.
func (s *OrchestratorService) getCheckRequirementsSpecs(request models.CheckMachineRequirementsRequest) (*models.CreateVirtualMachineSpecs, *models.ApiErrorResponse) {
	if request.Requirements != nil {
		requirements := request.Requirements
		if requirements.Cpu <= 0 {
			requirements.Cpu = 2
		}
		if requirements.Memory <= 0 {
			requirements.Memory = 2048
		}
		return &models.CreateVirtualMachineSpecs{
			Type:   requirements.GetSpecsType(),
			Cpu:    strconv.Itoa(requirements.Cpu),
			Memory: strconv.Itoa(requirements.Memory),
			Disk:   strconv.Itoa(requirements.Disk),
			Size:   requirements.Size,
		}, nil
	}

	version := request.Version
	if version == "" {
		version = constants.LATEST_TAG
	}
	createRequest := models.CreateVirtualMachineRequest{
		Architecture: request.Architecture,
		Owner:        request.Owner,
		CatalogManifest: &models.CreateCatalogVirtualMachineRequest{
			CatalogId:    request.CatalogId,
			Version:      version,
			Architecture: request.Architecture,
			Connection:   request.Connection,
			Owner:        request.Owner,
		},
	}
	specs, err := s.getSpecsFromRequest(createRequest)
	if err != nil {
		// Catalog lookup failures and requirement errors surface as 400s, matching
		// the VM creation path; GetSystemErrorCode defaults non-coded errors to 400.
		return nil, &models.ApiErrorResponse{Message: err.Error(), Code: models.GetSystemErrorCode(err)}
	}
	return specs, nil
}

// resolvedRequirements echoes the effective requirements back to the caller.
func resolvedRequirements(specs *models.CreateVirtualMachineSpecs, request models.CheckMachineRequirementsRequest) *models.MachineRequirements {
	if specs == nil {
		return nil
	}
	cpu, _ := strconv.Atoi(specs.Cpu)
	memory, _ := strconv.Atoi(specs.Memory)
	disk, _ := strconv.Atoi(specs.Disk)
	requirements := &models.MachineRequirements{
		Type:   specs.Type,
		Cpu:    cpu,
		Memory: memory,
		Disk:   disk,
		Size:   specs.Size,
	}
	if request.Requirements != nil {
		requirements.Os = request.Requirements.Os
	}
	return requirements
}

func failedCheck(name, message string) models.RequirementCheckResult {
	return models.RequirementCheckResult{Name: name, Passed: false, Message: message}
}

// dropRequirementCheck removes the named check from the result list.
func dropRequirementCheck(checks []models.RequirementCheckResult, name string) []models.RequirementCheckResult {
	filtered := make([]models.RequirementCheckResult, 0, len(checks))
	for _, check := range checks {
		if check.Name == name {
			continue
		}
		filtered = append(filtered, check)
	}
	return filtered
}

// isBlockingRequirementCheck reports whether the check list is a legacy short-circuit:
// a single failing host_state/architecture/vm_requirements check that must be reported
// verbatim instead of inside the aggregated "Insufficient host resources" message.
func isBlockingRequirementCheck(checks []models.RequirementCheckResult) bool {
	if len(checks) != 1 || checks[0].Passed {
		return false
	}
	switch checks[0].Name {
	case checkNameHostState, checkNameArchitecture, checkNameVmRequirements:
		return true
	}
	return false
}
