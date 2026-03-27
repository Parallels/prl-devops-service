package orchestrator

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/registry"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) CreateVirtualMachine(ctx basecontext.ApiContext, jobID string, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, *models.ApiErrorResponse) {
	var apiError *models.ApiErrorResponse

	jobManager := jobs.Get(ctx)
	updateJob := func(msg string) {
		if jobID != "" && jobManager != nil {
			_, _ = jobManager.UpdateJobMessage(jobID, msg)
		}
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		apiError = &models.ApiErrorResponse{
			Message: "There was an error getting the database",
			Code:    500,
		}
		updateJob(apiError.Message)
		return nil, apiError
	}

	specs := s.getSpecsFromRequest(request)

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		apiError = &models.ApiErrorResponse{
			Message: "There was an error getting the hosts from the database",
			Code:    500,
		}
		updateJob(apiError.Message)
		return nil, apiError
	}

	var validHosts []data_models.OrchestratorHost
	for _, orchestratorHost := range hosts {
		isOk, validateErr := s.validateHost(orchestratorHost, request, specs)
		if validateErr != nil || !isOk {
			msg := fmt.Sprintf("Host %s skipped", orchestratorHost.Host)
			if validateErr != nil {
				msg = fmt.Sprintf("Host %s skipped: %s", orchestratorHost.Host, validateErr.Message)
			}
			ctx.LogInfof("[Orchestrator] %s", msg)
			updateJob(msg)
			continue
		}
		validHosts = append(validHosts, orchestratorHost)
	}

	if len(validHosts) == 0 {
		apiError = &models.ApiErrorResponse{
			Message: "No host available to create the virtual machine",
			Code:    400,
		}
		updateJob(apiError.Message)
		return nil, apiError
	}

	validHosts, filterErr := filterAndSortHosts(validHosts, request, s.pingHostForLatency)
	if filterErr != nil {
		updateJob(filterErr.Message)
		return nil, filterErr
	}

	// Stage 5: Target Execution — dispatch async to the first willing host.
	reg := registry.Get()
	for _, host := range validHosts {
		updateJob(fmt.Sprintf("Dispatching to host %s", host.Host))
		hostJob, err := s.CallCreateHostVirtualMachineAsync(host, request)
		if err != nil {
			e := models.NewFromError(err)
			apiError = &e
			updateJob(fmt.Sprintf("Host %s failed: %s — trying next host", host.Host, e.Message))
			continue
		}
		reg.Register(hostJob.ID, jobID, host.ID)
		updateJob(fmt.Sprintf("Dispatched to host %s, tracking progress via job %s", host.Host, hostJob.ID))
		// Completion (success or failure) is forwarded by HostJobEventHandler.
		return nil, nil
	}

	// All hosts failed.
	if apiError == nil {
		apiError = &models.ApiErrorResponse{
			Message: "Failed to dispatch VM to any of the selected hosts",
			Code:    500,
		}
	}
	updateJob(apiError.Message)
	return nil, apiError
}

func (s *OrchestratorService) CreateHosVirtualMachine(ctx basecontext.ApiContext, jobID string, hostId string, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, *models.ApiErrorResponse) {
	jobManager := jobs.Get(ctx)
	updateJob := func(msg string) {
		if jobID != "" && jobManager != nil {
			_, _ = jobManager.UpdateJobMessage(jobID, msg)
		}
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		apiError := &models.ApiErrorResponse{
			Message: "There was an error getting the database",
			Code:    500,
		}
		updateJob(apiError.Message)
		return nil, apiError
	}

	specs := s.getSpecsFromRequest(request)
	if specs == nil {
		apiError := &models.ApiErrorResponse{
			Message: "There was an error getting the specs from the request",
			Code:    500,
		}
		updateJob(apiError.Message)
		return nil, apiError
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		apiError := &models.ApiErrorResponse{
			Message: "There was an error getting the host from the database",
			Code:    500,
		}
		updateJob(apiError.Message)
		return nil, apiError
	}

	isOk, validateErr := s.validateHost(*host, request, specs)
	if validateErr != nil {
		updateJob(fmt.Sprintf("Host %s failed validation: %s", host.Host, validateErr.Message))
		return nil, validateErr
	}

	if !isOk {
		apiError := &models.ApiErrorResponse{
			Message: fmt.Sprintf("Host %s is not available to create the virtual machine", host.Host),
			Code:    400,
		}
		updateJob(apiError.Message)
		s.Refresh()
		return nil, apiError
	}

	updateJob(fmt.Sprintf("Dispatching to host %s", host.Host))
	ctx.LogInfof("[Orchestrator] Dispatching async VM creation to host %s", host.Host)
	reg := registry.Get()
	hostJob, err := s.CallCreateHostVirtualMachineAsync(*host, request)
	if err != nil {
		e := models.NewFromError(err)
		updateJob(fmt.Sprintf("Host %s failed: %s", host.Host, e.Message))
		s.Refresh()
		return nil, &e
	}

	reg.Register(hostJob.ID, jobID, host.ID)
	updateJob(fmt.Sprintf("Dispatched to host %s, tracking progress via job %s", host.Host, hostJob.ID))
	// Completion (success or failure) is forwarded by HostJobEventHandler.
	return nil, nil
}

