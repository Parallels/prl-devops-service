package parity

// DomainSuite "roles_claims" (PBI #421 task 05): first two-engine differential
// over real SQL. Seed uses ONLY the exported DatabaseService surface; every
// check asserts legacy-vs-SQL observable equivalence (lists, order, sentinel
// identity, mutation cascades, renames).

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	repoerrors "github.com/Parallels/prl-devops-service/errors"
)

func seedRolesClaims(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	claims := []models.Claim{
		{Name: "read_users"},
		{Name: "write_users"},
		{Name: "system_admin", Internal: true},
		{Name: "view_metrics"},
		{Name: "audit_log"},
		{Name: "Space Name"}, // pins the ToUpper(NormalizeString(...)) = SPACE_NAME path
	}
	for _, c := range claims {
		if _, err := db.CreateClaim(ctx, c); err != nil {
			t.Fatalf("seed claim %s: %v", c.Name, err)
		}
	}
	roles := []models.Role{
		{Name: "auditor", Claims: []models.Claim{{ID: "READ_USERS"}, {ID: "AUDIT_LOG"}}},
		{Name: "operator", Claims: []models.Claim{{ID: "WRITE_USERS"}}},
		{Name: "internal_role", Internal: true},
		{Name: "space role", Description: "space case"},
	}
	for _, r := range roles {
		if _, err := db.CreateRole(ctx, r); err != nil {
			t.Fatalf("seed role %s: %v", r.Name, err)
		}
	}
}

func TestParityDomain_RolesClaims(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "roles_claims",
		Seed: seedRolesClaims,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkRoleLists,
			checkClaimLists,
			checkLookupAndSentinels,
			checkClaimAttachDetach,
			checkDeleteClaimCascade,
			checkRenameCascades,
			checkRoleUpdates,
			checkRoleDeletes,
		},
	})
}

func checkRoleLists(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	for _, f := range []string{"", "name=.*", "internal=true", "id=nonexistent"} {
		lr, lerr := legacy.GetRoles(ctx, f)
		sr, serr := sql.GetRoles(ctx, f)
		if lerr != nil || serr != nil {
			t.Fatalf("filter %q: errors %v / %v", f, lerr, serr)
		}
		AssertSameList(t, "GetRoles("+f+")", normaliseRoles(lr), normaliseRoles(sr))
	}
}

func checkClaimLists(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	for _, f := range []string{"", "name=.*", "internal=true", "id=nonexistent"} {
		lc, lerr := legacy.GetClaims(ctx, f)
		sc, serr := sql.GetClaims(ctx, f)
		if lerr != nil || serr != nil {
			t.Fatalf("filter %q: errors %v / %v", f, lerr, serr)
		}
		AssertSameList(t, "GetClaims("+f+")", normaliseClaims(lc), normaliseClaims(sc))
	}
}

func assertSameError(t testing.TB, what string, want, got error) {
	t.Helper()
	if want == nil && got == nil {
		return
	}
	if want == nil || got == nil {
		t.Errorf("%s: error mismatch: legacy=%v sql=%v", what, want, got)
		return
	}
	if want != got { // sentinel identity (==) is the preferred contract…
		w, wok := want.(*repoerrors.SystemError)
		g, gok := got.(*repoerrors.SystemError)
		if wok && gok && w.Code() == g.Code() && want.Error() == got.Error() {
			return
		}
		// Message-pinned plain errors (e.g. legacy's stdlib "ID already
		// exists" in SetId, task 08): neither side is a SystemError — the
		// contract is the exact message.
		if !wok && !gok && want.Error() == got.Error() {
			return
		}
		t.Errorf("%s: error mismatch: legacy=%q sql=%q", what, want.Error(), got.Error())
	}
}

