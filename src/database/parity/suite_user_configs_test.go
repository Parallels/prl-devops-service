package parity

// DomainSuite "user_configs" (PBI #421 task 09): the leakage matrix — two
// owners sharing the slug "theme" (different values), listings per owner
// (folded + unknown uid) across plain/regexp filters, cross-user probes by
// BOTH id and shared slug, and the mutation paths (create-dup incl. the
// id-vs-slug quirk, per-owner dup allowance, partial-copy update, upsert
// branches, owner-scoped delete). ctx is IGNORED by this domain on both
// engines (legacy takes the owner from the userID parameter) — the auth
// matrix is out of scope here; the spec's owner-predicate rule is the only
// scope.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func normaliseCfgs(in []models.UserConfig) []models.UserConfig {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.UserConfig(nil), in...)
	for i := range out {
		out[i].CreatedAt, out[i].UpdatedAt = "", "" // engine clocks
	}
	return out
}

func seedUserConfigs(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, r := range constants.DefaultRoles {
		if _, err := db.CreateRole(ctx, models.Role{Name: r}); err != nil {
			t.Fatalf("seed default role %s: %v", r, err)
		}
	}
	for _, c := range constants.DefaultClaims {
		if _, err := db.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed default claim %s: %v", c, err)
		}
	}
	for _, id := range []string{"u1", "u2"} {
		if _, err := db.CreateUser(ctx, models.User{ID: id, Username: id, Email: id + "@x.io", Name: id, Password: "Seed-Pass-1!"}); err != nil {
			t.Fatalf("seed user %s: %v", id, err)
		}
	}
	// The mandatory pair pattern: BOTH slugs owned by BOTH owners, values
	// different — a missing owner predicate changes answers, not just counts.
	mk := func(id, uid, slug, value string) {
		if _, err := db.CreateUserConfig(ctx, models.UserConfig{ID: id, UserID: uid, Slug: slug, Name: slug, Type: "string", Value: value}); err != nil {
			t.Fatalf("seed cfg %s: %v", id, err)
		}
	}
	mk("cfg-u1-theme", "u1", "theme", "dark")
	mk("cfg-u2-theme", "u2", "theme", "light")
	mk("cfg-u1-layout", "u1", "layout", "u1-grid")
	mk("cfg-u2-layout", "u2", "layout", "u2-grid")
}

func TestParityDomain_UserConfigs(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "user_configs",
		Seed: seedUserConfigs,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkUserConfigListings,
			checkUserConfigCrossUser,
			checkUserConfigCreate,
			checkUserConfigUpdateUpsert,
			checkUserConfigDelete,
		},
	})
}

func checkUserConfigListings(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, uid := range []string{"u1", "U1", "u2", "nobody", ""} {
		for _, f := range []string{"", "slug=the.*", "id=nope", "value=grid"} {
			lc, lerr := legacy.GetUserConfigs(ctx, uid, f)
			sc, serr := sql.GetUserConfigs(ctx, uid, f)
			assertSameError(t, "GetUserConfigs["+uid+"/"+f+"]", lerr, serr)
			AssertSameList(t, "GetUserConfigs["+uid+"/"+f+"]", normaliseCfgs(lc), normaliseCfgs(sc))
		}
	}
}

func checkUserConfigCrossUser(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	probes := []struct{ uid, slug string }{
		{"u1", "theme"},        // own — must see dark, never light
		{"u1", "cfg-u2-theme"}, // u2's id under u1 ⇒ NotFound both
		{"u1", "layout"},       // u1 owns layout (grid); u2's layout value must not leak
		{"u2", "theme"},        // sees light
		{"nobody", "theme"},
		{"u1", ""}, // empty needle — no empty-slug rows seeded ⇒ NotFound both
	}
	for _, p := range probes {
		lc, lerr := legacy.GetUserConfig(ctx, p.uid, p.slug)
		sc, serr := sql.GetUserConfig(ctx, p.uid, p.slug)
		assertSameError(t, "GetUserConfig["+p.uid+"/"+p.slug+"]", lerr, serr)
		if lerr == nil && (lc == nil || sc == nil || lc.Value != sc.Value || lc.ID != sc.ID) {
			t.Errorf("lookup drift %v: %+v vs %+v", p, lc, sc)
		}
	}
}

func checkUserConfigCreate(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	// dup: same owner same slug.
	_, lerr := legacy.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "theme", Value: "nope"})
	_, serr := sql.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "theme", Value: "nope"})
	assertSameError(t, "dup create", lerr, serr)
	// id-vs-slug quirk: needle hits an existing row's ID.
	_, lerr = legacy.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "cfg-u1-theme"})
	_, serr = sql.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "cfg-u1-theme"})
	assertSameError(t, "id-vs-slug quirk create", lerr, serr)
	// per-owner allowance: u2 may own slug "theme".
	_, lerr = legacy.CreateUserConfig(ctx, models.UserConfig{ID: "cfg-x", UserID: "u2", Slug: "shared", Value: "1"})
	_, serr = sql.CreateUserConfig(ctx, models.UserConfig{ID: "cfg-x", UserID: "u2", Slug: "shared", Value: "1"})
	assertSameError(t, "cross-owner create", lerr, serr)
	// generated-id create: ids differ per engine ⇒ normalise to "GEN".
	lg, lerr := legacy.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "gen", Value: "g"})
	sg, serr := sql.CreateUserConfig(ctx, models.UserConfig{UserID: "u1", Slug: "gen", Value: "g"})
	assertSameError(t, "generated-id create", lerr, serr)
	if (lg == nil) != (sg == nil) || (lg != nil && lg.ID == "") || (sg != nil && sg.ID == "") {
		t.Fatalf("generated id: %+v vs %+v", lg, sg)
	}
	lc, _ := legacy.GetUserConfigs(ctx, "u1", "")
	sc, _ := sql.GetUserConfigs(ctx, "u1", "")
	lc, sc = normaliseCfgs(lc), normaliseCfgs(sc)
	for i := range lc {
		if lc[i].Slug == "gen" {
			lc[i].ID = "GEN"
		}
	}
	for i := range sc {
		if sc[i].Slug == "gen" {
			sc[i].ID = "GEN"
		}
	}
	AssertSameList(t, "listings after creates", lc, sc)
}

