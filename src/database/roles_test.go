package database

import (
	"errors"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"gorm.io/gorm"
)

// allGormModels mirrors data/models test registry (the production AutoMigrate
// list ships with the task-18 boot wiring; duplicated here until then).
func allGormModels() []interface{} { return AllGormModels() }

// newSQLDB returns a migrated :memory: GormDatabase with NO legacy delegate:
// every roles/claims code path must be fully SQL (AC-6, zero reach-through).
func newSQLDB(t *testing.T) *GormDatabase {
	t.Helper()
	conn, err := NewDatabaseConnection(basecontext.NewBaseContext(), DatabaseConfig{Type: "sqlite", Path: MemoryPath})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := conn.AutoMigrate(allGormModels()...); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, e := conn.DB(); e == nil {
			_ = sqlDB.Close()
		}
	})
	return NewGormDatabase(basecontext.NewBaseContext(), conn)
}

func TestSQLRoles_CRUD(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	created, err := g.CreateRole(ctx, models.Role{Name: "auditor"})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if created.ID != "AUDITOR" { // ID = ToUpper(NormalizeString(name))
		t.Errorf("ID normalisation: got %q want AUDITOR", created.ID)
	}

	if _, err := g.CreateRole(ctx, models.Role{Name: "auditor"}); err == nil ||
		!strings.Contains(err.Error(), "already exists with ID") {
		t.Errorf("duplicate create: want 400 already-exists, got %v", err)
	}
	if _, err := g.CreateRole(ctx, models.Role{}); err != interfaces.ErrRoleEmptyName {
		t.Errorf("empty name: want ErrRoleEmptyName identity, got %v", err)
	}

	got, err := g.GetRole(ctx, "AUDITOR") // EqualFold by name
	if err != nil || got.ID != "AUDITOR" {
		t.Fatalf("GetRole by name: %v %v", got, err)
	}
	if _, err := g.GetRole(ctx, "nope"); err != interfaces.ErrRoleNotFound {
		t.Errorf("missing role: want ErrRoleNotFound identity, got %v", err)
	}

	if err := g.UpdateRoleDescription(ctx, "Auditor", "reads everything"); err != nil {
		t.Fatal(err)
	}
	if err := g.UpdateRoleInternalFlag(ctx, "auditor", true); err != nil {
		t.Fatal(err)
	}
	got, _ = g.GetRole(ctx, "auditor")
	if got.Description != "reads everything" || !got.Internal {
		t.Errorf("targeted updates lost: %+v", got)
	}

	// internal guard fires for non-root (UpdateRole path).
	if _, err := g.UpdateRole(ctx, &models.Role{ID: "auditor", Name: "auditor"}); err != interfaces.ErrUpdateInternalRole {
		t.Errorf("internal guard: want ErrUpdateInternalRole, got %v", err)
	}

	// uninternal it again, then full-object replacement incl. ID change.
	_ = g.UpdateRoleInternalFlag(ctx, "auditor", false)
	repl, err := g.UpdateRole(ctx, &models.Role{ID: "auditor", Name: "auditor", Description: "v2"})
	if err != nil || repl.Description != "v2" {
		t.Fatalf("UpdateRole same-id: %v %+v", err, repl)
	}
	if _, err := g.UpdateRole(ctx, &models.Role{ID: "ghost", Name: "ghost"}); err != interfaces.ErrRoleNotFound {
		t.Errorf("UpdateRole missing: want ErrRoleNotFound, got %v", err)
	}

	if err := g.DeleteRole(ctx, "AUDITOR"); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
	roles, _ := g.GetRoles(ctx, "")
	if len(roles) != 0 {
		t.Errorf("role survived delete: %+v", roles)
	}
	if err := g.DeleteRole(ctx, ""); err != interfaces.ErrRoleEmptyNameOrId {
		t.Errorf("empty delete guard: want ErrRoleEmptyNameOrId, got %v", err)
	}
}

