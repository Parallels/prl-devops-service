package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryDSNShape(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 3; i++ {
		dsn := memoryDSN(t)
		assert.Regexp(t, `^file:t425-[a-z0-9-]{1,48}-\d+\?mode=memory&cache=shared$`, dsn)
		assert.Falsef(t, seen[dsn], "seq must make repeats distinct: %s", dsn)
		seen[dsn] = true
	}
	// sanitiser: slashes/uppercase/dots from real subtest names → [a-z0-9-]
	t.Run("Sub/Test.Name", func(t *testing.T) {
		dsn := memoryDSN(t)
		name := regexp.MustCompile(`^file:(t425-.*)-\d+\?`).FindStringSubmatch(dsn)[1]
		assert.Regexp(t, `^[a-z0-9-]+$`, name)
	})
}

// The leak proof: same name ⇒ same logical DB (why unique DSNs are MANDATORY).
func TestSameNamedDSNShares(t *testing.T) {
	const shared = "file:t425-shape-share-probe?mode=memory&cache=shared"
	rc, ctx := RecordingContext()
	db1, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{Type: database.DBTypeSQLite, Path: shared})
	require.NoErrorf(t, err, "%v", rc.Lines())
	defer database.Close(ctx, db1)
	db2, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{Type: database.DBTypeSQLite, Path: shared})
	require.NoError(t, err)
	defer database.Close(ctx, db2)

	require.NoError(t, db1.Exec("CREATE TABLE probe (x int)").Error)
	assert.True(t, db2.Migrator().HasTable("probe"), "same-name DSNs must share one logical DB")
}

func TestDistinctDSNsIsolated(t *testing.T) {
	dbA := OpenMemoryDB(t)
	dbB := OpenMemoryDB(t)
	require.NoError(t, dbA.Exec("CREATE TABLE only_a (x int)").Error)
	assert.True(t, dbA.Migrator().HasTable("only_a"))
	assert.False(t, dbB.Migrator().HasTable("only_a"), "distinct DSNs leaked state")
}

func TestNoDiskWrites(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DATABASE_FOLDER", dir)
	db := OpenMemoryDB(t)
	MigrateSchema(t, db)
	require.NoError(t, db.Exec("INSERT INTO users (id, username) VALUES ('x','x')").Error)

	var files []string
	require.NoError(t, filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, p)
		}
		return nil
	}))
	assert.Empty(t, files, "memory legs must touch no disk: %v", files)
}

func TestRecordingContextCaptures(t *testing.T) {
	rc, ctx := RecordingContext()
	ctx.LogInfof("i %d", 1)
	ctx.LogErrorf("e %s", "boom")
	ctx.LogWarnf("w")
	ctx.LogDebugf("d")
	ctx.LogTracef("t")
	lines := rc.Lines()
	require.Len(t, lines, 5)
	assert.Equal(t, []string{"info: i 1", "error: e boom", "warn: w", "debug: d", "trace: t"}, lines)
	assert.True(t, rc.HasLevel("warn", "w"))
	assert.False(t, rc.HasLevel("warn", "boom"))
}

// TestOpenMigrateCost — D-9 budget probe. t.Logf ONLY (never fails); median
// is pasted in the PR description; > 40 ms ⇒ escalate fallback (a) now.
func TestOpenMigrateCost(t *testing.T) {
	const runs = 20
	var samples []time.Duration
	for i := 0; i < runs; i++ {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			start := time.Now()
			db := OpenMemoryDB(t)
			MigrateSchema(t, db)
			samples = append(samples, time.Since(start))
		})
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
	median := samples[len(samples)/2]
	t.Logf("D-9: median open+AutoMigrate over %d runs = %s (budget: ~100 scenarios × this ≤ +30 s)", runs, median.Round(time.Millisecond))
}
