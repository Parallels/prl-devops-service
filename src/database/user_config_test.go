package database

import (
	"errors"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func seedCfg(t *testing.T, g *GormDatabase, id, userID, slug, name, value string) {
	t.Helper()
	if err := g.gormDB.Table("user_configs").Create(&models.UserConfig{
		ID: id, UserID: userID, Slug: slug, Name: name, Type: "string", Value: value,
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}).Error; err != nil {
		t.Fatal(err)
	}
}

func cfgCount(t *testing.T, g *GormDatabase, userID string) int64 {
	t.Helper()
	var n int64
	if err := g.gormDB.Table("user_configs").Where("user_id = ?", userID).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestSQLUserConfigs_ListingAndOwnerFold(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedCfg(t, g, "c1", "u1", "theme", "Theme", "dark")
	seedCfg(t, g, "c2", "u1", "layout", "Layout", "grid")
	seedCfg(t, g, "c3", "u2", "theme", "Theme", "light") // same slug, other owner

	l, err := g.GetUserConfigs(ctx, "U1", "") // owner match is EqualFold (legacy :29)
	if err != nil || len(l) != 2 {
		t.Fatalf("folded owner listing: %v %d", err, len(l))
	}
	// nil-vs-empty pin: zero-owner answer is an EMPTY, non-nil slice (legacy :27).
	none, err := g.GetUserConfigs(ctx, "nobody", "")
	if err != nil || none == nil || len(none) != 0 {
		t.Fatalf("empty shape: %v %#v", err, none)
	}
	// Go regexp filter, unanchored: slug=the.* keeps c3 only.
	r, err := g.GetUserConfigs(ctx, "u2", "slug=the.*")
	if err != nil || len(r) != 1 || r[0].ID != "c3" {
		t.Fatalf("regexp filter: %v %+v", err, r)
	}
	if bad, err := g.GetUserConfigs(ctx, "u1", "id=nope"); err != nil || len(bad) != 0 {
		t.Fatalf("no-match filter: %v %d", err, len(bad))
	}
	if _, err := g.GetUserConfigs(ctx, "u1", "bogus~~filter"); err == nil {
		t.Error("want ParseFilter error")
	}
}

func TestSQLUserConfig_LookupAndSentinels(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedCfg(t, g, "c1", "u1", "theme", "Theme", "dark")
	seedCfg(t, g, "c2", "u2", "theme", "Theme", "light")
	seedCfg(t, g, "e1", "u3", "", "Empty", "x") // empty slug row for the "" needle quirk

	for _, probe := range []struct{ uid, slug string }{
		{"u1", "c1"}, {"u1", "C1"}, {"u1", "theme"}, {"U1", "THEME"},
	} {
		c, err := g.GetUserConfig(ctx, probe.uid, probe.slug)
		if err != nil || c.ID != "c1" {
			t.Fatalf("lookup %v: %v %+v", probe, err, c)
		}
	}
	// Cross-user probes: u1 cannot reach u2's row by id OR by shared slug.
	for _, probe := range []struct{ uid, slug string }{
		{"u1", "c2"}, {"u1", "c3-missing"}, {"nobody", "theme"}, {"u2", "c1"},
	} {
		if _, err := g.GetUserConfig(ctx, probe.uid, probe.slug); !errors.Is(err, interfaces.ErrUserConfigNotFound) {
			t.Fatalf("want NotFound for %v, got %v", probe, err)
		}
	}
	// Legacy has NO empty-input sentinel: "" matches a same-owner empty
	// ID-or-slug row through EqualFold (deviation from the spec text).
	hit, err := g.GetUserConfig(ctx, "u3", "")
	if err != nil || hit.ID != "e1" {
		t.Fatalf("empty-needle quirk: %v %+v", err, hit)
	}
}

func TestSQLUserConfig_CreateDupRules(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedCfg(t, g, "c1", "u1", "theme", "Theme", "dark")

	// Same owner, same slug ⇒ dup sentinel, no extra row.
	if _, err := g.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "theme"}); !errors.Is(err, interfaces.ErrUserConfigAlreadyExists) {
		t.Fatalf("want dup sentinel, got %v", err)
	}
	if n := cfgCount(t, g, "u1"); n != 1 {
		t.Fatalf("dup must not insert, count %d", n)
	}
	// Different owner, same slug ⇒ allowed (per-owner scoping).
	if _, err := g.CreateUserConfig(ctx, models.UserConfig{UserID: "u9", Slug: "theme"}); err != nil {
		t.Fatalf("cross-owner same slug: %v", err)
	}
	// Quirk: the create probe matches id OR slug — a new slug hitting an
	// existing row's ID is a dup in legacy (GetUserConfig needle semantics).
	if _, err := g.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "c1"}); !errors.Is(err, interfaces.ErrUserConfigAlreadyExists) {
		t.Fatalf("want id-vs-slug quirk dup, got %v", err)
	}
	// Empty ID gets generated before the probe; timestamps stamped.
	created, err := g.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "fresh"})
	if err != nil || created.ID == "" || created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("generate/stamp: %v %+v", err, created)
	}
}

