package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The hardware endpoint is how an operator reads what a replica thinks it is, and
// the two identity fields are a pair with a specific relationship: machine_id is
// the key that sets orchestrator_node_id. Their JSON names are therefore part of
// the public contract (the UI and any scripted fleet inventory read them), and
// both are empty on a host — which must stay omitted rather than appear as "".

func TestSystemUsageResponseIdentityFieldsAreOmittedWhenUnset(t *testing.T) {
	// A host answers this endpoint too and reports no replica identity.
	payload, err := json.Marshal(SystemUsageResponse{TotalVms: 3})
	require.NoError(t, err)

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &fields))

	assert.NotContains(t, fields, "orchestrator_node_id")
	assert.NotContains(t, fields, "machine_id")
	assert.Contains(t, fields, "total_vms")
}

func TestSystemUsageResponseIdentityFieldNames(t *testing.T) {
	payload, err := json.Marshal(SystemUsageResponse{
		OrchestratorNodeID: "orchestrator-1",
		MachineID:          "6E2BF1C4-1B2A",
		EnabledModules:     []string{"orchestrator", "placement"},
		SqlDatabaseType:    "sqlite",
	})
	require.NoError(t, err)

	// Renaming either tag is a breaking API change: it silently drops the value
	// for every reader, which for machine_id means nobody can address the
	// replica's settings any more.
	//
	// Three fields are always emitted (no omitempty) and are listed to keep the
	// comparison exact rather than to test them: they are the pre-existing shape.
	assert.JSONEq(t, `{
		"orchestrator_node_id": "orchestrator-1",
		"machine_id": "6E2BF1C4-1B2A",
		"enabled_modules": ["orchestrator", "placement"],
		"sql_database_type": "sqlite",
		"pending_mac_vms": 0,
		"is_reverse_proxy_enabled": false,
		"is_log_streaming_enabled": false
	}`, string(payload))

	var decoded SystemUsageResponse
	require.NoError(t, json.Unmarshal(payload, &decoded))
	assert.Equal(t, "orchestrator-1", decoded.OrchestratorNodeID)
	assert.Equal(t, "6E2BF1C4-1B2A", decoded.MachineID)
}
