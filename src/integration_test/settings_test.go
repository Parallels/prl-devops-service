//go:build integration

package integration_test

// Settings API QA suite — exercises the reshaped surface end to end through
// real HTTP against the in-process listener:
//   self console        : GET /settings/definitions, GET /settings/values,
//                         PUT/DELETE /settings/values/{key} (host local layer)
//   orchestrator console: PUT/DELETE /orchestrator/settings/values/{key},
//                         /orchestrator/hosts/settings/values/{key},
//                         /orchestrator/hosts/{id}/settings/values[/{key}],
//                         POST /orchestrator/settings/sync
//   host-as-service     : self writes land in the local layer, managed pushes
//                         (POST /settings/apply) shadow but never destroy them.
//
// The base harness boots the legacy JSON database for users/auth/config;
// these tests additionally inject a migrated GORM sqlite handle (the handle
// the settings store reads through serviceprovider.GetSQLDatabase) so the
// settings domain is live while everything else stays on JSON — exactly the
// posture the SQL migration (D-6/S2) ships today.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/controllers"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/restapi"
	bruteforceguard "github.com/Parallels/prl-devops-service/security/brute_force_guard"
	"github.com/Parallels/prl-devops-service/security/jwt"
	"github.com/Parallels/prl-devops-service/security/password"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/Parallels/prl-devops-service/settings"
	"github.com/Parallels/prl-devops-service/startup/seeds"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	orchestratorModules = "api,cors,orchestrator"
	hostModules         = "api,cors,host"
)

// newSettingsEnv boots the QA server with the settings domain live: legacy
// JSON database for auth/users/config (like newTestEnv) PLUS a freshly
// migrated GORM sqlite handle for the settings store. modules controls the
// role ("api,cors,orchestrator" or "api,cors,host").
func newSettingsEnv(t *testing.T, modules string) *testEnv {
	t.Helper()

	t.Setenv("JWT_HMACS_SECRET", testHMACSecret)
	t.Setenv("ROOT_PASSWORD", adminPassword)
	t.Setenv("SECURITY_PASSWORD_MIN_PASSWORD_LENGTH", "6")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_UPPERCASE", "false")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_LOWERCASE", "false")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_NUMBER", "false")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR", "false")
	t.Setenv("ENABLED_MODULES", modules)
	t.Setenv("DB_TYPE", "sqlite")

	tmpDir, err := os.MkdirTemp("", "settings-integration-*")
	require.NoError(t, err)
	dbFile := filepath.Join(tmpDir, "data.json")

	data.ResetForTesting()

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	config.New(ctx)
	password.New(ctx)
	jwt.New(ctx)
	bruteforceguard.New(ctx)

	provider := serviceprovider.NewMockProvider()
	provider.System = system.New(ctx)
	jsonDb := data.NewJsonDatabase(ctx, dbFile)
	provider.JsonDatabase = jsonDb
	provider.HardwareId = "settings-integration-hardware-id"
	require.NoError(t, jsonDb.Connect(ctx))

	// SQL handle for the settings domain (same engine/driver the product
	// uses for DB_TYPE=sqlite).
	sqliteDb, err := gorm.Open(sqlite.Open(filepath.Join(tmpDir, "settings.db")), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, sqliteDb.AutoMigrate(database.AllGormModels()...))
	provider.SQLDatabase = sqliteDb

	require.NoError(t, seeds.SeedDefaultClaims())
	require.NoError(t, seeds.SeedDefaultRoles())
	require.NoError(t, seeds.SeedDefaultRoleClaims())
	require.NoError(t, seeds.SeedDefaultUsers())

	// Boot the settings domain (catalog seed + resolver registration).
	_, err = settings.InitService(ctx)
	require.NoError(t, err)

	listener := restapi.NewHttpListener()
	listener.WithVersion("v1", "/v1", true)
	require.NoError(t, controllers.RegisterV1Handlers(ctx))
	for _, c := range listener.Controllers {
		require.NoError(t, c.Serve())
	}

	srv := httptest.NewServer(listener.Router)
	t.Cleanup(func() {
		srv.Close()
		settings.Shutdown()
		data.ResetForTesting()
		_ = os.RemoveAll(tmpDir)
	})

	return &testEnv{srv: srv, ctx: ctx}
}

func settingsToken(t *testing.T, env *testEnv) string {
	t.Helper()
	return env.mustGetToken(t, adminEmail, adminPassword)
}

type definitionEntry struct {
	Key             string `json:"key"`
	Type            string `json:"type"`
	DefaultValue    string `json:"default_value"`
	RequiresRestart bool   `json:"requires_restart"`
	AppliesTo       string `json:"applies_to"`
}

func findDefinition(defs []definitionEntry, key string) *definitionEntry {
	for i := range defs {
		if defs[i].Key == key {
			return &defs[i]
		}
	}
	return nil
}

