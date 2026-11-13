# Task 15 — Jobs on GORM (+ stale-job watcher function, timeout handler)

**PBI** [#421](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 06 · **PR size:** M
**Plan refs:** D-8 (stale-job relocation, always-on proposal OQ-421-4) · D-9 (no tx row listed for jobs — see §4) · D-11 · D-12 (`Job.Steps` serialised)

## Context

Legacy `src/data/jobs.go` (254 lines, untouched) is 10 methods — but it also owns the feature's *behavioural*
quirk: `DetectStaleJobs` :216 and `RecoverOngoingJobs` :183 run inside the legacy engine's own goroutine
(`GhostJobCheckIntervalSeconds` cadence from `src/data/main.go`, started by `Connect`). D-8 already decided:
**SQL mode gets its own watcher** (same interval constant, SQL-side goroutine) and **legacy keeps its
goroutine only in legacy mode** — with start/stop wired in task 18. This task delivers the *functions* pure:
watcher body, timeout handler registration, and the 8 CRUD methods — no goroutine-start side effects in
`GormDatabase` constructors.

## Objective

Jobs CRUD + owner filtering + bulk-delete on GORM with parity; `DetectStaleJobs`/`RecoverOngoingJobs`
callable and tested against fixtures (time-travel via injected clock, not sleeps); timeout-handler
registration semantics preserved; **zero goroutines started by this PR**.

## Files

| File | Op |
|---|---|
| `src/database/jobs.go` | create (8 CRUD + Recover + SetJobTimeoutHandler) |
| `src/database/jobs_watcher.go` | create (`func (g *GormDatabase) runStaleJobWatcher(ctx) bool` + `Stop()` hook — invoked by 18 only) |
| `src/database/delegate_jobs.go` | delete |
| `src/database/jobs_parity_test.go` | create |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/data/jobs.go` | 254 (measured) | untouched | ok |
| `src/database/jobs.go` + `jobs_watcher.go` | new | ~140 + ~70 | ok |

## Implementation spec

**Inventory:** `GetJobs` :20 · `GetJobsByOwner` :37 · `GetJob` :56 · `CreateJob` :74 · `UpdateJob` :101 ·
`DeleteJob` :130 · `DeleteJobsByState` :158 · `RecoverOngoingJobs` :183 · `SetJobTimeoutHandler` :208 ·
`DetectStaleJobs` :216.

1. **CRUD.** `GetJobs` filter/order/index + authz wraps exactly as legacy applied (:20-37 — confirm wrap);
   `GetJobsByOwner` :37 — `WHERE user_id = ?` (the GetOwnRecords analogue **implementer to confirm whether
   :37 is the own-records filter or an explicit-arg method**); `CreateJob` :74 — Steps serialised column per
   #419/D-12; `UpdateJob` :101 — legacy read-modify-write whole job incl. Steps; `DeleteJob` :130;
   `DeleteJobsByState` :158 — bulk `WHERE state = ?` (returns count parity: legacy returns deleted slice —
   **confirm whether callers need the deleted records back (:158-183)**; if yes, find-then-delete-in-tx).
2. **`RecoverOngoingJobs` :183.** Legacy: at startup, marks jobs stuck in running-ish states as failed after
   unclean shutdown. SQL port: same state predicate — **implementer to confirm the state set (:183-208)**;
   `UPDATE ... SET state=?, <failure fields> WHERE state IN (...)` in one tx, returns/counts parity. Called
   from startup in task 18 (D-8/OQ-421-4 always-on posture).
3. **`SetJobTimeoutHandler` :208.** Registration of a callback held by the engine. GORM side: same
   registration surface on `GormDatabase` (field + setter, mutex-guarded); the *caller* that fires it is the
   watcher (below) and task 18's start-up wiring — keep the handler invocation semantics (args, once-per-job)
   **confirm from :208-216 + jobs_timeout_test.go legacy expectations**.
4. **`DetectStaleJobs` :216 + watcher.** Predicate (stale-by-`UpdatedAt`-plus-timeout per legacy — **confirm
   exact staleness rule (:216-254)**) ported as one query + per-stale-job handler invocation (timeout handler
   §3, then state update). `jobs_watcher.go` body: ticker on the legacy interval constant
   (`constants.GhostJobCheckIntervalSeconds` — reuse, don't duplicate), context-cancellable, `stopCh`-style
   hook mirroring #418's save-loop stop idiom, **started only by task 18**. Time control: package-private
   `nowFn func() time.Time` field defaulting to `time.Now` — tests inject a clock and assert stale-window
   edges without sleeps.
5. **Tx posture.** D-9 lists no jobs transaction — but §1's find-then-delete (if caller-needs-records) and
   §4's detect-then-mark are check-then-act pairs: **use one tx each; PR note explains the addition**
   (D-9 rule: grep `dataMutex.Lock()` in `jobs.go` before deciding — implementer records the grep result in
   PR body).

## Out of scope

Goroutine start/stop (18), job *production* logic (job engine in `src/jobs/` — call-site retypes in 17),
schema (`jobs` table + `steps` column per #419), provider flip.

## Validation

- [ ] `cd src && go build ./... && go vet ./database/...`
- [ ] `cd src && go test ./database/... -run Jobs -count=1 -race` — clock-injected stale-window edge tests
      (fresh-but-old, stale exactly-at-boundary), recovery state-set matrix
- [ ] Legacy `jobs_timeout_test.go` re-run green (still tests the legacy engine — ER-5: untouched)
- [ ] `cd src && go test ./... -count=1` (ER-5)
- [ ] `git diff --stat src/data/jobs.go src/data/main.go` → empty (no goroutine, no legacy touch)

## Rollback

Revert restores `delegate_jobs.go`; watcher function unused (18 owns start-up) so nothing dangles; no schema
outreach.

## PBI acceptance mapping

**CRUD via GORM** — 10 methods; **Transactions** — the two check-then-act pairs (recorded against D-9);
**Error handling** — D-11; **Legacy bypassed** — jobs domain fully off-delegate after this PR; **D-8
pre-work** — task 18 only wires what this proves.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-15`.** Cut `slice/421-15` from `sql-migration` HEAD
(`.worktrees/slice-421-15`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