func TestSQLUserConfig_UpdatePartial(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedCfg(t, g, "c1", "u1", "theme", "Theme", "dark")

	// Legacy copies ONLY Name/Type/Value (+UpdatedAt): Slug/CreatedAt/UserID
	// of the payload are ignored — the classic GORM partial-update trap.
	up, err := g.UpdateUserConfig(ctx, models.UserConfig{
		ID: "c1", UserID: "u1", Slug: "hijack", Name: "New Name", Type: "int", Value: "42", CreatedAt: "bogus",
	})
	if err != nil {
		t.Fatal(err)
	}
	if up.Slug != "theme" || up.CreatedAt != "2026-01-01T00:00:00Z" || up.Name != "New Name" || up.Value != "42" {
		t.Fatalf("partial copy drift: %+v", up)
	}
	stored, _ := g.GetUserConfig(ctx, "u1", "theme") // slug survived
	if stored.Name != "New Name" {
		t.Fatalf("store drift: %+v", stored)
	}
	// Match is id AND owner, EqualFold; wrong owner / unknown id ⇒ NotFound.
	if _, err := g.UpdateUserConfig(ctx, models.UserConfig{ID: "C1", UserID: "U1", Name: "Fold"}); err != nil {
		t.Fatalf("folded update match: %v", err)
	}
	if _, err := g.UpdateUserConfig(ctx, models.UserConfig{ID: "c1", UserID: "u2"}); !errors.Is(err, interfaces.ErrUserConfigNotFound) {
		t.Fatalf("wrong-owner update: %v", err)
	}
	if _, err := g.UpdateUserConfig(ctx, models.UserConfig{ID: "nope", UserID: "u1"}); !errors.Is(err, interfaces.ErrUserConfigNotFound) {
		t.Fatalf("unknown-id update: %v", err)
	}
}

func TestSQLUserConfig_UpsertBranches(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedCfg(t, g, "c1", "u1", "theme", "Theme", "dark")

	// Update branch: pins the EXISTING id (payload ID ignored), keeps slug
	// and CreatedAt, copies the rest.
	up, err := g.UpsertUserConfig(ctx, models.UserConfig{
		ID: "payload-ignored", UserID: "u1", Slug: "theme", Name: "Upserted", Value: "v2", CreatedAt: "bogus",
	})
	if err != nil || up.ID != "c1" || up.Name != "Upserted" || up.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Fatalf("upsert update-branch: %v %+v", err, up)
	}
	// Insert branch: missing slug creates with generated id + stamped times.
	ins, err := g.UpsertUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "brand-new", Name: "New", Value: "v"})
	if err != nil || ins.ID == "" || ins.CreatedAt == "" || ins.CreatedAt == "bogus" {
		t.Fatalf("upsert insert-branch: %v %+v", err, ins)
	}
	if n := cfgCount(t, g, "u1"); n != 2 {
		t.Fatalf("want 2 rows, got %d", n)
	}
}

func TestSQLUserConfig_Delete(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedCfg(t, g, "c1", "u1", "theme", "Theme", "dark")
	seedCfg(t, g, "c2", "u2", "theme", "Theme", "light")

	// Owner-scoped: u2 cannot delete u1's row even with the shared slug.
	if err := g.DeleteUserConfig(ctx, "u2", "c1"); !errors.Is(err, interfaces.ErrUserConfigNotFound) {
		t.Fatalf("cross-owner delete: %v", err)
	}
	if err := g.DeleteUserConfig(ctx, "U1", "THEME"); err != nil { // folded both keys
		t.Fatalf("delete by folded slug: %v", err)
	}
	if n := cfgCount(t, g, "u1"); n != 0 {
		t.Fatalf("row survived, count %d", n)
	}
	if err := g.DeleteUserConfig(ctx, "u1", "theme"); !errors.Is(err, interfaces.ErrUserConfigNotFound) {
		t.Fatalf("double delete: %v", err)
	}
	if _, err := g.GetUserConfig(ctx, "u2", "theme"); err != nil { // other owner untouched
		t.Fatalf("collateral damage: %v", err)
	}
}

func TestSQLUserConfigs_NotConnected(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cold := NewGormDatabase(ctx, nil)

	if _, err := cold.GetUserConfigs(ctx, "u1", ""); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("list guard: %v", err)
	}
	if _, err := cold.GetUserConfig(ctx, "u1", "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("get guard: %v", err)
	}
	if _, err := cold.CreateUserConfig(ctx, models.UserConfig{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("create guard: %v", err)
	}
	if _, err := cold.UpsertUserConfig(ctx, models.UserConfig{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("upsert guard: %v", err)
	}
	if _, err := cold.UpdateUserConfig(ctx, models.UserConfig{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("update guard: %v", err)
	}
	if err := cold.DeleteUserConfig(ctx, "u1", "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("delete guard: %v", err)
	}
}
