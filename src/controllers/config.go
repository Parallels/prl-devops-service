package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/logs"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"

	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/websocket"
)

func registerConfigHandlers(ctx basecontext.ApiContext, version string) {
	provider := serviceprovider.Get()

	ctx.LogInfof("Registering version %s config handlers", version)
	if provider.System.GetOperatingSystem() == "macos" {
		restapi.NewController().
			WithMethod(restapi.POST).
			WithVersion(version).
			WithPath("/config/tools/install").
			WithHandler(Install3rdPartyToolsHandler()).
			Register()

		restapi.NewController().
			WithMethod(restapi.POST).
			WithVersion(version).
			WithPath("/config/tools/uninstall").
			WithHandler(Uninstall3rdPartyToolsHandler()).
			Register()
	}

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/config/tools/restart").
		WithHandler(RestartApiHandler()).
		Register()

	if provider.IsParallelsDesktopAvailable() {
		restapi.NewController().
			WithMethod(restapi.GET).
			WithVersion(version).
			WithPath("/config/parallels-desktop/license").
			WithHandler(GetParallelsDesktopLicenseHandler()).
			Register()
	}

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/config/hardware").
		WithRequiredClaim(constants.LIST_CLAIM).
		WithHandler(GetHardwareInfo()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/health/system").
		WithHandler(GetSystemHealth()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/logs/stream").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(StreamSystemLogs()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/logs").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(GetSystemLogs()).
		Register()
}

// @Summary		Gets Parallels Desktop active license
// @Description	This endpoint returns Parallels Desktop active license
// @Tags			Config
// @Produce		json
// @Success		200	{object}	models.ParallelsDesktopLicense
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/parallels_desktop/key [get]
func GetParallelsDesktopLicenseHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		if provider.ParallelsDesktopService == nil || !provider.ParallelsDesktopService.Installed() {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Parallels Desktop is not installed",
				Code:    http.StatusNotFound,
			})
			return
		}

		license, err := provider.ParallelsDesktopService.GetLicense()
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(license)
		ctx.LogInfof("Parallels Desktop License returned successfully")
	}
}

// @Summary		Installs API requires 3rd party tools
// @Description	This endpoint installs API requires 3rd party tools
// @Tags			Config
// @Produce		json
// @Param			installToolsRequest	body		models.InstallToolsRequest	true	"Install Tools Request"
// @Success		200					{object}	models.InstallToolsResponse
// @Failure		400					{object}	models.ApiErrorResponse
// @Failure		401					{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/config/tools/install [post]
func Install3rdPartyToolsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.InstallToolsRequest
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

		response := models.InstallToolsResponse{
			Success:        true,
			InstalledTools: make(map[string]models.InstallToolsResponseItem),
		}

		provider := serviceprovider.Get()
		if request.All {
			provider.InstallAllTools(request.RunAs, map[string]string{})
		} else {
			for tool, option := range request.Tools {
				result := provider.InstallTool(request.RunAs, tool, option.Version, option.Flags)
				modelResponse := models.InstallToolsResponseItem{
					Success: result.Result,
					Version: result.Version,
				}
				if !result.Result {
					modelResponse.ErrorMessage = result.Message
				}
				response.InstalledTools[tool] = modelResponse
			}
		}

		// Restarting the API Service
		restapi.Get().Restart()

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Tools install request successfully")
	}
}

// @Summary		Uninstalls API requires 3rd party tools
// @Description	This endpoint uninstalls API requires 3rd party tools
// @Tags			Config
// @Produce		json
// @Param			uninstallToolsRequest	body		models.UninstallToolsRequest	true	"Uninstall Tools Request"
// @Success		200						{object}	models.InstallToolsResponse
// @Failure		400						{object}	models.ApiErrorResponse
// @Failure		401						{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/config/tools/uninstall [post]
func Uninstall3rdPartyToolsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.UninstallToolsRequest
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

		response := models.UninstallToolsResponse{
			Success:          true,
			UninstalledTools: make(map[string]models.UninstallToolsResponseItem),
		}

		provider := serviceprovider.Get()
		if request.All {
			provider.UninstallAllTools(request.RunAs, request.UninstallDependencies, map[string]string{})
		} else {
			for tool, option := range request.Tools {
				result := provider.UninstallTool(request.RunAs, tool, request.UninstallDependencies, option.Flags)
				modelResponse := models.UninstallToolsResponseItem{
					Success: result.Result,
				}
				if !result.Result {
					modelResponse.ErrorMessage = result.Message
				}
				response.UninstalledTools[tool] = modelResponse
			}
		}

		restapi.Get().Restart()

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Tools uninstall request successfully")
	}
}

// @Summary		Restarts the API Service
// @Description	This endpoint restarts the API Service
// @Tags			Config
// @Produce		json
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/config/tools/restart [post]
func RestartApiHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		go restapi.Get().Restart()
		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Restart request accepted")
	}
}

