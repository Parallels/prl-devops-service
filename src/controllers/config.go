package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
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
		ctx := GetBaseContext(r)
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
		ctx := GetBaseContext(r)
		var request models.InstallToolsRequest
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

		response := models.InstallToolsResponse{
			Success:        true,
			InstalledTools: make(map[string]models.InstallToolsResponseItem),
		}

		provider := serviceprovider.Get()
		if request.All {
			provider.InstallAllTools(request.RunAs, map[string]string{})
		} else {
			for tool, option := range request.Tools {
				switch tool {
				case "packer":
					if err := provider.PackerService.Install(request.RunAs, option.Version, option.Flags); err != nil {
						response.Success = false
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success:      false,
							Version:      option.Version,
							ErrorMessage: err.Error(),
						}
					} else {
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success: true,
							Version: option.Version,
						}
					}
				case "vagrant":
					if err := provider.VagrantService.Install(request.RunAs, option.Version, option.Flags); err != nil {
						response.Success = false
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success:      false,
							Version:      option.Version,
							ErrorMessage: err.Error(),
						}
					} else {
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success: true,
							Version: option.Version,
						}
					}
				case "parallels":
					if err := provider.ParallelsDesktopService.Install(request.RunAs, option.Version, option.Flags); err != nil {
						response.Success = false
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success:      false,
							Version:      option.Version,
							ErrorMessage: err.Error(),
						}
					} else {
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success: true,
							Version: option.Version,
						}
					}
				case "git":
					if err := provider.GitService.Install(request.RunAs, option.Version, option.Flags); err != nil {
						response.Success = false
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success:      false,
							Version:      option.Version,
							ErrorMessage: err.Error(),
						}
					} else {
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success: true,
							Version: option.Version,
						}
					}
				case "brew":
					if err := provider.System.Install(request.RunAs, option.Version, option.Flags); err != nil {
						response.Success = false
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success:      false,
							Version:      option.Version,
							ErrorMessage: err.Error(),
						}
					} else {
						response.InstalledTools[tool] = models.InstallToolsResponseItem{
							Success: true,
							Version: option.Version,
						}
					}
				default:
					response.InstalledTools[tool] = models.InstallToolsResponseItem{
						Success:      false,
						Version:      option.Version,
						ErrorMessage: "Not Recognized Tool",
					}
				}
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
		ctx := GetBaseContext(r)
		var request models.UninstallToolsRequest
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

		response := models.UninstallToolsResponse{
			Success:          true,
			UninstalledTools: make(map[string]models.UninstallToolsResponseItem),
		}

		provider := serviceprovider.Get()
		if request.All {
			provider.UninstallAllTools(request.RunAs, request.UninstallDependencies, map[string]string{})
		} else {
			for tool, option := range request.Tools {
				switch tool {
				case "packer":
					if err := provider.PackerService.Uninstall(request.RunAs, option.UninstallDependencies); err != nil {
						response.Success = false
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success:      false,
							ErrorMessage: err.Error(),
						}
					} else {
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success: true,
						}
					}
				case "vagrant":
					if err := provider.VagrantService.Uninstall(request.RunAs, option.UninstallDependencies); err != nil {
						response.Success = false
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success:      false,
							ErrorMessage: err.Error(),
						}
					} else {
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success: true,
						}
					}
				case "parallels":
					if err := provider.ParallelsDesktopService.Uninstall(request.RunAs, option.UninstallDependencies); err != nil {
						response.Success = false
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success:      false,
							ErrorMessage: err.Error(),
						}
					} else {
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success: true,
						}
					}
				case "git":
					if err := provider.GitService.Uninstall(request.RunAs, option.UninstallDependencies); err != nil {
						response.Success = false
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success: false,

							ErrorMessage: err.Error(),
						}
					} else {
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success: true,
						}
					}
				case "brew":
					if err := provider.System.Uninstall(request.RunAs, option.UninstallDependencies); err != nil {
						response.Success = false
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success:      false,
							ErrorMessage: err.Error(),
						}
					} else {
						response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
							Success: true,
						}
					}
				default:
					response.UninstalledTools[tool] = models.UninstallToolsResponseItem{
						Success:      false,
						ErrorMessage: "Not Recognized Tool",
					}
				}
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
		ctx := GetBaseContext(r)
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
		ctx := GetBaseContext(r)
		provider := serviceprovider.Get()
		hardwareInfo, err := provider.ParallelsDesktopService.GetHardwareUsage(ctx)
		if err != nil {
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
