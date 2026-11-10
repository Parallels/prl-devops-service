package parity

// Hybrid-delegate parity (task 04, re-pointed in task 05): representative
// UNPORTED domains through GormDatabase must be observationally identical to
// the legacy handle. Roles/claims can no longer serve as the example — they
// are real SQL since task 05 — so the proof rides the users domain (delegates
// until task 06). Lives here, not in package database: see
// gorm_database_test.go header for the legacy-goroutine rationale.

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
			t.Fatalf("seed role %s: %v", r, err)
		}
	}
	for _, c := range constants.DefaultClaims {
		if _, err := legacy.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed claim %s: %v", c, err)
		}
	}
	user := models.User{ID: "delegate-u1", Email: "delegate@test.local", Username: "delegate1", Name: "Delegate One"}
	if _, err := legacy.CreateUser(ctx, user); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	// Hybrid engine: migrated :memory: SQL + this legacy handle.
	sql := newSQLTestEngine(t, legacy)

	lu, lerr := legacy.GetUsers(ctx, "username=delegate")
	gu, gerr := sql.GetUsers(ctx, "username=delegate")
	if lerr != nil || gerr != nil {
		t.Fatalf("GetUsers errors: %v %v", lerr, gerr)
	}
	if len(lu) != 1 || len(gu) != 1 || lu[0].ID != gu[0].ID {
		t.Errorf("users delegate drift: %v vs %v", lu, gu)
	}

	lu2, _ := legacy.GetUsers(ctx, "username=DELEGATE1,i")
	gu2, _ := sql.GetUsers(ctx, "username=DELEGATE1,i")
	if len(lu2) != len(gu2) {
		t.Errorf("filter-through-delegate drift: %d vs %d", len(lu2), len(gu2))
	}

	// A ported domain (roles) must NOT touch the delegate: roles were seeded
	// legacy-side only (DefaultRoles) — the SQL store is empty, so a non-empty
	// result here would mean reach-through into legacy.
	empty, err := sql.GetRoles(ctx, "")
	if err != nil || len(empty) != 0 {
		t.Errorf("ported domain reached through/failed: %v %+v", err, empty)
	}
}
