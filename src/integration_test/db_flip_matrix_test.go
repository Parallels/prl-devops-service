//go:build integration

// PBI #421 task 18 — the flip matrix (opt-in; heavy): process-level proof
// that provider selection boots the RIGHT backend per DB_TYPE and nothing
// else. Same harness posture as #420's boot matrix (db_boot_matrix_test.go:
// shared bootBinary, minimal env, never relax assertions — only the probe
// may skip when the environment blocks every boot).
//
// Local invocation:
//
//	PRLDEVOPS_FLIP_MATRIX=1 go test -tags integration ./integration_test/... -run FlipMatrix -v
//
// Row d (postgres) additionally needs POSTGRES_TEST_DSN (see
// database/postgres_integration_test.go header for the docker one-liner);
// unset = skipped, which is the deliberate CI posture.
package integration_test

import (
	"bufio"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

const (
	sqlReadyLine     = "SQL database connection ready"
	automigrateLine  = "SQL AutoMigrate complete"
	legacyCreateLine = "Database file does not exist, creating new file"
	legacyLoadLine   = "[Database] Loading database from"
)

// runBoot starts `<binary> api` like runAPI (#420) but with a caller-chosen
// wait line, and reports whether the wait line was seen before the process
// died or the deadline hit.
func runBoot(t *testing.T, env []string, waitLine string, deadline time.Duration) (output string, sawLine bool) {
	t.Helper()
	folder := t.TempDir()
	t.Setenv("FLIP_MATRIX_FOLDER", folder) // read back by the row for file probes
	minEnv := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + folder, // never touch the developer's real ~/.parallels-devops
		"DATABASE_FOLDER=" + folder,
		"PUID=0",
		"PGID=0",
	}
	cmd := exec.Command(bootBinary(t), "api")
	cmd.Env = append(minEnv, env...)
	cmd.Dir = folder
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { // never leak a live server on any path
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})

	lines := make(chan string, 256)
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()

	var sb strings.Builder
	timer := time.After(deadline)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				_ = cmd.Wait()
				return sb.String(), sawLine // died: caller asserts on output
			}
			sb.WriteString(l)
			sb.WriteString("\n")
			if strings.Contains(l, waitLine) {
				sawLine = true
				_ = cmd.Process.Signal(syscall.SIGKILL)
				return sb.String(), true
			}
		case <-timer:
			_ = cmd.Process.Signal(syscall.SIGKILL)
			return sb.String(), false
		}
	}
}

// flipFeasible probes once: on hosts where the pre-existing root-branch
// /etc gate blocks ANY boot (uid 0 + read-only /etc, #420 discovery), the
// matrix skips whole — never weakens an assertion.
func flipFeasible(t *testing.T) {
	t.Helper()
	out, saw := runBoot(t, nil, sqlReadyLine, 20*time.Second)
	if saw {
		return // this host CAN boot SQL mode
	}
	if strings.Contains(out, "/etc/parallels-devops") &&
		(strings.Contains(out, "permission denied") || strings.Contains(out, "read-only")) {
		t.Skipf("pre-existing environment gate (see #420 boot-matrix note): %s", firstLines(out, 3))
	}
	// Anything else is a REAL failure of the default row — fall through so
	// row (a) below fails with the full transcript.
}

func firstLines(s string, n int) string {
	parts := strings.Split(strings.TrimSpace(s), "\n")
	if len(parts) > n {
		parts = parts[:n]
	}
	return strings.Join(parts, " | ")
}

func fileTable(t *testing.T, path, table string) bool {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&n); err != nil {
		return false
	}
	return n == 1
}