func TestSettingsQA_Definitions(t *testing.T) {
	env := newSettingsEnv(t, orchestratorModules)
	token := settingsToken(t, env)

	resp := env.do(t, http.MethodGet, "/settings/definitions", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Definitions []definitionEntry `json:"definitions"`
	}
	mustDecodeBody(t, resp, &body)
	require.GreaterOrEqual(t, len(body.Definitions), 20)

	logLevel := findDefinition(body.Definitions, "LOG_LEVEL")
	require.NotNil(t, logLevel)
	require.True(t, logLevel.RequiresRestart, "LOG_LEVEL must be flagged requires_restart")
	require.Nil(t, findDefinition(body.Definitions, "DB_PATH"), "boot-critical keys must never be catalogued")
}

func TestSettingsQA_OrchestratorSelfWriteEffectiveAndDelete(t *testing.T) {
	env := newSettingsEnv(t, orchestratorModules)
	token := settingsToken(t, env)

	// PUT the orchestrator node's own values — under the orchestrator prefix,
	// like every other orchestrator-owned resource.
	resp := env.do(t, http.MethodPut, "/orchestrator/settings/values/LOG_LEVEL", map[string]string{"value": " warn "}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var echo struct {
		Scope string `json:"scope"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	mustDecodeBody(t, resp, &echo)
	require.Equal(t, "orchestrator", echo.Scope)
	require.Equal(t, "LOG_LEVEL", echo.Key, "key must be normalized")
	require.Equal(t, "warn", echo.Value, "value must be trimmed/normalized")

	// Effective view flips to the DB layer.
	resp = env.do(t, http.MethodGet, "/settings/values", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var state struct {
		Values []struct {
			Scope string `json:"scope"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"values"`
		Effective []effectiveEntry `json:"effective"`
	}
	mustDecodeBody(t, resp, &state)
	require.NotEmpty(t, state.Values)
	entry := findEffectiveEntry(state.Effective, "LOG_LEVEL")
	require.NotNil(t, entry)
	require.Equal(t, "warn", entry.Effective)
	require.Equal(t, "orchestrator_db", entry.Source)

	// DELETE removes it; the next layer answers.
	resp = env.do(t, http.MethodDelete, "/orchestrator/settings/values/LOG_LEVEL", nil, token)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	resp = env.do(t, http.MethodGet, "/settings/values", nil, token)
	mustDecodeBody(t, resp, &state)
	entry = findEffectiveEntry(state.Effective, "LOG_LEVEL")
	require.Equal(t, "info", entry.Effective)
	require.Equal(t, "default", entry.Source)

	// Second delete: 404.
	resp = env.do(t, http.MethodDelete, "/orchestrator/settings/values/LOG_LEVEL", nil, token)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	// The host-persona local path is NOT registered on a pure orchestrator.
	resp = env.do(t, http.MethodPut, "/settings/values/LOG_LEVEL", map[string]string{"value": "warn"}, token)
	require.Equal(t, http.StatusNotFound, resp.StatusCode, "local self-writes belong to the host persona only")
	resp.Body.Close()
}

type effectiveEntry struct {
	Key        string `json:"key"`
	Effective  string `json:"effective"`
	Source     string `json:"source"`
	ShadowedBy string `json:"shadowed_by"`
}

func findEffectiveEntry(entries []effectiveEntry, key string) *effectiveEntry {
	for i := range entries {
		if entries[i].Key == key {
			return &entries[i]
		}
	}
	return nil
}

func TestSettingsQA_WhitelistRejects(t *testing.T) {
	env := newSettingsEnv(t, orchestratorModules)
	token := settingsToken(t, env)

	cases := []struct {
		name  string
		path  string
		value string
	}{
		{"uncatalogued boot-critical key", "/orchestrator/settings/values/DB_PATH", "/tmp/evil"},
		{"value outside allow-list", "/orchestrator/settings/values/LOG_LEVEL", "verbose"},
		{"host-only key on the orchestrator", "/orchestrator/settings/values/MAX_VMS_PER_HOST", "4"},
		{"orchestrator-only key via hosts-global", "/orchestrator/hosts/settings/values/ORCHESTRATOR_PULL_FREQUENCY_SECONDS", "5"},
		{"non-integer for an int key", "/orchestrator/hosts/settings/values/MAX_VMS_PER_HOST", "abc"},
	}
	for _, tc := range cases {
		resp := env.do(t, http.MethodPut, tc.path, map[string]string{"value": tc.value}, token)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, tc.name)
		resp.Body.Close()
	}
}

func TestSettingsQA_HostsGlobalPerHostAndBundleView(t *testing.T) {
	env := newSettingsEnv(t, orchestratorModules)
	token := settingsToken(t, env)

	// Hosts-global tier.
	resp := env.do(t, http.MethodPut, "/orchestrator/hosts/settings/values/ENABLE_CANARY", map[string]string{"value": "true"}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var global struct {
		Scope string `json:"scope"`
	}
	mustDecodeBody(t, resp, &global)
	require.Equal(t, "hosts", global.Scope)

	// Per-host override wins over hosts-global.
	resp = env.do(t, http.MethodPut, "/orchestrator/hosts/h1/settings/values/ENABLE_CANARY", map[string]string{"value": "false"}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Both rows show up in the values listing.
	resp = env.do(t, http.MethodGet, "/settings/values", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var state struct {
		Values []struct {
			Scope string `json:"scope"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"values"`
	}
	mustDecodeBody(t, resp, &state)
	scopes := map[string]string{}
	for _, row := range state.Values {
		scopes[row.Scope+":"+row.Key] = row.Value
	}
	require.Equal(t, "true", scopes["hosts:ENABLE_CANARY"])
	require.Equal(t, "false", scopes["host:h1:ENABLE_CANARY"])
	resp.Body.Close()

	// Per-host bundle view route is live; unknown host answers 404 (h1 was
	// only ever a scope label, never a registered host).
	resp = env.do(t, http.MethodGet, "/orchestrator/hosts/h1/settings/values", nil, token)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestSettingsQA_AuthAndSyncGuards(t *testing.T) {
	env := newSettingsEnv(t, orchestratorModules)
	token := settingsToken(t, env)

	// No token → 401.
	resp := env.do(t, http.MethodGet, "/settings/definitions", nil, "")
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Bad sync target → 400; "all" with no registered hosts → 200.
	resp = env.do(t, http.MethodPost, "/orchestrator/settings/sync", map[string]string{"target": "banana"}, token)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	resp = env.do(t, http.MethodPost, "/orchestrator/settings/sync", map[string]string{"target": "all"}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestSettingsQA_HostRoleLocalLayerAndApply(t *testing.T) {
	// The platform rules strip the host module off non-darwin systems even
	// when ENABLED_MODULES requests it (host requires the Parallels stack),
	// so the host-as-service console is only QA-able on macOS.
	if runtime.GOOS != "darwin" {
		t.Skip("host module is darwin-only")
	}
	env := newSettingsEnv(t, hostModules)
	token := settingsToken(t, env)

	// Self write on a host lands in the LOCAL layer.
	resp := env.do(t, http.MethodPut, "/settings/values/LOG_LEVEL", map[string]string{"value": "error"}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var echo struct {
		Scope string `json:"scope"`
	}
	mustDecodeBody(t, resp, &echo)
	require.Equal(t, "local", echo.Scope)

	// Orchestrator-only keys are not writable locally on a host.
	resp = env.do(t, http.MethodPut, "/settings/values/ORCHESTRATOR_PULL_FREQUENCY_SECONDS", map[string]string{"value": "5"}, token)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// Apply a managed bundle — managed shadows local but does not destroy it.
	bundle := map[string]string{"LOG_LEVEL": "debug"}
	resp = env.do(t, http.MethodPost, "/settings/apply", map[string]interface{}{
		"revision": 1, "hash": settings.BundleHash(bundle), "values": bundle,
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var ack struct {
		Unchanged bool `json:"unchanged"`
	}
	mustDecodeBody(t, resp, &ack)
	require.False(t, ack.Unchanged)

	resp = env.do(t, http.MethodGet, "/settings/values", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var state struct {
		Values []struct {
			Scope string `json:"scope"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"values"`
		Effective []effectiveEntry `json:"effective"`
		Node      struct {
			Managed bool `json:"managed"`
		} `json:"node"`
	}
	mustDecodeBody(t, resp, &state)
	require.True(t, state.Node.Managed)

	entry := findEffectiveEntry(state.Effective, "LOG_LEVEL")
	require.NotNil(t, entry)
	require.Equal(t, "debug", entry.Effective, "managed must win")
	require.Equal(t, "managed_db", entry.Source)
	require.Equal(t, "managed_db", entry.ShadowedBy, "local edit must be marked shadowed")

	scopes := map[string]string{}
	for _, row := range state.Values {
		scopes[row.Scope+":"+row.Key] = row.Value
	}
	require.Equal(t, "debug", scopes["orchestrator:LOG_LEVEL"], "managed row present")
	require.Equal(t, "error", scopes["local:LOG_LEVEL"], "local row must survive the managed replace")

	// Idempotent re-apply answers unchanged.
	resp = env.do(t, http.MethodPost, "/settings/apply", map[string]interface{}{
		"revision": 1, "hash": settings.BundleHash(bundle), "values": bundle,
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	mustDecodeBody(t, resp, &ack)
	require.True(t, ack.Unchanged)

	// Corrupted bundle (hash mismatch) is rejected.
	resp = env.do(t, http.MethodPost, "/settings/apply", map[string]interface{}{
		"revision": 2, "hash": "deadbeef", "values": map[string]string{"MAX_VMS_PER_HOST": "1"},
	}, token)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}
