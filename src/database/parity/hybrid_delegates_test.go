package parity

// Hybrid-delegate parity (task 04; re-pointed in 05/06/07/08/09/10/11/12/14/15):
// representative UNPORTED domains through GormDatabase must be
// observationally identical to the legacy handle. Since task 15 jobs became
// ported SQL, so enrollment tokens are the unported representative
// (delegate pair CreateEnrollmentToken/GetEnrollmentToken); reverse-proxy
// hosts remain the zero-reach-through subject (ported reads AND writes must
// stay in the SQL store, never crossing to legacy). Lives here, not in
// package database: see gorm_database_test.go header for the legacy-
// goroutine rationale.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestHybridDelegatesMatchLegacy(t *testing.T) {
	legacy := NewLegacyEngine(t)
	ctx := basecontext.NewBaseContext()

	// Unported domain (enrollment tokens): delegate pass-through must
	// mirror the legacy handle exactly — same store, same mutation, both
	// directions.
	if _, err := legacy.GetEnrollmentToken(ctx, "none"); err == nil {
		t.Fatalf("legacy token read should report not-found")
	}

	// Hybrid engine: migrated :memory: SQL + this legacy handle.
	sql := newSQLTestEngine(t, legacy)

	tok, err := sql.CreateEnrollmentToken(ctx, "host-x", 15)
	if err != nil || tok == nil || tok.Token == "" {
		t.Fatalf("delegate write: %v %+v", err, tok)
	}
	ltok, err := legacy.GetEnrollmentToken(ctx, tok.Token)
	if err != nil || ltok.HostName != "host-x" {
		t.Fatalf("delegate write did not reach legacy store: %v %+v", err, ltok)
	}
	stok, err := sql.GetEnrollmentToken(ctx, tok.Token)
	if err != nil || stok.ID != ltok.ID {
		t.Fatalf("delegate read drift: %v %+v vs %+v", err, stok, ltok)
	}

	// Ported domain (reverse-proxy hosts since task 12) must NOT reach the
	// delegate in either direction: the legacy-created host stays invisible
	// to SQL, and the SQL-created host never materialises in legacy.
	if _, err := legacy.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Host: "ghost.local", Port: "80"}); err != nil {
		t.Fatalf("seed legacy host: %v", err)
	}
	if n := mustHostCount(legacy); n != 1 {
		t.Fatalf("legacy probe seed: %d", n)
	}
	if n := mustHostCount(sql); n != 0 {
		t.Errorf("SQL store saw the legacy-store host — reach-through: %d", n)
	}
	if _, err := sql.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Host: "sqlonly.local", Port: "80"}); err != nil {
		t.Fatal(err)
	}
	if n := mustHostCount(legacy); n != 1 {
		t.Errorf("legacy store grew through the ported domain — write reach-through: %d", n)
	}
}

func mustHostCount(db interface {
	GetReverseProxyHosts(ctx basecontext.ApiContext, filter string) ([]models.ReverseProxyHost, error)
}) int {
	rows, _ := db.GetReverseProxyHosts(basecontext.NewBaseContext(), "")
	return len(rows)
}
