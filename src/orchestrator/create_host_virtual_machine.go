package orchestrator

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
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

	return &response, apiError
}

func (s *OrchestratorService) CallCreateHostVirtualMachine(host data_models.OrchestratorHost, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, error) {
	httpClient := s.getApiClient(host)
	path := "/machines"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.CreateVirtualMachineResponse
	_, err = httpClient.Post(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	response.Host = host.GetHost()

	s.Refresh()
	return &response, nil
}

func (s *OrchestratorService) getSpecsFromRequest(request models.CreateVirtualMachineRequest) *models.CreateVirtualMachineSpecs {
	var specs *models.CreateVirtualMachineSpecs
	switch {
	case request.CatalogManifest != nil && request.CatalogManifest.Specs != nil:
		specs = request.CatalogManifest.Specs
	case request.VagrantBox != nil && request.VagrantBox.Specs != nil:
		specs = request.VagrantBox.Specs
	case request.PackerTemplate != nil && request.PackerTemplate.Specs != nil:
		specs = request.PackerTemplate.Specs
	default:
		specs = &models.CreateVirtualMachineSpecs{
			Cpu:    "2",
			Memory: "2048",
		}
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

	systemCPUThreshold := int64(1)
	systemMemoryThreshold := float64(1024)
	availableCpus := host.Resources.TotalAvailable.LogicalCpuCount - systemCPUThreshold
	availableMemory := host.Resources.TotalAvailable.MemorySize - systemMemoryThreshold

	if availableCpus <= specs.GetCpuCount() ||
		availableMemory <= specs.GetMemorySize() {
		if availableCpus <= specs.GetCpuCount() {
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
