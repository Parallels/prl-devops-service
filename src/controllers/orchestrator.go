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
		WithPath("/orchestrator/hosts/{id}").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(UpdateOrchestratorHostHandler()).
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
		WithPath("/orchestrator/machines/{id}/start").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(StartOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/stop").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(StopOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/execute").
		WithRequiredClaim(constants.EXECUTE_COMMAND_VM_CLAIM).
		WithHandler(ExecutesOrchestratorVirtualMachineHandler()).
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
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/start").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(StartOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/stop").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(StopOrchestratorHostVirtualMachineHandler()).
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

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts").
		WithRequiredClaim(constants.LIST_REVERSE_PROXY_HOSTS_CLAIM).
		WithHandler(GetOrchestratorHostReverseProxyHostsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}").
		WithRequiredClaim(constants.LIST_REVERSE_PROXY_HOSTS_CLAIM).
		WithHandler(GetOrchestratorHostReverseProxyHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts").
		WithRequiredClaim(constants.CREATE_REVERSE_PROXY_HOST_CLAIM).
		WithHandler(CreateOrchestratorHostReverseProxyHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}").
		WithRequiredClaim(constants.UPDATE_REVERSE_PROXY_HOST_CLAIM).
		WithHandler(UpdateOrchestratorHostReverseProxyHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}").
		WithRequiredClaim(constants.DELETE_REVERSE_PROXY_HOST_CLAIM).
		WithHandler(DeleteOrchestratorHostReverseProxyHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes").
		WithRequiredClaim(constants.UPDATE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM).
		WithHandler(UpsertOrchestratorHostReverseProxyHostHttpRouteHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}").
		WithRequiredClaim(constants.DELETE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM).
		WithHandler(DeleteOrchestratorHostReverseProxyHostHttpRouteHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route").
		WithRequiredClaim(constants.UPDATE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM).
		WithHandler(UpdateOrchestratorHostReverseProxyHostTcpRouteHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/restart").
		WithRequiredClaim(constants.CONFIGURE_REVERSE_PROXY_CLAIM).
		WithHandler(RestartsOrchestratorHostReverseProxyHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/enable").
		WithRequiredClaim(constants.CONFIGURE_REVERSE_PROXY_CLAIM).
		WithHandler(EnableOrchestratorHostReverseProxyHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy/disable").
		WithRequiredClaim(constants.CONFIGURE_REVERSE_PROXY_CLAIM).
		WithHandler(DisableOrchestratorHostReverseProxyHandler()).
		Register()
}

//	@Summary		Gets all hosts from the orchestrator
//	@Description	This endpoint returns all hosts from the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Success		200	{object}	[]models.OrchestratorHostResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts [get]
func GetOrchestratorHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			defer r.Body.Close()
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

//	@Summary		Gets a host from the orchestrator
//	@Description	This endpoint returns a host from the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Host ID"
//	@Success		200	{object}	models.OrchestratorHostResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id} [get]
func GetOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Register a Host in the orchestrator
//	@Description	This endpoint register a host in the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			hostRequest	body		models.OrchestratorHostRequest	true	"Host Request"
//	@Success		200			{object}	models.OrchestratorHostResponse
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts [post]
func RegisterOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.OrchestratorHostRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
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

//	@Summary		Unregister a host from the orchestrator
//	@Description	This endpoint unregister a host from the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Host ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id} [delete]
func UnregisterOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Enable a host in the orchestrator
//	@Description	This endpoint will enable an existing host in the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Success		200	{object}	models.OrchestratorHostResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/enable [put]
func EnableOrchestratorHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Disable a host in the orchestrator
//	@Description	This endpoint will disable an existing host in the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Success		200	{object}	models.OrchestratorHostResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/disable [put]
func DisableOrchestratorHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Update a Host in the orchestrator
//	@Description	This endpoint updates a host in the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			hostRequest	body		models.OrchestratorHostUpdateRequest	true	"Host Update Request"
//	@Success		200			{object}	models.OrchestratorHostResponse
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts [put]
func UpdateOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.OrchestratorHostUpdateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		svc := orchestrator.NewOrchestratorService(ctx)
		host, err := svc.GetDatabaseHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if request.Authentication != nil && request.Authentication.Username != "" {
			dtoRecord := mappers.ApiOrchestratorAuthenticationToDto(*request.Authentication)
			host.Authentication = &dtoRecord
		}

		if request.Description != "" {
			host.Description = request.Description
		}

		if request.Host != "" {
			host.Host = request.Host
			host.Schema = request.Schema
			host.Port = request.Port
			host.PathPrefix = request.Prefix
		}

		record, err := svc.UpdateHost(ctx, host)
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

//	@Summary		Get orchestrator resource overview
//	@Description	This endpoint returns orchestrator resource overview
//	@Tags			Orchestrator
//	@Produce		json
//	@Success		200	{object}	models.HostResourceOverviewResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/overview/resources [get]
func GetOrchestratorOverviewHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Get orchestrator host resources
//	@Description	This endpoint returns orchestrator host resources
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Host ID"
//	@Success		200	{object}	models.HostResourceOverviewResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/overview/{id}/resources [get]
func GetOrchestratorHostResourcesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Get orchestrator Virtual Machines
//	@Description	This endpoint returns orchestrator Virtual Machines
//	@Tags			Orchestrator
//	@Produce		json
//	@Success		200	{object}	[]models.ParallelsVM
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines [get]
func GetOrchestratorVirtualMachinesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Get orchestrator Virtual Machine
//	@Description	This endpoint returns orchestrator Virtual Machine by its ID
//	@Tags			Orchestrator
//	@Produce		json
//	@Success		200	{object}	models.ParallelsVM
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id} [get]
func GetOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Deletes orchestrator virtual machine
//	@Description	This endpoint deletes orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Virtual Machine ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id} [delete]
func DeleteOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Get orchestrator virtual machine status
//	@Description	This endpoint returns orchestrator virtual machine status
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.ParallelsVM
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{vmId}/status [get]
func GetOrchestratorVirtualMachineStatusHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Renames orchestrator virtual machine
//	@Description	This endpoint renames orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.ParallelsVM
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/rename [put]
func RenameOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Configures orchestrator virtual machine
//	@Description	This endpoint configures orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineConfigResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{vmId}/set [put]
func SetOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Starts orchestrator virtual machine
//	@Description	This endpoint starts orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineConfigResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{vmId}/start [put]
func StartOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		request := models.VirtualMachineConfigRequest{
			Operations: []*models.VirtualMachineConfigRequestOperation{
				{
					Group:     "state",
					Operation: "start",
				},
			},
		}

		response, err := orchestratorSvc.ConfigureVirtualMachine(ctx, id, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully started the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Stops orchestrator virtual machine
//	@Description	This endpoint sops orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineConfigResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{vmId}/stop [put]
func StopOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		request := models.VirtualMachineConfigRequest{
			Operations: []*models.VirtualMachineConfigRequestOperation{
				{
					Group:     "state",
					Operation: "stop",
				},
			},
		}

		response, err := orchestratorSvc.ConfigureVirtualMachine(ctx, id, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully stopped the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Executes a command in a orchestrator virtual machine
//	@Description	This endpoint executes a command in a orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineConfigResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{vmId}/execute [put]
func ExecutesOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Get orchestrator host virtual machines
//	@Description	This endpoint returns orchestrator host virtual machines
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Host ID"
//	@Success		200	{object}	[]models.ParallelsVM
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines [get]
func GetOrchestratorHostVirtualMachinesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Get orchestrator host virtual machine
//	@Description	This endpoint returns orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.ParallelsVM
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId} [get]
func GetOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Deletes orchestrator host virtual machine
//	@Description	This endpoint deletes orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path	string	true	"Host ID"
//	@Param			vmId	path	string	true	"Virtual Machine ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId} [delete]
func DeleteOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Get orchestrator host virtual machine status
//	@Description	This endpoint returns orchestrator host virtual machine status
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.ParallelsVM
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/status [get]
func GetOrchestratorHostVirtualMachineStatusHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

//	@Summary		Renames orchestrator host virtual machine
//	@Description	This endpoint renames orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.ParallelsVM
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/rename [put]
func RenameOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Configures orchestrator host virtual machine
//	@Description	This endpoint configures orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineConfigResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/set [put]
func SetOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Starts orchestrator host virtual machine
//	@Description	This endpoint starts orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineConfigResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/start [put]
func StartOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		request := models.VirtualMachineConfigRequest{
			Operations: []*models.VirtualMachineConfigRequestOperation{
				{
					Group:     "state",
					Operation: "start",
				},
			},
		}

		response, err := orchestratorSvc.ConfigureHostVirtualMachine(ctx, id, vmId, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully started the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Starts orchestrator host virtual machine
//	@Description	This endpoint starts orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineConfigResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/stop [put]
func StopOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		request := models.VirtualMachineConfigRequest{
			Operations: []*models.VirtualMachineConfigRequestOperation{
				{
					Group:     "state",
					Operation: "stop",
				},
			},
		}

		response, err := orchestratorSvc.ConfigureHostVirtualMachine(ctx, id, vmId, request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully stopped the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Executes a command in a orchestrator host virtual machine
//	@Description	This endpoint executes a command in a orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineConfigResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/execute [put]
func ExecutesOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Register a virtual machine in a orchestrator host
//	@Description	This endpoint registers a virtual machine in a orchestrator host
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string									true	"Host ID"
//	@Param			request	body		models.RegisterVirtualMachineRequest	true	"Register Virtual Machine Request"
//	@Success		200		{object}	models.ParallelsVM
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/register [post]
func RegisterOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Unregister a virtual machine in a orchestrator host
//	@Description	This endpoint unregister a virtual machine in a orchestrator host
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string									true	"Host ID"
//	@Param			vmId	path		string									true	"Virtual Machine ID"
//	@Param			request	body		models.UnregisterVirtualMachineRequest	true	"Register Virtual Machine Request"
//	@Success		200		{object}	models.ParallelsVM
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister [post]
func UnregisterOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Creates a orchestrator host virtual machine
//	@Description	This endpoint creates a orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string								true	"Host ID"
//	@Param			request	body		models.CreateVirtualMachineRequest	true	"Create Virtual Machine Request"
//	@Success		200		{object}	models.CreateVirtualMachineResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines [post]
func CreateOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
			return
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

//	@Summary		Creates a virtual machine in one of the hosts for the orchestrator
//	@Description	This endpoint creates a virtual machine in one of the hosts for the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			request	body		models.CreateVirtualMachineRequest	true	"Create Virtual Machine Request"
//	@Success		200		{object}	models.CreateVirtualMachineResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines [post]
func CreateOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.CreateVirtualMachineRequest

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
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

// region Orchestrator Reverse Proxy

//	@Summary		Gets orchestrator host reverse proxy hosts
//	@Description	This endpoint returns orchestrator host reverse proxy hosts
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Host ID"
//	@Success		200	{object}	[]models.ReverseProxyHost
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts [get]
func GetOrchestratorHostReverseProxyHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.GetHostReverseProxyHosts(ctx, id, "")
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully got orchestrator host %s reverse proxy hosts", id)
	}
}

//	@Summary		Gets orchestrator host reverse proxy hosts
//	@Description	This endpoint returns orchestrator host reverse proxy hosts
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Host ID"
//	@Success		200	{object}	models.ReverseProxyHost
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id} [get]
func GetOrchestratorHostReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]
		reverseProxyHostId := vars["reverse_proxy_host_id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.GetHostReverseProxyHost(ctx, id, reverseProxyHostId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully got orchestrator host %s reverse proxy hosts", id)
	}
}

//	@Summary		Creates a orchestrator host reverse proxy host
//	@Description	This endpoint creates a orchestrator host reverse proxy host
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			request	body		models.ReverseProxyHostCreateRequest	true	"Create Host Reverse Proxy Host Request"
//	@Success		200		{object}	models.ReverseProxyHost
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts [post]
func CreateOrchestratorHostReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.ReverseProxyHostCreateRequest

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.CreateHostReverseProxyHost(ctx, id, request)
		if err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully created the orchestrator host reverse proxy host %s", response.ID)
	}
}

