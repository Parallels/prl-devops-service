package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/orchestrator"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerOrchestratorHostsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Claims handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).WithPath("orchestrator/hosts").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorHostsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts").
		WithRequiredClaim(constants.CREATE_CLAIM).
		WithHandler(CreateOrchestratorHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}").
		WithRequiredClaim(constants.DELETE_CLAIM).
		WithHandler(DeleteOrchestratorHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/overview/resources").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorOverviewHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/overview/{id}/resources").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorHostResourcesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/machines").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorVirtualMachinesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorHostVirtualMachinesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}").
		WithRequiredClaim(constants.DELETE_CLAIM).
		WithHandler(DeleteOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/status").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorHostVirtualMachineStatusHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/rename").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(RenameOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/set").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(SetOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/execute").
		WithRequiredClaim(constants.EXECUTE_COMMAND_VM_CLAIM).
		WithHandler(ExecutesOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/register").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(RegisterOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/unregister").
		WithRequiredClaim(constants.UPDATE_VM_CLAIM).
		WithHandler(UnregisterOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(CreateOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/machines").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(CreateOrchestratorVirtualMachineHandler()).
		Register()
}

// @Summary		Gets all hosts from the orchestrator
// @Description	This endpoint returns all hosts from the orchestrator
// @Tags			Claims
// @Produce		json
// @Success		200	{object}	[]models.OrchestratorHostResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts [get]
func GetOrchestratorHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoOrchestratorHosts, err := dbService.GetOrchestratorHosts(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(dtoOrchestratorHosts) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.OrchestratorHostResponse, 0)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Hosts returned: %v", len(response))
			return
		}

		response := make([]models.OrchestratorHostResponse, 0)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		// // Checking the orchestrator hosts health
		// for _, host := range dtoOrchestratorHosts {
		// 	rHost := mappers.DtoOrchestratorHostToApiResponse(host)
		// 	rHost.State = orchestratorSvc.GetHostHealthCheckState(&host)

		// 	response = append(response, rHost)
		// }

		var wg sync.WaitGroup
		mutex := sync.Mutex{}

		for _, host := range dtoOrchestratorHosts {
			starTime := time.Now()
			wg.Add(1)
			go func(host data_models.OrchestratorHost) {
				ctx.LogDebugf("Processing Host: %v\n", host.Host)
				defer wg.Done()

				rHost := mappers.DtoOrchestratorHostToApiResponse(host)
				rHost.State = orchestratorSvc.GetHostHealthCheckState(&host)

				mutex.Lock()
				response = append(response, rHost)
				mutex.Unlock()
				ctx.LogDebugf("Processing Host: %v - Time: %v\n", host.Host, time.Since(starTime))
			}(host)
		}

		wg.Wait()

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Hosts returned successfully")
	}
}

// @Summary		Gets a host from the orchestrator
// @Description	This endpoint returns a host from the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path		string	true	"Host ID"
// @Success		200	{object}	models.OrchestratorHostResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id} [get]
func GetOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoOrchestratorHost, err := dbService.GetOrchestratorHost(ctx, helpers.NormalizeString(id))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		// Validating the Health check probe of the host
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		dtoOrchestratorHost.State = orchestratorSvc.GetHostHealthCheckState(dtoOrchestratorHost)
		response := mappers.DtoOrchestratorHostToApiResponse(*dtoOrchestratorHost)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Orchestrator host returned successfully")
	}
}

// @Summary		Creates a Host in the orchestrator
// @Description	This endpoint creates a host in the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Param			hostRequest	body		models.OrchestratorHostRequest	true	"Host Request"
// @Success		200			{object}	models.OrchestratorHostResponse
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts [post]
func CreateOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.OrchestratorHostRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		oSvc := orchestrator.NewOrchestratorService(ctx)
		// checking if we can connect to host before adding it
		dtoRecord := mappers.ApiOrchestratorRequestToDto(request)

		_, err = oSvc.GetHostHardwareInfo(&dtoRecord)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		record, err := dbService.CreateOrchestratorHost(ctx, dtoRecord)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoOrchestratorHostToApiResponse(*record)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Orchestrator Host created successfully")
	}
}

