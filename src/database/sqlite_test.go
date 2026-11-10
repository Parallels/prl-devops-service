package database

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCtx() basecontext.ApiContext {
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	return ctx
}

func TestNewDatabaseConnection_SQLite_CreatesFile(t *testing.T) {
	ctx := testCtx()
	dbPath := filepath.Join(t.TempDir(), "probe.db")

	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: dbPath})
	require.NoError(t, err)
	require.FileExists(t, dbPath)

	require.NoError(t, db.Exec("CREATE TABLE probe (id INTEGER PRIMARY KEY, val TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO probe (val) VALUES ('written')").Error)
	Close(ctx, db)

	// Reopen: data survives the lifecycle.
	db2, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: dbPath})
	require.NoError(t, err)
	defer Close(ctx, db2)

	var val string
	require.NoError(t, db2.Raw("SELECT val FROM probe LIMIT 1").Scan(&val).Error)
	assert.Equal(t, "written", val)
}

func TestNewDatabaseConnection_SQLite_CreatesMissingParentDir(t *testing.T) {
	ctx := testCtx()
	dbPath := filepath.Join(t.TempDir(), "a", "b", "c.db")

	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: dbPath})
	require.NoError(t, err)
	defer Close(ctx, db)

	assert.DirExists(t, filepath.Dir(dbPath))
	assert.FileExists(t, dbPath)
}

func TestNewDatabaseConnection_SQLite_Pragmas(t *testing.T) {
	ctx := testCtx()
	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: filepath.Join(t.TempDir(), "p.db")})
	require.NoError(t, err)
	defer Close(ctx, db)

	var mode string
	require.NoError(t, db.Raw("PRAGMA journal_mode").Scan(&mode).Error)
	assert.Equal(t, "wal", mode)

	var fk int64
	require.NoError(t, db.Raw("PRAGMA foreign_keys").Scan(&fk).Error)
	assert.Equal(t, int64(1), fk)
}

func TestNewDatabaseConnection_SQLite_PathWithSpaces(t *testing.T) {
	ctx := testCtx()
	dbPath := filepath.Join(t.TempDir(), "My Docs", "data.db")

	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: dbPath})
	require.NoError(t, err)
	defer Close(ctx, db)

	require.NoError(t, db.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY)").Error)
}

func TestNewDatabaseConnection_Memory(t *testing.T) {
	ctx := testCtx()
	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: MemoryPath})
	require.NoError(t, err)

	require.NoError(t, db.Exec("CREATE TABLE m (id INTEGER PRIMARY KEY, v TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO m (v) VALUES ('mem')").Error)
	var v string
	require.NoError(t, db.Raw("SELECT v FROM m LIMIT 1").Scan(&v).Error)
	assert.Equal(t, "mem", v)

	Close(ctx, db)

	// Private memory semantics: reopen is a brand-new empty database.
	db2, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: MemoryPath})
	require.NoError(t, err)
	defer Close(ctx, db2)

	var count int64
	require.NoError(t, db2.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='m'").Scan(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestNewDatabaseConnection_RejectsUnwritablePath(t *testing.T) {
	ctx := testCtx()
	blocker := filepath.Join(t.TempDir(), "probe.db")
	require.NoError(t, os.WriteFile(blocker, []byte("not a dir"), 0o644))
	badPath := filepath.Join(blocker, "sub", "x.db") // parent is a FILE

	cfg := DatabaseConfig{Type: DBTypeSQLite, Path: badPath}
	_, err := NewDatabaseConnection(ctx, cfg)
	require.Error(t, err)

	// Deterministic: a second identical call fails the same way, no leaked state.
	_, err2 := NewDatabaseConnection(ctx, cfg)
	require.Error(t, err2)
	assert.Equal(t, err.Error(), err2.Error())
}

// Concurrency guard (AC-adjacent "concurrent access" seed for #421): 20
// goroutines inserting through one file DB under WAL + busy_timeout.
func TestNewDatabaseConnection_SQLite_ConcurrentInserts(t *testing.T) {
	ctx := testCtx()
	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: filepath.Join(t.TempDir(), "conc.db")})
	require.NoError(t, err)
	defer Close(ctx, db)

	require.NoError(t, db.Exec("CREATE TABLE writes (id INTEGER PRIMARY KEY, worker TEXT)").Error)

	var wg sync.WaitGroup
	errCh := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := db.Exec("INSERT INTO writes (worker) VALUES (?)", string(rune('a'+i))).Error; err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		assert.NoError(t, err, "concurrent insert must not fail under busy_timeout")
	}

	var count int64
	require.NoError(t, db.Raw("SELECT COUNT(*) FROM writes").Scan(&count).Error)
	assert.Equal(t, int64(20), count)
}