//	@Summary		Updates an orchestrator host reverse proxy host
//	@Description	This endpoint updates an orchestrator host reverse proxy host
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			request	body		models.ReverseProxyHostUpdateRequest	true	"Update Host Reverse Proxy Host Request"
//	@Success		200		{object}	models.ReverseProxyHost
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id} [put]
func UpdateOrchestratorHostReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.ReverseProxyHostUpdateRequest

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		rpHostId := vars["reverse_proxy_host_id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.UpdateHostReverseProxyHost(ctx, id, rpHostId, request)
		if err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully updated the orchestrator host reverse proxy host %s", response.ID)
	}
}

//	@Summary		Deletes an orchestrator host reverse proxy host
//	@Description	This endpoint deletes an orchestrator host reverse proxy host
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id						path	string	true	"Host ID"
//	@Param			reverse_proxy_host_id	path	string	true	"Reverse Proxy Host ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id} [delete]
func DeleteOrchestratorHostReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		rpHostId := vars["reverse_proxy_host_id"]

		err := orchestratorSvc.DeleteHostReverseProxyHost(ctx, id, rpHostId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully deleted the orchestrator reverse proxy host %s", rpHostId)
	}
}

//	@Summary		Upserts an orchestrator host reverse proxy host http route
//	@Description	This endpoint upserts an orchestrator host reverse proxy host http route
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			request	body		models.ReverseProxyHostUpdateRequest	true	"Upsert Host Reverse Proxy Host Http Routes Request"
//	@Success		200		{object}	models.ReverseProxyHost
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes [post]
func UpsertOrchestratorHostReverseProxyHostHttpRouteHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.ReverseProxyHostHttpRouteCreateRequest

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		rpHostId := vars["reverse_proxy_host_id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.UpsertHostReverseProxyHostHttpRoute(ctx, id, rpHostId, request)
		if err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully updated the orchestrator host reverse proxy host %s", response.ID)
	}
}

