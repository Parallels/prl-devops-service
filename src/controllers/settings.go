package controllers

// Orchestrator-side settings API (registered only when the orchestrator
// module is on): the centralized configuration console — catalog, scoped
// overrides, this node's effective view with sources, per-host drift and the
// manual sync fan-out. Host-scoped writes here are pushed to the hosts by
// the settings sync engine; orchestrator-scoped writes apply to this process
// immediately through the config override resolver.

import (
	"encoding/json"
	stderrors "errors"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/settings"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerSettingsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Settings handlers", version)

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/settings").
		WithAuthorization().
		WithRequiredClaim(constants.LIST_SETTINGS_CLAIM).
		WithHandler(GetSettingsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/settings/values/{scope}/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(SetSettingValueHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/settings/values/{scope}/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(DeleteSettingValueHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/settings/hosts/{id}/effective").
		WithAuthorization().
		WithRequiredClaim(constants.LIST_SETTINGS_CLAIM).
		WithHandler(GetHostEffectiveSettingsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/settings/sync").
		WithAuthorization().
		WithRequiredClaim(constants.SYNC_SETTINGS_CLAIM).
		WithHandler(SyncSettingsHandler()).
		Register()
}

// settingsServiceOrFail resolves the settings service, answering 503 when
// the domain is inert (legacy JSON database mode or failed boot init).
func settingsServiceOrFail(ctx basecontext.ApiContext, w http.ResponseWriter, diag *errors.Diagnostics) (*settings.Service, bool) {
	svc := settings.Get()
	if svc == nil {
		diag.AddError(strconv.Itoa(http.StatusServiceUnavailable), "settings service is not available", "SettingsService")
		ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusServiceUnavailable))
		return nil, false
	}
	return svc, true
}

// settingWriteError answers service-side validation/store failures: anything
// that is not a store outage is a client-side 400 (unknown key, bad type,
// scope mismatch — all raised by the whitelist validator).
func settingWriteError(ctx basecontext.ApiContext, w http.ResponseWriter, diag *errors.Diagnostics, stage string, err error) {
	code := http.StatusBadRequest
	if stderrors.Is(err, settings.ErrStoreUnavailable) {
		code = http.StatusServiceUnavailable
	}
	diag.AddError(strconv.Itoa(code), err.Error(), stage)
	ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, code))
}

func settingUpdatedBy(ctx basecontext.ApiContext) string {
	if user := ctx.GetUser(); user != nil && user.Username != "" {
		return user.Username
	}
	return "service"
}

// @Summary		Gets the centralized settings overview
// @Description	Returns the definitions catalog, every stored override, this node's effective view (value + source + shadowing) and per-host sync/drift state.
// @Tags			Settings
// @Produce		json
// @Success		200	{object}	models.SettingsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		503	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings [get]
func GetSettingsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/settings")
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}

		response := models.SettingsResponse{
			Definitions: mappers.SettingDefinitionsToResponse(svc.Definitions()),
			Values:      mappers.SettingValuesToResponse(svc.Values()),
			Effective:   svc.OrchestratorNodeView(),
			Node: models.SettingsNodeInfo{
				Managed:     svc.NodeInfo().Managed,
				Revision:    svc.NodeInfo().Revision,
				ManagedHash: svc.NodeInfo().ManagedHash,
			},
		}

		states := svc.SyncStates()
		names := hostNamesForSyncStates(ctx, states)
		sync := make([]models.SettingSyncStateResponse, 0, len(states))
		for _, state := range states {
			sync = append(sync, mappers.SettingSyncStateToResponse(state, names[state.HostID]))
		}
		response.Sync = sync

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// hostNamesForSyncStates best-effort joins host display names onto drift
// rows (a missing hosts table must never break the listing).
func hostNamesForSyncStates(ctx basecontext.ApiContext, states []data_models.SettingSyncState) map[string]string {
	names := map[string]string{}
	if len(states) == 0 {
		return names
	}
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil || dbService == nil {
		return names
	}
	hosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return names
	}
	for _, host := range hosts {
		names[host.ID] = host.Host
	}
	return names
}

// @Summary		Sets a scoped settings override
// @Description	Creates or updates one override for the orchestrator ("orchestrator"), all hosts ("hosts") or a single host ("host:<hostId>"). Only catalogued keys are accepted; host-scoped changes are pushed to the hosts automatically.
// @Tags			Settings
// @Accept			json
// @Produce			json
// @Param			scope	path	string						true	"orchestrator | hosts | host:<hostId>"
// @Param			key		path	string						true	"Catalogued setting key"
// @Param			value	body	models.SettingValueRequest	true	"Override value"
// @Success		200		{object}	models.SettingValueResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/values/{scope}/{key} [put]
func SetSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		scope := html.UnescapeString(vars["scope"])
		key := html.UnescapeString(vars["key"])
		diag := errors.NewDiagnostics("/settings/values/" + scope + "/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}

		var request models.SettingValueRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
			return
		}

		if err := validateSettingScopeForOrchestrator(scope); err != nil {
			settingWriteError(ctx, w, diag, "ValidateScope", err)
			return
		}

		if _, err := svc.SetScopeValue(scope, key, request.Value, settingUpdatedBy(ctx)); err != nil {
			settingWriteError(ctx, w, diag, "SetScopeValue", err)
			return
		}

		value := data_models.SettingValue{Scope: scope, Key: strings.ToUpper(strings.TrimSpace(key)), Value: request.Value}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mappers.SettingValueToResponse(value))
		ctx.LogInfof("Settings override set: scope=%s key=%s", scope, key)
	}
}

