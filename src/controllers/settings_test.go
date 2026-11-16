package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/models"
	"github.com/cjlapao/common-go/helper/http_helper"
)

// The settings service is uninitialized in unit context (no SQL handle) —
// every handler must answer 503, never panic, and never write a body leak.

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
		{"orch listing", http.MethodGet, "/api/v1/settings", "", GetSettingsHandler()},
		{"orch write", http.MethodPut, "/api/v1/settings/values/hosts/MAX_VMS_PER_HOST", `{"value":"1"}`, SetSettingValueHandler()},
		{"orch delete", http.MethodDelete, "/api/v1/settings/values/hosts/MAX_VMS_PER_HOST", "", DeleteSettingValueHandler()},
		{"orch host effective", http.MethodGet, "/api/v1/settings/hosts/h1/effective", "", GetHostEffectiveSettingsHandler()},
		{"orch sync", http.MethodPost, "/api/v1/settings/sync", `{"target":"all"}`, SyncSettingsHandler()},
		{"host listing", http.MethodGet, "/api/v1/settings/node", "", GetNodeSettingsHandler()},
		{"host local write", http.MethodPut, "/api/v1/settings/node/MAX_VMS_PER_HOST", `{"value":"1"}`, SetNodeSettingHandler()},
		{"host local delete", http.MethodDelete, "/api/v1/settings/node/MAX_VMS_PER_HOST", "", DeleteNodeSettingHandler()},
		{"host apply", http.MethodPost, "/api/v1/settings/apply", `{"revision":1,"hash":"","values":{}}`, ApplySettingsBundleHandler()},
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
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/values/hosts/X", strings.NewReader(`{"value":" hello "}`))
	var request models.SettingValueRequest
	if err := http_helper.MapRequestBody(req, &request); err != nil {
		t.Fatalf("body mapping failed: %v", err)
	}
	if request.Value != " hello " {
		t.Fatalf("value must be passed through verbatim for service-side normalization, got %q", request.Value)
	}
}
