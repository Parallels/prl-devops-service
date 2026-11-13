package parity

// DomainSuite "users" (PBI #421 task 06): differential over the hub domain —
// cross-field lookups, default materialisation, join mutations, the
// rehydration pin, updates, root flows, deletion purges. Seed + checks use
// ONLY the exported DatabaseService surface.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/security/password"
)

// normaliseUsers canonicalises per-engine artefacts before differential
// compare: password hashes (salted — value is engine-private; verification is
// asserted separately), RFC3339Nano timestamps (wall-clock by nature), and
// the deep user-copies legacy embeds inside role/claim snapshots (json:"-",
// API-invisible). Content of every visible field still compares.
func normaliseUsers(in []models.User) []models.User {
	out := append([]models.User(nil), in...)
	for i := range out {
		if out[i].Password != "" {
			out[i].Password = "***"
		}
		if out[i].CreatedAt != "" {
			out[i].CreatedAt = "SET"
		}
		if out[i].UpdatedAt != "" {
			out[i].UpdatedAt = "SET"
		}
		out[i].Roles = normaliseNestedRoles(out[i].Roles)
		out[i].Claims = normaliseNestedClaims(out[i].Claims)
	}
	return out
}

// normaliseNestedRoles / normaliseNestedClaims canonicalise the user-embedded
// association entries: legacy STORES whatever the writing path handed it
// (caller stubs with empty Name/Description via CreateUser/UpdateUser; live
// deep copies via the Add* paths), while SQL projects the live row. The entry
// ID and the (rehydrated/preloaded alike) Claims list are the contract —
// scalar metadata inside user embeds is not comparable across engines.
func normaliseNestedRoles(in []models.Role) []models.Role {
	out := append([]models.Role(nil), in...)
	for i := range out {
		out[i].Name, out[i].Description, out[i].Internal = "", "", false
		out[i].Users = []models.User{}
		out[i].Claims = normaliseNestedClaims(out[i].Claims)
	}
	return out
}

func normaliseNestedClaims(in []models.Claim) []models.Claim {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.Claim(nil), in...)
	for i := range out {
		out[i].Name, out[i].Description = "", ""
		out[i].Internal, out[i].Group, out[i].Resource, out[i].Action = false, "", "", ""
		out[i].Users = []models.User{}
	}
	return out
}

func seedUsers(t testing.TB, db datainterfaces.DatabaseService) {
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
	for _, r := range []string{"auditor", "operator", "extra"} {
		if _, err := db.CreateRole(ctx, models.Role{Name: r}); err != nil {
			t.Fatalf("seed role %s: %v", r, err)
		}
	}
	for _, c := range []string{"extra claim", "late claim"} {
		if _, err := db.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed claim %s: %v", c, err)
		}
	}
	u := func(id, username, email, name string, extra ...func(*models.User)) {
		nu := models.User{ID: id, Username: username, Email: email, Name: name, Password: fixtureSecret("seed")}
		for _, e := range extra {
			e(&nu)
		}
		if _, err := db.CreateUser(ctx, nu); err != nil {
			t.Fatalf("seed user %s: %v", id, err)
		}
	}
	u("u1", "user1", "u1@x.io", "User One") // empty Roles/Claims ⇒ defaults
	u("u2", "user2", "u2@x.io", "User Two", func(x *models.User) {
		x.Roles = []models.Role{{ID: "OPERATOR"}}
		x.Claims = []models.Claim{{ID: "EXTRA_CLAIM"}}
	})
	u("u3", "user3", "u3@x.io", "User Three", func(x *models.User) {
		x.Blocked = true
		x.BlockedSince = "2026-01-01T00:00:00Z"
		x.BlockedReason = "manual hold"
		x.FailedLoginAttempts = 3
	})
}

func TestParityDomain_Users(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "users",
		Seed: seedUsers,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkUserLists,
			checkUserLookups,
			checkUserJoinMutations,
			checkUserRehydration,
			checkUserUpdates,
			checkRootFlows,
			checkUserDeletePurge,
		},
	})
}

