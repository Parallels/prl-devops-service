package database

import (
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/security/password"
)

// seedDefaults materialises the constants.DefaultRoles/DefaultClaims that
// CreateUser falls back to (legacy seeds do this; the engine does not).
func seedDefaults(t *testing.T, g *GormDatabase, ctx basecontext.ApiContext) {
	t.Helper()
	for _, r := range constants.DefaultRoles {
		if _, err := g.CreateRole(ctx, models.Role{Name: r}); err != nil {
			t.Fatalf("seed default role %s: %v", r, err)
		}
	}
	for _, c := range constants.DefaultClaims {
		if _, err := g.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed default claim %s: %v", c, err)
		}
	}
}

func sqlUser(id, username, email, name string) models.User {
	return models.User{ID: id, Username: username, Email: email, Name: name, Password: "S3cret-pass!"}
}

func TestSQLUsers_CRUD(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedDefaults(t, g, ctx)

	created, err := g.CreateUser(ctx, sqlUser("u1", "user1", "u1@x.io", "User One"))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Errorf("timestamps not stamped: %+v", created)
	}
	if created.Password == "S3cret-pass!" {
		t.Error("password stored unhashed")
	}
	if err := password.Get().Compare("S3cret-pass!", "u1", created.Password); err != nil {
		t.Errorf("hash does not verify with (pw, ID) salt: %v", err)
	}

	// defaults materialised (empty Roles/Claims ⇒ constants lists)
	if len(created.Roles) != len(constants.DefaultRoles) || len(created.Claims) != len(constants.DefaultClaims) {
		t.Errorf("defaults not materialised: %d roles %d claims", len(created.Roles), len(created.Claims))
	}

	// generated id path
	gen, err := g.CreateUser(ctx, sqlUser("", "user2", "u2@x.io", "User Two"))
	if err != nil || gen.ID == "" {
		t.Fatalf("generated id: %v %v", gen, err)
	}

	// empty-field guards (identity)
	if _, err := g.CreateUser(ctx, models.User{ID: "x1", Username: "x", Name: "x"}); err != interfaces.ErrUserEmailCannotBeEmpty {
		t.Errorf("want ErrUserEmailCannotBeEmpty, got %v", err)
	}
	if _, err := g.CreateUser(ctx, models.User{ID: "x2", Email: "x@x.io", Name: "x"}); err != interfaces.ErrUserUsernameCannotBeEmpty {
		t.Errorf("want ErrUserUsernameCannotBeEmpty, got %v", err)
	}
	if _, err := g.CreateUser(ctx, models.User{ID: "x3", Email: "y@x.io", Username: "y"}); err != interfaces.ErrUserNameCannotBeEmpty {
		t.Errorf("want ErrUserNameCannotBeEmpty, got %v", err)
	}

	// duplicate probes ×3 — collision across ANY of the three fields
	for _, dupe := range []models.User{
		sqlUser("u1", "other", "other@x.io", "Other"),
		sqlUser("n1", "user1", "other@x.io", "Other"), // username collision
		sqlUser("n2", "other2", "U1@X.IO", "Other"),   // email collision, case-skewed
	} {
		if _, err := g.CreateUser(ctx, dupe); err != interfaces.ErrUserAlreadyExists {
			t.Errorf("dup id=%s: want ErrUserAlreadyExists identity, got %v", dupe.ID, err)
		}
	}

	// cross-field EqualFold lookup
	for _, probe := range []string{"u1", "U1", "u1@x.io", "USER1"} {
		u, err := g.GetUser(ctx, probe)
		if err != nil || u.ID != "u1" {
			t.Errorf("GetUser(%q): %v %+v", probe, err, u)
		}
	}
	if _, err := g.GetUser(ctx, "nobody"); err != interfaces.ErrUserNotFound {
		t.Errorf("want ErrUserNotFound identity, got %v", err)
	}

	// unknown role/claim on explicit lists ⇒ constructed 400 (message+code)
	_, err = g.CreateUser(ctx, models.User{ID: "u9", Username: "u9", Email: "u9@x.io", Name: "Nine",
		Password: "pw12345678", Roles: []models.Role{{ID: "ghost-role", Name: "Ghostly"}}})
	if err == nil || err == interfaces.ErrRoleNotFound ||
		!strings.Contains(err.Error(), "role Ghostly does not exist") {
		t.Errorf("want 400 role Ghostly does not exist, got %v", err)
	}
	_, err = g.CreateUser(ctx, models.User{ID: "u9", Username: "u9", Email: "u9@x.io", Name: "Nine",
		Password: "pw12345678", Claims: []models.Claim{{ID: "ghost-claim", Name: "Ghosty"}}})
	if err == nil || !strings.Contains(err.Error(), "claim Ghosty does not exist") {
		t.Errorf("want 400 claim Ghosty does not exist, got %v", err)
	}
}

