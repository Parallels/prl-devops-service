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

// The tags field distinguishes "leave untouched" (nil) from "clear" (empty):
// Validate normalizes entries but must never collapse an explicit empty list
// into nil, and must never turn an absent field into one.
func TestOrchestratorHostUpdateRequestValidate_Tags(t *testing.T) {
	req := &OrchestratorHostUpdateRequest{}
	require.NoError(t, req.Validate())
	assert.Nil(t, req.Tags, "absent tags stay nil")

	req = &OrchestratorHostUpdateRequest{Tags: []string{}}
	require.NoError(t, req.Validate())
	require.NotNil(t, req.Tags, "an explicit clear must survive validation")
	assert.Empty(t, req.Tags)

	req = &OrchestratorHostUpdateRequest{Tags: []string{" Pool=Build ", "", "pool=build", "os=sonoma"}}
	require.NoError(t, req.Validate())
	assert.Equal(t, []string{"Pool=Build", "os=sonoma"}, req.Tags, "trimmed, blanks dropped, case-insensitive dedupe keeping the first spelling")
}