func checkUserConfigUpdateUpsert(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	// partial-copy update: payload carries bogus Slug/CreatedAt.
	lu, lerr := legacy.UpdateUserConfig(ctx, models.UserConfig{ID: "cfg-u1-theme", UserID: "u1", Slug: "hijack", Name: "Renamed", Type: "int", Value: "77", CreatedAt: "bogus"})
	su, serr := sql.UpdateUserConfig(ctx, models.UserConfig{ID: "cfg-u1-theme", UserID: "u1", Slug: "hijack", Name: "Renamed", Type: "int", Value: "77", CreatedAt: "bogus"})
	assertSameError(t, "partial update", lerr, serr)
	if lerr == nil && (lu.Slug != su.Slug || lu.Name != su.Name || lu.Value != su.Value ||
		lu.CreatedAt == "bogus" || su.CreatedAt == "bogus" || lu.CreatedAt == "" || su.CreatedAt == "") {
		// Cross-engine CreatedAt equality is impossible (seed clock); the
		// contract is "payload value ignored, original stamp survives".
		t.Errorf("partial update drift: %+v vs %+v", lu, su)
	}
	// wrong owner update ⇒ NotFound both.
	_, lerr = legacy.UpdateUserConfig(ctx, models.UserConfig{ID: "cfg-u1-theme", UserID: "u2"})
	_, serr = sql.UpdateUserConfig(ctx, models.UserConfig{ID: "cfg-u1-theme", UserID: "u2"})
	assertSameError(t, "wrong-owner update", lerr, serr)
	// upsert: existing slug ⇒ update branch (id pinned, CreatedAt preserved).
	lu, lerr = legacy.UpsertUserConfig(ctx, models.UserConfig{ID: "ignored", UserID: "u1", Slug: "theme", Name: "ViaUpsert", Value: "u"})
	su, serr = sql.UpsertUserConfig(ctx, models.UserConfig{ID: "ignored", UserID: "u1", Slug: "theme", Name: "ViaUpsert", Value: "u"})
	assertSameError(t, "upsert update", lerr, serr)
	if lerr == nil && (lu.ID != su.ID || lu.ID != "cfg-u1-theme" ||
		lu.CreatedAt == "bogus" || su.CreatedAt == "bogus") {
		t.Errorf("upsert update drift: %+v vs %+v", lu, su)
	}
	// upsert: missing slug ⇒ insert branch, equal shape modulo id/timestamps.
	lu, lerr = legacy.UpsertUserConfig(ctx, models.UserConfig{UserID: "u2", Slug: "brand", Name: "B", Value: "b"})
	su, serr = sql.UpsertUserConfig(ctx, models.UserConfig{UserID: "u2", Slug: "brand", Name: "B", Value: "b"})
	assertSameError(t, "upsert insert", lerr, serr)
	if lu == nil || su == nil || lu.ID == "" || su.ID == "" || lu.Slug != su.Slug || lu.CreatedAt == "" || su.CreatedAt == "" {
		t.Fatalf("upsert insert shape: %+v vs %+v", lu, su)
	}
}

func checkUserConfigDelete(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	// cross-owner delete attempt (u2 deleting u1's by shared slug) ⇒ NotFound both.
	assertSameError(t, "cross-owner delete",
		legacy.DeleteUserConfig(ctx, "u2", "cfg-u1-theme"),
		sql.DeleteUserConfig(ctx, "u2", "cfg-u1-theme"))
	assertSameError(t, "delete by slug (folded)",
		legacy.DeleteUserConfig(ctx, "U1", "THEME"),
		sql.DeleteUserConfig(ctx, "U1", "THEME"))
	assertSameError(t, "double delete",
		legacy.DeleteUserConfig(ctx, "u1", "theme"),
		sql.DeleteUserConfig(ctx, "u1", "theme"))
	AssertSameList(t, "listings after delete",
		normaliseGenIds(mustList(legacy, "u1")), normaliseGenIds(mustList(sql, "u1")))
}

// normaliseGenIds collapses engine-generated ids (different per engine) to
// a shared marker so listings can deep-compare.
func normaliseGenIds(in []models.UserConfig) []models.UserConfig {
	out := normaliseCfgs(in)
	for i := range out {
		if out[i].Slug == "gen" || out[i].Slug == "brand" {
			out[i].ID = "GEN"
		}
	}
	return out
}

func mustList(db datainterfaces.DatabaseService, uid string) []models.UserConfig {
	out, _ := db.GetUserConfigs(basecontext.NewBaseContext(), uid, "")
	return out
}
