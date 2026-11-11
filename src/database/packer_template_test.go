package database

import (
	"errors"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

func seedTpl(t *testing.T, g *GormDatabase, id, name string, extra ...func(*models.PackerTemplate)) {
	t.Helper()
	tpl := &models.PackerTemplate{ID: id, Name: name, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
		Variables: map[string]string{"a": "1"}}
	for _, e := range extra {
		e(tpl)
	}
	if err := g.gormDB.Table("packer_templates").Create(tpl).Error; err != nil {
		t.Fatal(err)
	}
}

func tplRootCtx() basecontext.ApiContext {
	return &fakeCtx{auth: &basecontext.AuthorizationContext{
		AuthorizedBy: "RootAuthorization", User: &apimodels.ApiUser{ID: "root", Username: "root"}}}
}

func TestSQLPackerTemplates_ListFilterAuthz(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	// Insertion order deliberately unsorted (zz before aa).
	seedTpl(t, g, "zz-one", "ubuntu-2204")
	seedTpl(t, g, "aa-two", "Windows-2022")
	seedTpl(t, g, "mm-guarded", "fedora-40", func(x *models.PackerTemplate) {
		x.RequiredRoles = []string{"operator"}
	})

	// Authz matrix row 1: ctx WITHOUT an authorization context sees NOTHING
	// — public records included (legacy authorize_layer denies nil-auth
	// before any record inspection). Pinned, not "fixed".
	all, err := g.GetPackerTemplates(ctx, "")
	if err != nil || len(all) != 0 {
		t.Fatalf("nil-auth ctx must see zero rows: %v %+v", err, all)
	}
	root, err := g.GetPackerTemplates(tplRootCtx(), "")
	if err != nil || len(root) != 3 || root[0].ID != "zz-one" || root[1].ID != "aa-two" {
		t.Fatalf("root list (insertion order): %v %+v", err, root)
	}
	// D-4 pinned shapes: unanchored, caller-anchored, i-option case fold.
	cases := []struct {
		filter string
		want   []string
	}{
		{"name=ubu", []string{"zz-one"}},     // unanchored substring
		{"name=win", []string{}},             // case-SENSITIVE: no hit on Windows
		{"name=WIN,i", []string{"aa-two"}},   // (?i)(?i) double-prefix path
		{"name=^ubuntu", []string{"zz-one"}}, // caller-anchored hit
		{"id=nope", []string{}},
		{"name=.,j", []string{"zz-one", "aa-two", "mm-guarded"}}, // unknown option silently ignored
	}
	for _, c := range cases {
		l, err := g.GetPackerTemplates(tplRootCtx(), c.filter)
		if err != nil || len(l) != len(c.want) {
			t.Fatalf("filter %q: %v %+v", c.filter, err, l)
		}
		for i := range c.want {
			if l[i].ID != c.want[i] {
				t.Fatalf("filter %q order: %d got %q want %q", c.filter, i, l[i].ID, c.want[i])
			}
		}
	}
	if _, err := g.GetPackerTemplates(tplRootCtx(), "bad~filter"); err == nil {
		t.Fatal("malformed filter must error")
	}
	// nil-vs-empty pin: zero-hit answer keeps legacy shape (non-nil empty).
	bare, err := g.GetPackerTemplates(tplRootCtx(), "id=nope")
	if err != nil || bare == nil || len(bare) != 0 {
		t.Fatalf("empty shape: %v %#v", err, bare)
	}
}

func TestSQLPackerTemplate_GetSentinels(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedTpl(t, g, "zz-one", "ubuntu-2204")
	seedTpl(t, g, "mm-guarded", "fedora-40", func(x *models.PackerTemplate) {
		x.RequiredRoles = []string{"operator"}
	})
	for _, probe := range []string{"zz-one", "ZZ-ONE", "Ubuntu-2204"} {
		tpl, err := g.GetPackerTemplate(tplRootCtx(), probe)
		if err != nil || tpl.ID != "zz-one" {
			t.Fatalf("probe %q: %v %+v", probe, err, tpl)
		}
	}
	// The lookup rides the AUTHZ-WRAPPED list: the guarded row is invisible
	// to a bare ctx even by exact id (legacy :44 quirk).
	if _, err := g.GetPackerTemplate(ctx, "mm-guarded"); !errors.Is(err, interfaces.ErrPackerTemplateNotFound) {
		t.Fatalf("authz-wrapped lookup: %v", err)
	}
	if tpl, err := g.GetPackerTemplate(tplRootCtx(), "mm-guarded"); err != nil || tpl.ID != "mm-guarded" {
		t.Fatalf("root lookup: %v %+v", err, tpl)
	}
	for _, probe := range []string{"nope", ""} {
		if _, err := g.GetPackerTemplate(ctx, probe); !errors.Is(err, interfaces.ErrPackerTemplateNotFound) {
			t.Fatalf("sentinel probe %q: %v", probe, err)
		}
	}
}

func TestSQLPackerTemplates_AddQuirks(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedTpl(t, g, "zz-one", "ubuntu-2204")

	if _, err := g.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "zz-one", Name: "dup"}); !errors.Is(err, interfaces.ErrPackerTemplateAlreadyExists) {
		t.Fatalf("dup id: %v", err)
	}
	// Duplicate NAME is allowed by legacy (probe checks ID only — the spec's
	// "dup by name" edge case is another stale claim; pinned as allowed).
	if _, err := g.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "new-id", Name: "ubuntu-2204"}); err != nil {
		t.Fatalf("same-name create: %v", err)
	}
	// Required claims/roles must exist; errors propagate verbatim.
	if _, err := g.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "c1", Name: "n", RequiredClaims: []string{"ghost-claim"}}); err == nil {
		t.Fatal("missing required claim must error")
	} else if !errors.Is(err, interfaces.ErrClaimNotFound) {
		t.Fatalf("claim error identity drifted: %v", err)
	}
	if _, err := g.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "c2", Name: "n", RequiredRoles: []string{"ghost-role"}}); err == nil {
		t.Fatal("missing required role must error")
	}
	// Timestamps stamped raw RFC3339Nano (legacy :67-68, not the helper).
	added, err := g.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "ok1", Name: "ok"})
	if err != nil || added.CreatedAt == "" || added.UpdatedAt == "" {
		t.Fatalf("stamp: %v %+v", err, added)
	}
	// Stricter-than-legacy pin (deviation): the authz-blind duplicate probe
	// passes on a bare ctx, but the SQL primary key refuses the second row
	// with the duplicate sentinel — legacy would have double-stored.
	var n int64
	if err := g.gormDB.Table("packer_templates").Where("id = ?", "zz-one").Count(&n).Error; err != nil || n != 1 {
		t.Fatalf("dup must not double-store: %v %d", err, n)
	}
}

