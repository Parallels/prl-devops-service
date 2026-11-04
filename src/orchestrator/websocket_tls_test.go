package orchestrator

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestHostWebSocketSelfSignedTLS(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Reading handles protocol ping/pong frames for the probe.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	config.New(ctx)
	host := &models.OrchestratorHost{ID: "tls-test", Host: u.Hostname(), Port: u.Port(), Schema: u.Scheme}
	for _, value := range []string{"", "true", "false"} {
		t.Run("validation_disabled="+value, func(t *testing.T) {
			t.Setenv(constants.TLS_DISABLE_VALIDATION_ENV_VAR, value)
			client := NewHostWebSocketClient(ctx, host, nil)
			defer client.Close()
			wantSuccess := value == "true"
			require.Equal(t, wantSuccess, client.Probe())
			err := client.establishConnection([]constants.EventType{constants.EventTypeJobManager})
			if wantSuccess {
				require.NoError(t, err)
				require.True(t, client.IsConnected())
			} else {
				require.Error(t, err)
				var certificateError *tls.CertificateVerificationError
				require.ErrorAs(t, err, &certificateError)
				require.False(t, client.IsConnected())
			}
		})
	}
}

// Hosts saved through the update endpoint used to store Host as a full URL
// (e.g. "https://host:3080/api"). The WebSocket client must still dial the
// real host instead of building "https://https://host:3080/api:3080" and
// resolving the literal hostname "https".
func TestHostWebSocketProbeWithFullUrlHost(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	config.New(ctx)
	t.Setenv(constants.TLS_DISABLE_VALIDATION_ENV_VAR, "true")

	host := &models.OrchestratorHost{
		ID:     "full-url-host",
		Host:   fmt.Sprintf("https://%s:%s/api", u.Hostname(), u.Port()),
		Port:   u.Port(),
		Schema: "https",
	}

	client := NewHostWebSocketClient(ctx, host, nil)
	defer client.Close()
	require.True(t, client.Probe())
	require.NoError(t, client.establishConnection([]constants.EventType{constants.EventTypeJobManager}))
	require.True(t, client.IsConnected())
}

func TestNormalizeHostEndpoint(t *testing.T) {
	cases := []struct {
		name       string
		host       models.OrchestratorHost
		wantName   string
		wantPort   string
		wantSchema string
	}{
		{
			name:       "full url stored by update endpoint extracts schema",
			host:       models.OrchestratorHost{Host: "https://host.example.com:3080/api", Port: "3080", Schema: "https"},
			wantName:   "host.example.com",
			wantPort:   "3080",
			wantSchema: "https",
		},
		{
			name:       "embedded url schema wins over stale schema field",
			host:       models.OrchestratorHost{Host: "http://host.example.com:9090/prefix", Port: "3080", Schema: "https"},
			wantName:   "host.example.com",
			wantPort:   "9090",
			wantSchema: "http",
		},
		{
			name:       "full url without port keeps port field",
			host:       models.OrchestratorHost{Host: "http://host.example.com/prefix", Port: "3080", Schema: ""},
			wantName:   "host.example.com",
			wantPort:   "3080",
			wantSchema: "http",
		},
		{
			name:       "url schema extracted when schema field empty",
			host:       models.OrchestratorHost{Host: "https://host.example.com:3080", Port: "3080", Schema: ""},
			wantName:   "host.example.com",
			wantPort:   "3080",
			wantSchema: "https",
		},
		{
			name:       "canonical bare hostname takes schema from field",
			host:       models.OrchestratorHost{Host: "host.example.com", Port: "3080", Schema: "https"},
			wantName:   "host.example.com",
			wantPort:   "3080",
			wantSchema: "https",
		},
		{
			name:       "schemeless host with port splits port, schema from field",
			host:       models.OrchestratorHost{Host: "host.example.com:3080", Port: "", Schema: "https"},
			wantName:   "host.example.com",
			wantPort:   "3080",
			wantSchema: "https",
		},
		{
			name:       "schemeless host with port and no schema defaults to http",
			host:       models.OrchestratorHost{Host: "host.example.com:8080", Port: "", Schema: ""},
			wantName:   "host.example.com",
			wantPort:   "8080",
			wantSchema: "http",
		},
		{
			name:       "bare hostname without any schema defaults to http",
			host:       models.OrchestratorHost{Host: "host.example.com", Port: "8080", Schema: ""},
			wantName:   "host.example.com",
			wantPort:   "8080",
			wantSchema: "http",
		},
		{
			name:       "bracketed ipv6 without schema",
			host:       models.OrchestratorHost{Host: "[::1]:8080", Port: "", Schema: ""},
			wantName:   "[::1]",
			wantPort:   "8080",
			wantSchema: "http",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, port, schema := normalizeHostEndpoint(&tc.host)
			require.Equal(t, tc.wantName, name)
			require.Equal(t, tc.wantPort, port)
			require.Equal(t, tc.wantSchema, schema)
		})
	}
}

// Hosts given without any schema must default to http/ws and still dial:
// both the "host:port" single-field shape and the split name+port shape.
func TestHostWebSocketProbeWithoutSchema(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	cases := []struct {
		name string
		host *models.OrchestratorHost
	}{
		{
			name: "host and port in one field, no schema",
			host: &models.OrchestratorHost{ID: "schemeless-combined", Host: u.Host},
		},
		{
			name: "host and port split, no schema",
			host: &models.OrchestratorHost{ID: "schemeless-split", Host: u.Hostname(), Port: u.Port()},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := basecontext.NewBaseContext()
			ctx.DisableLog()
			config.New(ctx)
			client := NewHostWebSocketClient(ctx, tc.host, nil)
			defer client.Close()
			require.True(t, client.Probe())
			require.NoError(t, client.establishConnection([]constants.EventType{constants.EventTypeJobManager}))
			require.True(t, client.IsConnected())
		})
	}
}
