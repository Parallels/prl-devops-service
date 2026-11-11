package parity

// DomainSuite "enrollment-tokens" (PBI #421 task 16): create/Get/Validate/
// Mark/Sweep across both engines — token-VALUE Fold reads (the id NEVER
// matches Get), cross-store isolation probes, used-before-expiry outcome,
// idempotent EXACT-id marking, and the sweep (used-always pruning; fresh
// survives; sweep-empties-all). Expired-clock edges cannot travel in
// parity (legacy has no clock hook, ER-5 forbids touching it): pinned in
// the clock-injected SQL units, mirroring the task-15 stale-window split.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func normaliseET(tok *models.OrchestratorEnrollmentToken) *models.OrchestratorEnrollmentToken {
	if tok == nil {
		return nil
	}
	out := *tok
	out.ID, out.Token = "ID", "TOK"
	out.ExpiresAt, out.CreatedAt = "SET", "SET"
	out.DbRecord = nil // legacy mints &DbRecord{}, SQL has no such device
	return &out
}

func foldUp(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'f' {
			b[i] -= 32
		}
	}
	return string(b)
}

func TestParityDomain_EnrollmentTokens(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "enrollment-tokens",
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkETLifecycle,
		},
	})
}

func checkETLifecycle(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	runPair := func(name string, fn func(db datainterfaces.DatabaseService) error) {
		le, se := fn(legacy), fn(sql)
		assertSameError(t, name, le, se)
	}

	// Zero-token state.
	for _, probe := range []string{"none", ""} {
		_, lerr := legacy.GetEnrollmentToken(ctx, probe)
		_, serr := sql.GetEnrollmentToken(ctx, probe)
		assertSameError(t, "GetEmpty["+probe+"]", lerr, serr)
		_, lerr = legacy.ValidateEnrollmentToken(ctx, probe)
		_, serr = sql.ValidateEnrollmentToken(ctx, probe)
		assertSameError(t, "ValidateEmpty["+probe+"]", lerr, serr)
	}
	runPair("SweepEmpty", func(db datainterfaces.DatabaseService) error { return db.DeleteExpiredEnrollmentTokens(ctx) })

	lc, lerr := legacy.CreateEnrollmentToken(ctx, "h-enroll", 15)
	sc, serr := sql.CreateEnrollmentToken(ctx, "h-enroll", 15)
	assertSameError(t, "Create", lerr, serr)
	if lc == nil || sc == nil {
		t.Fatalf("create views: %+v %+v", lc, sc)
	}
	AssertSameList(t, "Create-view",
		[]*models.OrchestratorEnrollmentToken{normaliseET(lc)},
		[]*models.OrchestratorEnrollmentToken{normaliseET(sc)})

	// Get: Fold token-value hits on the OWN engine only — id never matches,
	// and the other engine's token is a miss (store isolation).
	pairs := []struct {
		db         datainterfaces.DatabaseService
		own        *models.OrchestratorEnrollmentToken
		otherToken string
	}{{legacy, lc, sc.Token}, {sql, sc, lc.Token}}
	for _, p := range pairs {
		got, err := p.db.GetEnrollmentToken(ctx, foldUp(p.own.Token))
		if err != nil || got.ID != p.own.ID {
			t.Fatalf("fold get: %v %+v", err, got)
		}
		_, err = p.db.GetEnrollmentToken(ctx, p.own.ID)
		if err == nil {
			t.Fatalf("Get must never match the id")
		}
		_, err = p.db.GetEnrollmentToken(ctx, p.otherToken)
		if err == nil {
			t.Fatalf("cross-store leak")
		}
		v, _ := p.db.GetEnrollmentToken(ctx, p.own.Token)
		AssertSameList(t, "Get-view",
			[]*models.OrchestratorEnrollmentToken{normaliseET(lc)},
			[]*models.OrchestratorEnrollmentToken{normaliseET(v)})

		if v, err := p.db.ValidateEnrollmentToken(ctx, p.own.Token); err != nil || v.ID != p.own.ID {
			t.Fatalf("validate valid: %v %+v", err, v)
		}
		// Fresh + 15min TTL under the real clock: the sweep keeps it.
		if err := p.db.DeleteExpiredEnrollmentTokens(ctx); err != nil {
			t.Fatalf("sweep fresh: %v", err)
		}
		if _, err := p.db.GetEnrollmentToken(ctx, p.own.Token); err != nil {
			t.Fatalf("fresh pruned: %v", err)
		}
		// Mark used: idempotent nil; validation flips to USED (before any
		// expiry question); the sweep then ALWAYS prunes a used token.
		if err := p.db.MarkEnrollmentTokenUsed(ctx, p.own.ID); err != nil {
			t.Fatalf("mark: %v", err)
		}
		if err := p.db.MarkEnrollmentTokenUsed(ctx, p.own.ID); err != nil {
			t.Fatalf("re-mark: %v", err)
		}
		if _, err := p.db.ValidateEnrollmentToken(ctx, p.own.Token); err == nil {
			t.Fatalf("validate used must fail")
		}
	}
	// Both marked now: the outcome errors must be the same sentinel.
	_, lerr = legacy.ValidateEnrollmentToken(ctx, lc.Token)
	_, serr = sql.ValidateEnrollmentToken(ctx, sc.Token)
	assertSameError(t, "ValidateUsed", lerr, serr)

	// Sweep-empties-all.
	runPair("SweepAll", func(db datainterfaces.DatabaseService) error { return db.DeleteExpiredEnrollmentTokens(ctx) })
	for _, p := range pairs {
		if _, err := p.db.GetEnrollmentToken(ctx, p.own.Token); err == nil {
			t.Fatalf("sweep left a used token")
		}
	}
	// Sweep on the now-empty store: nil on both.
	runPair("SweepEmpty2", func(db datainterfaces.DatabaseService) error { return db.DeleteExpiredEnrollmentTokens(ctx) })
}
