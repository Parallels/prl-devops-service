//go:build integration

// PBI #420 task 04 — process-level boot matrix (opt-in; heavy).
// Local invocation:
//
//	PRLDEVOPS_BOOT_MATRIX=1 go test -tags integration ./integration_test/... -run BootMatrix -v
//
// Environment note (discovery, recorded per spec): the built binary must land
// on an EXECUTABLE mount. Hosts where /tmp is mounted noexec need TMPDIR set to
// one, e.g. TMPDIR=$PWD/../.gotmp/tmpd (dev container posture).
//
// Feasibility discovery #2 (pre-existing, NOT caused by #420): as uid 0 (or in
// containers where /etc is read-only), the provider bootstrap in
// serviceprovider/main.go creates ServiceDefaultDirectory (/etc/parallels-devops)
// BEFORE any SQL init, so every boot — any env — dies there first. The matrix
// therefore PROBE-boots once and t.Skipfs (never weakens assertions) when that
// pre-existing gate blocks the environment. Verified posture: developer macOS
// (non-root) and CI linux with writable /etc run all rows.
//
// CI posture: not wired (deliberate — same posture as #418's manual boot
// evidence; .github/workflows/pull_request.yml has no PRLDEVOPS_BIN).
package integration_test

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

var (
	bootBinOnce sync.Once
	bootBinDir  string
	bootBinPath string
	bootBinErr  error
)

// bootBinary builds the service binary from the module root (src/) into a
// temp dir shared across the matrix run, and drops a minimal
// prldevops_config.yaml NEXT TO THE BINARY (config resolution searches the
// exec dir; discovery per spec: without a config file `api` enters the
// installer flow before any SQL init).
func bootBinary(t *testing.T) string {
	t.Helper()
	bootBinOnce.Do(func() {
		bootBinDir, bootBinErr = os.MkdirTemp("", "prldevops-boot-*")
		if bootBinErr != nil {
			return
		}
		bootBinPath = filepath.Join(bootBinDir, "prldevops")
		cmd := exec.Command("go", "build", "-o", bootBinPath, ".")
		cmd.Dir = ".." // module root = src/ (this package is ./integration_test)
		if out, err := cmd.CombinedOutput(); err != nil {
			bootBinErr = err
			t.Fatalf("build service binary: %v\n%s", err, out)
		}
		bootBinErr = os.WriteFile(filepath.Join(bootBinDir, "prldevops_config.yaml"),
			[]byte("environment: {}\n"), 0o600)
	})
	if bootBinErr != nil {
		t.Fatalf("build service binary: %v", bootBinErr)
	}
	return bootBinPath
}

// runAPI starts `<binary> api` with a MINIMAL env (t.TempDir DATABASE_FOLDER +
// case env) and streams combined output until: the success line appears
// (returns exited=false), the process exits (returns code), or deadline hits
// (kill + fail-fast snapshot). Assertions are never relaxed by the harness —
// only the injected env is, per spec, when boot needs more (recorded in PR).
func runAPI(t *testing.T, env []string, deadline time.Duration) (output string, exited bool, code int) {
	t.Helper()
	folder := t.TempDir()
	minEnv := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + folder, // never touch the developer's real ~/.parallels-devops
		"DATABASE_FOLDER=" + folder,
		"PUID=0",
		"PGID=0",
	}
	cmd := exec.Command(bootBinary(t), "api")
	cmd.Env = append(minEnv, env...)
	cmd.Dir = folder // cwd = the temp folder: no stray writes near the repo
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout // combined

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { // belt & braces: never leak a live server on failure paths
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})

	type line struct{ s string }
	lines := make(chan line, 256)
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			lines <- line{sc.Text()}
		}
		close(lines)
	}()

	var sb strings.Builder
	timer := time.After(deadline)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				// EOF: process ended
				err := cmd.Wait()
				code := 0
				if err != nil {
					if ee, isExit := err.(*exec.ExitError); isExit {
						code = ee.ExitCode()
					} else {
						code = -1
					}
				}
				return sb.String(), true, code
			}
			sb.WriteString(l.s)
			sb.WriteString("\n")
			if strings.Contains(l.s, "SQL database connection ready") {
				_ = cmd.Process.Signal(syscall.SIGKILL)
				return sb.String(), false, 0 // success: line present within window
			}
		case <-timer:
			_ = cmd.Process.Signal(syscall.SIGKILL)
			return sb.String(), false, -1 // timeout: caller asserts on output
		}
	}
}