func checkUserLists(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, f := range []string{"", "username=u", "email=u1@x.io", "blocked=true", "id=nope", "name=.*"} {
		lu, lerr := legacy.GetUsers(ctx, f)
		su, serr := sql.GetUsers(ctx, f)
		if lerr != nil || serr != nil {
			t.Fatalf("filter %q: %v / %v", f, lerr, serr)
		}
		AssertSameList(t, "GetUsers("+f+")", normaliseUsers(lu), normaliseUsers(su))
	}
}

func checkUserLookups(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, probe := range []string{"u2", "U2", "u2@x.io", "USER2", "u3"} {
		lu, lerr := legacy.GetUser(ctx, probe)
		su, serr := sql.GetUser(ctx, probe)
		assertSameError(t, "GetUser("+probe+")", lerr, serr)
		if lerr == nil && (lu == nil || su == nil || lu.ID != su.ID) {
			t.Errorf("GetUser(%s): %+v vs %+v", probe, lu, su)
		}
	}
	if _, err := legacy.GetUser(ctx, "nope"); err != datainterfaces.ErrUserNotFound {
		t.Errorf("legacy sentinel identity broken: %v", err)
	}
	if _, err := sql.GetUser(ctx, "nope"); err != datainterfaces.ErrUserNotFound {
		t.Errorf("sql sentinel identity broken: %v", err)
	}
}

func checkUserJoinMutations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	assertSameError(t, "dup default role",
		errOf(legacy.AddRoleToUser(ctx, "u1", "USER")), // default role already on u1
		errOf(sql.AddRoleToUser(ctx, "u1", "USER")))
	for _, eng := range []datainterfaces.DatabaseService{legacy, sql} {
		if err := eng.AddRoleToUser(ctx, "u1", "auditor"); err != nil {
			t.Fatal(err)
		}
		if err := eng.AddClaimToUser(ctx, "u2", "late_claim"); err != nil {
			t.Fatal(err)
		}
	}
	assertSameError(t, "dup add role",
		errOf(legacy.AddRoleToUser(ctx, "u1", "AUDITOR")),
		errOf(sql.AddRoleToUser(ctx, "u1", "AUDITOR")))
	// case-skew gate quirk: resolves via EqualFold, exact gate answers.
	assertSameError(t, "case-skew add",
		errOf(legacy.AddRoleToUser(ctx, "U1", "OPERATOR")),
		errOf(sql.AddRoleToUser(ctx, "U1", "OPERATOR")))
	assertSameError(t, "case-skew add claim",
		errOf(legacy.AddClaimToUser(ctx, "U1", "LATE_CLAIM")),
		errOf(sql.AddClaimToUser(ctx, "U1", "LATE_CLAIM")))
	assertSameError(t, "remove absent claim",
		errOf(legacy.RemoveClaimFromUser(ctx, "u2", "auditor")), // claim id that isn't on user
		errOf(sql.RemoveClaimFromUser(ctx, "u2", "auditor")))
	if err := legacy.RemoveRoleFromUser(ctx, "u1", "auditor"); err != nil {
		t.Fatal(err)
	}
	if err := sql.RemoveRoleFromUser(ctx, "u1", "auditor"); err != nil {
		t.Fatal(err)
	}
	assertSameError(t, "remove twice",
		errOf(legacy.RemoveRoleFromUser(ctx, "u1", "auditor")),
		errOf(sql.RemoveRoleFromUser(ctx, "u1", "auditor")))
	lu, _ := legacy.GetUsers(ctx, "")
	su, _ := sql.GetUsers(ctx, "")
	AssertSameList(t, "GetUsers after join mutations", normaliseUsers(lu), normaliseUsers(su))
}

func checkUserRehydration(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	// re-attach auditor to u1 and push a new claim onto the role: both
	// engines must show it inside Roles[].Claims without touching the user.
	for _, eng := range []datainterfaces.DatabaseService{legacy, sql} {
		if err := eng.AddRoleToUser(ctx, "u1", "auditor"); err != nil {
			t.Fatal(err)
		}
		if err := eng.AddClaimToRole(ctx, "auditor", "late_claim"); err != nil {
			t.Fatal(err)
		}
	}
	lu, _ := legacy.GetUsers(ctx, "")
	su, _ := sql.GetUsers(ctx, "")
	AssertSameList(t, "GetUsers after live role-claim push", normaliseUsers(lu), normaliseUsers(su))
}

