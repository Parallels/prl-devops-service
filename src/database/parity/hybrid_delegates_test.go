package parity

// Hybrid-delegate parity (task 04): representative domains through
// GormDatabase must be observationally identical to the legacy handle.
// Lives here, not in package database, because the legacy engine's
// background save goroutine shares logger/config globals with this
// package's serialised engine tests — see gorm_database_test.go header.

import (
	"reflect"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
)

func TestHybridDelegatesMatchLegacy(t *testing.T) {
	legacy := NewLegacyEngine(t)
	ctx := basecontext.NewBaseContext()
	if _, err := legacy.CreateRole(ctx, models.Role{Name: "delegate-probe"}); err != nil {
		t.Fatalf("seed role: %v", err)
	}
	// roles/claims required by CreateUser backfill come from the defaults:
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

	g := database.NewGormDatabase(ctx, nil, legacy)
	lr, lerr := legacy.GetRoles(ctx, "")
	gr, gerr := g.GetRoles(ctx, "")
	if !reflect.DeepEqual(lr, gr) || lerr != nil || gerr != nil {
		t.Errorf("roles delegate drift: %v vs %v (%v/%v)", roleNames(lr), roleNames(gr), lerr, gerr)
	}

	lu, lerr := legacy.GetUsers(ctx, "username=delegate")
	gu, gerr := g.GetUsers(ctx, "username=delegate")
	if lerr != nil || gerr != nil {
		t.Fatalf("GetUsers errors: %v %v", lerr, gerr)
	}
	if len(lu) != 1 || len(gu) != 1 || lu[0].ID != gu[0].ID {
		t.Errorf("users delegate drift: %v vs %v", lu, gu)
	}

	// filter language flows through the delegate unchanged too (query parity
	// of the delegate path is guaranteed by construction — same code).
	lu2, _ := legacy.GetUsers(ctx, "username=DELEGATE1,i")
	gu2, _ := g.GetUsers(ctx, "username=DELEGATE1,i")
	if len(lu2) != len(gu2) {
		t.Errorf("filter-through-delegate drift: %d vs %d", len(lu2), len(gu2))
	}
}

func roleNames(rs []models.Role) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Name)
	}
	return out
}
