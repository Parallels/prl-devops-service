package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
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

// @Summary		Gets reverse proxy configuration
// @Description	This endpoint returns the reverse proxy configuration
// @Tags			ReverseProxy
// @Produce		json
// @Success		200	{object}	[]models.ReverseProxy
// @Failure		400	{object}	models.ApiErrorResponse
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
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts [get]
func GetReverseProxyHostsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		dtoRpHosts, err := dbService.GetReverseProxyHosts(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(dtoRpHosts) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.ReverseProxyHost, 0)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Reverse Proxy Hosts returned: %v", len(response))
			return
		}

		result := mappers.DtoReverseProxyHostsToApi(dtoRpHosts)

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
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}  [get]
func GetReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoRpHost, err := dbService.GetReverseProxyHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := mappers.DtoReverseProxyHostToApi(*dtoRpHost)

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
// @Failure		400								{object}	models.ApiErrorResponse
// @Failure		401								{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts [post]
func CreateReverseProxyHostHandler() restapi.ControllerHandler {
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

		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
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
// @Failure		400								{object}	models.ApiErrorResponse
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
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

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
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy host not found"), http.StatusNotFound))
			return
		}

		if dtoHost.TcpRoute != nil {
			if request.Cors != nil {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("cors is not allowed for tcp route"), http.StatusBadRequest))
				return
			}
			if request.Tls != nil {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("tls is not allowed for tcp route"), http.StatusBadRequest))
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
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoReverseProxyHostToApi(*resultDto)

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
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id} [delete]
func DeleteReverseProxyHostHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteReverseProxyHost(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
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
// @Failure		400									{object}	models.ApiErrorResponse
// @Failure		401									{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}/http_route [post]
func UpsertReverseProxyHostHttpRouteHandler() restapi.ControllerHandler {
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
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

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
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy host not found"), http.StatusNotFound))
			return
		}

		if dtoHost.TcpRoute != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("cannot update reverse proxy HTTP route when TCP routes are present"), http.StatusBadRequest))
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

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
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
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id} [delete]
func DeleteReverseProxyHostHttpRoutesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		httpRouteID := vars["http_route_id"]

		err = dbService.DeleteReverseProxyHostHttpRoute(ctx, id, httpRouteID)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Reverse Proxy Host http route deleted successfully")
	}
}

// @Summary		Updates a reverse proxy host TCP route
// @Description	This endpoint updates a reverse proxy host TCP route
// @Tags			ReverseProxy
// @Produce		json
// @Param			reverse_proxy_tcp_route_request	body		models.ReverseProxyHostTcpRouteCreateRequest	true	"Reverse Host Request"
// @Success		200								{object}	models.ReverseProxyHost
// @Failure		400								{object}	models.ApiErrorResponse
// @Failure		401								{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/hosts/{id}/http_routes [post]
func UpdateReverseProxyHostTcpRouteHandler() restapi.ControllerHandler {
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
		cfg := config.Get()
		if !cfg.IsReverseProxyEnabled() {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

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
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy host not found"), http.StatusNotFound))
			return
		}

		if dtoHost.HttpRoutes != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("cannot update reverse proxy TCP route when HTTP routes are present"), http.StatusBadRequest))
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

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
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
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/reverse-proxy/restart [put]
func RestartReverseProxyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		rpConfig := reverse_proxy.GetConfig()
		if !rpConfig.Enabled {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("reverse proxy is disabled"), http.StatusBadRequest))
			return
		}

		rps := reverse_proxy.Get(ctx)
		if err := rps.Restart(); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
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
// @Failure		400	{object}	models.ApiErrorResponse
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

		rps := reverse_proxy.Get(ctx)
		if cfg.IsReverseProxyEnabled() {
			if err := rps.Restart(); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			go func() {
				if err := rps.Start(); err != nil {
					ctx.LogErrorf("Error starting reverse proxy service: %v", err)
				}
			}()
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		if _, err := dbService.EnableProxyConfig(ctx); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		cfg.EnableReverseProxy(true)
		cfg.Save()

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
		cfg := config.Get()
		defer Recover(ctx, r, w)

		rps := reverse_proxy.Get(ctx)
		if cfg.IsReverseProxyEnabled() {
			_ = rps.Stop()
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		if _, err := dbService.DisableProxyConfig(ctx); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		cfg.EnableReverseProxy(false)
		cfg.Save()

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Reverse Proxy Config returned successfully")
	}
}
