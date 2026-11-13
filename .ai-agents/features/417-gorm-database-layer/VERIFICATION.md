# Slice verification ledger — feature #417 (`sql-migration`)

One entry per slice (task). The gate (ER-7): entry written by the implementer/lead → **presented to the
human** → human runs the test plan / reviews results → **verdict recorded here** → next slice may start.

Verdict vocabulary: `APPROVED` · `APPROVED WITH FOLLOW-UPS (list)` · `REJECTED (rework notes — slice stays
open in its worktree)` · `PENDING HUMAN VALIDATION (current state while gate is open)`.

Rules (§14): slices are serial by default; `sql-migration` must never receive a red merge (ER-5); push of
`sql-migration` only when the human asks.

---

## Slice index

| # | Slice | Branch | Verdict | Date |
|---|---|---|---|---|
| 1 | `418-01` add GORM deps | `slice/418-01` | **APPROVED** — merged `c96c4b5` | 2026-11-10 |
| 2 | `418-02` config contract | `slice/418-02` | **APPROVED** — merged `be3ca50` | 2026-11-10 |
| 3 | `418-03` SQLite factory | `slice/418-03` | **APPROVED** — merged `ef6fba1` | 2026-11-10 |
| 4 | `418-04` postgres + pool + redaction | `slice/418-04` | **PENDING HUMAN VALIDATION** | 2026-11-10 |

---

## Entry template (copy per slice)

```markdown
## slice/<pbi>-<nn> — <task title>            PENDING HUMAN VALIDATION

**Worktree**: `.worktrees/slice-<pbi>-<nn>` · **Base**: sql-migration @ <sha> · **Commits**: <sha..sha>
**Diff**: <git diff --stat summary>

### Self-gate results (implementer)
- [x] every `## Validation` checkbox of the task spec — <paste evidence lines>
- [x] `cd src && go build ./... && go test ./... -count=1` — <tail line>

