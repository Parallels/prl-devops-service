package placement

// Cross-dialect contract leg. The admission SQL is hand-written and the whole
// design depends on it being dialect-neutral (no rowid, no ON CONFLICT, no
// pg_*), so the same invariants run against Postgres when a DSN is available:
//
//	POSTGRES_TEST_DSN='postgres://postgres:postgres@127.0.0.1:5432/prl_test?sslmode=disable' \
//	  make test-integration
//
// This mirrors how database/testutil.PostgresEngine boots a throwaway schema,
// and deliberately does NOT go through testutil.RunOnEngines: those engines
// hand back a DatabaseService and run the T17 rowid probe, and placement never
// touches the frozen DatabaseService surface (#421 R-1).

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openPostgresStore(t *testing.T) (*sqlStore, *gorm.DB, basecontext.ApiContext) {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN unset — Postgres leg skipped (CI service container / make test-integration only)")
	}

	ctx := testutil.WithCleanConfig(t)
	schema := fmt.Sprintf("prl_placement_%d", time.Now().UnixNano())
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}

	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
		Type: database.DBTypePostgres,
		DSN:  dsn + sep + "search_path=" + schema,
	})
	require.NoError(t, err, "postgres boot failed on %s", schema)
	t.Cleanup(func() { database.Close(ctx, db) })
	require.NoError(t, db.Exec("CREATE SCHEMA IF NOT EXISTS "+schema).Error)
	t.Cleanup(func() {
		_ = db.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE").Error
	})
	require.NoError(t, database.RunAutoMigrate(ctx, db))

	return newSQLStore(db), db, ctx
}

// TestPostgresStoreContract runs the invariants the design cannot do without
// against the other engine: the guard-row lock really does serialise, the
// aggregate really is a single SUM, and the user column really is quoted
// rather than silently resolving to current_user.
func TestPostgresStoreContract(t *testing.T) {
	store, db, ctx := openPostgresStore(t)
	seedHost(t, db, "host-1", 10, 8192)
	seedHost(t, db, "host-2", 10, 8192)

	// Concurrent burst against a fixed snapshot: five 2-CPU creates fit ten
	// CPUs, on either engine.
	const attempts = 20
	var admitted atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			admission, err := store.Admit(ctx, "host-1",
				reserve(fmt.Sprintf("job-%02d", i), "host-1", fmt.Sprintf("vm-%02d", i), "root", 2, 0), CapacityView{})
			if err == nil && admission.Admitted {
				admitted.Add(1)
			}
		}(i)
	}
	wg.Wait()
	assert.Equal(t, int64(5), admitted.Load(), "Postgres must serialise admissions exactly like SQLite")

	pending, err := store.PendingByHost(ctx, "host-1")
	require.NoError(t, err)
	assert.Equal(t, int64(10), pending.CPU)
	assert.Equal(t, int64(5), pending.VMCount)

	// The unquoted "user" column is the trap this leg exists for: unquoted, it
	// is a reserved word in Postgres and the comparison silently becomes
	// current_user = 'root', which matches nothing.
	require.NoError(t, db.Create(&data_models.VirtualMachine{
		ID: "vm-1", Name: "raw-name", User: "root", HostId: "host-1",
	}).Error)

	taken, err := store.NameTaken(ctx, "root", "raw-name", "")
	require.NoError(t, err)
	assert.True(t, taken, "the existing-VM probe must match on the owner column")

	taken, err = store.NameTaken(ctx, "someone-else", "raw-name", "")
	require.NoError(t, err)
	assert.False(t, taken, "the name guard is per owner")

	// Refusal wording must be identical across engines — the QA runbook greps it.
	admission, err := store.Admit(ctx, "host-1", reserve("job-over", "host-1", "vm-over", "root", 6, 0), CapacityView{})
	require.NoError(t, err)
	assert.False(t, admission.Admitted)
	assert.Equal(t, "CPU: required 6 logical CPUs, available 0", admission.Message)

	// Re-placing onto the second host moves the charge and frees the first.
	moved, err := store.Admit(ctx, "host-2", reserve("job-00", "host-2", "vm-00", "root", 2, 0), CapacityView{})
	require.NoError(t, err)
	assert.True(t, moved.Admitted)

	byHost, err := store.PendingByHosts(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), byHost["host-2"].CPU, "the moved job's charge follows it")

	// Sweeps must be set-based deletes here too — that is what makes two
	// replicas reconciling at once harmless.
	removed, err := store.ExpireLeases(ctx, time.Now().UTC().Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(5), removed, "every claim predates a cutoff an hour ahead")
}
