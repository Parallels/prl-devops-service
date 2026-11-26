package database

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
)

// The engine branch is pure, so both legs are testable without a server: the
// PostgreSQL half of this cannot be connected to in a unit test, but the
// expression it hands to ORDER BY can be pinned exactly — which is the thing
// that was crashing a fresh install.

func TestInsertionOrderPerEngine(t *testing.T) {
	cases := []struct {
		dialect, table, want string
	}{
		{"sqlite", "jobs", "rowid"},
		{"", "jobs", "rowid"},           // unknown engine ⇒ the supported one
		{"postgres", "jobs", "jobs.id"}, // qualified: Preload builds joins
		{"postgres", "catalog_manifests", "catalog_manifests.id"},
		{"postgres", "schema_metadata", "schema_metadata.version"}, // PK is version
	}

	for _, c := range cases {
		if got := insertionOrderFor(c.dialect, c.table); got != c.want {
			t.Errorf("insertionOrderFor(%q, %q) = %q, want %q", c.dialect, c.table, got, c.want)
		}
	}
}

func TestInsertionOrderAsksTheConnectedEngine(t *testing.T) {
	g := newSQLDB(t)

	// The SQLite leg must stay byte-for-byte what every existing test pins.
	if got := insertionOrder(g.gormDB, "jobs"); got != "rowid" {
		t.Fatalf("sqlite leg changed: %q", got)
	}

	// A handle that does not exist yet (cold store, early boot) still answers,
	// and answers with the supported backend's expression.
	if got := insertionOrder(nil, "jobs"); got != "rowid" {
		t.Fatalf("cold handle: %q", got)
	}
}

func TestInsertionOrderIsUsableInAQuery(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()

	for _, id := range []string{"ffff", "0000", "aaaa"} {
		if _, err := g.CreateJob(root, models.Job{ID: id}); err != nil {
			t.Fatal(err)
		}
	}

	var ids []string
	if err := g.gormDB.Table("jobs").Order(insertionOrder(g.gormDB, "jobs")).Pluck("id", &ids).Error; err != nil {
		t.Fatalf("ordering a real query: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("rows: %v", ids)
	}

	// On SQLite this is append order; the point of the test is that the
	// expression GORM is handed is valid for the engine in play.
	if ids[0] != "ffff" {
		t.Fatalf("sqlite rowid order broken: %v", ids)
	}
}
