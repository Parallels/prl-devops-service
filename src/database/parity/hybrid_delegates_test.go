package parity

// Hybrid-delegate parity (task 04; re-pointed in 05/06/07/08/09/10):
// representative UNPORTED domains through GormDatabase must be
// observationally identical to the legacy handle. Since task 10 the
// reverse-proxy config domain is the unported representative; packer
// templates became the zero-reach-through subject (ported reads AND writes
// must stay in the SQL store, never crossing to legacy). Lives here, not in
// package database: see gorm_database_test.go header for the
// legacy-goroutine rationale.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestHybridDelegatesMatchLegacy(t *testing.T) {
	legacy := NewLegacyEngine(t)
	ctx := basecontext.NewBaseContext()

	// Unported domain (reverse-proxy config): delegate pass-through must
	// mirror the legacy handle exactly — same store, same mutation. The
	// singleton must exist first (Enable/Disable 404 on a nil config —
	// legacy :42/:56), so seed it through UpdateReverseProxy.
	if _, err := legacy.UpdateReverseProxy(ctx, models.ReverseProxy{}); err != nil {
		t.Fatalf("seed reverse-proxy config: %v", err)
	}
	lrp, err := legacy.EnableProxyConfig(ctx)
	if err != nil || lrp == nil || !lrp.Enabled {
		t.Fatalf("legacy EnableProxyConfig: %v %+v", err, lrp)
	}

	// Hybrid engine: migrated :memory: SQL + this legacy handle.
	sql := newSQLTestEngine(t, legacy)

	srp, err := sql.GetReverseProxyConfig(ctx)
	if err != nil || srp == nil || srp.Enabled != lrp.Enabled {
		t.Fatalf("delegate read drift: %v %+v vs %+v", err, srp, lrp)
	}
	if _, err := sql.DisableProxyConfig(ctx); err != nil {
		t.Fatal(err)
	}
	lrp2, err := legacy.GetReverseProxyConfig(ctx)
	if err != nil || lrp2 == nil || lrp2.Enabled {
		t.Fatalf("delegate write did not reach legacy store: %v %+v", err, lrp2)
	}

	// Ported domain (packer templates) must NOT reach the delegate in either
	// direction: the legacy-seeded template stays invisible to SQL, and the
	// SQL-added template never materialises in the legacy store.
	if _, err := legacy.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "legacy-tpl", Name: "Legacy"}); err != nil {
		t.Fatalf("seed legacy tpl: %v", err)
	}
	root := &parityCtx{auth: &basecontext.AuthorizationContext{AuthorizedBy: "RootAuthorization"}}
	if n := mustLen(legacy, root); n != 1 {
		t.Fatalf("legacy probe seed: %d", n)
	}
	if n := mustLen(sql, root); n != 0 {
		t.Errorf("SQL list saw the legacy-store template — reach-through: %d", n)
	}
	if _, err := sql.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "sql-tpl", Name: "SQL"}); err != nil {
		t.Fatal(err)
	}
	if n := mustLen(legacy, root); n != 1 {
		t.Errorf("legacy store grew through the ported domain — write reach-through: %d", n)
	}
}

func mustLen(db interface {
	GetPackerTemplates(ctx basecontext.ApiContext, filter string) ([]models.PackerTemplate, error)
}, ctx basecontext.ApiContext) int {
	rows, _ := db.GetPackerTemplates(ctx, "")
	return len(rows)
}
