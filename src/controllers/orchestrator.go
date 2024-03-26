package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator"
	"github.com/Parallels/prl-devops-service/restapi"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerOrchestratorHostsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Orchestrator handlers", version)
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
		WithHandler(RegisterOrchestratorHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}").
		WithRequiredClaim(constants.DELETE_CLAIM).
		WithHandler(UnregisterOrchestratorHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/enable").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(EnableOrchestratorHostsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/disable").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(DisableOrchestratorHostsHandler()).
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
		WithPath("/orchestrator/machines/{id}").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}").
		WithRequiredClaim(constants.DELETE_CLAIM).
		WithHandler(DeleteOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/status").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorVirtualMachineStatusHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/rename").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(RenameOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/set").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(SetOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/execute").
		WithRequiredClaim(constants.EXECUTE_COMMAND_VM_CLAIM).
		WithHandler(ExecuteCommandOnVirtualMachineHandler()).
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
// @Tags			Orchestrator
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
		defer Recover(ctx, r, w)
		defer Recover(ctx, r, w)
		filter := GetFilterHeader(r)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		dtoOrchestratorHosts, err := orchestratorSvc.GetHosts(ctx, filter)
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

		for _, host := range dtoOrchestratorHosts {
			rHost := mappers.DtoOrchestratorHostToApiResponse(*host)
			response = append(response, rHost)
		}

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
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		host, err := orchestratorSvc.GetHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoOrchestratorHostToApiResponse(*host)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Orchestrator host returned successfully")
	}
}

// @Summary		Register a Host in the orchestrator
// @Description	This endpoint register a host in the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Param			hostRequest	body		models.OrchestratorHostRequest	true	"Host Request"
// @Success		200			{object}	models.OrchestratorHostResponse
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts [post]
func RegisterOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
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
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		// checking if we can connect to host before adding it
		dtoRecord := mappers.ApiOrchestratorRequestToDto(request)

		record, err := orchestratorSvc.RegisterHost(ctx, &dtoRecord)
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

// @Summary		Unregister a host from the orchestrator
// @Description	This endpoint unregister a host from the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path	string	true	"Host ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id} [delete]
func UnregisterOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc.UnregisterHost(ctx, id)

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Orchestrator host deleted successfully")
	}
}

// @Summary		Enable a host in the orchestrator
// @Description	This endpoint will enable an existing host in the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Success		200	{object}	models.OrchestratorHostResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/enable [get]
func EnableOrchestratorHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		host, err := orchestratorSvc.EnableHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoOrchestratorHostToApiResponse(*host)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Host %v enabled successfully", id)
	}
}

// @Summary		Disable a host in the orchestrator
// @Description	This endpoint will disable an existing host in the orchestrator
// @Tags			Orchestrator
// @Produce		json
// @Success		200	{object}	models.OrchestratorHostResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{id}/disable [get]
func DisableOrchestratorHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		host, err := orchestratorSvc.DisableHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoOrchestratorHostToApiResponse(*host)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Host %v disabled successfully", id)
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
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		result := make([]models.HostResourceOverviewResponse, 0)
		resources, err := orchestratorSvc.GetResources(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		for _, value := range resources {
			item := models.HostResourceOverviewResponse{}
			item.Total = mappers.MapApiHostResourceItemFromHostResourceItem(value.Total)
			item.TotalAvailable = mappers.MapApiHostResourceItemFromHostResourceItem(value.TotalAvailable)
			item.TotalInUse = mappers.MapApiHostResourceItemFromHostResourceItem(value.TotalInUse)
			item.TotalReserved = mappers.MapApiHostResourceItemFromHostResourceItem(value.TotalReserved)
			item.CpuType = value.CpuType
			item.CpuBrand = value.CpuBrand
			result = append(result, item)
		}

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
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		resources, err := orchestratorSvc.GetHostResources(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.MapSystemUsageResponseFromHostResources(*resources)

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
		filter := GetFilterHeader(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vms, err := orchestratorSvc.GetVirtualMachines(ctx, filter)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := make([]models.ParallelsVM, 0)
		for _, vm := range vms {
			response = append(response, mappers.MapDtoVirtualMachineToApi(vm))
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned %v virtual machines from all hosts", len(response))
	}
}

// @Summary		Get orchestrator Virtual Machine
// @Description	This endpoint returns orchestrator Virtual Machine by its ID
// @Tags			Orchestrator
// @Produce		json
// @Success		200	{object}	models.ParallelsVM
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines/{id} [get]
func GetOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		vm, err := orchestratorSvc.GetVirtualMachine(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.MapDtoVirtualMachineToApi(*vm)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned virtual machine %v from host", vm.ID, vm.HostId)
	}
}

// @Summary		Deletes orchestrator virtual machine
// @Description	This endpoint deletes orchestrator virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path	string	true	"Virtual Machine ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines/{id} [delete]
func DeleteOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		err := orchestratorSvc.DeleteVirtualMachine(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully deleted the orchestrator virtual machine %s", id)
	}
}

