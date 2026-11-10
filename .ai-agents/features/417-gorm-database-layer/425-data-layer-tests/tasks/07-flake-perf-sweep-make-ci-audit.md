# Task 07 — Flake/perf sweep + Makefile targets + the measurable budget evidence

**PBI** [#425](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** tasks 01–06 (the suite exists; this task certifies it) · **PR size:** S
**Plan refs:** [D-3](./plan.md) isolation rules · [D-8](./plan.md) (CI/Makefile audit **record — follow what the record says, it differs from the common assumption**; read it first) · [D-9](./plan.md) baseline + cost model · AC 6 + ≤+30 s constraint

## Context

The two "quality-of-life" ACs become provable here: **no flaky tests** (the hard definition, not vibes) and
**the +30 s budget** (measured, not argued — D-9's baseline: `./data` alone ≈ 30.4 s warm, whole-suite
number from 01's first run). Also the developer-loop closure: D-8's record of what the Makefile *actually*
has vs needs (`test`, `test-race`, `test-integration` per task index — the audit record decides add-vs-edit
per target; this file is 07's **only** repo change, keeping Risk-9's single-owner rule).

## Objective

Zero flakes under stress; PR body carries before/after wall-time evidence; `make test` / `make test-race` /
`make test-integration` (final names per D-8) work for a new developer from the README-less repo state;
D-8's audit record appended-to with anything new discovered.

## Files

| File | Op |
|---|---|
| `Makefile` | modify per D-8 record (targets for unit / race / integration; integration target sets the canonical docker one-liner hint + `POSTGRES_TEST_DSN` default) |
| *(no other repo file)* | CI edits were 05's; sweeps and evidence live in the PR description |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `Makefile` | *(measure at task start)* | +≤15 | ok |

## Implementation spec

1. **Flake sweep (definition of "no flaky tests" — gate):**
   ```bash
   cd src
   go test ./database/... -count=5 -shuffle=on -race     # new + ported surfaces
   go test ./data/...     -count=3 -shuffle=on           # Tier 1: informational (AC 6 = new tests);
                                                          # legacy failures are NOT this PBI's fixes —
                                                          # record in PR, feed #426 ledger (D-1/Risk-6)
   ```
   Any `./database/...` failure = fix root cause or quarantine file with `//go:build flaky_quarantine` +
   issue link (quarantine needs PR approval; count must be 0 at merge — "retried away" is forbidden,
   Risk-3 discipline).
2. **Wall-time evidence:** warm-cache, same machine: `time go test ./... -count=1` on the **merge-base**
   commit and on this branch — both outputs pasted verbatim; delta must be ≤ +30 s against the AC's
   reference point (the pre-feature suite — the PR explains the two-reference-point honesty: merge-base
   delta *and* pre-#417 tag number if reachable cheaply). If over: D-9's fallback ladder (a) shared
   schema-migration cache in `testutil` behind a flag, (b) per-wave file parallelism audit
   (`go test -json | jq` slowest list pasted), (c) escalate — do not silently land.
3. **Makefile (per D-8 record — exact shape if record shows gaps):**
   ```makefile
   test:            ## run unit tests (default CI set)
   	cd src && go test ./... -count=1
   test-race:       ## unit tests with race detector
   	cd src && go test ./... -count=1 -race
   test-integration:## Postgres integration suite (needs docker or POSTGRES_TEST_DSN)
   	cd src && POSTGRES_TEST_DSN=$${POSTGRES_TEST_DSN:-postgres://postgres:postgres@127.0.0.1:5432/prl_test?sslmode=disable} \
   	  go test -tags integration -count=1 ./database/...
   ```
   (Adapt to existing style/phoney rules; the docker one-liner cross-reference comment goes beside
   `test-integration`.)
4. **D-8 record append:** anything new found while wiring (e.g. coverage-step reaction observed in 05,
   CI's actual Go version vs `go.mod`) lands in the plan's D-8 table — plan edit is documentation-only,
   allowed for the owning task.

## Out of scope

Fixing legacy flakiness (ledger → #426), CI topology (05), testutil internals (single-owner rule; perf
fallback (a) is the only sanctioned testutil edit, flagged in PR if triggered).

## Validation

- [ ] `make test`, `make test-race`, `make test-integration` from clean clone state (documented invocation pasted)
- [ ] Gate 1 outputs pasted in PR: `./database/...` `-count=5 -shuffle=on -race` × **two consecutive runs**, zero failures
- [ ] Gate 2: wall-time before/after pasted, delta ≤ 30 s, or fallback ladder step executed
- [ ] `git diff --stat` → `Makefile` (+ this plan's D-8 appendix lines) and nothing else

## Rollback

Makefile-only revert; the evidence stays in the (closed) PR history; quarantine files, if any existed,
would be deleted by the same revert.

## PBI acceptance mapping

**No flaky tests** — proven by the two-run stress gate, definition enforced; **+30 s budget** — measured
evidence, architecture defence (D-6) and measurement (here) together close the AC; developer ergonomics
(`make test-*`) as a bonus the feature plan's §12 asked for; #426 receives the Tier-1 findings ledger.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `425-07`.** Cut `slice/425-07` from `sql-migration` HEAD
(`.worktrees/slice-425-07`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