// @Summary		Delete a host from the orchestrator
// @Description	This endpoint deletes a host from the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path	string	true	"Host ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id} [delete]
func DeleteOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Orchestrator host deleted successfully")
	}
}

// @Summary		Get orchestrator resource overview
// @Description	This endpoint returns orchestrator resource overview
// @Tags			Orchestrator
// @Produce		json
// @Success		200	{object}	models.HostResourceOverviewResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/overview/resources [get]
func GetOrchestratorOverviewHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}
		response := models.HostResourceOverviewResponse{}
		result := make([]models.HostResourceOverviewResponse, 0)

		totalResources := dbService.GetOrchestratorTotalResources(ctx)
		inUseResources := dbService.GetOrchestratorInUseResources(ctx)
		availableResources := dbService.GetOrchestratorAvailableResources(ctx)
		reservedResources := dbService.GetOrchestratorReservedResources(ctx)

		for key, value := range totalResources {
			response.Total = mappers.MapApiHostResourceItemFromHostResourceItem(value)
			response.TotalAvailable = mappers.MapApiHostResourceItemFromHostResourceItem(availableResources[key])
			response.TotalInUse = mappers.MapApiHostResourceItemFromHostResourceItem(inUseResources[key])
			response.TotalReserved = mappers.MapApiHostResourceItemFromHostResourceItem(reservedResources[key])
			response.CpuType = key
			result = append(result, response)
		}
		// response.Total = mappers.MapApiHostResourceItemFromHostResourceItem(totalResources)
		// response.TotalAvailable = mappers.MapApiHostResourceItemFromHostResourceItem(availableResources)
		// response.TotalInUse = mappers.MapApiHostResourceItemFromHostResourceItem(inUseResources)
		// response.TotalReserved = mappers.MapApiHostResourceItemFromHostResourceItem(reservedResources)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Returned successfully the orchestrator overview")
	}
}

// @Summary		Get orchestrator host resources
// @Description	This endpoint returns orchestrator host resources
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path		string	true	"Host ID"
// @Success		200	{object}	models.HostResourceOverviewResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/overview/{id}/resources [get]
func GetOrchestratorHostResourcesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.MapSystemUsageResponseFromHostResources(*host.Resources)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned successfully the orchestrator host resources")
	}
}

// @Summary		Get orchestrator Virtual Machines
// @Description	This endpoint returns orchestrator Virtual Machines
// @Tags			Orchestrator
// @Produce		json
// @Success		200	{object}	[]models.ParallelsVM
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines [get]
func GetOrchestratorVirtualMachinesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vms, err := dbService.GetOrchestratorVirtualMachines(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := make([]models.ParallelsVM, 0)
		for _, vm := range vms {
			response = append(response, mappers.MapDtoVirtualMachineToApi(vm))
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned successfully the orchestrator virtual machines")
	}
}

// @Summary		Get orchestrator host virtual machines
// @Description	This endpoint returns orchestrator host virtual machines
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path		string	true	"Host ID"
// @Success		200	{object}	[]models.ParallelsVM
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines [get]
func GetOrchestratorHostVirtualMachinesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		vms, err := dbService.GetOrchestratorHostVirtualMachines(ctx, id, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := make([]models.ParallelsVM, 0)
		for _, vm := range vms {
			response = append(response, mappers.MapDtoVirtualMachineToApi(vm))
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned successfully %s orchestrator virtual machines", len(response))
	}
}

// @Summary		Get orchestrator host virtual machine
// @Description	This endpoint returns orchestrator host virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string	true	"Host ID"
// @Param			vmId	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/{vmId} [get]
func GetOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		vm, err := dbService.GetOrchestratorHostVirtualMachine(ctx, id, vmId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.MapDtoVirtualMachineToApi(*vm)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned successfully the orchestrator virtual machine")
	}
}

