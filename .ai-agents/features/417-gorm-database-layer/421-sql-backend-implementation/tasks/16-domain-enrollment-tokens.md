# Task 16 — Enrollment tokens on GORM (TTL expiry)

**PBI** [#421](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 03 (independent leaf — plan index: parallelisable with 05/06) · **PR size:** S
**Plan refs:** D-8 (expiry sweep cadence posture) · D-11 · D-12

## Context

Five small methods over `src/data/orchestrator_enrollment_token.go` (142 lines, untouched): host enrollment
tokens with TTL + single-use marking + expiry sweep. Storage per #419 (`orchestrator_enrollment_tokens`
table; token value uniqueness). The only subtlety: expiry is *evaluated* (`ValidateEnrollmentToken`) **and**
*swept* (`DeleteExpiredEnrollmentTokens`) — the sweep's invocation cadence in legacy must be discovered and
mirrored (it is the enrollment analog of the stale-job watcher; task 18 wires goroutines, this task supplies
the function).

## Objective

5 methods on GORM with parity; validate/expire edges pinned by clock injection; sweep callable + tested,
**not self-starting**.

## Files

| File | Op |
|---|---|
| `src/database/enrollment_token.go` | create (5 methods) |
| `src/database/delegate_enrollment_tokens.go` | delete |
| `src/database/enrollment_token_parity_test.go` | create |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/data/orchestrator_enrollment_token.go` | 142 (measured) | untouched | ok |
| `src/database/enrollment_token.go` | new | ~90 | ok |

## Implementation spec

**Inventory:** `CreateEnrollmentToken` :20 · `GetEnrollmentToken` :45 · `ValidateEnrollmentToken` :65 ·
`MarkEnrollmentTokenUsed` :89 · `DeleteExpiredEnrollmentTokens` :116.

1. `Create` :20 — TTL source from config/default (**implementer to confirm :20-45**); dup-token pre-check per
   D-11 row 2, constraint as backstop (#419 unique index on token value).
2. `Get` :45 — by token/id, D-11 not-found mapping.
3. `Validate` :65 — three outcomes (unknown / expired / used / valid): legacy's distinct error values map via
   D-11; expiry evaluated on read (`expires_at < now`) — do **not** rely on sweep ordering; clock via the
   same `nowFn` seam introduced in task 15 §4 (shared package-private helper — one implementation, PR note if
   the seam needs lifting).
4. `MarkEnrollmentTokenUsed` :89 — field-scoped update; double-mark parity (legacy behaviour on re-use:
   **confirm :89-116** — idempotent vs error) pinned in test.
5. `Sweep` :116 — `DELETE WHERE expires_at < now` (+ used-and-old if legacy prunes used rows too —
   **confirm predicate :116-142**); returns deleted count; cadence discovery: grep callers of
   `DeleteExpiredEnrollmentTokens` across `src/` (controller/timer location + interval constant) — record
   finding in PR body; if legacy cadence is API-side (not engine-side), the sweep needs **no** SQL-mode
   goroutine and 18 gains a note, not code.

## Out of scope

Goroutine wiring (18, pending §5 finding), host lifecycle (13/14), provider flip.

## Validation

- [ ] `cd src && go build ./... && go vet ./database/...`
- [ ] `cd src && go test ./database/... -run EnrollmentToken -count=1` — clock edges: valid/±1s expired/used;
      double-mark shape
- [ ] Parity fixtures: 0 tokens / mixed-validity set / sweep-empties-all
- [ ] `cd src && go test ./... -count=1` (ER-5)
- [ ] `git diff --stat src/data/orchestrator_enrollment_token.go` → empty

## Rollback

Revert restores the delegate; no goroutine existed to stop; zero schema outreach.

## PBI acceptance mapping

**CRUD via GORM** — 5/5 with expiry semantics; **Error handling** — three distinct outcome errors mapped;
**Transactions** — none needed (single-statement domain — grep `dataMutex` in the file to evidence the claim);
**Legacy bypassed** — domain off-delegate after this PR. With this file, **all 15 domain tasks (05–16) have
landed: every interface method answers from SQL**, which is 17's precondition.
