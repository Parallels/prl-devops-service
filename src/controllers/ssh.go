package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper/http_helper"
)

func registerSshHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s SSH handlers", version)
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/ssh/execute").
		WithRequiredClaim(constants.EXECUTE_SSH_CLAIM).
		WithHandler(ExecuteSshHandler()).
		Register()
}

type SshExecutionRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Key      string `json:"key,omitempty"`
	Command  string `json:"command"`
}

type SshExecutionResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func (r *SshExecutionRequest) Validate() error {
	if r.Host == "" {
		return errors.NewWithCode("host cannot be empty", 400)
	}
	if r.Port == 0 {
		return errors.NewWithCode("port cannot be zero", 400)
	}
	if r.Username == "" {
		return errors.NewWithCode("username cannot be empty", 400)
	}
	if r.Command == "" {
		return errors.NewWithCode("command cannot be empty", 400)
	}
	return nil
}

// @Summary		Execute SSH Command
// @Description	Executes a command on a remote host via SSH
// @Tags			SSH
// @Produce		json
// @Param			sshRequest	body		SshExecutionRequest	true	"Body"
// @Success		200			{object}	SshExecutionResponse
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Failure		403			{object}	models.ApiErrorResponse
// @Failure		500			{object}	models.ApiErrorResponse
// @Router			/v1/ssh/execute [post]
func ExecuteSshHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request SshExecutionRequest
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

		provider := serviceprovider.Get()
		if provider.SshService == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "SSH Service not available",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		output, err := provider.SshService.Execute(ctx, request.Host, request.Port, request.Username, request.Password, request.Key, request.Command)
		response := SshExecutionResponse{
			Output: output,
		}

		if err != nil {
			response.Error = err.Error()
			// We return 200 even if SSH execution failed (but connection worked),
			// or we can decide to return 500. Usually if command fails on remote end,
			// it's still a successful execution of the *request*.
			// If we want to signal failure, we can use 500.
			// Let's assume if there's an error connecting it's 500, if command fails it might be 200 with error.
			// The service returns error for both connection and run errors.
			// Let's return 500 if there is an err.

			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