// @Summary		Gets the Hardware Info
// @Description	This endpoint returns the Hardware Info
// @Tags			Config
// @Produce		json
// @Success		200	{object}	models.SystemUsageResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/config/hardware [get]
func GetHardwareInfo() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		cfg := config.Get()
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		os := system.Get().GetOperatingSystem()
		var hardwareInfo *models.SystemUsageResponse
		var err error
		if os == "macos" {
			hardwareInfo, err = provider.ParallelsDesktopService.GetHardwareUsage(ctx)
		} else {
			hardwareInfo, err = provider.System.GetHardwareUsage(ctx)
		}

		hardwareInfo.EnabledModules = cfg.GetEnabledModules()

		if cfg.IsReverseProxyEnabled() {
			hardwareInfo.IsReverseProxyEnabled = true
			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
				if reverseProxy, err := dbService.GetReverseProxyConfig(ctx); err == nil {
					hardwareInfo.ReverseProxy = &models.SystemReverseProxy{
						Enabled: true,
						Host:    reverseProxy.Host,
						Port:    reverseProxy.Port,
					}
					if hosts, err := dbService.GetReverseProxyHosts(ctx, ""); err == nil {
						if len(hosts) > 0 {
							apiHosts := mappers.DtoReverseProxyHostsToApi(hosts)
							hardwareInfo.ReverseProxy.Hosts = apiHosts
						}
					}
				}
			}
		} else {
			hardwareInfo.IsReverseProxyEnabled = false
		}
		if hardwareInfo.OsName == "macOS" {
			hardwareInfo.IsLogStreamingEnabled = true
		}

		if err != nil || hardwareInfo == nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(hardwareInfo)
	}
}

// @Summary		Gets the API System Health
// @Description	This endpoint returns the API Health Probe
// @Tags			Config
// @Produce		json
// @Param			full	query		bool	false	"Full Health Check"
// @Success		200		{object}	models.ServiceHealthCheck
// @Router			/health/system [get]
func GetSystemHealth() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		provider := serviceprovider.Get()
		result := models.ApiHealthCheck{}

		pdService := models.ServiceHealthCheck{
			Healthy: true,
			Name:    "Parallels Desktop Service",
		}

		_, err := provider.ParallelsDesktopService.GetInfo()
		if err != nil {
			pdService.Healthy = false
			pdService.ErrorMessage = err.Error()
		}
		result.Services = append(result.Services, pdService)

		fullHealthCheck := http_helper.GetHttpRequestBoolValue(r, "full", false)
		if fullHealthCheck {
			packerService := models.ServiceHealthCheck{
				Healthy: true,
				Name:    "Packer Service",
			}
			if version := provider.PackerService.Version(); version == "unknown" {
				packerService.Healthy = false
				packerService.ErrorMessage = "Packer Service not installed"
			} else {
				packerService.Message = version
			}
			result.Services = append(result.Services, packerService)

			vagrantService := models.ServiceHealthCheck{
				Healthy: true,
				Name:    "Vagrant Service",
			}
			if version := provider.VagrantService.Version(); version == "unknown" {
				vagrantService.Healthy = false
				vagrantService.ErrorMessage = "Packer Service not installed"
			} else {
				vagrantService.Message = version
			}
			result.Services = append(result.Services, vagrantService)

			gitService := models.ServiceHealthCheck{
				Healthy: true,
				Name:    "Git Service",
			}
			if version := provider.VagrantService.Version(); version == "unknown" {
				gitService.Healthy = false
				gitService.ErrorMessage = "Git Service not installed"
			} else {
				gitService.Message = version
			}
			result.Services = append(result.Services, gitService)
		}

		healthy, message := result.GetHealthStatus()
		result.Healthy = healthy
		if !result.Healthy {
			result.ErrorMessage = message
		} else {
			result.Message = message
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	}
}

// @Summary		Gets the system logs from the disk
// @Description	This endpoint returns the system logs from the disk
// @Tags			Config
// @Produce		plain
// @Success		200
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/logs [get]
func GetSystemLogs() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		cfg := config.Get()

		// Checking if we have the logs to file enabled so we can read the logs
		if !cfg.GetBoolKey(constants.LOG_TO_FILE_ENV_VAR) {
			err := errors.New("logs to file is not enabled, we cannot read the logs")
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Failed to read log file: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		logFile := logs.GetLogFilePath(ctx)
		if logFile == "" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Failed to read log file: log file path is empty",
				Code:    http.StatusBadRequest,
			})
			return
		}

		content, err := os.ReadFile(logFile)
		if err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Failed to read log file %s: %v", logFile, err.Error()),
				Code:    http.StatusBadRequest,
			})
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// @Summary		Streams the system logs via WebSocket
// @Description	This endpoint streams the system logs in real-time via WebSocket
// @Tags			Config
// @Produce		json
// @Success		101	"Switching Protocols to websocket"
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/logs/stream [get]
func StreamSystemLogs() restapi.ControllerHandler {
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

		// Upgrade HTTP connection to WebSocket
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ctx.LogErrorf("Failed to upgrade connection to WebSocket: %v", err)
			return
		}
		defer ws.Close()

		// Create a done channel to signal when to stop
		done := make(chan struct{})
		// Start a goroutine to handle client messages and disconnection
		go func() {
			defer close(done)
			for {
				_, _, err := ws.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						if err.Error() != "websocket: close 1000 (normal)" {
							ctx.LogErrorf("WebSocket read error: %v", err)
						}
					}
					return
				}
			}
		}()

		subscriptionId := "logsub"
		onMessage := func(msg log.LogMessage) {
			err = ws.WriteMessage(websocket.TextMessage, []byte(msg.Message))
			if err != nil {
				if err.Error() != "websocket: close sent" {
					ctx.LogErrorf("Error writing to WebSocket: %v", err)
				}
				return
			}
		}
		subscriptionId = ctx.Logger().OnMessage(subscriptionId, onMessage)

		<-done
		ctx.Logger().RemoveMessageHandler(subscriptionId)
	}
}
