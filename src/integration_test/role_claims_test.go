//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoleInheritedClaims_GrantsAndRevokes is the canonical integration test for
// role-based claim inheritance:
//
//  1. An admin creates a custom role "TESTER" and adds the LIST_USER claim to it.
//  2. A new user is created with only the TESTER role (no direct LIST_USER claim).
//  3. The user obtains a JWT — the JWT must contain LIST_USER via inheritance.
//  4. GET /auth/users succeeds (200).
//  5. The admin removes LIST_USER from the TESTER role.
//  6. The user obtains a fresh JWT — LIST_USER must no longer be present.
//  7. GET /auth/users is denied (401 or 403).
func TestRoleInheritedClaims_GrantsAndRevokes(t *testing.T) {
	e := newTestEnv(t)

	// ── Admin token ───────────────────────────────────────────────────────────
	adminToken := e.mustGetToken(t, adminEmail, adminPassword)

	// ── 1. Create custom role ─────────────────────────────────────────────────
	resp := e.do(t, http.MethodPost, "/auth/roles",
		map[string]interface{}{"name": "TESTER"},
		adminToken)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create TESTER role")

	var roleResp struct {
		ID string `json:"id"`
	}
	require.NoError(t, jsonDecode(resp.Body, &roleResp))
	roleID := roleResp.ID
	require.NotEmpty(t, roleID)

	// ── 2. Add LIST_USER claim to the TESTER role ─────────────────────────────
	resp2 := e.do(t, http.MethodPost, "/auth/roles/"+roleID+"/claims",
		map[string]string{"name": "LIST_USER"},
		adminToken)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusCreated, resp2.StatusCode, "add LIST_USER claim to TESTER role")

	// ── 3. Create test user with TESTER role only (no direct LIST_USER claim) ──
	const (
		testerEmail = "tester@integration.test"
		testerPass  = "tester123"
	)
	resp3 := e.do(t, http.MethodPost, "/auth/users",
		map[string]interface{}{
			"email":    testerEmail,
			"username": "tester",
			"name":     "Tester User",
			"password": testerPass,
			"roles":    []string{roleID},
		},
		adminToken)
	defer resp3.Body.Close()
	require.Equal(t, http.StatusCreated, resp3.StatusCode, "create tester user")

	// ── 4. Tester obtains a JWT — LIST_USER must be present via inheritance ────
	testerToken1 := e.mustGetToken(t, testerEmail, testerPass)
	require.NotEmpty(t, testerToken1)

	// ── 5. GET /auth/users — expect 200 ───────────────────────────────────────
	resp4 := e.do(t, http.MethodGet, "/auth/users", nil, testerToken1)
	defer resp4.Body.Close()
	assert.Equal(t, http.StatusOK, resp4.StatusCode,
		"tester should be able to list users via inherited LIST_USER claim")

	// ── 6. Admin removes LIST_USER from TESTER role ───────────────────────────
	resp5 := e.do(t, http.MethodDelete, "/auth/roles/"+roleID+"/claims/LIST_USER",
		nil, adminToken)
	defer resp5.Body.Close()
	require.Equal(t, http.StatusAccepted, resp5.StatusCode, "remove LIST_USER from TESTER role")

	// ── 7. Tester gets a FRESH token — LIST_USER must no longer appear ─────────
	testerToken2 := e.mustGetToken(t, testerEmail, testerPass)
	require.NotEmpty(t, testerToken2)
	require.NotEqual(t, testerToken1, testerToken2, "expected a new JWT after role change")

	// ── 8. GET /auth/users — expect 401 or 403 ────────────────────────────────
	resp6 := e.do(t, http.MethodGet, "/auth/users", nil, testerToken2)
	defer resp6.Body.Close()
	assert.True(t,
		resp6.StatusCode == http.StatusUnauthorized || resp6.StatusCode == http.StatusForbidden,
		"tester should be denied after LIST_USER claim removed from role, got %d", resp6.StatusCode)
}
