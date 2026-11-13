package database

import (
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestSQLClaims_CRUD(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	created, err := g.CreateClaim(ctx, models.Claim{Name: "read users"})
	if err != nil {
		t.Fatalf("CreateClaim: %v", err)
	}
	if created.ID != "READ_USERS" { // space normalises to _
		t.Errorf("normalisation: got %q want READ_USERS", created.ID)
	}
	if _, err := g.CreateClaim(ctx, models.Claim{Name: "read users"}); err == nil ||
		!strings.Contains(err.Error(), "already exists with ID") {
		t.Errorf("dup create: want 400 already-exists, got %v", err)
	}
	if _, err := g.CreateClaim(ctx, models.Claim{}); err != interfaces.ErrClaimEmptyName {
		t.Errorf("want ErrClaimEmptyName identity, got %v", err)
	}

	got, err := g.GetClaim(ctx, "READ_users")
	if err != nil || got.ID != "READ_USERS" {
		t.Fatalf("GetClaim EqualFold: %v %v", got, err)
	}
	if _, err := g.GetClaim(ctx, "nope"); err != interfaces.ErrClaimNotFound {
		t.Errorf("want ErrClaimNotFound identity, got %v", err)
	}

	if err := g.UpdateClaimMetadata(ctx, "read_users", "reads them", "users", "user", "read"); err != nil {
		t.Fatal(err)
	}
	got, _ = g.GetClaim(ctx, "READ_USERS")
	if got.Description != "reads them" || got.Group != "users" || got.Resource != "user" || got.Action != "read" {
		t.Errorf("metadata update lost: %+v", got)
	}
	// metadata helper blanks what you hand it (legacy overwrites unconditionally)
	if err := g.UpdateClaimMetadata(ctx, "read_users", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	got, _ = g.GetClaim(ctx, "READ_USERS")
	if got.Description != "" || got.Group != "" {
		t.Errorf("metadata should have been cleared: %+v", got)
	}

	if _, err := g.UpdateClaim(ctx, &models.Claim{ID: "ghost", Name: "ghost"}); err != interfaces.ErrClaimNotFound {
		t.Errorf("want ErrClaimNotFound, got %v", err)
	}
	if err := g.DeleteClaim(ctx, ""); err != interfaces.ErrClaimEmptyNameOrId {
		t.Errorf("want ErrClaimEmptyNameOrId, got %v", err)
	}
}

func TestSQLClaims_InternalGuardAndDeleteCascade(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	if _, err := g.CreateClaim(ctx, models.Claim{Name: "system_admin", Internal: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateClaim(ctx, models.Claim{Name: "attach_me"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateRole(ctx, models.Role{Name: "auditor"}); err != nil {
		t.Fatal(err)
	}

	if err := g.DeleteClaim(ctx, "system_admin"); err != interfaces.ErrRemoveInternalClaim {
		t.Errorf("want ErrRemoveInternalClaim, got %v", err)
	}
	if _, err := g.UpdateClaim(ctx, &models.Claim{ID: "SYSTEM_ADMIN", Name: "SYSTEM_ADMIN"}); err != interfaces.ErrUpdateInternalClaim {
		t.Errorf("want ErrUpdateInternalClaim, got %v", err)
	}

	// attach to a role AND a user; usage scan blocks, then after user detach
	// the delete cascades through role_claims.
	if err := g.AddClaimToRole(ctx, "auditor", "attach_me"); err != nil {
		t.Fatal(err)
	}
	u := models.User{ID: "u-claim", Username: "uc", Email: "uc@x.local"}
	if err := g.gormDB.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	if err := g.gormDB.Model(&u).Association("Claims").Append([]models.Claim{{ID: "ATTACH_ME"}}); err != nil {
		t.Fatal(err)
	}
	if err := g.DeleteClaim(ctx, "attach_me"); err != interfaces.ErrClaimHasUsers {
		t.Errorf("want ErrClaimHasUsers, got %v", err)
	}
	if err := g.gormDB.Exec("DELETE FROM user_claims WHERE claim_id = 'ATTACH_ME'").Error; err != nil {
		t.Fatal(err)
	}
	if err := g.DeleteClaim(ctx, "attach_me"); err != nil {
		t.Fatalf("cascade delete: %v", err)
	}
	var n int64
	g.gormDB.Table("role_claims").Where("claim_id = ?", "ATTACH_ME").Count(&n)
	if n != 0 {
		t.Errorf("cascade left %d role_claims rows", n)
	}
	roles, _ := g.GetRoles(ctx, "")
	if len(roles) == 1 && len(roles[0].Claims) != 0 {
		t.Errorf("role still lists deleted claim: %+v", roles[0].Claims)
	}
}

func TestSQLClaims_RenameCascade(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	if _, err := g.CreateClaim(ctx, models.Claim{Name: "view_metrics"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateRole(ctx, models.Role{Name: "auditor"}); err != nil {
		t.Fatal(err)
	}
	u := models.User{ID: "u-rename", Username: "rn", Email: "rn@x.local"}
	if err := g.gormDB.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	if err := g.gormDB.Model(&u).Association("Claims").Append([]models.Claim{{ID: "VIEW_METRICS"}}); err != nil {
		t.Fatal(err)
	}
	if err := g.AddClaimToRole(ctx, "auditor", "view_metrics"); err != nil {
		t.Fatal(err)
	}

	updated, err := g.UpdateClaim(ctx, &models.Claim{ID: "VIEW_METRICS", Name: "view usage"})
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if updated.ID != "VIEW_USAGE" {
		t.Fatalf("rename side effect: got ID %q", updated.ID)
	}
	if _, err := g.GetClaim(ctx, "VIEW_METRICS"); err != interfaces.ErrClaimNotFound {
		t.Errorf("old id must be gone, got %v", err)
	}
	if _, err := g.GetClaim(ctx, "view_usage"); err != nil {
		t.Errorf("new id (EqualFold) must resolve, got %v", err)
	}

	// joins rewritten to the new ID (legacy replaced the embedded copies).
	var n int64
	g.gormDB.Table("role_claims").Where("claim_id = ?", "VIEW_USAGE").Count(&n)
	if n != 1 {
		t.Errorf("role_claims rewrite: %d rows", n)
	}
	g.gormDB.Table("user_claims").Where("claim_id = ?", "VIEW_USAGE").Count(&n)
	if n != 1 {
		t.Errorf("user_claims rewrite: %d rows", n)
	}
	g.gormDB.Table("claims").Where("id = ?", "VIEW_METRICS").Count(&n)
	if n != 0 {
		t.Errorf("stale claim row remains")
	}
}
