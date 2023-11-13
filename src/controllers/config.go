package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
)

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
func GetParallelsDesktopLicenseController() restapi.ControllerHandler {
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
		json.NewEncoder(w).Encode(license)
		ctx.LogInfo("Parallels Desktop License returned successfully")
	}
}

// @Summary		Installs API requires 3rd party tools
// @Description	This endpoint installs API requires 3rd party tools
// @Tags			Config
// @Produce		json
// @Param			installToolsRequest	body	models.InstallToolsRequest true "Install Tools Request"
// @Success		200	{object}	models.InstallToolsResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/config/tools/install [post]
func InstallToolsController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.InstallToolsRequest
		http_helper.MapRequestBody(r, &request)
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
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Tools install request successfully")
	}
}

// @Summary		Uninstalls API requires 3rd party tools
// @Description	This endpoint uninstalls API requires 3rd party tools
// @Tags			Config
// @Produce		json
// @Param			uninstallToolsRequest	body	models.UninstallToolsRequest true "Uninstall Tools Request"
// @Success		200	{object}	models.InstallToolsResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/config/tools/install [post]
func UninstallToolsController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UninstallToolsRequest
		http_helper.MapRequestBody(r, &request)
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
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Tools uninstall request successfully")
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
// @Router			/v1/config/tools/install [post]
func RestartController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		go restapi.Get().Restart()
		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Restart request accepted")
	}
}