func checkLookupAndSentinels(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	for _, key := range []string{"auditor", "AUDITOR", "space role", "nope"} {
		lr, lerr := legacy.GetRole(ctx, key)
		sr, serr := sql.GetRole(ctx, key)
		assertSameError(t, "GetRole("+key+")", lerr, serr)
		if lerr == nil && (lr == nil || sr == nil || lr.ID != sr.ID) {
			t.Errorf("GetRole(%s): %v vs %v", key, lr, sr)
		}
	}
	for _, key := range []string{"read_users", "READ_USERS", "Space Name", "nope"} {
		lc, lerr := legacy.GetClaim(ctx, key)
		sc, serr := sql.GetClaim(ctx, key)
		assertSameError(t, "GetClaim("+key+")", lerr, serr)
		if lerr == nil && (lc == nil || sc == nil || lc.ID != sc.ID) {
			t.Errorf("GetClaim(%s): %v vs %v", key, lc, sc)
		}
	}
	// sentinel identity, not just message parity:
	if _, err := legacy.GetRole(ctx, "nope"); err != datainterfaces.ErrRoleNotFound {
		t.Errorf("legacy GetRole sentinel identity broken: %v", err)
	}
	if _, err := sql.GetRole(ctx, "nope"); err != datainterfaces.ErrRoleNotFound {
		t.Errorf("sql GetRole sentinel identity broken: %v", err)
	}
}

func checkClaimAttachDetach(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	// ORDER QUIRK: garbage role + missing claim → ErrClaimNotFound on BOTH.
	assertSameError(t, "add-order-quirk",
		errOf(legacy.AddClaimToRole(ctx, "ghost", "ghost-claim")),
		errOf(sql.AddClaimToRole(ctx, "ghost", "ghost-claim")))
	// internal guard (non-root):
	assertSameError(t, "internal guard",
		errOf(legacy.AddClaimToRole(ctx, "internal_role", "view_metrics")),
		errOf(sql.AddClaimToRole(ctx, "internal_role", "view_metrics")))
	if err := legacy.AddClaimToRole(ctx, "operator", "read_users"); err != nil {
		t.Fatal(err)
	}
	if err := sql.AddClaimToRole(ctx, "operator", "read_users"); err != nil {
		t.Fatal(err)
	}
	assertSameError(t, "dup add",
		errOf(legacy.AddClaimToRole(ctx, "operator", "READ_USERS")),
		errOf(sql.AddClaimToRole(ctx, "operator", "READ_USERS")))

	// remove by NAME (EqualFold both sides)
	if err := legacy.RemoveClaimFromRole(ctx, "operator", "READ_USERS"); err != nil {
		t.Fatal(err)
	}
	if err := sql.RemoveClaimFromRole(ctx, "operator", "READ_USERS"); err != nil {
		t.Fatal(err)
	}
	assertSameError(t, "remove absent",
		errOf(legacy.RemoveClaimFromRole(ctx, "operator", "read_users")),
		errOf(sql.RemoveClaimFromRole(ctx, "operator", "read_users")))

	lr, _ := legacy.GetRoles(ctx, "")
	sr, _ := sql.GetRoles(ctx, "")
	AssertSameList(t, "GetRoles after attach/detach", normaliseRoles(lr), normaliseRoles(sr))
}

func checkDeleteClaimCascade(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	assertSameError(t, "delete internal",
		errOf(legacy.DeleteClaim(ctx, "system_admin")),
		errOf(sql.DeleteClaim(ctx, "system_admin")))
	// view_metrics is attached to nothing; audit_log is on the auditor role
	// from seed — deleting it must cascade identically.
	if err := legacy.DeleteClaim(ctx, "audit_log"); err != nil {
		t.Fatal(err)
	}
	if err := sql.DeleteClaim(ctx, "audit_log"); err != nil {
		t.Fatal(err)
	}
	assertSameError(t, "delete gone",
		errOf(legacy.DeleteClaim(ctx, "audit_log")),
		errOf(sql.DeleteClaim(ctx, "audit_log")))
	lr, _ := legacy.GetRoles(ctx, "")
	sr, _ := sql.GetRoles(ctx, "")
	AssertSameList(t, "GetRoles after claim cascade delete", normaliseRoles(lr), normaliseRoles(sr))
	lc, _ := legacy.GetClaims(ctx, "")
	sc, _ := sql.GetClaims(ctx, "")
	AssertSameList(t, "GetClaims after claim cascade delete", normaliseClaims(lc), normaliseClaims(sc))
}

