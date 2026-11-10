package parity

// Hybrid-delegate parity (task 04; re-pointed in 05 and 06): representative
// UNPORTED domains through GormDatabase must be observationally identical to
// the legacy handle. Since task 06 the users domain is ported too, so the
// delegate proof rides api-keys (unported until task 07) and the
// zero-reach-through assertion rides users. Lives here, not in package
// database: see gorm_database_test.go header for the legacy-goroutine
// rationale.

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
	if _, err := legacy.CreateApiKey(ctx, models.ApiKey{ID: "probe-key", Name: "probe", Key: "k-1", Secret: "s-1", UserID: "delegate-u1"}); err != nil {
		t.Fatalf("seed api key: %v", err)
	}

	// Hybrid engine: migrated :memory: SQL + this legacy handle.
	sql := newSQLTestEngine(t, legacy)

	// Unported domain (api keys): delegate pass-through must mirror legacy,
	// including the filter string reaching the legacy engine untouched.
	lk, lerr := legacy.GetApiKeys(ctx, "name=probe")
	gk, gerr := sql.GetApiKeys(ctx, "name=probe")
	if lerr != nil || gerr != nil {
		t.Fatalf("GetApiKeys errors: %v %v", lerr, gerr)
	}
	if len(lk) != len(gk) {
		t.Fatalf("api-keys delegate drift: %d vs %d", len(lk), len(gk))
	}
	if len(lk) > 0 && lk[0].ID != gk[0].ID {
		t.Errorf("api-keys delegate identity drift: %v vs %v", lk[0].ID, gk[0].ID)
	}

	// A ported domain (users) must NOT touch the delegate: the user was seeded
	// legacy-side only; the SQL store is empty, so a non-empty result here
	// would mean reach-through into legacy.
	empty, err := sql.GetUsers(ctx, "")
	if err != nil || len(empty) != 0 {
		t.Errorf("ported domain reached through/failed: %v %+v", err, empty)
	}
}