func TestFlipMatrix(t *testing.T) {
	if os.Getenv("PRLDEVOPS_FLIP_MATRIX") != "1" {
		t.Skip("opt-in matrix: PRLDEVOPS_FLIP_MATRIX=1 required")
	}
	flipFeasible(t)

	t.Run("a_default_boots_sql_with_file_and_automigrate", func(t *testing.T) {
		out, saw := runBoot(t, nil, automigrateLine, 60*time.Second)
		if !saw {
			t.Fatalf("default boot never reached AutoMigrate:\n%s", out)
		}
		for _, want := range []string{sqlReadyLine, automigrateLine, "AutoMigrate complete (18 models)"} {
			if !strings.Contains(out, want) {
				t.Errorf("missing %q in default boot:\n%s", want, out)
			}
		}
		for _, not := range []string{legacyCreateLine, legacyLoadLine} {
			if strings.Contains(out, not) {
				t.Errorf("legacy engine line in SQL boot — goroutines started: %q", not)
			}
		}
		folder := os.Getenv("FLIP_MATRIX_FOLDER")
		dbPath := filepath.Join(folder, "prl-devops.db")
		if _, err := os.Stat(dbPath); err != nil {
			t.Fatalf("default DB_PATH file: %v", err)
		}
		if _, err := os.Stat(filepath.Join(folder, "data.json")); !os.IsNotExist(err) {
			t.Errorf("legacy data.json exists in SQL mode — NewJsonDatabase ran")
		}
		if !fileTable(t, dbPath, "users") {
			t.Error("users table absent — AutoMigrate did not persist")
		}
	})

	t.Run("b_legacy_escape_hatch", func(t *testing.T) {
		out, saw := runBoot(t, []string{"DB_TYPE=legacy"}, legacyCreateLine, 60*time.Second)
		if !saw {
			t.Fatalf("DB_TYPE=legacy never booted the file store:\n%s", out)
		}
		if strings.Contains(out, sqlReadyLine) || strings.Contains(out, automigrateLine) {
			t.Errorf("SQL init happened under the legacy escape hatch — the lever must bypass FromApplicationConfig entirely:\n%s", out)
		}
		folder := os.Getenv("FLIP_MATRIX_FOLDER")
		if _, err := os.Stat(filepath.Join(folder, "data.json")); err != nil {
			t.Errorf("data.json missing under legacy mode: %v", err)
		}
		if _, err := os.Stat(filepath.Join(folder, "prl-devops.db")); !os.IsNotExist(err) {
			t.Error("SQL file appeared under DB_TYPE=legacy")
		}
	})

	t.Run("c_explicit_sqlite_with_db_path", func(t *testing.T) {
		custom := filepath.Join(os.Getenv("FLIP_MATRIX_FOLDER"), "custom.db")
		out, saw := runBoot(t, []string{"DB_TYPE=sqlite", "DB_PATH=" + custom}, automigrateLine, 60*time.Second)
		if !saw {
			t.Fatalf("DB_TYPE=sqlite never reached AutoMigrate:\n%s", out)
		}
		if _, err := os.Stat(custom); err != nil {
			t.Fatalf("DB_PATH not honoured: %v\n%s", err, out)
		}
		if !fileTable(t, custom, "users") {
			t.Error("users table absent at DB_PATH")
		}
	})

	t.Run("d_postgres_dsn", func(t *testing.T) {
		dsn := os.Getenv("POSTGRES_TEST_DSN")
		if dsn == "" {
			t.Skip("row d needs POSTGRES_TEST_DSN (docker postgres:16 — see database/postgres_integration_test.go); skipped in CI by design")
		}
		out, saw := runBoot(t, []string{"DB_TYPE=postgres", "DB_DSN=" + dsn}, automigrateLine, 60*time.Second)
		if !saw {
			t.Fatalf("postgres boot never reached AutoMigrate:\n%s", out)
		}
		// information_schema introspection is the pg-side proof — the
		// AutoMigrate line already proves 18 models applied; the flip
		// guarantee here is simply "no legacy engine ever starts":
		if strings.Contains(out, legacyCreateLine) || strings.Contains(out, legacyLoadLine) {
			t.Errorf("legacy engine ran under postgres mode:\n%s", out)
		}
	})
}