func (s *OrchestratorService) pingHostForLatency(host data_models.OrchestratorHost) time.Duration {
	client := s.getApiClient(host)
	client.WithTimeout(2 * time.Second)
	path := "/api/v1/config/health"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return 10 * time.Second
	}

	start := time.Now()
	var res interface{}
	resp, err := client.Get(url.String(), &res)
	duration := time.Since(start)
	if err != nil || resp.StatusCode != 200 {
		return 10 * time.Second
	}
	return duration
}

func filterAndSortHosts(validHosts []data_models.OrchestratorHost, request models.CreateVirtualMachineRequest, getPing func(host data_models.OrchestratorHost) time.Duration) ([]data_models.OrchestratorHost, *models.ApiErrorResponse) {
	// Stage 2: Selection Tags Filter
	if len(request.SelectionTags) > 0 {
		var tagMatchedHosts []data_models.OrchestratorHost
		for _, host := range validHosts {
			matched := false
			for _, tag := range request.SelectionTags {
				for _, hostTag := range host.Tags {
					if strings.EqualFold(tag, hostTag) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if matched {
				tagMatchedHosts = append(tagMatchedHosts, host)
			}
		}
		if len(tagMatchedHosts) == 0 {
			return nil, &models.ApiErrorResponse{
				Message: "Did not find any available host that meets the tag condition",
				Code:    400,
			}
		}
		validHosts = tagMatchedHosts
	}

	// Stage 3: Cache Locality Check
	// this will be used to select the host that has the cache and make it the first choice
	// it will make the creation of the VM faster if the host has the cache as it does not
	// need to download the cache from the catalog
	if request.CatalogManifest != nil {
		var cachedHosts []data_models.OrchestratorHost
		for _, host := range validHosts {
			hasCache := false
			for _, cacheItem := range host.CacheItems {
				if strings.EqualFold(cacheItem.CatalogId, request.CatalogManifest.CatalogId) &&
					strings.EqualFold(cacheItem.Version, request.CatalogManifest.Version) &&
					strings.EqualFold(cacheItem.Architecture, request.Architecture) {
					hasCache = true
					break
				}
			}
			if hasCache {
				cachedHosts = append(cachedHosts, host)
			}
		}
		// Only filter down if at least one host has the cache, otherwise we allow them all to download fresh
		if len(cachedHosts) > 0 {
			validHosts = cachedHosts
		}
	}

	// Stage 4: Ping Latency Sorting
	// This is for global load balancing, we decide what is the closest host to the orchestrator and the less busy
	// and make it the first choice
	type hostPing struct {
		host data_models.OrchestratorHost
		ping time.Duration
	}

	var pings []hostPing
	for _, host := range validHosts {
		duration := getPing(host)
		pings = append(pings, hostPing{host: host, ping: duration})
	}

	sort.Slice(pings, func(i, j int) bool {
		return pings[i].ping < pings[j].ping
	})

	sortedHosts := make([]data_models.OrchestratorHost, 0)
	for _, p := range pings {
		sortedHosts = append(sortedHosts, p.host)
	}

	return sortedHosts, nil
}

// CallCreateHostVirtualMachineAsync calls the host's async machine-creation
// endpoint and returns the host job response immediately (HTTP 202).
func (s *OrchestratorService) CallCreateHostVirtualMachineAsync(host data_models.OrchestratorHost, request models.CreateVirtualMachineRequest) (*models.JobResponse, error) {
	httpClient := s.getApiClient(host)
	httpClient.WithTimeout(30 * time.Second)

	path := "/machines/async"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

  tempPayload, _ := json.Marshal(request)
  s.ctx.LogInfof("[Orchestrator] Sending async VM creation request to %s with payload: %s", url, string(tempPayload))

	var response models.JobResponse
	apiResponse, err := httpClient.Post(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 202 {
		return nil, errors.NewWithCodef(apiResponse.StatusCode, "error dispatching async VM creation to host %s: status %d", host.Host, apiResponse.StatusCode)
	}

	s.ctx.LogInfof("[Orchestrator] Dispatched async VM creation to host %s, host job ID: %s", host.Host, response.ID)
	return &response, nil
}

func (s *OrchestratorService) CallCreateHostVirtualMachine(host data_models.OrchestratorHost, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, error) {
	httpClient := s.getApiClient(host)
	timeout := 5 * time.Hour
	s.ctx.LogInfof("[Orchestrator] Setting timeout of %v for VM creation request to host %s", timeout, host.Host)
	httpClient.WithTimeout(timeout)

	path := "/machines"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	s.ctx.LogInfof("[Orchestrator] Starting VM creation request to %s", url)
	var response models.CreateVirtualMachineResponse
	apiResponse, err := httpClient.Post(url.String(), request, &response)
	if err != nil {
		s.ctx.LogErrorf("[Orchestrator] VM creation request failed: %v (Status: %d)", err, apiResponse.StatusCode)
		return nil, err
	}

	response.Host = host.GetHost()
	s.ctx.LogInfof("[Orchestrator] Successfully created VM on host %s", host.Host)

	s.Refresh()
	return &response, nil
}

func (s *OrchestratorService) getSpecsFromRequest(request models.CreateVirtualMachineRequest) *models.CreateVirtualMachineSpecs {
	var specs *models.CreateVirtualMachineSpecs
	var err error
	switch {
	case request.CatalogManifest != nil:
		specs, err = s.getCatalogSpecs(request.CatalogManifest.Connection, request.CatalogManifest.CatalogId, request.CatalogManifest.Version, request.Architecture)
		if err != nil {
			// Unable to reach the catalog (e.g. local catalog, no connection string, or
			// older host that doesn't expose the endpoint). Use safe defaults so host
			// selection can still proceed based on CPU/memory alone.
			s.ctx.LogWarnf("[Orchestrator] Could not retrieve catalog specs for %s/%s: %v — using defaults for host selection", request.CatalogManifest.CatalogId, request.CatalogManifest.Version, err)
			specs = &models.CreateVirtualMachineSpecs{
				Type:   "pvm",
				Cpu:    "2",
				Memory: "2048",
			}
		}
		if request.CatalogManifest.Specs != nil {
			if request.CatalogManifest.Specs.Cpu != "" && request.CatalogManifest.Specs.Cpu != "0" {
				specs.Cpu = request.CatalogManifest.Specs.Cpu
			}
			if request.CatalogManifest.Specs.Memory != "" && request.CatalogManifest.Specs.Memory != "0" {
				specs.Memory = request.CatalogManifest.Specs.Memory
			}
			if request.CatalogManifest.Specs.Disk != "" && request.CatalogManifest.Specs.Disk != "0" {
				specs.Disk = request.CatalogManifest.Specs.Disk
			}
		}
	case request.VagrantBox != nil && request.VagrantBox.Specs != nil:
		specs = request.VagrantBox.Specs
	case request.PackerTemplate != nil && request.PackerTemplate.Specs != nil:
		specs = request.PackerTemplate.Specs
	default:
		specs = &models.CreateVirtualMachineSpecs{
			Type:   "pvm",
			Cpu:    "2",
			Memory: "2048",
		}
	}

	if specs.Type == "" {
		specs.Type = "pvm"
	}

	return specs
}

func (s *OrchestratorService) validateHost(host data_models.OrchestratorHost, request models.CreateVirtualMachineRequest, specs *models.CreateVirtualMachineSpecs) (bool, *models.ApiErrorResponse) {
	var apiError *models.ApiErrorResponse
	if !host.Enabled {
		apiError = &models.ApiErrorResponse{
			Message: "Host is not enabled",
			Code:    400,
		}
		return false, apiError
	}

	if host.State != "healthy" {
		apiError = &models.ApiErrorResponse{
			Message: "Host is not healthy",
			Code:    400,
		}
		return false, apiError
	}

	if host.Resources == nil {
		apiError = &models.ApiErrorResponse{
			Message: "Host does not have resources information",
			Code:    400,
		}
		return false, apiError
	}

	if !strings.EqualFold(host.Architecture, request.Architecture) {
		apiError = &models.ApiErrorResponse{
			Message: "Host does not have the same architecture",
			Code:    400,
		}

		return false, apiError
	}

	// We will trust that the host has the reserved cpus setup correctly
	// otherwise we would potentially go above the reserved cpus
	availableCpus := host.Resources.TotalAvailable.LogicalCpuCount
	availableMemory := host.Resources.TotalAvailable.MemorySize

	if specs != nil && specs.Size > 0 {
		diskSpace, diskErr := s.getHostDiskSpace(s.ctx, host, request.Owner)
		if diskErr != nil {
			// Host may be running an older version that doesn't expose the disk-space
			// endpoint. Log a warning and skip the check so the host remains eligible.
			s.ctx.LogWarnf("[Orchestrator] Could not get disk space info for host %s (may be older version, skipping check): %v", host.Host, diskErr)
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
			if diskSpace.ParallelsHome < requiredSpace {
				return false, &models.ApiErrorResponse{
					Message: fmt.Sprintf("Host does not have enough disk space: available %d MB, required %d MB, "+
						"we need 3x / 2x space of vm size depending on the volume configuration", diskSpace.ParallelsHome, requiredSpace),
					Code: 400,
				}
			}
		}
	}

	// Checking for the maximum number of Apple VMs
	if strings.EqualFold(specs.Type, "macvm") {
		if host.Resources.TotalAppleVms >= MaxNumberAppleVms {
			apiError = &models.ApiErrorResponse{
				Message: "Host has reached the maximum number of Apple VMs",
				Code:    400,
			}

			return false, apiError
		}
	}

	if availableCpus < specs.GetCpuCount() ||
		availableMemory < specs.GetMemorySize() {
		if availableCpus < specs.GetCpuCount() {
			apiError = &models.ApiErrorResponse{
				Message: "Host does not have enough CPU resources",
				Code:    400,
			}

			return false, apiError
		}
		if availableMemory < specs.GetMemorySize() {
			apiError = &models.ApiErrorResponse{
				Message: "Host does not have enough Memory resources",
				Code:    400,
			}

			return false, apiError
		}
	}

	return true, nil
}

func (s *OrchestratorService) getCatalogSpecs(connection string, catalogId string, version string, architecture string) (*models.CreateVirtualMachineSpecs, error) {
	provider := catalog_models.CatalogManifestProvider{}
	if err := provider.Parse(connection); err != nil {
		return nil, err
	}

	host := data_models.OrchestratorHost{
		Host: provider.GetUrl(),
		Authentication: &data_models.OrchestratorHostAuthentication{
			Username: provider.Username,
			Password: provider.Password,
			ApiKey:   provider.ApiKey,
		},
	}

	httpClient := s.getApiClient(host)
	path := "/api/v1/catalog/" + catalogId + "/" + version + "/" + architecture
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.CatalogManifest
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting hardware info for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	result := models.CreateVirtualMachineSpecs{}
	result.Type = response.Type

	if response.MinimumSpecRequirements != nil {
		result.Cpu = strconv.Itoa(response.MinimumSpecRequirements.Cpu)
		result.Memory = strconv.Itoa(response.MinimumSpecRequirements.Memory)
		result.Disk = strconv.Itoa(response.MinimumSpecRequirements.Disk)
		result.Size = response.Size
	}

	// Setting the default values
	if response.Type == "" {
		result.Type = "pvm"
	}
	if result.Cpu == "" || result.Cpu == "0" {
		result.Cpu = "2"
	}
	if result.Memory == "" || result.Memory == "0" {
		result.Memory = "2048"
	}
	if result.Disk == "" || result.Disk == "0" {
		if response.PackSize > 0 {
			result.Disk = strconv.Itoa(int(response.PackSize))
		} else {
			result.Disk = "35000"
		}
	}

	return &result, nil
}

// isSameVolume reports whether two filesystem paths reside on the same volume.
// On macOS, volumes under /Volumes/<name>/… are identified by their name.
// Paths on the root volume (or empty paths) are treated as the root volume.
func isSameVolume(path1, path2 string) bool {
	getVolume := func(p string) string {
		parts := strings.Split(p, "/")
		// /Volumes/<name>/... → parts = ["", "Volumes", "<name>", ...]
		if len(parts) > 2 && strings.EqualFold(parts[1], "Volumes") {
			return strings.ToLower(parts[2])
		}
		return ""
	}
	return getVolume(path1) == getVolume(path2)
}
