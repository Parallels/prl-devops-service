package controllers

// Host-only settings endpoint (registered when the host module is on): the
// orchestrator push channel. The self console (GET/PUT/DELETE under
// /settings/definitions and /settings/values) is shared with the orchestrator
// and registered by registerSettingsHandlers — on a host node, self writes
// land in the local layer, where orchestrator-managed values keep shadowing
// them (shadowed_by) and every sync preserves them.
//
// POST /settings/apply installs a full managed bundle. The orchestrator
// authenticates with the per-host API key minted at enrollment — an
// unassigned service key, which passes claim gates as a micro-service
// principal (claim checks apply to user-assigned identities only).

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/settings"

	"github.com/cjlapao/common-go/helper/http_helper"
)

func registerHostSettingsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Host Settings handlers", version)

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/settings/apply").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(ApplySettingsBundleHandler()).
		Register()
}

// @Summary		Applies a managed settings bundle
// @Description	Orchestrator-only push channel: replaces this node's managed layer wholesale (local layer untouched). Idempotent — a bundle whose hash already matches answers unchanged. Unknown keys, orchestrator-only keys and hash mismatches are rejected.
// @Tags			Settings
// @Accept			json
// @Produce		json
// @Param			bundle	body		models.SettingsSyncRequest	true	"Managed bundle"
// @Success		200		{object}	models.SettingsSyncResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/apply [post]
func ApplySettingsBundleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/settings/apply")
		svc := settings.Get()
		if svc == nil {
			diag.AddError(strconv.Itoa(http.StatusServiceUnavailable), "settings service is not available", "SettingsService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusServiceUnavailable))
			return
		}

		var request models.SettingsSyncRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
			return
		}
		if request.Values == nil {
			request.Values = map[string]string{}
		}

		response, err := svc.ApplyBundle(request.Revision, request.Hash, request.Values)
		if err != nil {
			settingWriteError(ctx, w, diag, "ApplyBundle", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Settings bundle applied (revision=%d unchanged=%t count=%d)",
			response.AppliedRevision, response.Unchanged, response.AppliedCount)
	}
}
