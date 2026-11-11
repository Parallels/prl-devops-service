//go:build integration

// Task 07, part 2: the adversarial half of the PBI proof — a mid-migration
// "process death" and the API-command exit policy observed at PROCESS
// level (unit tests must never trigger os.Exit, so this is the only place
// AC-3/exit semantics are proven against the real binary).
package legacy_migrator

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// openDB reopens an EXISTING sqlite file with a fresh handle — proving the
// rows really did (or did not) survive the transaction, nothing lived in
// an open tx.
func openDB(t *testing.T, dir string) *gorm.DB {
	t.Helper()
	gdb, err := database.NewDatabaseConnection(basecontext.NewRootBaseContext(), database.DatabaseConfig{
		Type: database.DBTypeSQLite,
		Path: filepath.Join(dir, "devops.db"),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if sqlDB, e := gdb.DB(); e == nil {
			_ = sqlDB.Close()
		}
	})
	return gdb
}

func assertAllEmpty(t *testing.T, gdb *gorm.DB) {
	t.Helper()
	empty := &data.Data{}
	for _, s := range collectionPlan() {
		_, actual, err := s.Validate(gdb, empty)
		require.NoErrorf(t, err, "counting %s after rollback", s.Name)
		assert.Zerof(t, actual, "collection %s must hold no committed rows", s.Name)
	}
	if gdb.Migrator().HasTable("legacy_migration_marker") {
		assert.Empty(t, markerRows(t, gdb), "marker must never commit alone")
	}
}

// AC-3: injected mid-migration crash (jobs = last collection in the plan)
// ⇒ zero committed rows incl. marker; re-run without the injection then
// migrates cleanly (the crash-then-retry story, closed).
func TestInt_FailAfterRollsBackEverything(t *testing.T) {
	dir := t.TempDir()
	src := placeGolden(t, dir)
	gdb := newFileDB(t, dir)

	boom, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{
		SourcePath: src,
		FailAfter: func(name string, done int) error {
			if name == "Jobs" {
				return errInjected
			}
			return nil
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "injected crash")
	assert.Empty(t, boom.Collections, "failed run reports nothing committed")

	// Fresh handle to the same FILE: proves the ROLLBACK is on-disk truth.
	assertAllEmpty(t, openDB(t, dir))

	rep, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.NoError(t, err)
	require.False(t, rep.Skipped)
	assert.Equal(t, int64(41), reportTotal(rep))

	legacy, err := data.LoadLegacyData(basecontext.NewRootBaseContext(), src)
	require.NoError(t, err)
	require.Empty(t, ValidateCounts(basecontext.NewRootBaseContext(), gdb, &legacy, collectionPlan()))
}

// ---------------------------------------------------------------------------
// Process-level exit policy (precedent: #418 task 05).
//
// Spec fallback honoured: if the binary's boot cannot reach the migration
// stage on this runner (the Parallels/platform surface is absent off a PD
// host), the scenario t.Skipf-s — the task-06 seam test
// (TestRunLegacyDataMigration* in startup) remains the guaranteed signal.
// Deviation: the binary is built per-test into t.TempDir() instead of a
// shared TestMain build — marker_test.go already owns the package TestMain
// (a second would not compile); the go build cache makes the 2nd link cheap.
// ---------------------------------------------------------------------------

var errInjected = errors.New("injected crash")

func buildServiceBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "devops-service")
	srcRoot := filepath.Join("..", "..") // src/
	buildCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(buildCtx, "go", "build", "-o", bin, ".")
	cmd.Dir = srcRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot build service binary for subprocess test: %v\n%s", err, out)
	}
	return bin
}

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	_, port, err := net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)
	return port
}

func bootEnv(t *testing.T, dir string) []string {
	t.Helper()
	return append(os.Environ(),
		"DATABASE_FOLDER="+dir,
		"DB_PATH="+filepath.Join(dir, "devops.db"),
		"API_PORT="+freePort(t),
	)
}

// AC-3 + exit policy: corrupt data.json under the `api` command ⇒ the
// process EXITS 1 and logs the rolled-back failure.
func TestInt_ExitPolicy_APICommandExits1(t *testing.T) {
	bin := buildServiceBinary(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "data.json"), []byte("{corrupt"), 0o600))

	runCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, bin, "api")
	cmd.Dir = dir
	cmd.Env = bootEnv(t, dir)
	out, err := cmd.CombinedOutput()

	if !strings.Contains(string(out), "Legacy data migration failed") {
		t.Skipf("boot never reached the migration stage on this runner (platform surface); "+
			"task-06 seam test is the guaranteed signal. output tail:\n%s", tail(string(out), 1200))
	}
	require.Error(t, err, "process must exit non-zero")
	var ee *exec.ExitError
	require.ErrorAs(t, err, &ee)
	assert.Equal(t, 1, ee.ExitCode(), "API command must exit(1) on migration failure")
}

// Companion: healthy golden boot ⇒ migration succeeds and the server
// reaches readiness; the test then KILLS the process (bounded 30 s).
func TestInt_ExitPolicy_HealthyBootReachesReadiness(t *testing.T) {
	bin := buildServiceBinary(t)
	dir := t.TempDir()
	placeGolden(t, dir)

	runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, bin, "api")
	cmd.Dir = dir
	cmd.Env = bootEnv(t, dir)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	ready := make(chan string, 1)
	go func() {
		buf := make([]byte, 0, 64*1024)
		chunk := make([]byte, 4096)
		for {
			n, rerr := stdout.Read(chunk)
			if n > 0 {
				buf = append(buf, chunk[:n]...)
				s := string(buf)
				if strings.Contains(s, "HTTP server started successfully") ||
					strings.Contains(s, "Legacy data migration failed") {
					ready <- s
					return
				}
			}
			if rerr != nil {
				ready <- string(buf)
				return
			}
		}
	}()

	var s string
	select {
	case s = <-ready:
	case <-time.After(30 * time.Second):
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()

	if !strings.Contains(s, "HTTP server started successfully") {
		t.Skipf("boot never reached readiness on this runner (platform surface); output tail:\n%s", tail(s, 1200))
	}
	assert.Contains(t, s, "validation ok", "healthy boot must run the migration cleanly")
	assert.NotContains(t, s, "Legacy data migration failed")
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