### Test plan for the human
1. <exact command / manual step> → expect: <observable outcome>
2. ...
   (derive from the task's Validation section + one end-to-end smoke where the slice is user-observable)

### Observed results
<implementer's own runs of the same plan, output snippets>

### Deviations / open-question changes
<anything shipped differently than spec'd, or OQ moved — "none" otherwise>

### Verdict
<human fills: APPROVED / APPROVED WITH FOLLOW-UPS / REJECTED + notes>  · <date>
```

---

## slice/418-01 — Add GORM dependencies (pure manifest PR)        APPROVED → merged `c96c4b5`

**Worktree**: `.worktrees/slice-418-01` · **Base**: sql-migration @ `cc7d3d4` · **Commit**: `fd2bbc3`
**Diff**: `src/database/doc.go` +16 (new) · `src/go.mod` +16 · `src/go.sum` +40 — 3 files, +72, 0 deletions
**Versions pinned**: gorm.io/gorm **v1.31.2** · github.com/glebarez/sqlite **v1.11.0** ·
gorm.io/driver/postgres **v1.6.3** — zero churn to any pre-existing require; `go.work.sum` untouched.

### Self-gate results (implementer — all six spec checks)
- `go build ./...` GREEN · `go vet ./database/...` GREEN · `gofmt -l database/` empty
- `go list -m` lists all three with the versions above
- `go test ./... -count=1`: baseline (pristine HEAD) 33 pkgs / 0 FAIL → post-change pass/fail set
  identical, **modulo one pre-existing flake** (below)
- `CGO_ENABLED=0 go build ./...` and `./database/...` GREEN (AD-2 proven)

### Pre-existing flake — first Tier-1 ledger entry for #425
`orchestrator.TestHostSelectionPreservesResourceFailuresInJobs`: 3 intermittent failures observed
across suite/count runs **with the slice applied**; 0/50 on `-count=50` stress; **slice exonerated
mechanically** — `go test -c ./orchestrator/` binary is sha256-identical pristine vs slice
(`2361d1e5…`). File to #425's task-07 flake sweep; not this PR's debt.

### Test plan for the human (native machine, ~2 min)
```bash
cd .worktrees/slice-418-01/src
go build ./... && echo GREEN                          # expect GREEN
go vet ./database/...                                 # expect silent
go list -m gorm.io/gorm github.com/glebarez/sqlite gorm.io/driver/postgres
# expect: gorm.io/gorm v1.31.2 · github.com/glebarez/sqlite v1.11.0 · gorm.io/driver/postgres v1.6.3
CGO_ENABLED=0 go build -o /tmp/prl-418-01 . && echo "cgo-off binary OK"   # real service binary, CGO off
go test ./... -count=1                                # expect 33 ok / 0 FAIL (~1-2 min; if the named
                                                      # orchestrator flake trips, rerun — see ledger note)
git show --stat HEAD                                  # expect exactly 3 files, +72
```
Runtime observation: none expected — nothing imports `database` yet; that is the point of the slice.

### Deviations / spec adjustments
1. **Spec order fixed** (task file edited): `doc.go` must precede `go mod tidy`, else tidy prunes the
   just-`go get`-ed deps — the first tidy silently reverted them; re-ran in corrected order.
2. **AD-2 issue comment (spec step 4) awaits this gate** — external side effect. Draft, ready to post
   on `Parallels/prl-devops-service#418` on your go:
   > **Deviation note (AD-2, feature #417):** the SQLite dialect ships as `github.com/glebarez/sqlite`
   > (pure-Go, modernc.org/sqlite engine) rather than the CGO `gorm.io/driver/sqlite` named in the AC.
   > Rationale: the service ships cross-compiled, CGO-free release binaries; the pure-Go dialect keeps
   > `CGO_ENABLED=0` builds green (proven in slice 418-01, commit `fd2bbc3`). The AC's intent — "SQLite
   > via GORM" — is fully met; only the driver package differs.
3. Environment note (agent container only): Go commands need the harness `GIT_CONFIG_*` vars unset
   (breaks buildvcs' git subprocess) and `GOTMPDIR` in-worktree (host `/tmp` is noexec). Irrelevant on
   your native shell.

### Verdict
**APPROVED** — human ran the test plan: "all passed correctly" (2026-11-10, chat).
Merged `--no-ff` into `sql-migration` as **`c96c4b5`**; worktree + branch removed.
*(Still pending on explicit go: posting the AD-2 deviation comment on issue #418.)*

---

## slice/418-02 — `DatabaseConfig` contract, AD-3 defaults, validation   PENDING HUMAN VALIDATION

**Worktree**: `.worktrees/slice-418-02` · **Base**: sql-migration @ `c96c4b5` · **Commit**: `0277f71`
**Diff**: `src/database/config.go` +136 · `config_test.go` +108 · `errors.go` +10 — 3 new files, +254,
zero deletions, zero touches outside `src/database/`.

### Self-gate results (all five spec checks)
- `go build ./...` GREEN · `go vet ./database/...` GREEN · `gofmt -l database/` empty
- `go test ./database/... -count=1 -v`: full spec matrix green (7 Validate cases, pool both branches,
  seam equality, path branches). One path-branch test SKIPs by uid (root vs non-root) — between my
  environment and yours **both** branches execute.
- Seam documented: `config.go:135` "MUST return DefaultDatabaseConfig()" grep-verified
- ER-5 full suite: package set identical to baseline + `database` line, **0 real FAILs** —
  orchestrator flake re-tripped (below)

### Flake ledger — same suspect, new data point
`TestHostSelectionPreservesResourceFailuresInJobs`: during this slice's validation it failed **6/6** in a
back-to-back loop (was intermittent). Still **not slice-related**, harder proof this time: exact
`go list -deps ./orchestrator/` filter for `prl-devops-service/database|gorm|glebarez` is **empty**, and
slice 02 changed **no module deps** → orchestrator's compiled test binary is unchanged by construction
(matches 418-01's sha256-identity proof). Failing-rate depends on machine state, not our code —
strongly feeding #425 task 07's flake sweep (suspect timing/order dependency in that test).

### Test plan for the human (~2 min, non-root shell)
```bash
cd .worktrees/slice-418-02/src
go build ./... && echo GREEN
go test ./database/... -count=1 -v    # expect all PASS; RootUsesEtc SKIPs for you, Home branch RUNS
go vet ./database/...                 # expect silent
git show --stat HEAD                  # expect exactly 3 files, +254 insertions
grep -n "MUST return" database/config.go   # expect line 135-ish comment (#420 seam declared)
# optional, mirrors the AD-3 default in your environment:
DATABASE_FOLDER=/tmp/x go test ./database/... -run UsesDatabaseFolder -v   # PASS
```
Runtime observation: none — still no consumers; `database` package is imported by nothing but its tests.

### Deviations / additions beyond the spec matrix (additive only)
- Added: root-branch path test (uid-guarded, complements the HOME case), `"postgres ok"` and
  `sqlite MemoryPath` happy-path Validate cases, seam-delegation test. No spec case removed or weakened.
- Flake note above (not a deviation of code, of environment).

### Verdict
_(yours: APPROVED / APPROVED WITH FOLLOW-UPS / REJECTED + notes)_ · ____
## slice/418-02 verdict applied
**APPROVED** — human: "approved, all passed" (2026-11-10). Merged `--no-ff` as **`be3ca50`**; worktree
+ branch removed.

---

## slice/418-03 — SQLite connection factory + WAL pragmas                 PENDING HUMAN VALIDATION

**Worktree**: `.worktrees/slice-418-03` · **Base**: sql-migration @ `be3ca50` · **Commit**: `8a93f8f`
**Diff**: 5 files, +339 (4 new: `connection.go` 60 · `sqlite.go` 92 · `sqlite_test.go` 149 ·
`connection_test.go` 34; `errors.go` +4 sentinel `ErrPostgresNotWiredYet`)

### Self-gate results (all five spec checks)
- `go build ./...` GREEN · `go vet ./database/...` GREEN · `gofmt -l` empty
- `go test ./database/... -count=1 -race` **ok 1.117s** — includes the 20-goroutine concurrent-insert
  guard on a file DB (WAL + busy_timeout), race-clean
- Banner check: 0 GORM noise lines in `-v` output (Silent config proven)
- Markers present: `connection.go:14/25` ErrPostgresNotWiredYet · `:52` TODO(#418 task 05)
- ER-5 full suite: only failure = the known flake (below); re-run of `./orchestrator/` alone immediately
  after = **ok** — intermittent, unchanged suspect

### Flake ledger — third sighting, same test
`TestHostSelectionPreservesResourceFailuresInJobs` tripped the 418-03 full-suite run. Same test name;
`go list -deps ./orchestrator/` exact-filter for `database|gorm|glebarez` = **0** for this slice too
(unchanged by construction: slice touches only `src/database/`). Immediate re-run: ok.

### Test plan for the human (~3 min)
```bash
cd .worktrees/slice-418-03/src
go build ./... && echo GREEN
go test ./database/... -count=1 -race -v     # all PASS incl. TestNewDatabaseConnection_SQLite_ConcurrentInserts
ls -la /tmp/manual-check.db* 2>/dev/null      # (nothing yet — see next line for the manual smoke)
cat > /tmp/wal_check.go <<'GO'
package main
import ( "fmt"; "github.com/Parallels/prl-devops-service/database" )
func main() {
  ctx := basecontext.NewBaseContext()
  db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{Type:"sqlite", Path:"/tmp/manual-check.db"})
  if err != nil { panic(err) }
  var mode string; db.Raw("PRAGMA journal_mode").Scan(&mode)
  fmt.Println("journal_mode =", mode)   // expect: wal
  database.Close(ctx, db)
}
GO
# (simpler equivalent — a single test already asserts this: TestNewDatabaseConnection_SQLite_Pragmas)
go test ./database/... -run 'Pragmas|Memory|PathWithSpaces' -count=1 -v   # PASS ×3
# observe the artifacts your run creates:
ls /var/folders/**/TestNewDatabaseConnection* 2>/dev/null || true   # t.TempDir trees; *.db + *.db-wal + *.db-shm = WAL live
git show --stat HEAD                        # expect 5 files, +339
```
Runtime observation worth making: while a test writes the file DB, `-wal`/`-shm` siblings appear next to
the `.db` — visible proof WAL is really on (not just reported).

### Deviations
- None from spec. `redactPath` shipped as documented passthrough (D-8 placeholder);
  `ErrPostgresNotWiredYet` carries code 501 (Not Implemented — the interim state's true semantics).

### Verdict
_(yours: APPROVED / APPROVED WITH FOLLOW-UPS / REJECTED + notes)_ · ____
## slice/418-03 verdict applied
**APPROVED** — human: "approved, lets move next" (2026-11-10). Merged `--no-ff` as **`ef6fba1`**; worktree
+ branch removed. (Process note: branch-delete ran before worktree-teardown — stale branch deleted after,
sequence corrected for later slices: remove worktree first, then `branch -D`.)

---

## slice/418-04 — postgres branch, connection pool, DSN redaction    PENDING HUMAN VALIDATION

**Worktree**: `.worktrees/slice-418-04` · **Base**: sql-migration @ `ef6fba1` · **Commit**: `f4fdead`
**Diff**: 8 files, +398/−11 (`postgres.go` 63 · `redact.go` 65 · `postgres_test.go` 68 ·
`redact_test.go` 81 · `postgres_integration_test.go` 111 · connection.go/test + sentinel edits)

### Self-gate results
- build GREEN · `gofmt -l` 0 · `go vet` + `go vet -tags integration` GREEN
- `go test ./database/ -count=1` **ok 0.086s** — offline (unroutable port 1 = instant refusal; total
  pkg wall 1.2s < 1s-per-spec-subset bar met; zero network attempted)
- Integration suite: compiles under tag, **3/3 SKIP** cleanly with `POSTGRES_TEST_DSN` unset
  (agent env has no docker — the real-server leg is YOUR plan item B below, optional at this gate)
- Sentinel audit: `grep -rn ErrPostgresNotWiredYet src/database/` → **0 hits** (incl. comments/tests)
- DSN audit: every non-test `cfg.DSN` use is either `postgres.Open(cfg.DSN)` (driver) or routed
  through `RedactDSN` — pasted verbatim in the slice commit session, reproduced in test plan below
- ER-5 full suite: 33 ok + known flake (4th sighting, same test, solo re-run pattern unchanged)

### Deviations (spec reality-fixes — each proven by test)
1. **`gorm.io/driver/postgres` v1.6 dials during `gorm.Open`** — spec comment said "does NOT dial —
   lazy". Evidence: unreachable DSN produced the Open-failure branch with pgx dial errors. Fix: that
   branch wraps `ErrDatabaseUnreachable` too (boot-failure semantics), Ping kept as engine-agnostic gate.
2. **`url.UserPassword` escapes `***` to `%2A%2A%2A`** — RedactDSN now splices the URL manually, so
   non-credential bytes survive byte-exact (`p%40ss` → gone, `u:***` literal, params untouched).
3. **applyPool offline proof uses memory-sqlite handle** — because postgres.Open now dials (1), the
   pure-helper path (spec's "preferred" option) is proven through the engine-agnostic database/sql
   setters: `Stats().MaxOpenConnections` == 25 default / 7 custom, no server.
4. **`connection_test.go` retired the interim test** — spec's file table omitted it; sentinel removal
   (its own Validation item) made that file red otherwise.
5. **Integration unreachable variant = wrong password, not bad search_path** — search_path is a plain
   session GUC that never fails Ping; wrong-password is the faithful unreachable-at-boot shape on a
   live server. Documented in-file.

### Test plan for the human (~3 min + optional 5 min)
**A. Offline (gate-blocking):**
```bash
cd .worktrees/slice-418-04/src
go build ./... && echo GREEN
go test ./database/... -count=1 -v        # all PASS, fast (<2s, no network) — watch RedactDSN_Matrix 10 cases
go vet -tags integration ./database/...   # integration file compiles
go test -tags integration ./database/... -count=1   # 3× SKIP (no POSTGRES_TEST_DSN), ok
grep -rn "ErrPostgresNotWiredYet" database/ ; echo "exit=$? (expect 1 = gone)"
grep -rn "cfg.DSN" database/*.go | grep -v _test   # expect: 3 lines, all Open() or RedactDSN
git show --stat HEAD                                  # 8 files, +398/−11
```
**B. Integration (optional today, canonical before #425):**
```bash
docker run --rm -d --name prl-pg418 -e POSTGRES_PASSWORD=*** -e POSTGRES_DB=prl_test -p 5432:5432 postgres:16
POSTGRES_TEST_DSN='postgres://postgres:prl@127.0.0.1:5432/prl_test?sslmode=disable' \
  go test -tags integration ./database/... -count=1 -v   # 3 PASS + 9 sqlite PASS
docker rm -f prl-pg418
```

### Verdict
_(yours: APPROVED / APPROVED WITH FOLLOW-UPS / REJECTED + notes)_ · ____