func TestSQLUsers_RoleClaimAssignment(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedDefaults(t, g, ctx)
	if _, err := g.CreateRole(ctx, models.Role{Name: "extra role"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateClaim(ctx, models.Claim{Name: "extra claim"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateUser(ctx, sqlUser("u1", "user1", "u1@x.io", "User One")); err != nil {
		t.Fatal(err)
	}

	// empty-id guards
	if err := g.AddRoleToUser(ctx, "", "EXTRA_ROLE"); err != interfaces.ErrUserIDCannotBeEmpty {
		t.Errorf("want ErrUserIDCannotBeEmpty, got %v", err)
	}
	if err := g.AddRoleToUser(ctx, "u1", ""); err != interfaces.ErrRoleIDCannotBeEmpty {
		t.Errorf("want ErrRoleIDCannotBeEmpty, got %v", err)
	}
	if err := g.AddClaimToUser(ctx, "", "EXTRA_CLAIM"); err != interfaces.ErrUserIDCannotBeEmpty {
		t.Errorf("want ErrUserIDCannotBeEmpty, got %v", err)
	}
	if err := g.AddClaimToUser(ctx, "u1", ""); err != interfaces.ErrClaimIDCannotBeEmpty {
		t.Errorf("want ErrClaimIDCannotBeEmpty, got %v", err)
	}

	if err := g.AddRoleToUser(ctx, "u1", "extra_role"); err != nil {
		t.Fatalf("AddRoleToUser: %v", err)
	}
	if err := g.AddRoleToUser(ctx, "u1", "EXTRA_ROLE"); err != interfaces.ErrUserAlreadyContainsRole {
		t.Errorf("dup role: want ErrUserAlreadyContainsRole, got %v", err)
	}
	var n int64
	g.gormDB.Table("user_roles").Where("user_id = ? AND role_id = ?", "u1", "EXTRA_ROLE").Count(&n)
	if n != 1 {
		t.Errorf("want exactly one join row, got %d", n)
	}
	if err := g.AddRoleToUser(ctx, "u1", "nope"); err != interfaces.ErrRoleNotFound {
		t.Errorf("want ErrRoleNotFound, got %v", err)
	}
	if err := g.AddRoleToUser(ctx, "ghost", "extra_role"); err != interfaces.ErrUserNotFound {
		t.Errorf("want ErrUserNotFound, got %v", err)
	}

	if err := g.AddClaimToUser(ctx, "u1", "extra_claim"); err != nil {
		t.Fatalf("AddClaimToUser: %v", err)
	}
	if err := g.AddClaimToUser(ctx, "u1", "extra_claim"); err != interfaces.ErrUserAlreadyContainsClaim {
		t.Errorf("dup claim: want ErrUserAlreadyContainsClaim, got %v", err)
	}

	// remove round-trips
	if err := g.RemoveRoleFromUser(ctx, "u1", "EXTRA_ROLE"); err != nil {
		t.Fatalf("RemoveRoleFromUser: %v", err)
	}
	if err := g.RemoveRoleFromUser(ctx, "u1", "EXTRA_ROLE"); err != interfaces.ErrRoleNotFound {
		t.Errorf("remove twice: want ErrRoleNotFound, got %v", err)
	}
	if err := g.RemoveClaimFromUser(ctx, "u1", "extra_claim"); err != nil {
		t.Fatal(err)
	}
	if err := g.RemoveClaimFromUser(ctx, "u1", "extra_claim"); err != interfaces.ErrClaimNotFound {
		t.Errorf("remove twice: want ErrClaimNotFound, got %v", err)
	}

	// CASE-SKEW QUIRK (users.go:298-306/376-384): EqualFold resolve succeeds,
	// the exact-id append gate then answers per-domain fall-through errors.
	if _, err := g.CreateUser(ctx, sqlUser("Mixed2", "mixed2", "m2@x.io", "Mixed Two")); err != nil {
		t.Fatal(err)
	}
	if err := g.AddRoleToUser(ctx, "mixed2", "EXTRA_ROLE"); err != interfaces.ErrUserNotFound {
		t.Errorf("case-skew add-role: want ErrUserNotFound fall-through, got %v", err)
	}
	if err := g.AddClaimToUser(ctx, "mixed2", "EXTRA_CLAIM"); err != interfaces.ErrClaimNotFound {
		t.Errorf("case-skew add-claim: want ErrClaimNotFound fall-through, got %v", err)
	}
	if err := g.RemoveRoleFromUser(ctx, "mixed2", "EXTRA_ROLE"); err != interfaces.ErrRoleNotFound {
		t.Errorf("case-skew remove-role: want ErrRoleNotFound, got %v", err)
	}
	// exact casing works fine
	if err := g.AddRoleToUser(ctx, "Mixed2", "EXTRA_ROLE"); err != nil {
		t.Errorf("exact-case add: %v", err)
	}
}

func TestSQLUsers_Updates(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedDefaults(t, g, ctx)
	if _, err := g.CreateRole(ctx, models.Role{Name: "extra role"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateUser(ctx, sqlUser("u1", "user1", "u1@x.io", "User One")); err != nil {
		t.Fatal(err)
	}

	// non-empty-field assignment; Username ignored (legacy :171-192); exact-ID scan.
	err := g.UpdateUser(ctx, models.User{ID: "u1", Name: "Renamed", Username: "hijack", Email: "new@x.io",
		Roles: []models.Role{{ID: "EXTRA_ROLE"}}})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	u, _ := g.GetUser(ctx, "u1")
	if u.Name != "Renamed" || u.Email != "new@x.io" {
		t.Errorf("fields not applied: %+v", u)
	}
	if u.Username != "user1" {
		t.Errorf("username must NOT update: %q", u.Username)
	}
	if len(u.Roles) != 1 || u.Roles[0].ID != "EXTRA_ROLE" {
		t.Errorf("roles join not replaced: %+v", u.Roles)
	}
	if err := g.UpdateUser(ctx, models.User{ID: "U1", Name: "CaseSkip"}); err != interfaces.ErrUserNotFound {
		t.Errorf("exact-ID scan: want ErrUserNotFound for U1, got %v", err)
	}

	// password rotation verifies with (pw, key.ID) salt
	if err := g.UpdateUser(ctx, models.User{ID: "u1", Password: "Rotated-1!"}); err != nil {
		t.Fatal(err)
	}
	u, _ = g.GetUser(ctx, "u1")
	if err := password.Get().Compare("Rotated-1!", "u1", u.Password); err != nil {
		t.Errorf("rotated hash does not verify: %v", err)
	}

	// block status — zero values must clear
	if err := g.UpdateUserBlockStatus(ctx, models.User{ID: "u1", Blocked: true, BlockedSince: "2026-01-01T00:00:00Z",
		BlockedReason: "abuse", FailedLoginAttempts: 7}); err != nil {
		t.Fatal(err)
	}
	u, _ = g.GetUser(ctx, "u1")
	if !u.Blocked || u.BlockedReason != "abuse" || u.FailedLoginAttempts != 7 {
		t.Errorf("block fields not written: %+v", u)
	}
	if err := g.UpdateUserBlockStatus(ctx, models.User{ID: "u1"}); err != nil {
		t.Fatal(err)
	}
	u, _ = g.GetUser(ctx, "u1")
	if u.Blocked || u.BlockedReason != "" || u.FailedLoginAttempts != 0 {
		t.Errorf("block fields must clear to zero: %+v", u)
	}
	if err := g.UpdateUserBlockStatus(ctx, models.User{ID: "ghost"}); err != interfaces.ErrUserNotFound {
		t.Errorf("want ErrUserNotFound, got %v", err)
	}

	// root-shaped UpdateUser without RootAuthorization
	if err := g.UpdateUser(ctx, models.User{ID: "u1", Name: "root"}); err != interfaces.ErrCannotUpdateRootUser {
		t.Errorf("want ErrCannotUpdateRootUser, got %v", err)
	}

	// UpdateRootPassword targets the root@localhost row (exact email)
	if err := g.UpdateRootPassword(ctx, "nobody-here-1!"); err != interfaces.ErrUserNotFound {
		t.Errorf("no root user yet: want ErrUserNotFound, got %v", err)
	}
	root, err := g.CreateUser(ctx, sqlUser("the-root", "rooty", "root@localhost", "Root Admin"))
	if err != nil {
		t.Fatal(err)
	}
	if err := g.UpdateRootPassword(ctx, "Rotated-Root-1!"); err != nil {
		t.Fatalf("UpdateRootPassword: %v", err)
	}
	root, _ = g.GetUser(ctx, "the-root")
	if err := password.Get().Compare("Rotated-Root-1!", "the-root", root.Password); err != nil {
		t.Errorf("root hash not rotated with user.ID salt: %v", err)
	}
}

func TestSQLUsers_DeleteAndConnection(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedDefaults(t, g, ctx)
	if _, err := g.CreateRole(ctx, models.Role{Name: "extra role"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateClaim(ctx, models.Claim{Name: "extra claim"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateUser(ctx, sqlUser("u1", "user1", "u1@x.io", "User One")); err != nil {
		t.Fatal(err)
	}
	if err := g.AddRoleToUser(ctx, "u1", "EXTRA_ROLE"); err != nil {
		t.Fatal(err)
	}
	if err := g.AddClaimToUser(ctx, "u1", "EXTRA_CLAIM"); err != nil {
		t.Fatal(err)
	}

	if err := g.DeleteUser(ctx, ""); err != nil {
		t.Errorf("empty id must be silent no-op, got %v", err)
	}

	if err := g.DeleteUser(ctx, "USER1"); err != nil { // EqualFold by username
		t.Fatalf("DeleteUser: %v", err)
	}
	var n int64
	g.gormDB.Table("user_roles").Count(&n)
	if n != 0 {
		t.Errorf("orphan user_roles rows: %d", n)
	}
	g.gormDB.Table("user_claims").Count(&n)
	if n != 0 {
		t.Errorf("orphan user_claims rows: %d", n)
	}
	if _, err := g.GetUser(ctx, "u1"); err != interfaces.ErrUserNotFound {
		t.Errorf("want ErrUserNotFound after delete, got %v", err)
	}

	// root-shaped user refuses deletion with the UPDATE sentinel (quirk)
	if _, err := g.CreateUser(ctx, sqlUser("the-root", "rooty", "root@localhost", "Root Admin")); err != nil {
		t.Fatal(err)
	}
	if err := g.DeleteUser(ctx, "the-root"); err != interfaces.ErrCannotUpdateRootUser {
		t.Errorf("want ErrCannotUpdateRootUser quirk, got %v", err)
	}

	// not-connected guard (engine without SQL handle)
	cold := NewGormDatabase(ctx, nil, nil)
	if _, err := cold.GetUsers(ctx, ""); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("want ErrDatabaseNotConnected identity, got %v", err)
	}
	if err := cold.DeleteUser(ctx, "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("DeleteUser guard: %v", err)
	}
}

// TestSQLUsers_LiveRoleClaims pins the rehydration equivalence: a claim added
// to a role AFTER user creation shows up inside Roles[].Claims on read.
func TestSQLUsers_LiveRoleClaims(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedDefaults(t, g, ctx)
	if _, err := g.CreateRole(ctx, models.Role{Name: "auditor"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateClaim(ctx, models.Claim{Name: "late_claim"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateUser(ctx, models.User{ID: "u1", Username: "user1", Email: "u1@x.io", Name: "One",
		Password: "pw12345678", Roles: []models.Role{{ID: "AUDITOR"}}}); err != nil {
		t.Fatal(err)
	}
	u, _ := g.GetUser(ctx, "u1")
	if len(u.Roles[0].Claims) != 0 {
		t.Fatalf("precondition: auditor has claims already: %+v", u.Roles[0].Claims)
	}
	if err := g.AddClaimToRole(ctx, "auditor", "late_claim"); err != nil {
		t.Fatal(err)
	}
	u, _ = g.GetUser(ctx, "u1")
	if len(u.Roles[0].Claims) != 1 || u.Roles[0].Claims[0].ID != "LATE_CLAIM" {
		t.Errorf("live claims not visible: %+v", u.Roles[0].Claims)
	}
}