//	@Summary		Deletes an orchestrator host reverse proxy host http route
//	@Description	This endpoint deletes an orchestrator host reverse proxy host http route
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id						path	string	true	"Host ID"
//	@Param			reverse_proxy_host_id	path	string	true	"Reverse Proxy Host ID"
//	@Param			route_id				path	string	true	"Route ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id} [delete]
func DeleteOrchestratorHostReverseProxyHostHttpRouteHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		rpHostId := vars["reverse_proxy_host_id"]
		routeId := vars["route_id"]

		err := orchestratorSvc.DeleteHostReverseProxyHostHttpRoute(ctx, id, rpHostId, routeId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully deleted the orchestrator reverse proxy host %s", rpHostId)
	}
}

//	@Summary		Update an orchestrator host reverse proxy host tcp route
//	@Description	This endpoint updates an orchestrator host reverse proxy host tcp route
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			request	body		models.ReverseProxyHostUpdateRequest	true	"Update Host Reverse Proxy Host tcp Routes Request"
//	@Success		200		{object}	models.ReverseProxyHost
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route [post]
func UpdateOrchestratorHostReverseProxyHostTcpRouteHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.ReverseProxyHostTcpRouteCreateRequest

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		rpHostId := vars["reverse_proxy_host_id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.UpdateHostReverseProxyHostTcpRoute(ctx, id, rpHostId, request)
		if err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully updated the orchestrator host reverse proxy host %s", response.ID)
	}
}

//	@Summary		Restarts orchestrator host reverse proxy
//	@Description	This endpoint restarts orchestrator host reverse proxy
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Host ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/restart [put]
func RestartsOrchestratorHostReverseProxyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		err := orchestratorSvc.RestartHostReverseProxy(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully restarted host %s reverse proxy", id)
	}
}

//	@Summary		Enables orchestrator host reverse proxy
//	@Description	This endpoint enables orchestrator host reverse proxy
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Host ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/enable [put]
func EnableOrchestratorHostReverseProxyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		err := orchestratorSvc.EnableHostReverseProxy(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully enabled orchestrator host %s reverse proxy", id)
	}
}

//	@Summary		Disables orchestrator host reverse proxy
//	@Description	This endpoint disables orchestrator host reverse proxy
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Host ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy/disable [put]
func DisableOrchestratorHostReverseProxyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		err := orchestratorSvc.DisableHostReverseProxy(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully disabled orchestrator host %s reverse proxy", id)
	}
}

// endregion
