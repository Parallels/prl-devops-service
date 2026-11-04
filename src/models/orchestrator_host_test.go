package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Validate() must split a full host URL into its parts so that callers can
// persist the bare hostname in Host (the canonical storage form used by the
// register mapper). Storing the full URL in Host breaks consumers that build
// schema://host:port from the parts.
func TestOrchestratorHostUpdateRequestValidate_FullUrl(t *testing.T) {
	req := &OrchestratorHostUpdateRequest{
		Host: "https://usseapdbmm4262.ads.autodesk.com:3080/api",
	}
	require.NoError(t, req.Validate())

	assert.Equal(t, "usseapdbmm4262.ads.autodesk.com", req.HostName)
	assert.Equal(t, "3080", req.Port)
	assert.Equal(t, "https", req.Schema)
	assert.Equal(t, "/api", req.Prefix)
}

func TestOrchestratorHostUpdateRequestValidate_SchemelessHostWithPort(t *testing.T) {
	req := &OrchestratorHostUpdateRequest{
		Host: "host.example.com:8080",
	}
	require.NoError(t, req.Validate())

	assert.Equal(t, "host.example.com", req.HostName)
	assert.Equal(t, "8080", req.Port)
	assert.Equal(t, "http", req.Schema)
	assert.Equal(t, "/api", req.Prefix)
}
