package restapi

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/require"
)

func TestHttpListenerStart_HTTPSOnlyRegistersControllers(t *testing.T) {
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api")
	t.Setenv(constants.ENABLE_CORS_ENV_VAR, "true")

	// Use a valid test certificate so startup takes the HTTPS-only branch.
	fixture := httptest.NewTLSServer(http.NotFoundHandler())
	certificate := fixture.TLS.Certificates[0]
	fixture.Close()
	privateKey, err := x509.MarshalPKCS8PrivateKey(certificate.PrivateKey)
	require.NoError(t, err)

	previousListener, previousInitialized := globalHttpListener, Initialized
	previousShutdown, previousRestart := shutdown, needsRestart
	globalHttpListener = nil
	Initialized = make(chan bool, 1)
	t.Cleanup(func() {
		globalHttpListener, Initialized = previousListener, previousInitialized
		shutdown, needsRestart = previousShutdown, previousRestart
	})

	listener := NewHttpListener()
	listener.Options.ApiPrefix = "/api"
	listener.Options.HttpPort = "0"
	listener.Options.EnableTLS = true
	listener.Options.DisableHttpWhenTls = true
	listener.Options.TLSPort = "0"
	listener.Options.TLSCertificate = string(pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE", Bytes: certificate.Certificate[0],
	}))
	listener.Options.TLSPrivateKey = string(pem.EncodeToMemory(&pem.Block{
		Type: "PRIVATE KEY", Bytes: privateKey,
	}))
	listener.WithVersion("Version 1", "v1", true)
	// Follow the same deferred registration path used by the auth controller.
	NewController().WithMethod(POST).WithVersion("v1").WithPath("/auth/token").
		WithHandler(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}).Register()

	stopped := make(chan struct{})
	go func() {
		listener.Start("test", "test")
		close(stopped)
	}()
	t.Cleanup(func() {
		select {
		case listener.shutdownRequest <- true:
		case <-time.After(5 * time.Second):
			t.Fatal("listener did not accept shutdown request")
		}
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			t.Fatal("listener did not stop")
		}
	})
	select {
	case <-Initialized:
	case <-time.After(5 * time.Second):
		t.Fatal("listener did not initialize")
	}

	require.Len(t, listener.Servers, 1, "HTTPS-only mode must not start HTTP")
	require.NotNil(t, listener.Servers[0].TLSConfig)
	server := httptest.NewUnstartedServer(listener.Servers[0].Handler)
	server.TLS = listener.Servers[0].TLSConfig.Clone()
	server.StartTLS()
	defer server.Close()
	client := server.Client()
	client.Timeout = 5 * time.Second

	for _, path := range []string{"/api/v1/auth/token", "/api/auth/token"} {
		response, err := client.Post(server.URL+path, "application/json",
			strings.NewReader(`{"email":"root","password":"test"}`))
		require.NoError(t, err)
		response.Body.Close()
		require.Equal(t, http.StatusNoContent, response.StatusCode, path)
	}
}
