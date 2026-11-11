package parity

// Zero-reach-through pin (task 04 as hybrid-delegate parity, re-pointed in
// 05-16 as domains ported; task 18 closed the architecture): with the legacy
// handle DELETED from GormDatabase, the only surviving hybrid assertion is
// that the two stores stay separate — legacy-created rows stay invisible to
// SQL and SQL-created rows never materialise in legacy. (The former SaveAs
// probe retired with the delegate; T5 records the legacy defect it bumped
// into, which now only concerns legacy mode.) Lives here, not in package
// database: see gorm_database_test.go header for the legacy-goroutine
// rationale.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestZeroReachThrough(t *testing.T) {
	legacy := NewLegacyEngine(t)
	ctx := basecontext.NewBaseContext()

	// SQL engine: migrated :memory:, booted exactly like production SQL mode.
	sql := newSQLTestEngine(t, legacy)

	// Reverse-proxy hosts (ported since task 12): the legacy-created host
	// stays invisible to SQL, and the SQL-created host never reaches legacy.
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
