package parity

// SQL engine registration for every DomainSuite (task 05 onward). Lives in
// the test binary: production boot registers its own engine in task 18.
//
// The GormDatabase is built hybrid (D-1): migrated :memory: sqlite + the
// test's legacy handle as delegate, so checks may touch unported domains
// (they reach the SAME legacy store the legacy engine uses — true hybrid
// observability). Ported domains bypass the delegate entirely.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/database"
)

func init() {
	SetSQLEngineFactory(newSQLTestEngine)
}

// allGormModels mirrors the data/models test registry; the production
// AutoMigrate list ships with the task-18 boot wiring.
func allGormModels() []interface{} { return database.AllGormModels() }

func newSQLTestEngine(t testing.TB, legacy *data.JsonDatabase) datainterfaces.DatabaseService {
	t.Helper()
	conn, err := database.NewDatabaseConnection(basecontext.NewBaseContext(),
		database.DatabaseConfig{Type: "sqlite", Path: database.MemoryPath})
	if err != nil {
		t.Fatalf("parity sql connect: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, e := conn.DB(); e == nil {
			_ = sqlDB.Close()
		}
	})
	if err := conn.AutoMigrate(allGormModels()...); err != nil {
		t.Fatalf("parity sql automigrate: %v", err)
	}
	return database.NewGormDatabase(basecontext.NewBaseContext(), conn)
}