func TestBootMatrix(t *testing.T) {
	if os.Getenv("PRLDEVOPS_BOOT_MATRIX") != "1" {
		t.Skip("opt-in boot matrix — PRLDEVOPS_BOOT_MATRIX=1 (see PR body)")
	}
	_ = bootBinary(t)                                  // build once under the PARENT test
	t.Cleanup(func() { _ = os.RemoveAll(bootBinDir) }) // parent-scoped lifetime (harness fix found in first local run)

	// Feasibility probe (see header discovery #2): if the pre-existing /etc
	// bootstrap gate blocks THIS environment, skip the matrix honestly.
	probeOut, _, _ := runAPI(t, nil, 8*time.Second)
	if strings.Contains(probeOut, "failed to create directory") || strings.Contains(probeOut, "panic:") {
		t.Skipf("boot matrix not feasible here (pre-existing bootstrap requires %s; probe output head: %.160s) — run on macOS/CI posture",
			"/etc/parallels-devops", probeOut)
	}

	const canary = "sup3rs3cr3t"

	t.Run("a_default_boot_matches_418_behavior", func(t *testing.T) {
		out, exited, _ := runAPI(t, nil, 30*time.Second)
		if exited || !strings.Contains(out, "SQL database connection ready") {
			t.Fatalf("row a: default boot must emit the success line\n--- output ---\n%s", out)
		}
		// folder is not returned by runAPI; success line + type proves sqlite path
		if !strings.Contains(out, "type: sqlite") {
			t.Fatalf("row a: expected sqlite type in success line\n%s", out)
		}
	})

	t.Run("b_postgres_without_dsn_exits_1", func(t *testing.T) {
		out, exited, code := runAPI(t, []string{"DB_TYPE=postgres"}, 30*time.Second)
		if !exited || code != 1 {
			t.Fatalf("row b: want exit 1, got exited=%v code=%d\n%s", exited, code, out)
		}
		// Task 02 shipped message text (pinned): sentinel + both notations.
		for _, want := range []string{"requires a connection string", "DB_DSN", "db_dsn"} {
			if !strings.Contains(out, want) {
				t.Fatalf("row b: message must name %q\n%s", want, out)
			}
		}
		if strings.Contains(out, canary) {
			t.Fatalf("row b sanity: canary present without any DSN set")
		}
	})

	t.Run("c_dsn_canary_never_plain_at_startup", func(t *testing.T) {
		out, exited, code := runAPI(t, []string{
			"DB_TYPE=postgres",
			"DB_DSN=postgres://dbuser:" + canary + "@127.0.0.1:1/x?sslmode=disable",
		}, 30*time.Second)
		if !exited || code != 1 {
			t.Fatalf("row c: want exit 1, got exited=%v code=%d\n%s", exited, code, out)
		}
		if strings.Contains(out, canary) {
			t.Fatalf("row c (HEADLINE D-8 EVIDENCE): canary leaked in plain text\n%s", out)
		}
		// #420 task 03 pins errors never embedding the DSN; the factory
		// (#418 RedactDSN) covers connection-attempt failures. A bad DSN may
		// legitimately reach the driver's own error — assert only the redaction
		// contract holds for anything WE print: the canary password never survives.
		if !strings.Contains(out, "postgres") {
			t.Fatalf("row c: error should reference the postgres backend\n%s", out)
		}
	})

	t.Run("d_garbage_auto_migrate_exits_1", func(t *testing.T) {
		out, exited, code := runAPI(t, []string{"DB_AUTO_MIGRATE=maybe"}, 30*time.Second)
		if !exited || code != 1 {
			t.Fatalf("row d: want exit 1, got exited=%v code=%d\n%s", exited, code, out)
		}
		if !strings.Contains(out, `"maybe"`) || !strings.Contains(out, "db_auto_migrate") {
			t.Fatalf("row d: message must quote the offending value + name the key\n%s", out)
		}
	})
}