func checkUserUpdates(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, eng := range []datainterfaces.DatabaseService{legacy, sql} {
		if err := eng.UpdateUser(ctx, models.User{ID: "u2", Name: "Renamed Two", Username: "hijack",
			Email: "two@x.io", Roles: []models.Role{{ID: "EXTRA"}}}); err != nil {
			t.Fatal(err)
		}
		if err := eng.UpdateUserBlockStatus(ctx, models.User{ID: "u3"}); err != nil { // clear
			t.Fatal(err)
		}
		if err := eng.UpdateUser(ctx, models.User{ID: "u2", Password: fixtureSecret("rotated-2")}); err != nil {
			t.Fatal(err)
		}
	}
	// the rotated hash must verify against the engine's own stored copy
	lu, _ := legacy.GetUser(ctx, "u2")
	su, _ := sql.GetUser(ctx, "u2")
	if err := password.Get().Compare(fixtureSecret("rotated-2"), "u2", lu.Password); err != nil {
		t.Errorf("legacy rotated hash fails verification: %v", err)
	}
	if err := password.Get().Compare(fixtureSecret("rotated-2"), "u2", su.Password); err != nil {
		t.Errorf("sql rotated hash fails verification: %v", err)
	}
	if lu.Username != "user2" || su.Username != "user2" {
		t.Errorf("username must not update: %q / %q", lu.Username, su.Username)
	}
	all1, _ := legacy.GetUsers(ctx, "")
	all2, _ := sql.GetUsers(ctx, "")
	AssertSameList(t, "GetUsers after updates", normaliseUsers(all1), normaliseUsers(all2))
}

func checkRootFlows(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, eng := range []datainterfaces.DatabaseService{legacy, sql} {
		if _, err := eng.CreateUser(ctx, models.User{ID: "the-root", Username: "rooty",
			Email: "root@localhost", Name: "Root Admin", Password: fixtureSecret("seed")}); err != nil {
			t.Fatal(err)
		}
	}
	for _, eng := range []datainterfaces.DatabaseService{legacy, sql} {
		if err := eng.UpdateRootPassword(ctx, "Rotated-Root-1!"); err != nil {
			t.Fatal(err)
		}
	}
	lr, _ := legacy.GetUser(ctx, "the-root")
	sr, _ := sql.GetUser(ctx, "the-root")
	if err := password.Get().Compare("Rotated-Root-1!", "the-root", lr.Password); err != nil {
		t.Errorf("legacy root hash not rotated: %v", err)
	}
	if err := password.Get().Compare("Rotated-Root-1!", "the-root", sr.Password); err != nil {
		t.Errorf("sql root hash not rotated: %v", err)
	}
	// UpdateUser against a root-shaped payload without RootAuthorization:
	assertSameError(t, "update root guard",
		errOf(legacy.UpdateUser(ctx, models.User{ID: "u1", Name: "root"})),
		errOf(sql.UpdateUser(ctx, models.User{ID: "u1", Name: "root"})))
	// deletion answers with the UPDATE sentinel on both (quirk pin):
	assertSameError(t, "delete root",
		errOf(legacy.DeleteUser(ctx, "the-root")),
		errOf(sql.DeleteUser(ctx, "the-root")))
}

func checkUserDeletePurge(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	assertSameError(t, "delete empty",
		errOf(legacy.DeleteUser(ctx, "")), // silent no-op — nil both
		errOf(sql.DeleteUser(ctx, "")))
	for _, eng := range []datainterfaces.DatabaseService{legacy, sql} {
		if err := eng.DeleteUser(ctx, "USER1"); err != nil { // EqualFold username
			t.Fatal(err)
		}
	}
	if _, err := legacy.GetUser(ctx, "u1"); err != datainterfaces.ErrUserNotFound {
		t.Errorf("legacy user survived: %v", err)
	}
	if _, err := sql.GetUser(ctx, "u1"); err != datainterfaces.ErrUserNotFound {
		t.Errorf("sql user survived: %v", err)
	}
	lu, _ := legacy.GetUsers(ctx, "")
	su, _ := sql.GetUsers(ctx, "")
	AssertSameList(t, "GetUsers after purge", normaliseUsers(lu), normaliseUsers(su))
}