// validateSettingScopeForOrchestrator accepts the three canonical scopes;
// host ids may not be empty.
func validateSettingScopeForOrchestrator(scope string) error {
	switch {
	case scope == data_models.SettingScopeOrchestrator, scope == data_models.SettingScopeHosts:
		return nil
	case data_models.IsHostSettingScope(scope):
		if data_models.HostIDFromSettingScope(scope) == "" {
			return errors.New("host scope requires a host id (host:<hostId>)")
		}
		return nil
	}
	return errors.New("invalid scope: expected orchestrator, hosts or host:<hostId>")
}

// @Summary		Removes a scoped settings override
// @Description	Hard-removes one override; the key falls back to the next layer. Host-scoped removals are re-pushed to the host automatically.
// @Tags			Settings
// @Produce			json
// @Param			scope	path	string	true	"orchestrator | hosts | host:<hostId>"
// @Param			key		path	string	true	"Catalogued setting key"
// @Success		202
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/values/{scope}/{key} [delete]
func DeleteSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		scope := html.UnescapeString(vars["scope"])
		key := html.UnescapeString(vars["key"])
		diag := errors.NewDiagnostics("/settings/values/" + scope + "/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}
		if err := validateSettingScopeForOrchestrator(scope); err != nil {
			settingWriteError(ctx, w, diag, "ValidateScope", err)
			return
		}

		existed, err := svc.DeleteScopeValue(scope, key)
		if err != nil {
			settingWriteError(ctx, w, diag, "DeleteScopeValue", err)
			return
		}
		if !existed {
			diag.AddError(strconv.Itoa(http.StatusNotFound), "no override stored for this scope and key", "DeleteScopeValue")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusNotFound))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Settings override removed: scope=%s key=%s", scope, key)
	}
}

// @Summary		Gets a host's resolved settings bundle
// @Description	Shows exactly what the given host will hold after the next successful sync (hosts-global plus per-host overrides, host wins) together with its drift state.
// @Tags			Settings
// @Produce		json
// @Param			id	path	string	true	"Host ID"
// @Success		200	{object}	models.SettingsHostEffectiveResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/hosts/{id}/effective [get]
func GetHostEffectiveSettingsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		hostID := html.UnescapeString(vars["id"])
		diag := errors.NewDiagnostics("/settings/hosts/" + hostID + "/effective")
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil || dbService == nil {
			diag.AddError(strconv.Itoa(http.StatusServiceUnavailable), "database unavailable", "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusServiceUnavailable))
			return
		}
		host, err := dbService.GetOrchestratorHost(ctx, hostID)
		if err != nil || host == nil {
			diag.AddError(strconv.Itoa(http.StatusNotFound), "host not found", "GetOrchestratorHost")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusNotFound))
			return
		}

		bundle, hash, revision := svc.DesiredBundleForHost(hostID)
		state := svc.SyncStateFor(hostID)
		response := models.SettingsHostEffectiveResponse{
			HostID:       hostID,
			HostName:     host.Host,
			Bundle:       bundle,
			Effective:    svc.HostEffectiveFor(hostID),
			Revision:     revision,
			ContentHash:  state.ContentHash,
			Status:       state.Status,
			LastError:    state.LastError,
			LastSyncedAt: state.LastSyncedAt,
		}
		if response.Status == "" {
			response.Status = data_models.SettingSyncStatusNever
		}
		if state.Status != data_models.SettingSyncStatusSynced || state.ContentHash != hash {
			response.Status = data_models.SettingSyncStatusPending
			if state.Status == data_models.SettingSyncStatusError {
				response.Status = data_models.SettingSyncStatusError
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Syncs settings to hosts
// @Description	Immediately pushes the current managed bundles to all hosts or to a single host ("host:<hostId>"). Idempotent: hosts already holding the current bundle answer unchanged.
// @Tags			Settings
// @Accept			json
// @Produce			json
// @Param			target	body	models.SettingsSyncTargetRequest	true	"Sync target"
// @Success		200		{object}	[]models.SettingsSyncResult
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/sync [post]
func SyncSettingsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/settings/sync")
		if _, ok := settingsServiceOrFail(ctx, w, diag); !ok {
			return
		}

		var request models.SettingsSyncTargetRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
			return
		}

		target := strings.TrimSpace(request.Target)
		if target == "" {
			target = "all"
		}
		if target != "all" && !data_models.IsHostSettingScope(target) {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "target must be \"all\" or \"host:<hostId>\"", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
			return
		}

		var results []models.SettingsSyncResult
		if target == "all" {
			results = settings.SyncAllNow(ctx, "")
		} else {
			results = []models.SettingsSyncResult{settings.SyncHostNow(ctx, data_models.HostIDFromSettingScope(target))}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(results)
		ctx.LogInfof("Settings sync finished for target %s (%d results)", target, len(results))
	}
}
