package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/reverse_proxy"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerReverseProxyHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s ReverseProxy handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).WithPath("/reverse-proxy").
		WithRequiredClaim(constants.CONFIGURE_REVERSE_PROXY_CLAIM).
		WithHandler(GetReverseProxyConfigHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).WithPath("/reverse-proxy/hosts").
		WithRequiredClaim(constants.LIST_REVERSE_PROXY_HOSTS_CLAIM).
		WithHandler(GetReverseProxyHostsHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).WithPath("/reverse-proxy/hosts").
		WithRequiredClaim(constants.CREATE_REVERSE_PROXY_HOST_CLAIM).
		WithHandler(CreateReverseProxyHostHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).WithPath("/reverse-proxy/hosts/{id}").
		WithRequiredClaim(constants.LIST_REVERSE_PROXY_HOSTS_CLAIM).
		WithHandler(GetReverseProxyHostHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).WithPath("/reverse-proxy/hosts/{id}").
		WithRequiredClaim(constants.UPDATE_REVERSE_PROXY_HOST_CLAIM).
		WithHandler(UpdateReverseProxyHostHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).WithPath("/reverse-proxy/hosts/{id}").
		WithRequiredClaim(constants.DELETE_REVERSE_PROXY_HOST_CLAIM).
		WithHandler(DeleteReverseProxyHostHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).WithPath("/reverse-proxy/hosts/{id}/http_routes").
		WithRequiredClaim(constants.UPDATE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM).
		WithHandler(UpsertReverseProxyHostHttpRouteHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).WithPath("/reverse-proxy/hosts/{id}/http_routes/{http_route_id}").
		WithRequiredClaim(constants.DELETE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM).
		WithHandler(DeleteReverseProxyHostHttpRoutesHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).WithPath("/reverse-proxy/hosts/{id}/http_routes/order").
		WithRequiredClaim(constants.UPDATE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM).
		WithHandler(UpdateReverseProxyHostHttpRouteOrderHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).WithPath("/reverse-proxy/hosts/{id}/tcp_route").
		WithRequiredClaim(constants.UPDATE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM).
		WithHandler(UpdateReverseProxyHostTcpRouteHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).WithPath("/reverse-proxy/restart").
		WithRequiredClaim(constants.CONFIGURE_REVERSE_PROXY_CLAIM).
		WithHandler(RestartReverseProxyHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).WithPath("/reverse-proxy/enable").
		WithRequiredClaim(constants.CONFIGURE_REVERSE_PROXY_CLAIM).
		WithHandler(EnableReverseProxyHandler()).
		Register()
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).WithPath("/reverse-proxy/disable").
		WithRequiredClaim(constants.CONFIGURE_REVERSE_PROXY_CLAIM).
		WithHandler(DisableReverseProxyHandler()).
		Register()
}

func enrichHostWithVmDetails(ctx basecontext.ApiContext, host *models.ReverseProxyHost) {
	provider := serviceprovider.Get()
	svc := provider.ParallelsDesktopService

	if host.TcpRoute != nil && host.TcpRoute.TargetVmId != "" {
		vm, err := svc.GetVmSync(ctx, host.TcpRoute.TargetVmId)
		if err == nil && vm != nil {
			host.TcpRoute.TargetVmDetails = &models.ReverseProxyRouteVmDetails{
				Name:                  vm.Name,
				State:                 vm.State,
				OS:                    vm.OS,
				Uptime:                vm.Uptime,
				GuestToolsState:       vm.GuestTools.State,
				GuestToolsVersion:     vm.GuestTools.Version,
				InternalIpAddress:     vm.InternalIpAddress,
				HostExternalIpAddress: vm.HostExternalIpAddress,
			}
		}
	}

	for _, route := range host.HttpRoutes {
		if route.TargetVmId != "" {
			vm, err := svc.GetVmSync(ctx, route.TargetVmId)
			if err == nil && vm != nil {
				route.TargetVmDetails = &models.ReverseProxyRouteVmDetails{
					Name:                  vm.Name,
					State:                 vm.State,
					OS:                    vm.OS,
					Uptime:                vm.Uptime,
					GuestToolsState:       vm.GuestTools.State,
					GuestToolsVersion:     vm.GuestTools.Version,
					InternalIpAddress:     vm.InternalIpAddress,
					HostExternalIpAddress: vm.HostExternalIpAddress,
				}
			}
		}
	}
}

