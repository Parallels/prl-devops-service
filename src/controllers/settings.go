package controllers

// Settings API — two consoles on every node that runs the settings domain:
//   - /settings/definitions            : the curated catalog (static metadata)
//   - /settings/values                 : stored overrides + this node's
//     effective view (value + source + shadowing) + drift + node posture
//   - PUT/DELETE /settings/values/{key}: the host persona's LOCAL layer
//     (managed pushes keep shadowing local edits, marked shadowed_by).
//     Registered on host nodes only — "self" without a scope is always us.
//
// Everything the orchestrator owns — its own values and the management of
// OTHER nodes — follows the established /orchestrator/... convention and is
// registered only on the orchestrator (a host running as a service never
// exposes it):
//   - /orchestrator/settings/values/{key}                    : orchestrator's own values
//   - /orchestrator/hosts/settings/values/{key}              : hosts-global tier
//   - /orchestrator/hosts/{hostId}/settings/values/{key}     : per-host tier
//   - /orchestrator/hosts/{hostId}/settings/values           : what that host will hold
//   - /orchestrator/settings/sync                            : manual fan-out push
//
// Host-scoped writes are re-pushed to the hosts by the settings sync engine;
// node-owned writes apply to this process immediately through the config
// override resolver (keys read only at boot carry requires_restart=true in
// the catalog so the UI can offer an explicit restart instead of silently
// ignoring the change).

