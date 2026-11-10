package models_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/require"
)

// Golden-JSON harness (PBI #419): pins json.Marshal output BYTE-for-byte
// (key order, typos, omitempty behaviour included) — deliberately not
// assert.JSONEq, which forgives exactly what we must not forgive.
//
// Capture-first protocol: fixtures + golden files land BEFORE the tag commit
// (GOLDEN_CAPTURE=1 go test -run TestCaptureGolden), and the tag commit must
// leave both this file and testdata/ untouched.

func fixtureUser(full bool) models.User {
	u := models.User{ID: "u-1", Username: "alice", Name: "Alice", Email: "alice@example.com"}
	if !full {
		return u
	}
	u.Password = "hashed-secret"
	u.CreatedAt = "2024-01-02 03:04:05"
	u.UpdatedAt = "2024-02-03 04:05:06"
	u.Roles = []models.Role{fixtureRole(true)}
	u.Claims = []models.Claim{fixtureClaim(true)}
	u.FailedLoginAttempts = 3
	u.Blocked = true
	u.BlockedSince = "2024-03-04 05:06:07"
	u.BlockedReason = "manual block"
	return u
}

func fixtureRole(full bool) models.Role {
	r := models.Role{ID: "r-1", Name: "admin"}
	if !full {
		return r
	}
	r.Description = "administrators"
	r.Internal = true
	r.Claims = []models.Claim{fixtureClaim(true)}
	return r
}

func fixtureClaim(full bool) models.Claim {
	c := models.Claim{ID: "c-1", Name: "vm.read"}
	if !full {
		return c
	}
	c.Description = "read virtual machines"
	c.Internal = true
	c.Group = "vm"
	c.Resource = "virtual_machine"
	c.Action = "read"
	return c
}

func fixtureApiKey(full bool) models.ApiKey {
	k := models.ApiKey{ID: "k-1", Name: "ci-key"}
	if !full {
		return k
	}
	k.Key = "pub-key-material"
	k.Secret = "secret-material"
	k.Revoked = true
	k.CreatedAt = "2024-04-05 06:07:08"
	k.UpdatedAt = "2024-05-06 07:08:09"
	k.RevokedAt = "2024-06-07 08:09:10"
	k.ExpiresAt = "2025-07-08 09:10:11"
	k.UserID = "u-1"
	k.Type = "internal"
	k.DbRecord = &models.DbRecord{IsLocked: true, LockedBy: "u-1", LockedAt: "2024-08-09 10:11:12"}
	return k
}

func fixtureRoleClaims() models.RoleClaims {
	return models.RoleClaims{RoleID: "r-1", ClaimID: "c-1"}
}

func goldenTargets() map[string]func() interface{} {
	return map[string]func() interface{}{
		"user.json":         func() interface{} { return fixtureUser(true) },
		"user_minimal.json": func() interface{} { return fixtureUser(false) },
		"role.json":         func() interface{} { return fixtureRole(true) },
		"claim.json":        func() interface{} { return fixtureClaim(true) },
		"api_key.json":      func() interface{} { return fixtureApiKey(true) },
		"role_claims.json":  func() interface{} { return fixtureRoleClaims() },
	}
}

// TestCaptureGolden writes the golden files; active only under
// GOLDEN_CAPTURE=1 (capture commits only; CI runs stay hermetic).
//
//go:generate env GOLDEN_CAPTURE=1 go test -run TestCaptureGolden -count=1
func TestCaptureGolden(t *testing.T) {
	if os.Getenv("GOLDEN_CAPTURE") != "1" {
		t.Skip("set GOLDEN_CAPTURE=1 to (re)generate fixtures")
	}
	require.NoError(t, os.MkdirAll("testdata/golden", 0o755))
	for name, build := range goldenTargets() {
		out, err := json.Marshal(build())
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join("testdata", "golden", name), out, 0o644))
	}
}

func assertGolden(t *testing.T, file string, v interface{}) {
	t.Helper()
	got, err := json.Marshal(v)
	require.NoError(t, err)
	want, err := os.ReadFile(filepath.Join("testdata", "golden", file))
	require.NoError(t, err, "golden fixture missing — capture protocol broken")
	require.Equal(t, strings.TrimSpace(string(want)), string(got), "golden byte drift: %s", file)
}

func TestUser_JSONGolden(t *testing.T) { assertGolden(t, "user.json", fixtureUser(true)) }
func TestUserMinimal_JSONGolden(t *testing.T) {
	assertGolden(t, "user_minimal.json", fixtureUser(false))
}
func TestRole_JSONGolden(t *testing.T)   { assertGolden(t, "role.json", fixtureRole(true)) }
func TestClaim_JSONGolden(t *testing.T)  { assertGolden(t, "claim.json", fixtureClaim(true)) }
func TestApiKey_JSONGolden(t *testing.T) { assertGolden(t, "api_key.json", fixtureApiKey(true)) }

func TestRoleClaims_RoundTrip(t *testing.T) {
	first, err := json.Marshal(fixtureRoleClaims())
	require.NoError(t, err)
	var back models.RoleClaims
	require.NoError(t, json.Unmarshal(first, &back))
	second, err := json.Marshal(back)
	require.NoError(t, err)
	assertGolden(t, "role_claims.json", back)
	require.Equal(t, string(first), string(second))
}