// @Summary		Gets reverse proxy configuration
// @Description	This endpoint returns the reverse proxy configuration
// @Tags			ReverseProxy
// @Produce		json
// @Success		200	{object}	models.ReverseProxy
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy [get]
func GetReverseProxyConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)

		defer Recover(ctx, r, w)

		config := reverse_proxy.GetConfig()
		if !config.Enabled {
			config.Host = ""
			config.Port = ""
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(config)
		ctx.LogInfof("Reverse Proxy Config returned successfully")
	}
}

// @Summary		Gets all the reverse proxy hosts
// @Description	This endpoint returns all the reverse proxy hosts
// @Tags			ReverseProxy
// @Produce		json
// @Success		200	{object}	[]models.ReverseProxyHost
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts [get]
func GetReverseProxyHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getReverseProxyHostsDiag := errors.NewDiagnostics("/reverse-proxy/hosts")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			getReverseProxyHostsDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getReverseProxyHostsDiag, http.StatusBadRequest))
			return
		}

		dtoRpHosts, err := dbService.GetReverseProxyHosts(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := mappers.DtoReverseProxyHostsToApi(dtoRpHosts)

		for i := range result {
			enrichHostWithVmDetails(ctx, &result[i])
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Claims returned successfully")
	}
}

// @Summary		Gets all the reverse proxy host
// @Description	This endpoint returns a reverse proxy host
// @Tags			ReverseProxy
// @Produce		json
// @Param			id	path		string	true	"Reverse Proxy Host ID"
// @Success		200	{object}	models.ReverseProxyHost
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}  [get]
func GetReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := vars["id"]
		getReverseProxyHostDiag := errors.NewDiagnostics("/reverse-proxy/hosts/" + id)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			getReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getReverseProxyHostDiag, http.StatusBadRequest))
			return
		}

		dtoRpHost, err := dbService.GetReverseProxyHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := mappers.DtoReverseProxyHostToApi(*dtoRpHost)

		enrichHostWithVmDetails(ctx, &result)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Claims returned successfully")
	}
}

// @Summary		Creates a reverse proxy host
// @Description	This endpoint creates a reverse proxy host
// @Tags			ReverseProxy
// @Produce		json
// @Param			reverse_proxy_create_request	body		models.ReverseProxyHostCreateRequest	true	"Reverse Host Request"
// @Success		200								{object}	models.ReverseProxyHost
// @Failure		400								{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401								{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts [post]
func CreateReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		createReverseProxyHostDiag := errors.NewDiagnostics("/reverse-proxy/hosts")
		var request models.ReverseProxyHostCreateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			createReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createReverseProxyHostDiag, http.StatusBadRequest))
			return
		}
		request.Validate(createReverseProxyHostDiag)
		if createReverseProxyHostDiag.HasErrors() {
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createReverseProxyHostDiag, http.StatusBadRequest))
			return
		}

		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			createReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createReverseProxyHostDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoHost := mappers.ApiCreateRequestReverseProxyHostToDto(request)

		resultDto, err := dbService.CreateReverseProxyHost(ctx, dtoHost)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoReverseProxyHostToApi(*resultDto)
		enrichHostWithVmDetails(ctx, &response)

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Reverse Proxy Host created successfully")
	}
}

// @Summary		Updates a reverse proxy host
// @Description	This endpoint creates a reverse proxy host
// @Tags			ReverseProxy
// @Produce		json
// @Param			reverse_proxy_update_request	body		models.ReverseProxyHostUpdateRequest	true	"Reverse Host Request"
// @Success		200								{object}	models.ReverseProxyHost
// @Failure		400								{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401								{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id} [put]
func UpdateReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.ReverseProxyHostUpdateRequest
		vars := mux.Vars(r)
		id := vars["id"]
		updateReverseProxyHostDiag := errors.NewDiagnostics("/reverse-proxy/hosts/" + id)
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			updateReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostDiag, http.StatusBadRequest))
			return
		}
		request.Validate(updateReverseProxyHostDiag)
		if updateReverseProxyHostDiag.HasErrors() {
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostDiag, http.StatusBadRequest))
			return
		}
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			updateReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoHost, err := dbService.GetReverseProxyHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		if dtoHost == nil {
			updateReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusNotFound), "reverse proxy host not found", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostDiag, http.StatusNotFound))
			return
		}

		if dtoHost.TcpRoute != nil {
			if request.Cors != nil {
				updateReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "cors is not allowed for tcp route", "")
				ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostDiag, http.StatusBadRequest))
				return
			}
			if request.Tls != nil {
				updateReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "tls is not allowed for tcp route", "")
				ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostDiag, http.StatusBadRequest))
				return
			}
		}

		mappedDtoHost := mappers.ApiUpdateRequestReverseProxyHostToDto(request)
		mappedDtoHost.ID = id

		resultDto, err := dbService.UpdateReverseProxyHost(ctx, &mappedDtoHost)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			rsp := models.NewFromError(err)
			updateReverseProxyHostDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostDiag, http.StatusInternalServerError))
			return
		}

		response := mappers.DtoReverseProxyHostToApi(*resultDto)
		enrichHostWithVmDetails(ctx, &response)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Reverse Proxy Host updated successfully")
	}
}

