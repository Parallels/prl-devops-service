package orchestrator

import (
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) CreateVirtualMachine(ctx basecontext.ApiContext, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, *models.ApiErrorResponse) {
	var apiError *models.ApiErrorResponse
	var response models.CreateVirtualMachineResponse

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		apiError = &models.ApiErrorResponse{
			Message: "There was an error getting the database",
			Code:    500,
		}
		return nil, apiError
	}

	specs := s.getSpecsFromRequest(request)

	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		apiError = &models.ApiErrorResponse{
			Message: "There was an error getting the hosts from the database",
			Code:    500,
		}
		return nil, apiError
	}

	var selectedHost *data_models.OrchestratorHost
	for _, orchestratorHost := range hosts {
		isOk, err := s.validateHost(orchestratorHost, request.Architecture, specs)
		if err != nil {
			continue
		}

		if isOk {
			resp, err := s.CallCreateHostVirtualMachine(orchestratorHost, request)
			if err != nil {
				e := models.NewFromError(err)
				apiError = &e
				continue
			} else {
				response = *resp
				selectedHost = &orchestratorHost
				break
			}
		}
	}

	if selectedHost == nil {
		if apiError != nil {
			return nil, apiError
		}

		apiError = &models.ApiErrorResponse{
			Message: "No host available to create the virtual machine",
			Code:    400,
		}
	}

	s.Refresh()
	return &response, apiError
}

func (s *OrchestratorService) CreateHosVirtualMachine(ctx basecontext.ApiContext, hostId string, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, *models.ApiErrorResponse) {
	var apiError *models.ApiErrorResponse
	var response models.CreateVirtualMachineResponse

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		apiError = &models.ApiErrorResponse{
			Message: "There was an error getting the database",
			Code:    500,
		}
		return nil, apiError
	}

	specs := s.getSpecsFromRequest(request)
	if specs == nil {
		apiError = &models.ApiErrorResponse{
			Message: "There was an error getting the specs from the request",
			Code:    500,
		}
		return nil, apiError
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		apiError = &models.ApiErrorResponse{
			Message: "There was an error getting the hosts from the database",
			Code:    500,
		}
		return nil, apiError
	}

	var selectedHost *data_models.OrchestratorHost
	isOk, validateErr := s.validateHost(*host, request.Architecture, specs)
	if validateErr != nil {
		return nil, validateErr
	}

	if isOk {
		ctx.LogInfof("Creating virtual machine on host %s", host.Host)
		resp, err := s.CallCreateHostVirtualMachine(*host, request)
		if err != nil {
			e := models.NewFromError(err)
			apiError = &e
			return nil, apiError
		} else {
			response = *resp
			selectedHost = host
		}
	}

	if selectedHost == nil {
		apiError = &models.ApiErrorResponse{
			Message: "No host available to create the virtual machine",
			Code:    400,
		}
	}

	s.Refresh()
	return &response, apiError
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
			return nil
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

func (s *OrchestratorService) validateHost(host data_models.OrchestratorHost, architecture string, specs *models.CreateVirtualMachineSpecs) (bool, *models.ApiErrorResponse) {
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

	if !strings.EqualFold(host.Architecture, architecture) {
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