// @Summary		Deletes orchestrator host virtual machine
// @Description	This endpoint deletes orchestrator host virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path	string	true	"Host ID"
// @Param			vmId	path	string	true	"Virtual Machine ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/{vmId} [delete]
func DeleteOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}
		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		if err := orchestratorSvc.DeleteHostVirtualMachine(host, vmId); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully deleted the orchestrator virtual machine %s", vmId)
	}
}

// @Summary		Get orchestrator host virtual machine status
// @Description	This endpoint returns orchestrator host virtual machine status
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string	true	"Host ID"
// @Param			vmId	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/status [get]
func GetOrchestratorHostVirtualMachineStatusHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
			return
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.GetHostVirtualMachineStatus(host, vmId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned successfully the orchestrator virtual machine")
	}
}

// @Summary		Renames orchestrator host virtual machine
// @Description	This endpoint renames orchestrator host virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string	true	"Host ID"
// @Param			vmId	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/rename [put]
func RenameOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.RenameVirtualMachineRequest
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		request.ID = vmId

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
			return
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.RenameHostVirtualMachine(host, vmId, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully renamed the orchestrator virtual machine %s", vmId)
	}
}

// @Summary		Configures orchestrator host virtual machine
// @Description	This endpoint configures orchestrator host virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string	true	"Host ID"
// @Param			vmId	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.VirtualMachineConfigResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/set [put]
func SetOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.VirtualMachineConfigRequest
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
			return
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.ConfigureHostVirtualMachine(host, vmId, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully configured the orchestrator virtual machine %s", vmId)
	}
}

// @Summary		Executes a command in a orchestrator host virtual machine
// @Description	This endpoint executes a command in a orchestrator host virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string	true	"Host ID"
// @Param			vmId	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.VirtualMachineConfigResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/execute [put]
func ExecutesOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.VirtualMachineExecuteCommandRequest
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
			return
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.ExecuteOnHostVirtualMachine(host, vmId, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully executed command in the orchestrator virtual machine %s", vmId)
	}
}

// @Summary		Register a virtual machine in a orchestrator host
// @Description	This endpoint registers a virtual machine in a orchestrator host
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string									true	"Host ID"
// @Param			request	body		models.RegisterVirtualMachineRequest	true	"Register Virtual Machine Request"
// @Success		200		{object}	models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/register [post]
func RegisterOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.RegisterVirtualMachineRequest
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
			return
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.RegisterHostVirtualMachine(host, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully registered virtual machine %s in the orchestrator", response.ID)
	}
}

// @Summary		Unregister a virtual machine in a orchestrator host
// @Description	This endpoint unregister a virtual machine in a orchestrator host
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string									true	"Host ID"
// @Param			vmId	path		string									true	"Virtual Machine ID"
// @Param			request	body		models.UnregisterVirtualMachineRequest	true	"Register Virtual Machine Request"
// @Success		200		{object}	models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister [post]
func UnregisterOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UnregisterVirtualMachineRequest
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
			return
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		_, err = orchestratorSvc.UnregisterHostVirtualMachine(host, vmId, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		ReturnApiCommonResponse(w)
		ctx.LogInfof("Successfully unregistered virtual machine %s in the orchestrator", vmId)
	}
}

// @Summary		Creates a orchestrator host virtual machine
// @Description	This endpoint creates a orchestrator host virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id		path		string								true	"Host ID"
// @Param			request	body		models.CreateVirtualMachineRequest	true	"Create Virtual Machine Request"
// @Success		200		{object}	models.CreateVirtualMachineResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/machines [post]
func CreateOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.CreateVirtualMachineRequest
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		host, err := dbService.GetOrchestratorHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		if host == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host not found",
				Code:    404,
			})
			return
		}
		if host.State != "healthy" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host is not healthy",
				Code:    400,
			})
			return
		}

		var specs *models.CreateVirtualMachineSpecs
		if request.CatalogManifest != nil && request.CatalogManifest.Specs != nil {
			specs = request.CatalogManifest.Specs
		} else if request.VagrantBox != nil && request.VagrantBox.Specs != nil {
			specs = request.VagrantBox.Specs
		} else if request.PackerTemplate != nil && request.PackerTemplate.Specs != nil {
			specs = request.PackerTemplate.Specs
		} else {
			specs = &models.CreateVirtualMachineSpecs{
				Cpu:    "1",
				Memory: "2048",
			}
		}

		if host.Resources.TotalAvailable.LogicalCpuCount <= specs.GetCpuCount() {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host does not have enough CPU resources",
				Code:    400,
			})
			return
		}
		if host.Resources.TotalAvailable.MemorySize <= specs.GetMemorySize() {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Host does not have enough Memory resources",
				Code:    400,
			})
			return
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.CreateHostVirtualMachine(*host, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully configured the orchestrator virtual machine %s", response.ID)
	}
}