// @Summary		Delete a a reverse proxy host
// @Description	This endpoint Deletes a reverse proxy host
// @Tags			ReverseProxy
// @Produce		json
// @Param			id	path	string	true	"Reverse Proxy Host ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id} [delete]
func DeleteReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := vars["id"]
		deleteReverseProxyHostDiag := errors.NewDiagnostics("/reverse-proxy/hosts/" + id)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			deleteReverseProxyHostDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteReverseProxyHostDiag, http.StatusBadRequest))
			return
		}

		err = dbService.DeleteReverseProxyHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			rsp := models.NewFromError(err)
			deleteReverseProxyHostDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteReverseProxyHostDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Reverse Proxy Host deleted successfully")
	}
}

// @Summary		Creates or updates a reverse proxy host HTTP route
// @Description	This endpoint creates or updates a reverse proxy host HTTP route
// @Tags			ReverseProxy
// @Produce		json
// @Param			reverse_proxy_http_route_request	body		models.ReverseProxyHostHttpRouteCreateRequest	true	"Reverse Host Request"
// @Success		200									{object}	models.ReverseProxyHost
// @Failure		400									{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401									{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}/http_routes [post]
func UpsertReverseProxyHostHttpRouteHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := vars["id"]
		upsertReverseProxyHostHttpRouteDiag := errors.NewDiagnostics("/reverse-proxy/hosts/" + id + "/http_routes")
		var request models.ReverseProxyHostHttpRouteCreateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			upsertReverseProxyHostHttpRouteDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(upsertReverseProxyHostHttpRouteDiag, http.StatusBadRequest))
			return
		}
		request.Validate(upsertReverseProxyHostHttpRouteDiag)
		if upsertReverseProxyHostHttpRouteDiag.HasErrors() {
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(upsertReverseProxyHostHttpRouteDiag, http.StatusBadRequest))
			return
		}
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			upsertReverseProxyHostHttpRouteDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(upsertReverseProxyHostHttpRouteDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoHost, err := dbService.GetReverseProxyHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if dtoHost == nil {
			upsertReverseProxyHostHttpRouteDiag.AddError(strconv.Itoa(http.StatusNotFound), "reverse proxy host not found", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(upsertReverseProxyHostHttpRouteDiag, http.StatusNotFound))
			return
		}

		if dtoHost.TcpRoute != nil {
			upsertReverseProxyHostHttpRouteDiag.AddError(strconv.Itoa(http.StatusBadRequest), "cannot update reverse proxy HTTP route when TCP routes are present", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(upsertReverseProxyHostHttpRouteDiag, http.StatusBadRequest))
			return
		}

		httpRouteID := ""
		for i, httpRoute := range dtoHost.HttpRoutes {
			if strings.EqualFold(httpRoute.GetRoute(), request.GetRoute()) {
				httpRouteID = dtoHost.HttpRoutes[i].ID
				break
			}
		}

		dtoUpsertHttpRoute := mappers.ApiReverseProxyHostCreateHttpRouteToDto(request)
		if request.TargetVmId != "" {
			dtoUpsertHttpRoute.TargetHost = ""
		} else {
			dtoUpsertHttpRoute.TargetVmId = ""
		}
		var resultDto *data_models.ReverseProxyHostHttpRoute
		var resultErr error
		if httpRouteID != "" {
			dtoUpsertHttpRoute.ID = httpRouteID
			resultDto, resultErr = dbService.UpdateReverseProxyHostHttpRoute(ctx, id, dtoUpsertHttpRoute)
		} else {
			resultDto, resultErr = dbService.CreateReverseProxyHostHttpRoute(ctx, id, dtoUpsertHttpRoute)
		}

		if resultErr != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		if resultDto == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy host http route was not upserted successfully"), http.StatusNotFound))
			return
		}

		dtoHost, _ = dbService.GetReverseProxyHost(ctx, id)
		response := mappers.DtoReverseProxyHostToApi(*dtoHost)
		enrichHostWithVmDetails(ctx, &response)

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			rsp := models.NewFromError(err)
			upsertReverseProxyHostHttpRouteDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(upsertReverseProxyHostHttpRouteDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Reverse Proxy Host Http Route upserted successfully")
	}
}

