package controllers

// Host-side settings API (registered only when the host module is on). Two
// audiences:
//   - the UI connecting directly to a host (GET /settings/node + PUT/DELETE
//     /settings/node/{key}): reads this host's effective config and edits
//     the LOCAL layer — orchestrator-managed values shadow local edits but
//     local edits survive every sync (they are never touched by the
//     source-scoped managed replace);
//   - the orchestrator push channel (POST /settings/apply): installs a full
//     managed bundle. The orchestrator authenticates with the per-host API
//     key minted at enrollment — an unassigned service key, which passes
//     claim gates as a micro-service principal (claim checks apply to
//     user-assigned identities only).

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/settings"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerHostSettingsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Host Settings handlers", version)

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/settings/node").
		WithAuthorization().
		WithRequiredClaim(constants.LIST_SETTINGS_CLAIM).
		WithHandler(GetNodeSettingsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/settings/node/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(SetNodeSettingHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/settings/node/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(DeleteNodeSettingHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/settings/apply").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(ApplySettingsBundleHandler()).
		Register()
}

// @Summary		Gets this node's effective settings
// @Description	Returns the definitions catalog plus this node's effective view: managed (orchestrator-pushed) layer, local layer, env/config-file/default fallbacks, the winning source and shadowing markers.
// @Tags			Settings
// @Produce		json
// @Success		200	{object}	models.SettingsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		503	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/node [get]
func GetNodeSettingsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/settings/node")
		svc := settings.Get()
		if svc == nil {
			diag.AddError(strconv.Itoa(http.StatusServiceUnavailable), "settings service is not available", "SettingsService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusServiceUnavailable))
			return
		}

		meta := svc.NodeInfo()
		response := models.SettingsResponse{
			Definitions: mappers.SettingDefinitionsToResponse(svc.Definitions()),
			Effective:   svc.HostNodeView(),
			Values:      []models.SettingValueResponse{},
			Node: models.SettingsNodeInfo{
				Managed:     meta.Managed,
				Revision:    meta.Revision,
				ManagedHash: meta.ManagedHash,
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Sets a node-local settings override
// @Description	Edits this node's local layer (direct UI against a host). If the orchestrator also manages this key the managed value keeps winning (shadowed_by marks it); the local edit still survives future syncs.
// @Tags			Settings
// @Accept			json
// @Produce			json
// @Param			key		path	string					true	"Catalogued setting key"
// @Param			value	body	models.SettingValueRequest	true	"Override value"
// @Success		200
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/node/{key} [put]
func SetNodeSettingHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		key := html.UnescapeString(vars["key"])
		diag := errors.NewDiagnostics("/settings/node/" + key)
		svc := settings.Get()
		if svc == nil {
			diag.AddError(strconv.Itoa(http.StatusServiceUnavailable), "settings service is not available", "SettingsService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusServiceUnavailable))
			return
		}

		var request models.SettingValueRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
			return
		}

		if _, err := svc.SetLocalValue(key, request.Value); err != nil {
			settingWriteError(ctx, w, diag, "SetLocalValue", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		ctx.LogInfof("Node-local setting updated: key=%s", strings.ToUpper(strings.TrimSpace(key)))
	}
}

// @Summary		Removes a node-local settings override
// @Description	Drops this node's local override for the key (the managed layer, env or default takes over).
// @Tags			Settings
// @Produce			json
// @Param			key	path	string	true	"Catalogued setting key"
// @Success		202
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/node/{key} [delete]
func DeleteNodeSettingHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		key := html.UnescapeString(vars["key"])
		diag := errors.NewDiagnostics("/settings/node/" + key)
		svc := settings.Get()
		if svc == nil {
			diag.AddError(strconv.Itoa(http.StatusServiceUnavailable), "settings service is not available", "SettingsService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusServiceUnavailable))
			return
		}

		existed, err := svc.DeleteLocalValue(key)
		if err != nil {
			settingWriteError(ctx, w, diag, "DeleteLocalValue", err)
			return
		}
		if !existed {
			diag.AddError(strconv.Itoa(http.StatusNotFound), "no local override stored for this key", "DeleteLocalValue")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusNotFound))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Node-local setting removed: key=%s", strings.ToUpper(strings.TrimSpace(key)))
	}
}

// @Summary		Applies a managed settings bundle
// @Description	Orchestrator-only push channel: replaces this node's managed layer wholesale (local layer untouched). Idempotent — a bundle whose hash already matches answers unchanged. Unknown keys, orchestrator-only keys and hash mismatches are rejected.
// @Tags			Settings
// @Accept			json
// @Produce			json
// @Param			bundle	body	models.SettingsSyncRequest	true	"Managed bundle"
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