import (
	"encoding/json"
	stderrors "errors"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
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

// registerSettingsHandlers wires the SELF console. Registered on the
// orchestrator and on hosts alike — same paths, the running process decides
// which layer "self" means.
func registerSettingsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Settings handlers", version)

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/settings/definitions").
		WithAuthorization().
		WithRequiredClaim(constants.LIST_SETTINGS_CLAIM).
		WithHandler(GetSettingsDefinitionsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/settings/values").
		WithAuthorization().
		WithRequiredClaim(constants.LIST_SETTINGS_CLAIM).
		WithHandler(GetSettingsValuesHandler()).
		Register()

	// Self writes hit the LOCAL layer — the host persona's own edits. The
	// orchestrator's own values live under /orchestrator/settings/values,
	// mirroring how every other orchestrator-owned resource is routed.
	if config.Get().IsHost() {
		restapi.NewController().
			WithMethod(restapi.PUT).
			WithVersion(version).
			WithPath("/settings/values/{key}").
			WithAuthorization().
			WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
			WithHandler(SetLocalSettingValueHandler()).
			Register()

		restapi.NewController().
			WithMethod(restapi.DELETE).
			WithVersion(version).
			WithPath("/settings/values/{key}").
			WithAuthorization().
			WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
			WithHandler(DeleteLocalSettingValueHandler()).
			Register()
	}
}

// registerOrchestratorSettingsHandlers wires the privileged console: the
// orchestrator's own values plus the routes used to configure OTHER nodes.
// Paths follow the established orchestrator convention — host-scoped
// resources hang under /orchestrator/hosts/{hostId}/..., node-owned ones
// under /orchestrator/settings/.... Orchestrator module only.
func registerOrchestratorSettingsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Orchestrator Settings handlers", version)

	// The orchestrator node's own values.
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/settings/values/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(SetOrchestratorSettingValueHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/settings/values/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(DeleteOrchestratorSettingValueHandler()).
		Register()

	// Hosts-global tier.
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/settings/values/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(SetHostsGlobalSettingValueHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/settings/values/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(DeleteHostsGlobalSettingValueHandler()).
		Register()

	// Per-host overrides.
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{hostId}/settings/values/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(SetHostSettingValueHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{hostId}/settings/values/{key}").
		WithAuthorization().
		WithRequiredClaim(constants.UPDATE_SETTINGS_CLAIM).
		WithHandler(DeleteHostSettingValueHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/orchestrator/hosts/{hostId}/settings/values").
		WithAuthorization().
		WithRequiredClaim(constants.LIST_SETTINGS_CLAIM).
		WithHandler(GetHostValuesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/settings/sync").
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

// @Summary		Gets the settings catalog
// @Description	Returns the curated definitions: type, default, allow-list, applies_to and requires_restart per key. Static metadata — stored values and effective answers live under /settings/values.
// @Tags			Settings
// @Produce		json
// @Success		200	{object}	models.SettingsDefinitionsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		503	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/definitions [get]
func GetSettingsDefinitionsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/settings/definitions")
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}

		response := models.SettingsDefinitionsResponse{
			Definitions: mappers.SettingDefinitionsToResponse(svc.Definitions()),
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Gets the settings state
// @Description	Stored overrides plus this node's effective view (value + source + shadowing), drift rows and the node posture block. On the orchestrator: every stored scope (orchestrator/hosts/host:<id>) with the orchestrator's own effective answers. On a host: the managed and local layers with the host's effective answers.
// @Tags			Settings
// @Produce		json
// @Success		200	{object}	models.SettingsValuesResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		503	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/values [get]
func GetSettingsValuesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/settings/values")
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}

		meta := svc.NodeInfo()
		response := models.SettingsValuesResponse{
			Values:    []models.SettingValueResponse{},
			Effective: []models.SettingEffectiveEntry{},
			Node: models.SettingsNodeInfo{
				Managed:     meta.Managed,
				Revision:    meta.Revision,
				ManagedHash: meta.ManagedHash,
			},
		}

		if config.Get().IsOrchestrator() {
			response.Values = mappers.SettingValuesToResponse(svc.Values())
			response.Effective = svc.OrchestratorNodeView()
			states := svc.SyncStates()
			names := hostNamesForSyncStates(ctx, states)
			for _, state := range states {
				response.Sync = append(response.Sync, mappers.SettingSyncStateToResponse(state, names[state.HostID]))
			}
		} else {
			// Host node: the two applied layers, flattened into rows.
			response.Effective = svc.HostNodeView()
			for _, entry := range response.Effective {
				if entry.ManagedValue != nil {
					response.Values = append(response.Values, models.SettingValueResponse{
						ID:    data_models.NodeSettingSourceOrchestrator + ":" + entry.Key,
						Scope: data_models.NodeSettingSourceOrchestrator,
						Key:   entry.Key,
						Value: *entry.ManagedValue,
					})
				}
				if entry.LocalValue != nil {
					response.Values = append(response.Values, models.SettingValueResponse{
						ID:    data_models.NodeSettingSourceLocal + ":" + entry.Key,
						Scope: data_models.NodeSettingSourceLocal,
						Key:   entry.Key,
						Value: *entry.LocalValue,
					})
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Sets a local setting on this host
// @Description	Writes the host's LOCAL layer — no scope in the path, it is always "us". Orchestrator-managed keys keep winning (marked shadowed_by) and local edits survive every sync. Orchestrator-owned values are written under /orchestrator/settings/values. Body: {"value":"..."}
// @Tags			Settings
// @Accept			json
// @Produce		json
// @Param			key		path		string						true	"Catalogued setting key"
// @Param			value	body		models.SettingValueRequest	true	"Override value"
// @Success		200		{object}	models.SettingValueResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/values/{key} [put]
func SetLocalSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		diag := errors.NewDiagnostics("/settings/values/" + key)
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

		if _, err := svc.SetLocalValue(key, request.Value); err != nil {
			settingWriteError(ctx, w, diag, "SetLocalValue", err)
			return
		}

		// Echo what the store actually kept (the service trims/normalizes).
		normalized := data_models.SettingValue{Scope: data_models.NodeSettingSourceLocal, Key: strings.ToUpper(strings.TrimSpace(key)), Value: storedLocalValue(svc, key)}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mappers.SettingValueToResponse(normalized))
		ctx.LogInfof("Settings local write: key=%s", key)
	}
}

// @Summary		Removes a local setting from this host
// @Description	Drops the host's LOCAL override (the next layer answers: managed, env, config file or catalog default). 404 when no local override was stored.
// @Tags			Settings
// @Produce		json
// @Param			key	path	string	true	"Catalogued setting key"
// @Success		202
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/settings/values/{key} [delete]
func DeleteLocalSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		diag := errors.NewDiagnostics("/settings/values/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
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
		ctx.LogInfof("Settings local delete: key=%s", strings.ToUpper(strings.TrimSpace(key)))
	}
}

func selfSettingKey(r *http.Request) string {
	return html.UnescapeString(mux.Vars(r)["key"])
}

// storedSettingValue reads back the just-written value so the write echo
// shows what the store actually kept (the service trims/normalizes values).
func storedSettingValue(svc *settings.Service, scope, key string) string {
	normalizedKey := strings.ToUpper(strings.TrimSpace(key))
	for _, row := range svc.Values() {
		if row.Scope == scope && row.Key == normalizedKey {
			return row.Value
		}
	}
	return ""
}

// storedLocalValue is the host-role counterpart of storedSettingValue.
func storedLocalValue(svc *settings.Service, key string) string {
	normalizedKey := strings.ToUpper(strings.TrimSpace(key))
	for _, entry := range svc.HostNodeView() {
		if entry.Key == normalizedKey && entry.LocalValue != nil {
			return *entry.LocalValue
		}
	}
	return ""
}

// scopedSettingWrite maps the body, applies the write at the given scope and
// echoes the normalized row. Shared by the hosts-global and per-host routes.
func scopedSettingWrite(ctx basecontext.ApiContext, w http.ResponseWriter, r *http.Request, diag *errors.Diagnostics, svc *settings.Service, scope, key string) {
	var request models.SettingValueRequest
	if err := http_helper.MapRequestBody(r, &request); err != nil {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
		ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
		return
	}
	if _, err := svc.SetScopeValue(scope, key, request.Value, settingUpdatedBy(ctx)); err != nil {
		settingWriteError(ctx, w, diag, "SetScopeValue", err)
		return
	}
	value := data_models.SettingValue{Scope: scope, Key: strings.ToUpper(strings.TrimSpace(key)), Value: storedSettingValue(svc, scope, key)}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mappers.SettingValueToResponse(value))
	ctx.LogInfof("Settings override set: scope=%s key=%s", scope, key)
}

func scopedSettingDelete(ctx basecontext.ApiContext, w http.ResponseWriter, diag *errors.Diagnostics, svc *settings.Service, scope, key string) {
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

// @Summary		Sets a setting on the orchestrator node
// @Description	The orchestrator's own values — scope=orchestrator, the layer that wins over its local layer and shadows env/file/default for the orchestrator process. Applied live, no restart. Body: {"value":"..."}
// @Tags			Settings
// @Accept			json
// @Produce		json
// @Param			key		path		string						true	"Catalogued setting key"
// @Param			value	body		models.SettingValueRequest	true	"Override value"
// @Success		200		{object}	models.SettingValueResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/settings/values/{key} [put]
func SetOrchestratorSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		diag := errors.NewDiagnostics("/orchestrator/settings/values/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}
		scopedSettingWrite(ctx, w, r, diag, svc, data_models.SettingScopeOrchestrator, key)
	}
}

// @Summary		Removes a setting from the orchestrator node
// @Description	Drops the orchestrator's own override (env, config file or catalog default take over again).
// @Tags			Settings
// @Produce		json
// @Param			key	path	string	true	"Catalogued setting key"
// @Success		202
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/settings/values/{key} [delete]
func DeleteOrchestratorSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		diag := errors.NewDiagnostics("/orchestrator/settings/values/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}
		scopedSettingDelete(ctx, w, diag, svc, data_models.SettingScopeOrchestrator, key)
	}
}

// @Summary		Sets a hosts-global setting
// @Description	Value for ALL hosts (the tier every per-host override wins over). Applied by the sync engine on its next push. Body: {"value":"..."}
// @Tags			Settings
// @Accept			json
// @Produce		json
// @Param			key		path		string						true	"Catalogued setting key"
// @Param			value	body		models.SettingValueRequest	true	"Override value"
// @Success		200		{object}	models.SettingValueResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/settings/values/{key} [put]
func SetHostsGlobalSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		diag := errors.NewDiagnostics("/orchestrator/hosts/settings/values/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}
		scopedSettingWrite(ctx, w, r, diag, svc, data_models.SettingScopeHosts, key)
	}
}

// @Summary		Removes a hosts-global setting
// @Description	Drops the hosts-global override (hosts fall back to the catalog default unless a per-host override exists).
// @Tags			Settings
// @Produce		json
// @Param			key	path	string	true	"Catalogued setting key"
// @Success		202
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/settings/values/{key} [delete]
func DeleteHostsGlobalSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		diag := errors.NewDiagnostics("/orchestrator/hosts/settings/values/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}
		scopedSettingDelete(ctx, w, diag, svc, data_models.SettingScopeHosts, key)
	}
}