// @Summary		Delete a a reverse proxy host HTTP route
// @Description	This endpoint Deletes a reverse proxy host HTTP route
// @Tags			ReverseProxy
// @Produce		json
// @Param			id				path	string	true	"Reverse Proxy Host ID"
// @Param			http_route_id	path	string	true	"Reverse Proxy Host HTTP Route ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id} [delete]
func DeleteReverseProxyHostHttpRoutesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := vars["id"]
		httpRouteID := vars["http_route_id"]
		deleteReverseProxyHostHttpRouteDiag := errors.NewDiagnostics("/reverse-proxy/hosts/" + id + "/http_routes/" + httpRouteID)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			deleteReverseProxyHostHttpRouteDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteReverseProxyHostHttpRouteDiag, http.StatusBadRequest))
			return
		}

		err = dbService.DeleteReverseProxyHostHttpRoute(ctx, id, httpRouteID)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			rsp := models.NewFromError(err)
			deleteReverseProxyHostHttpRouteDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteReverseProxyHostHttpRouteDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Reverse Proxy Host http route deleted successfully")
	}
}

// @Summary		Updates the order of a reverse proxy host HTTP route
// @Description	This endpoint reorders HTTP routes for a reverse proxy host
// @Tags			ReverseProxy
// @Produce		json
// @Param			id											path		string											true	"Reverse Proxy Host ID"
// @Param			reverse_proxy_http_route_reorder_request	body		models.ReverseProxyHostHttpRouteReorderRequest	true	"Reverse Proxy Host HTTP Route Reorder Request"
// @Success		200											{object}	models.ReverseProxyHost
// @Failure		400											{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401											{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}/http_routes/order [put]
func UpdateReverseProxyHostHttpRouteOrderHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := vars["id"]
		updateReverseProxyHostHttpRouteOrderDiag := errors.NewDiagnostics("/reverse-proxy/hosts/" + id + "/http_routes/order")
		var request models.ReverseProxyHostHttpRouteReorderRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			updateReverseProxyHostHttpRouteOrderDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostHttpRouteOrderDiag, http.StatusBadRequest))
			return
		}
		request.Validate(updateReverseProxyHostHttpRouteOrderDiag)
		if updateReverseProxyHostHttpRouteOrderDiag.HasErrors() {
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostHttpRouteOrderDiag, http.StatusBadRequest))
			return
		}

		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			updateReverseProxyHostHttpRouteOrderDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostHttpRouteOrderDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		resultDto, err := dbService.ReorderReverseProxyHostHttpRoute(ctx, id, request.ID, request.Order)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			rsp := models.NewFromError(err)
			updateReverseProxyHostHttpRouteOrderDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostHttpRouteOrderDiag, http.StatusInternalServerError))
			return
		}

		go rps.BroadcastHostUpdated(id)

		response := mappers.DtoReverseProxyHostToApi(*resultDto)
		enrichHostWithVmDetails(ctx, &response)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Reverse Proxy Host HTTP route order updated successfully")
	}
}

