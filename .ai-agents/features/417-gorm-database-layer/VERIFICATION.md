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
| 1 | `418-01` add GORM deps | `slice/418-01` | **PENDING HUMAN VALIDATION** | 2026-11-10 |

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

## slice/418-01 — Add GORM dependencies (pure manifest PR)        PENDING HUMAN VALIDATION

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
_(yours: APPROVED / APPROVED WITH FOLLOW-UPS / REJECTED + notes)_ · ____