func TestSQLPackerTemplates_Delete(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedTpl(t, g, "zz-one", "ubuntu-2204")
	seedTpl(t, g, "int-one", "internal-tpl", func(x *models.PackerTemplate) { x.Internal = true })

	if err := g.DeletePackerTemplate(ctx, ""); err != nil { // nil, NOT a sentinel (:94)
		t.Fatalf("empty delete must answer nil: %v", err)
	}
	if err := g.DeletePackerTemplate(ctx, "nope"); !errors.Is(err, interfaces.ErrPackerTemplateNotFound) {
		t.Fatalf("missing delete: %v", err)
	}
	if err := g.DeletePackerTemplate(ctx, "int-one"); !errors.Is(err, interfaces.ErrRemovingInternalPackerTemplate) {
		t.Fatalf("internal delete guard (bare): %v", err)
	}
	if err := g.DeletePackerTemplate(tplRootCtx(), "INT-ONE"); err != nil { // folded + root
		t.Fatalf("internal delete by root: %v", err)
	}
	if err := g.DeletePackerTemplate(ctx, "UBUNTU-2204"); err != nil { // folded name
		t.Fatalf("delete by folded name: %v", err)
	}
	l, _ := g.GetPackerTemplates(tplRootCtx(), "")
	if len(l) != 0 {
		t.Fatalf("rows survived: %+v", l)
	}
}

func TestSQLPackerTemplates_UpdateFullReplacement(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	seedTpl(t, g, "ubt-1", "old-name", func(x *models.PackerTemplate) {
		x.Variables = map[string]string{"old": "row"}
		x.Addons = []string{"x"}
	})

	// Legacy assigns the WHOLE struct (:123): payload wins every column,
	// CreatedAt included, and the id case lands verbatim.
	payload := &models.PackerTemplate{ID: "UBT-1", Name: "new-name", Owner: "team",
		CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-06-01T00:00:00Z",
		Variables: map[string]string{"new": "map"}, Specs: map[string]string{"memory": "2048"}}
	got, err := g.UpdatePackerTemplate(ctx, payload)
	if err != nil || got != payload {
		t.Fatalf("update return: %v", err)
	}
	stored, err := g.GetPackerTemplate(tplRootCtx(), "ubt-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.ID != "UBT-1" || stored.Name != "new-name" || stored.CreatedAt != "2025-01-01T00:00:00Z" ||
		stored.Owner != "team" || stored.Variables["new"] != "map" || stored.Variables["old"] != "" ||
		stored.Addons != nil {
		t.Fatalf("full-replacement drift: %+v", stored)
	}
	if stored.Specs["memory"] != "2048" {
		t.Fatalf("serializer round-trip: %+v", stored.Specs)
	}
	// Internal rows are unupdatable for EVERYONE (root included, :120-122).
	seedTpl(t, g, "int-1", "internal", func(x *models.PackerTemplate) { x.Internal = true })
	_, err = g.UpdatePackerTemplate(tplRootCtx(), &models.PackerTemplate{ID: "int-1", Name: "nope"})
	if !errors.Is(err, interfaces.ErrUpdatingInternalPackerTemplate) {
		t.Fatalf("internal update guard: %v", err)
	}
	if _, err := g.UpdatePackerTemplate(ctx, &models.PackerTemplate{ID: "ghost"}); !errors.Is(err, interfaces.ErrPackerTemplateNotFound) {
		t.Fatalf("missing update: %v", err)
	}
}

func TestSQLPackerTemplates_NotConnected(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cold := NewGormDatabase(ctx, nil)

	if _, err := cold.GetPackerTemplates(ctx, ""); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("list guard: %v", err)
	}
	if _, err := cold.GetPackerTemplate(ctx, "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("get guard: %v", err)
	}
	if _, err := cold.AddPackerTemplate(ctx, &models.PackerTemplate{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("add guard: %v", err)
	}
	if err := cold.DeletePackerTemplate(ctx, "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("delete guard: %v", err)
	}
	if _, err := cold.UpdatePackerTemplate(ctx, &models.PackerTemplate{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("update guard: %v", err)
	}
}
