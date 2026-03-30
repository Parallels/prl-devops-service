package controllers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/hardware").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetOrchestratorHostHardwareInfoHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts").
		WithRequiredClaim(constants.CREATE_CLAIM).
		WithExtraAdapter(restapi.EnrollmentTokenAuthorizationMiddlewareAdapter()).
		WithHandler(RegisterOrchestratorHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/enrollment-token").
		WithRequiredClaim(constants.CREATE_CLAIM).
		WithHandler(CreateEnrollmentTokenHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/enrollment-token/{token}/validate").
		WithHandler(ValidateEnrollmentTokenHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/deploy").
		WithRequiredClaim(constants.CREATE_CLAIM).
		WithHandler(DeployOrchestratorHostHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/hosts/deploy/async").
		WithRequiredClaim(constants.CREATE_CLAIM).
		WithHandler(AsyncDeployOrchestratorHostHandler()).
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
		WithPath("/orchestrator/machines/{id}/restart").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(RestartOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/suspend").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(SuspendOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/resume").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(ResumeOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/reset").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(ResetOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/pause").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(PauseOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/machines/{id}/clone").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(CloneOrchestratorVirtualMachineHandler()).
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
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/restart").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(RestartOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/suspend").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(SuspendOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/resume").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(ResumeOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/reset").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(ResetOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/pause").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(PauseOrchestratorHostVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/clone").
		WithRequiredClaim(constants.UPDATE_CLAIM).
		WithHandler(CloneOrchestratorHostVirtualMachineHandler()).
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
		WithPath("/orchestrator/hosts/{id}/machines/async").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(AsyncCreateOrchestratorHostVirtualMachineHandler()).
		Register()

	// Snapshot endpoints for orchestrator host virtual machines
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/snapshots").
		WithRequiredClaim(constants.LIST_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.LIST_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(ListOrchestratorHostVirtualMachineSnapshots()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/snapshots").
		WithRequiredClaim(constants.CREATE_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.CREATE_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(CreateOrchestratorHostVirtualMachineSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/snapshots").
		WithRequiredClaim(constants.DELETE_ALL_SNAPSHOTS_VM_CLAIM).
		WithRequiredClaim(constants.DELETE_ALL_OWN_VM_SNAPSHOTS_CLAIM).
		WithHandler(DeleteAllOrchestratorHostVirtualMachineSnapshots()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}").
		WithRequiredClaim(constants.DELETE_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.DELETE_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(DeleteOrchestratorHostVirtualMachineSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert").
		WithRequiredClaim(constants.REVERT_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.REVERT_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(RevertOrchestratorHostVirtualMachineSnapshot()).
		Register()

	// Snapshot endpoints for orchestrator virtual machines (host resolved automatically)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/machines/{id}/snapshots").
		WithRequiredClaim(constants.LIST_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.LIST_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(ListOrchestratorVirtualMachineSnapshots()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/machines/{id}/snapshots").
		WithRequiredClaim(constants.CREATE_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.CREATE_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(CreateOrchestratorVirtualMachineSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/machines/{id}/snapshots").
		WithRequiredClaim(constants.DELETE_ALL_SNAPSHOTS_VM_CLAIM).
		WithRequiredClaim(constants.DELETE_ALL_OWN_VM_SNAPSHOTS_CLAIM).
		WithHandler(DeleteAllOrchestratorVirtualMachineSnapshots()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/machines/{id}/snapshots/{snapshot_id}").
		WithRequiredClaim(constants.DELETE_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.DELETE_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(DeleteOrchestratorVirtualMachineSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithOrClaims().
		WithPath("/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert").
		WithRequiredClaim(constants.REVERT_SNAPSHOT_VM_CLAIM).
		WithRequiredClaim(constants.REVERT_OWN_VM_SNAPSHOT_CLAIM).
		WithHandler(RevertOrchestratorVirtualMachineSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/machines").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(CreateOrchestratorVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/machines/async").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(AsyncCreateOrchestratorVirtualMachineHandler()).
		Register()

	// region Catalog Cache
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/cache").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(GetOrchestratorHostCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/cache").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteOrchestratorHostCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/cache/{catalog_id}").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteOrchestratorHostCatalogCacheItemHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/cache/{catalog_id}/{catalog_version}").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteOrchestratorHostCatalogCacheItemVersionHandler()).
		Register()
		// endregion

	// region Reverse Proxy
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/reverse-proxy").
		WithRequiredClaim(constants.LIST_REVERSE_PROXY_HOSTS_CLAIM).
		WithHandler(GetOrchestratorHostReverseProxyConfigHandler()).
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
	// endregion

	// region Logs
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/logs/stream").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(StreamOrchestratorHostSystemLogs()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{id}/logs").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(GetOrchestratorHostSystemLogs()).
		Register()
	// endregion
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

//	@Summary		Gets a host hardware info from the orchestrator
//	@Description	This endpoint returns a host hardware info from the orchestrator
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Host ID"
//	@Success		200	{object}	models.SystemUsageResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/hardware [get]
func GetOrchestratorHostHardwareInfoHandler() restapi.ControllerHandler {
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

		hw, err := orchestratorSvc.GetHostHardwareInfo(host)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(hw)
		ctx.LogInfof("Orchestrator host hardware info returned successfully")
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

		// If the request was authenticated via an enrollment token, validate that
		// the token's intended host_name matches the Description field and then
		// mark the token as consumed so it cannot be reused.
		if tokenValue := r.Header.Get(constants.ENROLLMENT_TOKEN_HEADER); tokenValue != "" {
			db := serviceprovider.Get().JsonDatabase
			_ = db.Connect(ctx)
			token, err := db.ValidateEnrollmentToken(ctx, tokenValue)
			if err != nil {
				ReturnApiError(ctx, w, models.ApiErrorResponse{
					Message: "Enrollment token validation failed: " + err.Error(),
					Code:    http.StatusUnauthorized,
				})
				return
			}
			if token.HostName != "" && token.HostName != request.Description {
				ReturnApiError(ctx, w, models.ApiErrorResponse{
					Message: fmt.Sprintf("enrollment token is bound to host %q but request description is %q", token.HostName, request.Description),
					Code:    http.StatusForbidden,
				})
				return
			}
			// Mark as used before registration to prevent races.
			if err := db.MarkEnrollmentTokenUsed(ctx, token.ID); err != nil {
				ctx.LogWarnf("Failed to mark enrollment token as used: %v", err)
			}
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		resources, err := orchestratorSvc.GetResources(ctx, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		for _, value := range resources {
			item := models.HostResourceOverviewResponse{}
			item.SystemReserved = mappers.MapApiHostResourceItemFromHostResourceItem(value.SystemReserved)
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		vms, err := orchestratorSvc.GetVirtualMachines(ctx, filter, noCache)
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		vm, err := orchestratorSvc.GetVirtualMachine(ctx, id, noCache)
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
//	@Param			id		path	string	true	"Virtual Machine ID"
//	@Param			force	query	bool	false	"Force Delete"
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

		force := false
		if r.URL.Query().Get("force") == "true" {
			force = true
		}

		err := orchestratorSvc.DeleteVirtualMachine(ctx, id, force)
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
//	@Router			/v1/orchestrator/machines/{id}/status [get]
func GetOrchestratorVirtualMachineStatusHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.GetVirtualMachineStatus(ctx, id, noCache)
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
//	@Router			/v1/orchestrator/machines/{id}/set [put]
func SetOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineConfigRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

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

		response, err := orchestratorSvc.ConfigureVirtualMachine(ctx, id, request, noCache)
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
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/start [put]
func StartOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.StartVirtualMachine(ctx, id, noCache)
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
//	@Param			id		path		string	true	"Virtual Machine ID"
//	@Param			force	query		bool	false	"Force Stop"
//	@Success		200		{object}	models.VirtualMachineOperationResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/stop [put]
func StopOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		force := false
		if r.URL.Query().Get("force") == "true" {
			force = true
		}

		response, err := orchestratorSvc.StopVirtualMachine(ctx, id, force, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully stopped the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Restarts orchestrator virtual machine
//	@Description	This endpoint restarts orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/restart [put]
func RestartOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.RestartVirtualMachine(ctx, id, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully restarted the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Suspends orchestrator virtual machine
//	@Description	This endpoint suspends orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/suspend [put]
func SuspendOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.SuspendVirtualMachine(ctx, id, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully suspended the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Resumes orchestrator virtual machine
//	@Description	This endpoint resumes orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/resume [put]
func ResumeOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.ResumeVirtualMachine(ctx, id, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully resumed the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Resets orchestrator virtual machine
//	@Description	This endpoint resets orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/reset [put]
func ResetOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.ResetVirtualMachine(ctx, id, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully reset the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Pauses orchestrator virtual machine
//	@Description	This endpoint pauses orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/pause [put]
func PauseOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.PauseVirtualMachine(ctx, id, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully paused the orchestrator virtual machine %s", id)
	}
}

//	@Summary		Clones orchestrator virtual machine
//	@Description	This endpoint clones orchestrator virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id				path		string										true	"Virtual Machine ID"
//	@Param			configRequest	body		models.VirtualMachineCloneCommandRequest	true	"Machine Clone Request"
//	@Success		200				{object}	models.VirtualMachineCloneCommandResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/clone [put]
func CloneOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var req models.VirtualMachineCloneCommandRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		if err := http_helper.MapRequestBody(r, &req); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := req.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		response, err := orchestratorSvc.CloneVirtualMachine(ctx, id, req, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully cloned the orchestrator virtual machine %s", id)
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
//	@Router			/v1/orchestrator/machines/{id}/execute [put]
func ExecutesOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineExecuteCommandRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

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

		response, err := orchestratorSvc.ExecuteOnVirtualMachine(ctx, id, request, noCache)
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		vms, err := orchestratorSvc.GetHostVirtualMachines(ctx, id, GetFilterHeader(r), noCache)
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
		ctx.LogInfof("Returned successfully %v orchestrator virtual machines", len(response))
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		vm, err := orchestratorSvc.GetHostVirtualMachine(ctx, id, vmId, noCache)
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
//	@Param			force	query	bool	false	"Force Delete"
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

		force := false
		if r.URL.Query().Get("force") == "true" {
			force = true
		}

		err := orchestratorSvc.DeleteHostVirtualMachine(ctx, id, vmId, force)
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.GetHostVirtualMachineStatus(ctx, id, vmId, noCache)
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

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

		response, err := orchestratorSvc.ConfigureHostVirtualMachine(ctx, id, vmId, request, noCache)
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
//	@Success		200		{object}	models.VirtualMachineOperationResponse
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.StartHostVirtualMachine(ctx, id, vmId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully started the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Stops orchestrator host virtual machine
//	@Description	This endpoint stops orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Param			force	query		bool	false	"Force Stop"
//	@Success		200		{object}	models.VirtualMachineOperationResponse
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		force := false
		if r.URL.Query().Get("force") == "true" {
			force = true
		}

		response, err := orchestratorSvc.StopHostVirtualMachine(ctx, id, vmId, force, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully stopped the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Restarts orchestrator host virtual machine
//	@Description	This endpoint restarts orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineOperationResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/restart [put]
func RestartOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.RestartHostVirtualMachine(ctx, id, vmId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully restarted the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Suspends orchestrator host virtual machine
//	@Description	This endpoint suspends orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineOperationResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/suspend [put]
func SuspendOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.SuspendHostVirtualMachine(ctx, id, vmId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully suspended the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Resumes orchestrator host virtual machine
//	@Description	This endpoint resumes orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineOperationResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/resume [put]
func ResumeOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.ResumeHostVirtualMachine(ctx, id, vmId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully resumed the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Resets orchestrator host virtual machine
//	@Description	This endpoint resets orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineOperationResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/reset [put]
func ResetOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.ResetHostVirtualMachine(ctx, id, vmId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully reset the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Pauses orchestrator host virtual machine
//	@Description	This endpoint pauses orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.VirtualMachineOperationResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/pause [put]
func PauseOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		response, err := orchestratorSvc.PauseHostVirtualMachine(ctx, id, vmId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully paused the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Clones orchestrator host virtual machine
//	@Description	This endpoint clones orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id				path		string										true	"Host ID"
//	@Param			vmId			path		string										true	"Virtual Machine ID"
//	@Param			configRequest	body		models.VirtualMachineCloneCommandRequest	true	"Machine Clone Request"
//	@Success		200				{object}	models.VirtualMachineCloneCommandResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/clone [put]
func CloneOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var req models.VirtualMachineCloneCommandRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		if err := http_helper.MapRequestBody(r, &req); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := req.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		response, err := orchestratorSvc.CloneHostVirtualMachine(ctx, id, vmId, req, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully cloned the orchestrator virtual machine %s", vmId)
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
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

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

		response, err := orchestratorSvc.ExecuteOnHostVirtualMachine(ctx, id, vmId, request, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully executed command in the orchestrator virtual machine %s", vmId)
	}
}

//	@Summary		Lists snapshots of orchestrator host virtual machine
//	@Description	This endpoint lists snapshots of orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string	true	"Host ID"
//	@Param			vmId	path		string	true	"Virtual Machine ID"
//	@Success		200		{object}	models.ListVMSnapshotResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots [get]
func ListOrchestratorHostVirtualMachineSnapshots() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestrator := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]

		ctx.LogInfof("[controllers/orchestrator][snapshots] Listing snapshots for host %s machine %s", id, vmId)

		response, err := orchestrator.GetHostVirtualMachineSnapshots(ctx, id, vmId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully listed snapshots for orchestrator host %s virtual machine %s", id, vmId)
	}
}

//	@Summary		Creates a snapshot for orchestrator host virtual machine
//	@Description	This endpoint creates a snapshot for orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id				path		string							true	"Host ID"
//	@Param			vmId			path		string							true	"Virtual Machine ID"
//	@Param			createRequest	body		models.CreateVMSnapshotRequest	true	"Create Snapshot Request"
//	@Success		202				{object}	models.CreateVMSnapshotResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots [post]
func CreateOrchestratorHostVirtualMachineSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.CreateVMSnapshotRequest
		orchestrator := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			ctx.LogErrorf("[controllers/orchestrator][snapshots] Error decoding JSON: %v", err)
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Creating snapshot for host %s machine %s", id, vmId)

		response, err := orchestrator.CreateHostVirtualMachineSnapshot(ctx, id, vmId, request, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully created snapshot for orchestrator host %s virtual machine %s", id, vmId)
	}
}

//	@Summary		Deletes all snapshots of orchestrator host virtual machine
//	@Description	This endpoint deletes all snapshots of orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path	string	true	"Host ID"
//	@Param			vmId	path	string	true	"Virtual Machine ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots [delete]
func DeleteAllOrchestratorHostVirtualMachineSnapshots() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestrator := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Deleting all snapshots for host %s machine %s", id, vmId)

		err := orchestrator.DeleteAllHostVirtualMachineSnapshots(ctx, id, vmId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully deleted all snapshots for orchestrator host %s virtual machine %s", id, vmId)
	}
}

//	@Summary		Deletes a snapshot of orchestrator host virtual machine
//	@Description	This endpoint deletes a snapshot of orchestrator host virtual machine
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id			path	string	true	"Host ID"
//	@Param			vmId		path	string	true	"Virtual Machine ID"
//	@Param			snapshot_id	path	string	true	"Snapshot ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id} [delete]
func DeleteOrchestratorHostVirtualMachineSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestrator := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		snapshotId := vars["snapshot_id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Deleting snapshot %s for host %s machine %s", snapshotId, id, vmId)

		err := orchestrator.DeleteHostVirtualMachineSnapshot(ctx, id, vmId, snapshotId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully deleted snapshot %s for orchestrator host %s virtual machine %s", snapshotId, id, vmId)
	}
}

//	@Summary		Reverts orchestrator host virtual machine to a snapshot
//	@Description	This endpoint reverts orchestrator host virtual machine to a snapshot
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id				path		string							true	"Host ID"
//	@Param			vmId			path		string							true	"Virtual Machine ID"
//	@Param			snapshot_id		path		string							true	"Snapshot ID"
//	@Param			revertRequest	body		models.RevertVMSnapshotRequest	false	"Revert Snapshot Request"
//	@Success		202				{object}	models.ApiCommonResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert [post]
func RevertOrchestratorHostVirtualMachineSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.RevertVMSnapshotRequest
		orchestrator := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		vmId := vars["vmId"]
		snapshotId := vars["snapshot_id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			ctx.LogDebugf("[controllers/orchestrator][snapshots] No request body provided for revert, proceeding with empty request")
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Reverting to snapshot %s for host %s machine %s", snapshotId, id, vmId)

		err := orchestrator.RevertHostVirtualMachineSnapshot(ctx, id, vmId, snapshotId, request, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.ApiCommonResponse{
			Success: true,
			Data:    "Snapshot revert operation completed successfully",
		})
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully reverted to snapshot %s for orchestrator host %s virtual machine %s", snapshotId, id, vmId)
	}
}

//	@Summary		Lists snapshots of an orchestrator virtual machine
//	@Description	This endpoint lists snapshots of an orchestrator virtual machine (host resolved automatically)
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Virtual Machine ID"
//	@Success		200	{object}	models.ListVMSnapshotResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/snapshots [get]
func ListOrchestratorVirtualMachineSnapshots() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Listing snapshots for machine %s", id)

		response, err := orchestratorSvc.GetVirtualMachineSnapshots(ctx, id, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully listed snapshots for orchestrator virtual machine %s", id)
	}
}

//	@Summary		Creates a snapshot for an orchestrator virtual machine
//	@Description	This endpoint creates a snapshot for an orchestrator virtual machine (host resolved automatically)
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id				path		string							true	"Virtual Machine ID"
//	@Param			createRequest	body		models.CreateVMSnapshotRequest	true	"Create Snapshot Request"
//	@Success		202				{object}	models.CreateVMSnapshotResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/snapshots [post]
func CreateOrchestratorVirtualMachineSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.CreateVMSnapshotRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			ctx.LogErrorf("[controllers/orchestrator][snapshots] Error decoding JSON: %v", err)
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Creating snapshot for machine %s", id)

		response, err := orchestratorSvc.CreateVirtualMachineSnapshot(ctx, id, request, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully created snapshot for orchestrator virtual machine %s", id)
	}
}

//	@Summary		Deletes all snapshots of an orchestrator virtual machine
//	@Description	This endpoint deletes all snapshots of an orchestrator virtual machine (host resolved automatically)
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Virtual Machine ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/snapshots [delete]
func DeleteAllOrchestratorVirtualMachineSnapshots() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Deleting all snapshots for machine %s", id)

		err := orchestratorSvc.DeleteAllVirtualMachineSnapshots(ctx, id, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully deleted all snapshots for orchestrator virtual machine %s", id)
	}
}

//	@Summary		Deletes a snapshot of an orchestrator virtual machine
//	@Description	This endpoint deletes a snapshot of an orchestrator virtual machine (host resolved automatically)
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id			path	string	true	"Virtual Machine ID"
//	@Param			snapshot_id	path	string	true	"Snapshot ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/snapshots/{snapshot_id} [delete]
func DeleteOrchestratorVirtualMachineSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		snapshotId := vars["snapshot_id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Deleting snapshot %s for machine %s", snapshotId, id)

		err := orchestratorSvc.DeleteVirtualMachineSnapshot(ctx, id, snapshotId, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully deleted snapshot %s for orchestrator virtual machine %s", snapshotId, id)
	}
}

//	@Summary		Reverts an orchestrator virtual machine to a snapshot
//	@Description	This endpoint reverts an orchestrator virtual machine to a snapshot (host resolved automatically)
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id				path		string							true	"Virtual Machine ID"
//	@Param			snapshot_id		path		string							true	"Snapshot ID"
//	@Param			revertRequest	body		models.RevertVMSnapshotRequest	false	"Revert Snapshot Request"
//	@Success		202				{object}	models.ApiCommonResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert [post]
func RevertOrchestratorVirtualMachineSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.RevertVMSnapshotRequest
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		snapshotId := vars["snapshot_id"]
		noCache := false
		if r.Header.Get("X-No-Cache") == "true" {
			noCache = true
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			ctx.LogDebugf("[controllers/orchestrator][snapshots] No request body provided for revert, proceeding with empty request")
		}

		ctx.LogInfof("[controllers/orchestrator][snapshots] Reverting to snapshot %s for machine %s", snapshotId, id)

		err := orchestratorSvc.RevertVirtualMachineSnapshot(ctx, id, snapshotId, request, noCache)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.ApiCommonResponse{
			Success: true,
			Data:    "Snapshot revert operation completed successfully",
		})
		ctx.LogInfof("[controllers/orchestrator][snapshots] Successfully reverted to snapshot %s for orchestrator virtual machine %s", snapshotId, id)
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
		ownerWasEmpty := request.Owner == ""
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if ownerWasEmpty {
			request.Owner = ""
			if request.CatalogManifest != nil {
				request.CatalogManifest.Owner = ""
			}
		}

		if request.CatalogManifest != nil {
			catalogConnection, connErr := resolveCatalogMachineConnection(ctx, request.CatalogManifest)
			if connErr != nil {
				ReturnApiError(ctx, w, models.NewFromError(connErr))
				return
			}
			request.CatalogManifest.Connection = catalogConnection
			request.CatalogManifest.CatalogManagerId = ""
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.CreateHosVirtualMachine(ctx, "", id, request)
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

		ownerWasEmpty := request.Owner == ""
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if ownerWasEmpty {
			request.Owner = ""
			if request.CatalogManifest != nil {
				request.CatalogManifest.Owner = ""
			}
		}

		if request.CatalogManifest != nil {
			catalogConnection, connErr := resolveCatalogMachineConnection(ctx, request.CatalogManifest)
			if connErr != nil {
				ReturnApiError(ctx, w, models.NewFromError(connErr))
				return
			}
			request.CatalogManifest.Connection = catalogConnection
			request.CatalogManifest.CatalogManagerId = ""
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.CreateVirtualMachine(ctx, "", request)
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

//	@Summary		Gets orchestrator host reverse proxy configuration
//	@Description	This endpoint returns orchestrator host reverse proxy configuration
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path		string	true	"Host ID"
//	@Success		200	{object}	models.ReverseProxy
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/reverse-proxy [get]
func GetOrchestratorHostReverseProxyConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		response, err := orchestratorSvc.GetHostReverseProxyConfig(ctx, id, "", true)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Successfully got orchestrator host %s reverse proxy config", id)
	}
}

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
		response, err := orchestratorSvc.GetHostReverseProxyHosts(ctx, id, "", true)
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
		response, err := orchestratorSvc.GetHostReverseProxyHost(ctx, id, reverseProxyHostId, true)
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

// region Orchestrator Hosts Catalog Cache

//	@Summary		Gets orchestrator host catalog cache
//	@Description	This endpoint returns orchestrator host catalog cache
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Host ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/catalog/cache [get]
func GetOrchestratorHostCatalogCacheHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		cacheItems, err := orchestratorSvc.GetHostCatalogCache(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cacheItems)
		ctx.LogInfof("Successfully got host %s cached Items", id)
	}
}

//	@Summary		Deletes an orchestrator host cache items
//	@Description	This endpoint deletes an orchestrator host cache items
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id	path	string	true	"Host ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/catalog/cache [delete]
func DeleteOrchestratorHostCatalogCacheHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		err := orchestratorSvc.DeleteHostCatalogCacheItem(ctx, id, "", "")
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully deleted the orchestrator host %s catalog cache", id)
	}
}

//	@Summary		Deletes an orchestrator host cache item and all its children
//	@Description	This endpoint deletes an orchestrator host cache item and all its children
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id			path	string	true	"Host ID"
//	@Param			catalog_id	path	string	true	"Catalog ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id} [delete]
func DeleteOrchestratorHostCatalogCacheItemHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		catalogId := vars["catalog_id"]

		err := orchestratorSvc.DeleteHostCatalogCacheItem(ctx, id, catalogId, "")
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully deleted the orchestrator host %s catalog cache item %v", id, catalogId)
	}
}

//	@Summary		Deletes an orchestrator host cache item and all its children
//	@Description	This endpoint deletes an orchestrator host cache item and all its children
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id				path	string	true	"Host ID"
//	@Param			catalog_id		path	string	true	"Catalog ID"
//	@Param			catalog_version	path	string	true	"Catalog Version"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version} [delete]
func DeleteOrchestratorHostCatalogCacheItemVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		catalogId := vars["catalog_id"]
		catalogVersion := vars["catalog_version"]

		err := orchestratorSvc.DeleteHostCatalogCacheItem(ctx, id, catalogId, catalogVersion)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Successfully deleted the orchestrator host %s catalog cache item %v and version", id, catalogId, catalogVersion)
	}
}

// endregion

// region Logs

//	@Summary		Gets the orchestrator host system logs from the disk
//	@Description	This endpoint returns the orchestrator host system logs from the disk
//	@Tags			Config
//	@Produce		plain
//	@Param			id	path	string	true	"Host ID"
//	@Success		200
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Router			/v1/orchestrator/hosts/{id}/logs [get]
func GetOrchestratorHostSystemLogs() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		logs, err := orchestratorSvc.GetHostLogs(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(logs))
		ctx.LogInfof("Successfully got host %s cached Items", id)
	}
}

//	@Summary		Streams the system logs via WebSocket
//	@Description	This endpoint streams the system logs in real-time via WebSocket
//	@Tags			Config
//	@Produce		json
//	@Success		101	"Switching Protocols to websocket"
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/logs/stream [get]
func StreamOrchestratorHostSystemLogs() restapi.ControllerHandler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		targetHostWebsocketUrl, err := orchestratorSvc.GetHostWebsocketBaseUrl(ctx, id)
		if err != nil || targetHostWebsocketUrl == "" {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		targetWebsocketUrl := fmt.Sprintf("%s/logs/stream", targetHostWebsocketUrl)
		authKey, authToken, err := orchestratorSvc.GetHostToken(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		// Upgrade the client connection to WebSocket
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ctx.LogErrorf("Failed to upgrade connection: %v", err)
			return
		}
		defer wsConn.Close()

		// Connect to the target WebSocket server
		dialer := websocket.Dialer{}
		// Disable TLS validation if configured
		cfg := config.Get()
		if cfg.DisableTlsValidation() {
			dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		remoteWs, _, err := dialer.Dial(targetWebsocketUrl, http.Header{
			authKey: {authToken},
		})
		if err != nil {
			ctx.LogErrorf("Failed to connect to remote WebSocket: %v", err)
			return
		}
		defer remoteWs.Close()

		// Channel to signal when either connection closes
		done := make(chan struct{})

		// Goroutine to copy messages from client WebSocket to remote WebSocket
		go func() {
			defer func() {
				close(done) // Signal the main routine to stop
			}()
			for {
				messageType, message, err := wsConn.ReadMessage()
				if err != nil {
					ctx.LogErrorf("Error reading from client WebSocket: %v", err)
					return
				}
				if err := remoteWs.WriteMessage(messageType, message); err != nil {
					ctx.LogErrorf("Error writing to remote WebSocket: %v", err)
					return
				}
			}
		}()

		// Main routine to copy messages from remote WebSocket to client WebSocket
		for {
			messageType, message, err := remoteWs.ReadMessage()
			if err != nil {
				ctx.LogErrorf("Error reading from remote WebSocket: %v", err)
				break
			}
			if err := wsConn.WriteMessage(messageType, message); err != nil {
				ctx.LogErrorf("Error writing to client WebSocket: %v", err)
				break
			}
		}

		// Wait for the other goroutine to finish before returning
		<-done
	}
}

//	@Summary		Create an enrollment token
//	@Description	Generates a short-lived, single-use token that allows a freshly installed agent to register itself with the orchestrator without requiring a permanent credential.
//	@Tags			Orchestrator
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateEnrollmentTokenRequest	true	"Enrollment token request"
//	@Success		201		{object}	models.CreateEnrollmentTokenResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/enrollment-token [post]
func CreateEnrollmentTokenHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var req models.CreateEnrollmentTokenRequest
		if err := http_helper.MapRequestBody(r, &req); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := req.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		db := serviceprovider.Get().JsonDatabase
		_ = db.Connect(ctx)
		token, err := db.CreateEnrollmentToken(ctx, req.HostName, req.TTLMinutes)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.CreateEnrollmentTokenResponse{
			Token:     token.Token,
			HostName:  token.HostName,
			ExpiresAt: token.ExpiresAt,
		})
		ctx.LogInfof("Enrollment token created for host %s", req.HostName)
	}
}

//	@Summary		Validate an enrollment token
//	@Description	Public endpoint that checks whether an enrollment token is valid, unused, and not expired. Used by agents before starting the registration flow.
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			token	path		string	true	"Enrollment token value"
//	@Success		200		{object}	models.ValidateEnrollmentTokenResponse
//	@Router			/v1/orchestrator/enrollment-token/{token}/validate [get]
func ValidateEnrollmentTokenHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		tokenValue := vars["token"]

		db := serviceprovider.Get().JsonDatabase
		_ = db.Connect(ctx)
		token, err := db.ValidateEnrollmentToken(ctx, tokenValue)

		resp := models.ValidateEnrollmentTokenResponse{}
		if err != nil {
			resp.Valid = false
			resp.Reason = err.Error()
		} else {
			resp.Valid = true
			resp.HostName = token.HostName
			resp.ExpiresAt = token.ExpiresAt
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

//	@Summary		Deploy and register an agent via SSH (synchronous)
//	@Description	SSHes into a remote host, installs the devops agent, and registers it with this orchestrator. Blocks until the operation completes.
//	@Tags			Orchestrator
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.DeployOrchestratorHostRequest	true	"Deploy request"
//	@Success		201		{object}	models.DeployOrchestratorHostResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/deploy [post]
// checkDuplicateDeployHost returns a non-nil error if a host with the same
// name or SSH address already exists, so both sync and async handlers can
// reject the request before doing any work.
func DeployOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var req models.DeployOrchestratorHostRequest
		if err := http_helper.MapRequestBody(r, &req); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := req.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := checkDuplicateDeployHost(ctx, req.HostName, req.SshHost); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		orchSvc := orchestrator.NewOrchestratorService(ctx)
		resp, err := orchSvc.DeployAndRegisterAgent(ctx, req, nil, nil)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
		ctx.LogInfof("Agent deployed successfully, host_id=%s", resp.HostID)
	}
}

//	@Summary		Deploy and register an agent via SSH (asynchronous)
//	@Description	SSHes into a remote host, installs the devops agent, and registers it with this orchestrator. Returns a job ID immediately; poll /jobs/{id} for status.
//	@Tags			Orchestrator
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.DeployOrchestratorHostRequest	true	"Deploy request"
//	@Success		202		{object}	models.JobResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/deploy/async [post]
func AsyncDeployOrchestratorHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var req models.DeployOrchestratorHostRequest
		if err := http_helper.MapRequestBody(r, &req); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := req.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		callerID, ok := getEffectiveCallerID(ctx)
		if !ok {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		if err := checkDuplicateDeployHost(ctx, req.HostName, req.SshHost); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		jobManager := jobs.Get(ctx)
		if jobManager == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(fmt.Errorf("job manager not available"), http.StatusInternalServerError))
			return
		}

		localJob, err := jobManager.CreateNewJob(callerID, "orchestrator", "deploy", "Deploying agent "+req.HostName)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		jobID := localJob.ID
		asyncCtx := basecontext.NewRootBaseContext()
		go func() {
			_, _ = jobManager.InitJob(jobID)
			_, _ = jobManager.UpdateJobProgress(jobID, 0, constants.JobStateRunning)
			orchSvc := orchestrator.NewOrchestratorService(asyncCtx)
			onOutput := func(line string) {
				_, _ = jobManager.UpdateJobMessage(jobID, line)
			}
			onProgress := func(pct int, msg string) {
				_, _ = jobManager.UpdateJobProgress(jobID, pct, constants.JobStateRunning)
				if msg != "" {
					_, _ = jobManager.UpdateJobMessage(jobID, msg)
				}
			}
			resp, deployErr := orchSvc.DeployAndRegisterAgent(asyncCtx, req, onOutput, onProgress)
			if deployErr != nil {
				_ = jobManager.MarkJobError(jobID, deployErr)
				return
			}
      
			deploymentMessage := fmt.Sprintf("Agent %s deployed successfully", req.HostName)
			_ = jobManager.MarkJobCompleteWithRecord(jobID, deploymentMessage, resp.HostID, "orchestrator_host")
		}()

		response := mappers.MapJobToApiJob(*localJob)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Async deploy job %s started for host %s", localJob.ID, req.HostName)
	}
}

// endregion

// region Orchestrator Async Machine Creation

//	@Summary		Creates a virtual machine in one of the orchestrator hosts asynchronously
//	@Description	This endpoint creates a virtual machine in one of the orchestrator hosts in the background and returns a Job ID to track progress
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			request	body		models.CreateVirtualMachineRequest	true	"Create Virtual Machine Request"
//	@Success		202		{object}	models.JobResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/machines/async [post]
func AsyncCreateOrchestratorVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		callerID, ok := getEffectiveCallerID(ctx)
		if !ok {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		var request models.CreateVirtualMachineRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		ownerWasEmpty := request.Owner == ""
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if ownerWasEmpty {
			request.Owner = ""
			if request.CatalogManifest != nil {
				request.CatalogManifest.Owner = ""
			}
		}

		if request.CatalogManifest != nil {
			catalogConnection, connErr := resolveCatalogMachineConnection(ctx, request.CatalogManifest)
			if connErr != nil {
				ReturnApiError(ctx, w, models.NewFromError(connErr))
				return
			}
			request.CatalogManifest.Connection = catalogConnection
			request.CatalogManifest.CatalogManagerId = ""
		}

		jobManager := jobs.Get(ctx)
		if jobManager == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(fmt.Errorf("job manager not available"), http.StatusInternalServerError))
			return
		}

		job, err := jobManager.CreateNewJob(callerID, "orchestrator", "create", "Initializing orchestrator virtual machine creation")
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		go func(jobID string, req models.CreateVirtualMachineRequest) {
			asyncCtx := basecontext.NewRootBaseContext()
			defer func() {
				if rec := recover(); rec != nil {
					asyncCtx.LogErrorf("[Orchestrator] Panic in async create goroutine for job %s: %v", jobID, rec)
					_ = jobManager.MarkJobError(jobID, fmt.Errorf("internal error: %v", rec))
				}
			}()
			_, _ = jobManager.UpdateJobProgress(jobID, 1, constants.JobStateRunning)
			orchSvc := orchestrator.NewOrchestratorService(asyncCtx)
			result, apiErr := orchSvc.CreateVirtualMachine(asyncCtx, jobID, req)
			if apiErr != nil {
				_ = jobManager.MarkJobError(jobID, fmt.Errorf("%s", apiErr.Message))
				return
			}
			if result == nil {
				// Async dispatch succeeded — HostJobEventHandler will complete the job.
				return
			}
			_ = jobManager.MarkJobCompleteWithRecord(jobID, fmt.Sprintf("Virtual machine %s created", result.ID), result.ID, "virtual_machine")
		}(job.ID, request)

		response := mappers.MapJobToApiJob(*job)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Async orchestrator machine create started, job ID: %v", response.ID)
	}
}

