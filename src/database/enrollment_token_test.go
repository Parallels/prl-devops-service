package database

// Enrollment tokens on GORM (PBI #421 task 16): minting/TTL, token-value
// Fold reads, used-before-expiry validation, EXACT-id idempotent marking,
// and the sweep predicate (used-always, expired-by-clock, unparseable and
// empty stamps survive). All clock edges via the shared nowFn seam.

import (
	"errors"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
)

func TestSQLEnrollmentTokens_CRUDValidate(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	t0 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g.nowFn = func() time.Time { return t0 }

	cold := NewGormDatabase(ctx, nil)
	if _, err := cold.CreateEnrollmentToken(ctx, "h", 5); !errors.Is(err, interfaces.ErrDatabaseNotConnected) {
		t.Fatalf("cold create: %v", err)
	}
	if _, err := cold.GetEnrollmentToken(ctx, "x"); !errors.Is(err, interfaces.ErrDatabaseNotConnected) {
		t.Fatalf("cold get: %v", err)
	}
	if _, err := cold.ValidateEnrollmentToken(ctx, "x"); !errors.Is(err, interfaces.ErrDatabaseNotConnected) {
		t.Fatalf("cold validate: %v", err)
	}
	if err := cold.MarkEnrollmentTokenUsed(ctx, "x"); !errors.Is(err, interfaces.ErrDatabaseNotConnected) {
		t.Fatalf("cold mark: %v", err)
	}
	if err := cold.DeleteExpiredEnrollmentTokens(ctx); !errors.Is(err, interfaces.ErrDatabaseNotConnected) {
		t.Fatalf("cold sweep: %v", err)
	}

	tok, err := g.CreateEnrollmentToken(ctx, "host-a", 5)
	if err != nil || tok.ID == "" || len(tok.Token) != 72 || tok.Used {
		t.Fatalf("mint: %v %+v", err, tok)
	}
	if tok.ExpiresAt != t0.Add(5*time.Minute).UTC().Format(time.RFC3339) || tok.CreatedAt != t0.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("ttl/stamps: %+v", tok)
	}
	tok15, err := g.CreateEnrollmentToken(ctx, "host-b", 0) // ttl<=0 -> 15
	if err != nil || tok15.ExpiresAt != t0.Add(15*time.Minute).UTC().Format(time.RFC3339) {
		t.Fatalf("ttl default: %v %+v", err, tok15)
	}

	got, err := g.GetEnrollmentToken(ctx, foldUpper(tok.Token)) // Fold on token VALUE
	if err != nil || got.ID != tok.ID {
		t.Fatalf("fold get: %v %+v", err, got)
	}
	if _, err := g.GetEnrollmentToken(ctx, tok.ID); !errors.Is(err, interfaces.ErrEnrollmentTokenNotFound) {
		t.Fatalf("Get must never match the id: %v", err)
	}
	if _, err := g.GetEnrollmentToken(ctx, "nope"); !errors.Is(err, interfaces.ErrEnrollmentTokenNotFound) {
		t.Fatalf("miss: %v", err)
	}

	if v, err := g.ValidateEnrollmentToken(ctx, tok.Token); err != nil || v.ID != tok.ID {
		t.Fatalf("validate valid: %v %+v", err, v)
	}
	setExpiry := func(id, stamp string) {
		if err := g.gormDB.Table("orchestrator_enrollment_tokens").Where("id = ?", id).
			Update("expires_at", stamp).Error; err != nil {
			t.Fatal(err)
		}
	}
	// Boundary: expires_at EXACTLY now is VALID (strictly-After rule :81).
	setExpiry(tok.ID, t0.Format(time.RFC3339))
	if _, err := g.ValidateEnrollmentToken(ctx, tok.Token); err != nil {
		t.Fatalf("boundary must be valid: %v", err)
	}
	setExpiry(tok.ID, t0.Add(-time.Second).UTC().Format(time.RFC3339))
	if _, err := g.ValidateEnrollmentToken(ctx, tok.Token); !errors.Is(err, interfaces.ErrEnrollmentTokenExpired) {
		t.Fatalf("expired: %v", err)
	}
	setExpiry(tok.ID, "garbage-stamp") // unparseable -> VALID (:81 err-skip)
	if _, err := g.ValidateEnrollmentToken(ctx, tok.Token); err != nil {
		t.Fatalf("unparseable must validate: %v", err)
	}

	// Mark: EXACT id, idempotent; token value never matches.
	if err := g.MarkEnrollmentTokenUsed(ctx, tok.Token); !errors.Is(err, interfaces.ErrEnrollmentTokenNotFound) {
		t.Fatalf("mark by token must miss: %v", err)
	}
	if err := g.MarkEnrollmentTokenUsed(ctx, tok.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.MarkEnrollmentTokenUsed(ctx, tok.ID); err != nil { // re-mark idempotent
		t.Fatalf("re-mark: %v", err)
	}
	// used checked BEFORE expiry (:75): expired+used reports USED.
	if _, err := g.ValidateEnrollmentToken(ctx, tok.Token); !errors.Is(err, interfaces.ErrEnrollmentTokenUsed) {
		t.Fatalf("used-before-expired: %v", err)
	}
}