// @Summary		Creates a virtual machine in one of the hosts for the orchestrator
// @Description	This endpoint creates a virtual machine in one of the hosts for the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Param			request	body		models.CreateVirtualMachineRequest	true	"Create Virtual Machine Request"
// @Success		200		{object}	models.CreateVirtualMachineResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines [post]
func CreateOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.CreateVirtualMachineRequest
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		var specs *models.CreateVirtualMachineSpecs
		if request.CatalogManifest != nil && request.CatalogManifest.Specs != nil {
			specs = request.CatalogManifest.Specs
		} else if request.VagrantBox != nil && request.VagrantBox.Specs != nil {
			specs = request.VagrantBox.Specs
		} else if request.PackerTemplate != nil && request.PackerTemplate.Specs != nil {
			specs = request.PackerTemplate.Specs
		} else {
			specs = &models.CreateVirtualMachineSpecs{
				Cpu:    "1",
				Memory: "2048",
			}
		}

		hosts, err := dbService.GetOrchestratorHosts(ctx, "")
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		var hostErr error
		var response models.CreateVirtualMachineResponse
		var apiError *models.ApiErrorResponse

		for _, orchestratorHost := range hosts {
			if orchestratorHost.State != "healthy" {
				apiError = &models.ApiErrorResponse{
					Message: "Host is not healthy",
					Code:    400,
				}
				continue
			}
			if orchestratorHost.Resources == nil {
				apiError = &models.ApiErrorResponse{
					Message: "Host does not have resources information",
					Code:    400,
				}
				continue
			}
			if !strings.EqualFold(orchestratorHost.Architecture, request.Architecture) {
				apiError = &models.ApiErrorResponse{
					Message: "Host does not have the same architecture",
					Code:    400,
				}
				continue
			}
			if orchestratorHost.Resources.TotalAvailable.LogicalCpuCount > specs.GetCpuCount() &&
				orchestratorHost.Resources.TotalAvailable.MemorySize > specs.GetMemorySize() {

				if orchestratorHost.State != "healthy" {
					apiError = &models.ApiErrorResponse{
						Message: "Host is not healthy",
						Code:    400,
					}
					hostErr = errors.New("host is not healthy")
					continue
				}

				if orchestratorHost.Resources.TotalAvailable.LogicalCpuCount <= 1 {
					apiError = &models.ApiErrorResponse{
						Message: "Host does not have enough CPU resources",
						Code:    400,
					}
					hostErr = errors.New("host does not have enough CPU resources")
					continue
				}
				if orchestratorHost.Resources.TotalAvailable.MemorySize < 2048 {
					apiError = &models.ApiErrorResponse{
						Message: "Host does not have enough Memory resources",
						Code:    400,
					}
					hostErr = errors.New("host does not have enough Memory resources")
					continue
				}

				orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
				resp, err := orchestratorSvc.CreateHostVirtualMachine(orchestratorHost, request)
				if err != nil {
					e := models.NewFromError(err)
					apiError = &e
					hostErr = err
					break
				} else {
					response = *resp
					break
				}
			}
		}

		if hostErr == nil {
			if apiError != nil {
				ReturnApiError(ctx, w, *apiError)
			} else {
				ReturnApiError(ctx, w, models.ApiErrorResponse{
					Message: "No host available to create the virtual machine",
					Code:    400,
				})
			}

			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully configured the orchestrator virtual machine %s", response.ID)
	}
}