//	@Summary		Creates a virtual machine in a specific orchestrator host asynchronously
//	@Description	This endpoint creates a virtual machine in a specific orchestrator host in the background and returns a Job ID to track progress
//	@Tags			Orchestrator
//	@Produce		json
//	@Param			id		path		string								true	"Host ID"
//	@Param			request	body		models.CreateVirtualMachineRequest	true	"Create Virtual Machine Request"
//	@Success		202		{object}	models.JobResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/orchestrator/hosts/{id}/machines/async [post]
func AsyncCreateOrchestratorHostVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		callerID, ok := getEffectiveCallerID(ctx)
		if !ok {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		var request models.CreateVirtualMachineRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		ownerWasEmpty := request.Owner == ""
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if ownerWasEmpty {
			request.Owner = ""
			if request.CatalogManifest != nil {
				request.CatalogManifest.Owner = ""
			}
		}

		if request.CatalogManifest != nil {
			catalogConnection, connErr := resolveCatalogMachineConnection(ctx, request.CatalogManifest)
			if connErr != nil {
				ReturnApiError(ctx, w, models.NewFromError(connErr))
				return
			}
			request.CatalogManifest.Connection = catalogConnection
			request.CatalogManifest.CatalogManagerId = ""
		}

		jobManager := jobs.Get(ctx)
		if jobManager == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(fmt.Errorf("job manager not available"), http.StatusInternalServerError))
			return
		}

		job, err := jobManager.CreateNewJob(callerID, "orchestrator", "create", fmt.Sprintf("Initializing virtual machine creation on host %s", id))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		go func(jobID string, hostID string, req models.CreateVirtualMachineRequest) {
			asyncCtx := basecontext.NewRootBaseContext()
			defer func() {
				if rec := recover(); rec != nil {
					asyncCtx.LogErrorf("[Orchestrator] Panic in async host create goroutine for job %s on host %s: %v", jobID, hostID, rec)
					_ = jobManager.MarkJobError(jobID, fmt.Errorf("internal error: %v", rec))
				}
			}()
			_, _ = jobManager.UpdateJobProgress(jobID, 1, constants.JobStateRunning)
			orchSvc := orchestrator.NewOrchestratorService(asyncCtx)
			result, apiErr := orchSvc.CreateHosVirtualMachine(asyncCtx, jobID, hostID, req)
			if apiErr != nil {
				_ = jobManager.MarkJobError(jobID, fmt.Errorf("%s", apiErr.Message))
				return
			}
			if result == nil {
				// Async dispatch succeeded — HostJobEventHandler will complete the job.
				return
			}
			_ = jobManager.MarkJobCompleteWithRecord(jobID, fmt.Sprintf("Virtual machine %s created on host %s", result.ID, hostID), result.ID, "virtual_machine")
		}(job.ID, id, request)

		response := mappers.MapJobToApiJob(*job)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Async orchestrator host machine create started on host %s, job ID: %v", id, response.ID)
	}
}
// endregion Orchestrator Async Machine Creation

func checkDuplicateDeployHost(ctx basecontext.ApiContext, hostName, sshHost string) error {
	db := serviceprovider.Get().JsonDatabase
	_ = db.Connect(ctx)
	existing, err := db.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to check existing hosts: %w", err)
	}
	for _, h := range existing {
		if strings.EqualFold(h.Description, hostName) {
			return fmt.Errorf("a host with the name %q already exists (id: %s)", hostName, h.ID)
		}
		if strings.EqualFold(h.Host, sshHost) {
			return fmt.Errorf("a host with the address %q already exists (id: %s)", sshHost, h.ID)
		}
	}
	return nil
}