// @Summary		Get orchestrator virtual machine status
// @Description	This endpoint returns orchestrator virtual machine status
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines/{vmId}/status [get]
func GetOrchestratorVirtualMachineStatusHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		response, err := orchestratorSvc.GetVirtualMachineStatus(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned successfully the orchestrator virtual machine status")
	}
}

// @Summary		Renames orchestrator virtual machine
// @Description	This endpoint renames orchestrator virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines/{id}/rename [put]
func RenameOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.RenameVirtualMachineRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		request.ID = id

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

		response, err := orchestratorSvc.RenameVirtualMachine(ctx, id, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully renamed the orchestrator virtual machine %s", id)
	}
}

// @Summary		Configures orchestrator virtual machine
// @Description	This endpoint configures orchestrator virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.VirtualMachineConfigResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines/{vmId}/set [put]
func SetOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineConfigRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

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

		response, err := orchestratorSvc.ConfigureVirtualMachine(ctx, id, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully configured the orchestrator virtual machine %s", id)
	}
}

// @Summary		Executes a command in a orchestrator virtual machine
// @Description	This endpoint executes a command in a orchestrator virtual machine
// @Tags			Orchestrator
// @Produce		json
// @Param			id	path		string	true	"Virtual Machine ID"
// @Success		200		{object}	models.VirtualMachineConfigResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines/{vmId}/execute [put]
func ExecutesOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineExecuteCommandRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

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

		response, err := orchestratorSvc.ExecuteOnVirtualMachine(ctx, id, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully executed command in the orchestrator virtual machine %s", id)
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
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		vms, err := orchestratorSvc.GetHostVirtualMachines(ctx, id, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := make([]models.ParallelsVM, 0)
		for _, vm := range vms {
			response = append(response, mappers.MapDtoVirtualMachineToApi(*vm))
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
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		vm, err := orchestratorSvc.GetHostVirtualMachine(ctx, id, vmId)
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
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		err := orchestratorSvc.DeleteHostVirtualMachine(ctx, id, vmId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
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
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		response, err := orchestratorSvc.GetHostVirtualMachineStatus(ctx, id, vmId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Returned successfully the orchestrator virtual machine status")
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
		defer Recover(ctx, r, w)
		var request models.RenameVirtualMachineRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

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

		response, err := orchestratorSvc.RenameHostVirtualMachine(ctx, id, vmId, request)
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
		defer Recover(ctx, r, w)
		var request models.VirtualMachineConfigRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

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

		response, err := orchestratorSvc.ConfigureHostVirtualMachine(ctx, id, vmId, request)
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
		defer Recover(ctx, r, w)
		var request models.VirtualMachineExecuteCommandRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

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

		response, err := orchestratorSvc.ExecuteOnHostVirtualMachine(ctx, id, vmId, request)
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
		defer Recover(ctx, r, w)
		var request models.RegisterVirtualMachineRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

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

		response, err := orchestratorSvc.RegisterHostVirtualMachine(ctx, id, request)
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
		defer Recover(ctx, r, w)
		var request models.UnregisterVirtualMachineRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

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

		_, err := orchestratorSvc.UnregisterHostVirtualMachine(ctx, id, vmId, request)
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
		defer Recover(ctx, r, w)
		var request models.CreateVirtualMachineRequest

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

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.CreateHosVirtualMachine(ctx, id, request)
		if err != nil {
			ReturnApiError(ctx, w, *err)
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
		defer Recover(ctx, r, w)
		var request models.CreateVirtualMachineRequest

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

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.CreateVirtualMachine(ctx, request)
		if err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully configured the orchestrator virtual machine %s", response.ID)
	}
}
