package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/cjlapao/common-go/helper/http_helper"
)

// The settings service is uninitialized in unit context (no SQL handle) —
// every handler must answer 503, never panic, and never leak a body.

func doRequest(t *testing.T, handler func(http.ResponseWriter, *http.Request), method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	rw := httptest.NewRecorder()
	handler(rw, req)
	return rw
}

func TestSettingsHandlers_503WhenServiceInert(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		target  string
		body    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"definitions", http.MethodGet, "/api/v1/settings/definitions", "", GetSettingsDefinitionsHandler()},
		{"values listing", http.MethodGet, "/api/v1/settings/values", "", GetSettingsValuesHandler()},
		{"local write", http.MethodPut, "/api/v1/settings/values/MAX_VMS_PER_HOST", `{"value":"1"}`, SetLocalSettingValueHandler()},
		{"local delete", http.MethodDelete, "/api/v1/settings/values/MAX_VMS_PER_HOST", "", DeleteLocalSettingValueHandler()},
		{"orchestrator self write", http.MethodPut, "/api/v1/orchestrator/settings/values/LOG_LEVEL", `{"value":"warn"}`, SetOrchestratorSettingValueHandler()},
		{"orchestrator self delete", http.MethodDelete, "/api/v1/orchestrator/settings/values/LOG_LEVEL", "", DeleteOrchestratorSettingValueHandler()},
		{"hosts global write", http.MethodPut, "/api/v1/orchestrator/hosts/settings/values/ENABLE_CANARY", `{"value":"true"}`, SetHostsGlobalSettingValueHandler()},
		{"hosts global delete", http.MethodDelete, "/api/v1/orchestrator/hosts/settings/values/ENABLE_CANARY", "", DeleteHostsGlobalSettingValueHandler()},
		{"host write", http.MethodPut, "/api/v1/orchestrator/hosts/h1/settings/values/ENABLE_CANARY", `{"value":"true"}`, SetHostSettingValueHandler()},
		{"host delete", http.MethodDelete, "/api/v1/orchestrator/hosts/h1/settings/values/ENABLE_CANARY", "", DeleteHostSettingValueHandler()},
		{"host values", http.MethodGet, "/api/v1/orchestrator/hosts/h1/settings/values", "", GetHostValuesHandler()},
		{"sync", http.MethodPost, "/api/v1/orchestrator/settings/sync", `{"target":"all"}`, SyncSettingsHandler()},
		{"apply", http.MethodPost, "/api/v1/settings/apply", `{"revision":1,"hash":"","values":{}}`, ApplySettingsBundleHandler()},
	}
	for _, tc := range cases {
		rw := doRequest(t, tc.handler, tc.method, tc.target, tc.body)
		if rw.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: expected 503 with inert service, got %d (%s)", tc.name, rw.Code, rw.Body.String())
		}
	}
}

// Sanity: request body mapping used by the write handlers behaves through
// http_helper for the settings request shapes.
func TestSettingValueRequestMapsBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/values/LOG_LEVEL", strings.NewReader(`{"value":" hello "}`))
	var request models.SettingValueRequest
	if err := http_helper.MapRequestBody(req, &request); err != nil {
		t.Fatalf("body mapping failed: %v", err)
	}
	if request.Value != " hello " {
		t.Fatalf("value must be passed through verbatim for service-side normalization, got %q", request.Value)
	}
}

// The orchestrator value endpoints gained one way to address a single replica.
// The default must be untouched — every existing caller and UI still writes the
// shared layer — and ?node= must land on the per-replica layer keyed by the
// machine id, which is the only place a per-replica identity may be stored.
func TestOrchestratorWriteScopeAddressing(t *testing.T) {
	cases := []struct {
		name   string
		target string
		want   string
	}{
		{"default stays the shared layer",
			"/api/v1/orchestrator/settings/values/LOG_LEVEL", data_models.SettingScopeOrchestrator},
		{"empty node is not a machine",
			"/api/v1/orchestrator/settings/values/LOG_LEVEL?node=", data_models.SettingScopeOrchestrator},
		{"blank node is not a machine either",
			"/api/v1/orchestrator/settings/values/LOG_LEVEL?node=%20%20", data_models.SettingScopeOrchestrator},
		{"one replica by machine id",
			"/api/v1/orchestrator/settings/values/ORCHESTRATOR_NODE_ID?node=6E2BF1C4-1B2A",
			"node:6E2BF1C4-1B2A"},
		{"space around the id is not a different machine",
			"/api/v1/orchestrator/settings/values/ORCHESTRATOR_NODE_ID?node=%206E2BF1C4-1B2A%20",
			"node:6E2BF1C4-1B2A"},
	}

	for _, tc := range cases {
		if got := orchestratorWriteScope(httptest.NewRequest(http.MethodPut, tc.target, nil)); got != tc.want {
			t.Errorf("%s: scope %q, want %q", tc.name, got, tc.want)
		}
	}
}