func checkRenameCascades(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	// same-id update (position and content stable):
	_, _ = legacy.UpdateClaim(ctx, &models.Claim{ID: "SPACE_NAME", Name: "Space Name", Description: "same id"})
	_, _ = sql.UpdateClaim(ctx, &models.Claim{ID: "SPACE_NAME", Name: "Space Name", Description: "same id"})
	// true rename: READ_USERS -> READ FILES (auditor holds it; position kept):
	_, errL := legacy.UpdateClaim(ctx, &models.Claim{ID: "READ_USERS", Name: "read files"})
	_, errS := sql.UpdateClaim(ctx, &models.Claim{ID: "READ_USERS", Name: "read files"})
	assertSameError(t, "rename", errL, errS)
	lc, _ := legacy.GetClaims(ctx, "")
	sc, _ := sql.GetClaims(ctx, "")
	AssertSameList(t, "GetClaims after rename (position pinned)", normaliseClaims(lc), normaliseClaims(sc))
	lr, _ := legacy.GetRoles(ctx, "")
	sr, _ := sql.GetRoles(ctx, "")
	AssertSameList(t, "GetRoles after rename (join rewrite)", normaliseRoles(lr), normaliseRoles(sr))
}

func checkRoleUpdates(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	_ = legacy.UpdateRoleDescription(ctx, "auditor", "reads all")
	_ = sql.UpdateRoleDescription(ctx, "auditor", "reads all")
	_ = legacy.UpdateRoleInternalFlag(ctx, "SPACE_ROLE", true)
	_ = sql.UpdateRoleInternalFlag(ctx, "SPACE_ROLE", true)
	// internal guard on UpdateRole (non-root):
	_, eL := legacy.UpdateRole(ctx, &models.Role{ID: "SPACE_ROLE", Name: "SPACE_ROLE"})
	_, eS := sql.UpdateRole(ctx, &models.Role{ID: "SPACE_ROLE", Name: "SPACE_ROLE"})
	assertSameError(t, "updaterole guard", eL, eS)
	_ = legacy.UpdateRoleInternalFlag(ctx, "SPACE_ROLE", false)
	_ = sql.UpdateRoleInternalFlag(ctx, "SPACE_ROLE", false)
	_, _ = legacy.UpdateRole(ctx, &models.Role{ID: "OPERATOR", Name: "OPERATOR", Description: "v2"})
	_, _ = sql.UpdateRole(ctx, &models.Role{ID: "OPERATOR", Name: "OPERATOR", Description: "v2"})
	lr, _ := legacy.GetRoles(ctx, "")
	sr, _ := sql.GetRoles(ctx, "")
	AssertSameList(t, "GetRoles after role updates", normaliseRoles(lr), normaliseRoles(sr))
}

func checkRoleDeletes(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	assertSameError(t, "delete internal role",
		errOf(legacy.DeleteRole(ctx, "internal_role")),
		errOf(sql.DeleteRole(ctx, "internal_role")))
	assertSameError(t, "delete empty",
		errOf(legacy.DeleteRole(ctx, "")),
		errOf(sql.DeleteRole(ctx, "")))
	if err := legacy.DeleteRole(ctx, "space_role"); err != nil {
		t.Fatal(err)
	}
	if err := sql.DeleteRole(ctx, "space_role"); err != nil {
		t.Fatal(err)
	}
	lr, _ := legacy.GetRoles(ctx, "")
	sr, _ := sql.GetRoles(ctx, "")
	AssertSameList(t, "GetRoles after role deletes", normaliseRoles(lr), normaliseRoles(sr))
}

func errOf(err error) error { return err }
