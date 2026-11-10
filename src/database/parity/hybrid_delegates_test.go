package parity

// Hybrid-delegate parity (task 04; re-pointed in 05/06/07): representative
// UNPORTED domains through GormDatabase must be observationally identical to
// the legacy handle. Since task 07 the config-id trio (GetId/SeedId) is the
// unported representative; api-keys became the zero-reach-through subject.
// Lives here, not in package database: see gorm_database_test.go header for
// the legacy-goroutine rationale.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestHybridDelegatesMatchLegacy(t *testing.T) {
	legacy := NewLegacyEngine(t)
	ctx := basecontext.NewBaseContext()
	for _, r := range constants.DefaultRoles {
		if _, err := legacy.CreateRole(ctx, models.Role{Name: r}); err != nil {
			t.Fatalf("seed default role: %v", err)
		}
	}
	for _, c := range constants.DefaultClaims {
		if _, err := legacy.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed default claim: %v", err)
		}
	}
	user := models.User{ID: "delegate-u1", Email: "delegate@test.local", Username: "delegate1", Name: "Delegate One", Password: "x"}
	if _, err := legacy.CreateUser(ctx, user); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := legacy.CreateApiKey(ctx, models.ApiKey{ID: "probe-key", Name: "probe", Key: "k-1", Secret: "s-1", UserID: "delegate-u1"}); err != nil {
		t.Fatalf("seed api key: %v", err)
	}

	// Hybrid engine: migrated :memory: SQL + this legacy handle.
	sql := newSQLTestEngine(t, legacy)

	// Unported domain (config id): the read/write trio must ride the legacy
	// handle. SetId is single-write on a fresh store; SeedId would then hit
	// the "ID already exists" guard, so one write through each direction.
	written, err := sql.SetId(ctx, "hybrid-written-id")
	if err != nil || written != "hybrid-written-id" {
		t.Fatalf("hybrid SetId: %v %q", err, written)
	}
	legacyId, err := legacy.GetId(ctx)
	if err != nil || legacyId != "hybrid-written-id" {
		t.Errorf("delegate write not visible legacy-side: %q (%v)", legacyId, err)
	}
	hybridId, err := sql.GetId(ctx)
	if err != nil || hybridId != "hybrid-written-id" {
		t.Errorf("delegate read drift: %q (%v)", hybridId, err)
	}
	if _, err := legacy.SetId(ctx, "second-write"); err == nil {
		t.Error("legacy store must reject second write — proves both engines share ONE store")
	}

	// A ported domain (api keys) must NOT touch the delegate: the key was
	// seeded legacy-side only; the SQL store is empty, so a non-empty result
	// here would mean reach-through into legacy.
	empty, err := sql.GetApiKeys(ctx, "")
	if err != nil || len(empty) != 0 {
		t.Errorf("ported domain reached through/failed: %v %+v", err, empty)
	}
}