func TestSQLRoles_GuardQuirks(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	if _, err := g.CreateClaim(ctx, models.Claim{Name: "read_users"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateRole(ctx, models.Role{Name: "auditor"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateRole(ctx, models.Role{Name: "internal_role", Internal: true}); err != nil {
		t.Fatal(err)
	}

	// ORDER QUIRK: garbage role + missing claim must report ErrClaimNotFound
	// (claim resolves first, legacy roles.go:132-135).
	if err := g.AddClaimToRole(ctx, "does-not-exist", "no-such-claim"); err != interfaces.ErrClaimNotFound {
		t.Errorf("check-order quirk: want ErrClaimNotFound, got %v", err)
	}

	if err := g.AddClaimToRole(ctx, "auditor", "read_users"); err != nil {
		t.Fatal(err)
	}
	if err := g.AddClaimToRole(ctx, "auditor", "read_users"); err != interfaces.ErrRoleAlreadyContainsClaim {
		t.Errorf("dup add: want ErrRoleAlreadyContainsClaim, got %v", err)
	}
	if err := g.AddClaimToRole(ctx, "internal_role", "read_users"); err != interfaces.ErrUpdateInternalRole {
		t.Errorf("internal guard: want ErrUpdateInternalRole, got %v", err)
	}

	// Remove by claim NAME (EqualFold both sides of roles.go:166).
	if err := g.RemoveClaimFromRole(ctx, "auditor", "READ_USERS"); err != nil {
		t.Fatalf("remove by name: %v", err)
	}
	if err := g.RemoveClaimFromRole(ctx, "auditor", "read_users"); err != interfaces.ErrRoleDoesNotContainClaim {
		t.Errorf("want ErrRoleDoesNotContainClaim, got %v", err)
	}

	// DeleteRole usage scan: attach a user through the join directly.
	_ = g.AddClaimToRole(ctx, "auditor", "read_users")
	u := models.User{ID: "u-rolecheck", Username: "rc", Email: "rc@x.local"}
	if err := g.gormDB.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	if err := g.gormDB.Model(&u).Association("Roles").Append([]models.Role{{ID: "AUDITOR"}}); err != nil {
		t.Fatal(err)
	}
	if err := g.DeleteRole(ctx, "auditor"); err != interfaces.ErrRoleHasUsers {
		t.Errorf("want ErrRoleHasUsers, got %v", err)
	}
	if err := g.DeleteRole(ctx, "internal_role"); err != interfaces.ErrRemoveInternalRole {
		t.Errorf("want ErrRemoveInternalRole, got %v", err)
	}

	// role_claims rows are cleaned by DeleteRole.
	if err := g.DeleteRole(ctx, "auditor"); err != interfaces.ErrRoleHasUsers {
		t.Fatalf("expected still blocked by user, got %v", err)
	}
	var n int64
	g.gormDB.Table("user_roles").Where("role_id = ?", "AUDITOR").Count(&n)
	_ = g.gormDB.Exec("DELETE FROM user_roles WHERE role_id = ?", "AUDITOR")
	if err := g.DeleteRole(ctx, "auditor"); err != nil {
		t.Fatal(err)
	}
	g.gormDB.Table("role_claims").Where("role_id = ?", "AUDITOR").Count(&n)
	if n != 0 {
		t.Errorf("DeleteRole left %d role_claims rows", n)
	}
}

// TestSQLRoles_TxRollback injects a failure at the role_claims INSERT inside
// AddClaimToRole and asserts the transaction rolled back (D-9/AC-3).
func TestSQLRoles_TxRollback(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	if _, err := g.CreateClaim(ctx, models.Claim{Name: "boom_claim"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateRole(ctx, models.Role{Name: "auditor"}); err != nil {
		t.Fatal(err)
	}

	boom := errors.New("injected constraint failure")
	err := g.gormDB.Callback().Create().Before("gorm:create").Register("test:boom",
		func(d *gorm.DB) {
			if _, ok := d.Statement.Dest.(*models.RoleClaims); ok {
				_ = d.AddError(boom)
			}
		})
	if err != nil {
		t.Fatalf("callback register: %v", err)
	}

	// The injected failure surfaces through mapGormError as a 500 with the
	// original message embedded (D-11 generic branch).
	if err := g.AddClaimToRole(ctx, "auditor", "boom_claim"); err == nil ||
		!strings.Contains(err.Error(), "injected constraint failure") {
		t.Fatalf("want injected error surfaced, got %v", err)
	}
	var n int64
	g.gormDB.Table("role_claims").Count(&n)
	if n != 0 {
		t.Errorf("tx not rolled back: %d join rows remain", n)
	}
}