// @Summary		Updates a reverse proxy host TCP route
// @Description	This endpoint updates a reverse proxy host TCP route
// @Tags			ReverseProxy
// @Produce		json
// @Param			reverse_proxy_tcp_route_request	body		models.ReverseProxyHostTcpRouteCreateRequest	true	"Reverse Host Request"
// @Success		200								{object}	models.ReverseProxyHost
// @Failure		400								{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401								{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}/tcp_route [post]
func UpdateReverseProxyHostTcpRouteHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := vars["id"]
		updateReverseProxyHostTcpRouteDiag := errors.NewDiagnostics("/reverse-proxy/hosts/" + id + "/tcp_routes")
		var request models.ReverseProxyHostTcpRouteCreateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			updateReverseProxyHostTcpRouteDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostTcpRouteDiag, http.StatusBadRequest))
			return
		}
		request.Validate(updateReverseProxyHostTcpRouteDiag)
		if updateReverseProxyHostTcpRouteDiag.HasErrors() {
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostTcpRouteDiag, http.StatusBadRequest))
			return
		}
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			updateReverseProxyHostTcpRouteDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostTcpRouteDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoHost, err := dbService.GetReverseProxyHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if dtoHost == nil {
			updateReverseProxyHostTcpRouteDiag.AddError(strconv.Itoa(http.StatusNotFound), "reverse proxy host not found", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostTcpRouteDiag, http.StatusNotFound))
			return
		}

		if dtoHost.HttpRoutes != nil {
			updateReverseProxyHostTcpRouteDiag.AddError(strconv.Itoa(http.StatusBadRequest), "cannot update reverse proxy TCP route when HTTP routes are present", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostTcpRouteDiag, http.StatusBadRequest))
			return
		}

		dtoUpdateTcpRoute := mappers.ApiReverseProxyHostTcpRouteCreateRequestToDto(request)
		if request.TargetVmId != "" {
			dtoUpdateTcpRoute.TargetHost = ""
		} else {
			dtoUpdateTcpRoute.TargetVmId = ""
		}
		resultDto, resultErr := dbService.UpdateReverseProxyHostTcpRoute(ctx, id, dtoUpdateTcpRoute)

		if resultErr != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		if resultDto == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy host tcp route was not updated successfully"), http.StatusNotFound))
			return
		}

		dtoHost, _ = dbService.GetReverseProxyHost(ctx, id)
		response := mappers.DtoReverseProxyHostToApi(*dtoHost)
		enrichHostWithVmDetails(ctx, &response)

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			rsp := models.NewFromError(err)
			updateReverseProxyHostTcpRouteDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateReverseProxyHostTcpRouteDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Reverse Proxy Host TCP Route upserted successfully")
	}
}

// @Summary		Restarts the reverse proxy
// @Description	This endpoint will restart the reverse proxy
// @Tags			ReverseProxy
// @Produce		json
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/restart [put]
func RestartReverseProxyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		restartReverseProxyDiag := errors.NewDiagnostics("/reverse-proxy/restart")
		rpConfig := reverse_proxy.GetConfig()
		if !rpConfig.Enabled {
			restartReverseProxyDiag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy is disabled", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(restartReverseProxyDiag, http.StatusBadRequest))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			rsp := models.NewFromError(err)
			restartReverseProxyDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(restartReverseProxyDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Reverse Proxy Config returned successfully")
	}
}

// @Summary		Enable the reverse proxy
// @Description	This endpoint will enable the reverse proxy
// @Tags			ReverseProxy
// @Produce		json
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/enable [put]
func EnableReverseProxyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		cfg := config.Get()
		defer Recover(ctx, r, w)
		enableReverseProxyDiag := errors.NewDiagnostics("/reverse-proxy/enable")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		rps := reverse_proxy.Get(ctx)
		rpConfig := reverse_proxy.GetConfig()
		wasEnabled := rpConfig.Enabled

		if _, err := dbService.EnableProxyConfig(ctx); err != nil {
			if sysErr, ok := err.(*errors.SystemError); ok && sysErr.Code() == http.StatusNotFound {
				// Entity doesn't exist, create it and set Enabled to true
				dto := data_models.ReverseProxy{
					ID:      helpers.GenerateId(),
					Enabled: true,
					Host:    cfg.ReverseProxyHost(),
					Port:    cfg.ReverseProxyPort(),
				}
				if _, err := dbService.UpdateReverseProxy(ctx, dto); err != nil {
					ReturnApiError(ctx, w, models.NewFromError(err))
					return
				}
			} else {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		}

		if wasEnabled {
			if err := rps.Restart(); err != nil {
				rsp := models.NewFromError(err)
				enableReverseProxyDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Restart")
				ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(enableReverseProxyDiag, http.StatusInternalServerError))
				return
			}
		} else {
			go func() {
				if err := rps.Start(); err != nil {
					ctx.LogErrorf("Error starting reverse proxy service: %v", err)
				}
			}()
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Reverse Proxy Config returned successfully")
	}
}

// @Summary		Disable the reverse proxy
// @Description	This endpoint will disable the reverse proxy
// @Tags			ReverseProxy
// @Produce		json
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/disable [put]
func DisableReverseProxyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		rps := reverse_proxy.Get(ctx)
		rpConfig := reverse_proxy.GetConfig()

		if rpConfig.Enabled {
			_ = rps.Stop()
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		if _, err := dbService.DisableProxyConfig(ctx); err != nil {
			// Ignore 404 because if it doesn't exist, it's already "disabled" effectively
			if sysErr, ok := err.(*errors.SystemError); !ok || sysErr.Code() != http.StatusNotFound {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Reverse Proxy Config returned successfully")
	}
}