// @Summary		Sets a per-host setting
// @Description	Override for a single host (wins over the hosts-global tier). Applied by the sync engine on its next push. Body: {"value":"..."}
// @Tags			Settings
// @Accept			json
// @Produce		json
// @Param			hostId	path		string						true	"Host ID"
// @Param			key		path		string						true	"Catalogued setting key"
// @Param			value	body		models.SettingValueRequest	true	"Override value"
// @Success		200		{object}	models.SettingValueResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{hostId}/settings/values/{key} [put]
func SetHostSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		hostID := html.UnescapeString(mux.Vars(r)["hostId"])
		diag := errors.NewDiagnostics("/orchestrator/hosts/" + hostID + "/settings/values/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}
		if hostID == "" {
			settingWriteError(ctx, w, diag, "ValidateHost", stderrors.New("host id is required"))
			return
		}
		scopedSettingWrite(ctx, w, r, diag, svc, data_models.HostSettingScope(hostID), key)
	}
}

// @Summary		Removes a per-host setting
// @Description	Drops the per-host override (the hosts-global tier takes over).
// @Tags			Settings
// @Produce		json
// @Param			hostId	path	string	true	"Host ID"
// @Param			key		path	string	true	"Catalogued setting key"
// @Success		202
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{hostId}/settings/values/{key} [delete]
func DeleteHostSettingValueHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		key := selfSettingKey(r)
		hostID := html.UnescapeString(mux.Vars(r)["hostId"])
		diag := errors.NewDiagnostics("/orchestrator/hosts/" + hostID + "/settings/values/" + key)
		svc, ok := settingsServiceOrFail(ctx, w, diag)
		if !ok {
			return
		}
		if hostID == "" {
			settingWriteError(ctx, w, diag, "ValidateHost", stderrors.New("host id is required"))
			return
		}
		scopedSettingDelete(ctx, w, diag, svc, data_models.HostSettingScope(hostID), key)
	}
}

// @Summary		Gets a host's resolved settings values
// @Description	Shows exactly what the given host will hold after the next successful sync (hosts-global plus per-host overrides, host wins) together with its drift state.
// @Tags			Settings
// @Produce		json
// @Param			hostId	path		string	true	"Host ID"
// @Success		200		{object}	models.SettingsHostValuesResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/hosts/{hostId}/settings/values [get]
func GetHostValuesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		hostID := html.UnescapeString(mux.Vars(r)["hostId"])
		diag := errors.NewDiagnostics("/orchestrator/hosts/" + hostID + "/settings/values")
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
		response := models.SettingsHostValuesResponse{
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
// @Produce		json
// @Param			target	body		models.SettingsSyncTargetRequest	true	"Sync target"
// @Success		200		{object}	[]models.SettingsSyncResult
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/settings/sync [post]
func SyncSettingsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/orchestrator/settings/sync")
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
