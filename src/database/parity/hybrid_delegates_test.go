package parity

// Hybrid-delegate parity (task 04; re-pointed in 05/06/07/08): representative
// UNPORTED domains through GormDatabase must be observationally identical to
// the legacy handle. Since task 08 the packer-templates domain is the
// unported representative; config/schema singletons became the
// zero-reach-through subjects (seed-on-read must mint per-store identities,
// never cross engines). Lives here, not in package database: see
// gorm_database_test.go header for the legacy-goroutine rationale.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestHybridDelegatesMatchLegacy(t *testing.T) {
	legacy := NewLegacyEngine(t)
	ctx := basecontext.NewBaseContext()
	if _, err := legacy.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "probe-tpl", Name: "Probe Template", Owner: "team-a"}); err != nil {
		t.Fatalf("seed packer template: %v", err)
	}

	// Hybrid engine: migrated :memory: SQL + this legacy handle.
	sql := newSQLTestEngine(t, legacy)

	// Unported domain (packer templates): delegate pass-through must mirror
	// legacy exactly — same store, same filter, same values. Root ctx: the
	// domain is authz-wrapped and a bare ctx legitimately sees zero records
	// on BOTH engines (equal but vacuous).
	rootCtx := &parityCtx{auth: &basecontext.AuthorizationContext{AuthorizedBy: "RootAuthorization"}}
	lt, lerr := legacy.GetPackerTemplates(rootCtx, "")
	gt, gerr := sql.GetPackerTemplates(rootCtx, "")
	if lerr != nil || gerr != nil {
		t.Fatalf("GetPackerTemplates: %v %v", lerr, gerr)
	}
	if len(lt) != len(gt) || len(lt) == 0 || lt[0].ID != gt[0].ID {
		t.Fatalf("packer-template delegate drift: %d vs %d", len(lt), len(gt))
	}

	// Ported singleton domains must NOT reach the delegate: the legacy-side
	// installation id (lazily minted here) must stay independent of the SQL
	// engine's own seed-on-read id, and legacy store must not mutate.
	legacyId, err := legacy.GetId(ctx)
	if err != nil || legacyId == "" {
		t.Fatalf("legacy GetId seed: %v %q", err, legacyId)
	}
	cfg, err := sql.GetConfiguration(ctx)
	if err != nil || cfg == nil || cfg.ID == "" {
		t.Fatalf("hybrid GetConfiguration seed-on-read: %v %+v", err, cfg)
	}
	if cfg.ID == legacyId {
		t.Errorf("SQL singleton read the legacy id — reach-through")
	}
	if again, _ := legacy.GetId(ctx); again != legacyId {
		t.Errorf("legacy store mutated through hybrid: %q -> %q", legacyId, again)
	}
	if err := sql.UpdateSchemaVersion(ctx, "sql-only-version"); err != nil {
		t.Fatal(err)
	}
	if lv, err := legacy.GetSchemaVersion(ctx); err != nil || lv != "" {
		t.Errorf("schema version reached into legacy store: %q (%v)", lv, err)
	}
}