func TestSQLEnrollmentTokens_Sweep(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	t0 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g.nowFn = func() time.Time { return t0 }

	mk := func(host string, used bool, expiry string) string {
		tok, err := g.CreateEnrollmentToken(ctx, host, 60)
		if err != nil {
			t.Fatal(err)
		}
		if expiry != "KEEP" { // KEEP = retain the created +60m stamp; "" explicitly cleared
			if err := g.gormDB.Table("orchestrator_enrollment_tokens").Where("id = ?", tok.ID).
				Update("expires_at", expiry).Error; err != nil {
				t.Fatal(err)
			}
		}
		if used {
			if err := g.MarkEnrollmentTokenUsed(ctx, tok.ID); err != nil {
				t.Fatal(err)
			}
		}
		return tok.Token
	}
	expired := mk("x1", false, t0.Add(-time.Second).UTC().Format(time.RFC3339))
	usedFresh := mk("x2", true, t0.Add(time.Hour).UTC().Format(time.RFC3339)) // used ALWAYS pruned (:128)
	fresh := mk("x3", false, "KEEP")
	emptyExp := mk("x4", false, "")
	_ = emptyExp                              // "" expiry: never pruned on time, keep
	garbage := mk("x5", false, "not-a-stamp") // unparseable stays (:133)
	// empty-expiry + used goes (used gate wins regardless):
	usedEmpty := mk("x6", true, "")
	_ = usedEmpty

	if err := g.DeleteExpiredEnrollmentTokens(ctx); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct {
		token string
		want  bool
	}{{expired, false}, {usedFresh, false}, {usedEmpty, false}, {fresh, true}, {emptyExp, true}, {garbage, true}} {
		_, err := g.GetEnrollmentToken(ctx, probe.token)
		if probe.want && err != nil {
			t.Fatalf("survivor pruned: %v", err)
		}
		if !probe.want && !errors.Is(err, interfaces.ErrEnrollmentTokenNotFound) {
			t.Fatalf("doomed survived: %v", err)
		}
	}
	if err := g.DeleteExpiredEnrollmentTokens(ctx); err != nil { // no-op second pass
		t.Fatal(err)
	}
	if err := g.MarkEnrollmentTokenUsed(ctx, mustRowID(t, g, fresh)); err != nil {
		t.Fatal(err)
	}
	if err := g.MarkEnrollmentTokenUsed(ctx, mustRowID(t, g, emptyExp)); err != nil {
		t.Fatal(err)
	}
	if err := g.MarkEnrollmentTokenUsed(ctx, mustRowID(t, g, garbage)); err != nil {
		t.Fatal(err)
	}
	if err := g.DeleteExpiredEnrollmentTokens(ctx); err != nil { // sweep-empties-all
		t.Fatal(err)
	}
	for _, probe := range []string{fresh, emptyExp, garbage} {
		if _, err := g.GetEnrollmentToken(ctx, probe); !errors.Is(err, interfaces.ErrEnrollmentTokenNotFound) {
			t.Fatalf("sweep-all: %v", err)
		}
	}
}

func mustRowID(t *testing.T, g *GormDatabase, token string) string {
	t.Helper()
	row, err := g.GetEnrollmentToken(basecontext.NewBaseContext(), token)
	if err != nil {
		t.Fatal(err)
	}
	return row.ID
}

func foldUpper(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'f' {
			b[i] -= 32
		}
	}
	return string(b)
}
