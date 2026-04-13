package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

func newCORSPreflightRequest(origin, method, headers string) *http.Request {
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/token", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", method)
	if headers != "" {
		req.Header.Set("Access-Control-Request-Headers", headers)
	}

	return req
}

func TestBuildCORSHandler_MergesConfiguredHeadersWithDefaults(t *testing.T) {
	t.Setenv(constants.CORS_ALLOWED_HEADERS_ENV_VAR, "Authorization")
	t.Setenv(constants.CORS_ALLOWED_ORIGINS_ENV_VAR, "http://localhost:1421")

	cfg := config.New(basecontext.NewBaseContext())
	handler := buildCORSHandler(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := newCORSPreflightRequest("http://localhost:1421", http.MethodPost, "content-type")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "http://localhost:1421", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Content-Type", rr.Header().Get("Access-Control-Allow-Headers"))
}

func TestBuildCORSHandler_AllowsPatchPreflightByDefault(t *testing.T) {
	cfg := config.New(basecontext.NewBaseContext())
	handler := buildCORSHandler(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := newCORSPreflightRequest("http://localhost:1421", http.MethodPatch, "content-type")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.MethodPatch, rr.Header().Get("Access-Control-Allow-Methods"))
}

func TestBuildCORSHandler_TrimsConfiguredOrigins(t *testing.T) {
	t.Setenv(constants.CORS_ALLOWED_ORIGINS_ENV_VAR, "http://localhost:3000, http://localhost:1421")

	cfg := config.New(basecontext.NewBaseContext())
	handler := buildCORSHandler(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := newCORSPreflightRequest("http://localhost:1421", http.MethodPost, "content-type")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "http://localhost:1421", rr.Header().Get("Access-Control-Allow-Origin"))
}
