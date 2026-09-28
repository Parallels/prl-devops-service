package orchestrator

import (
	"crypto/tls"
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
