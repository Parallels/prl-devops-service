# Task 05 — Postgres: offline gap-fill + integration parity + CI service-container job

**PBI** [#425](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01 (harness; task 04 completes the *full* parity surface this job runs — merge after 04 if sequencing allows, else run partial set) · **PR size:** M
**Plan refs:** [D-6](./plan.md) (offline/integration split + CI posture — follow verbatim) · [D-4](./plan.md) constraint rows · [D-9](./plan.md) budget defence · D-8 (coverage-step note)

## Context

AC 3's "behaviour may differ (type handling, constraint errors)" needs a real Postgres — and D-6 decided
*where*: offline gap-fill rides the default `go test ./...` step (server-free; string-date fidelity +
#418's pool/DSN/redaction matrix re-asserted at suite level), while the server-dependent leg is a **separate
CI job** (`postgres:16` service container) running `//go:build integration` files. Deliberately **not** in
`docker-compose.yml` (D-6: that's the user-facing product; a half-shipped postgres service there ships
half-shipped support). The budget AC constrains the *unit* step only — build-tagged files are excluded from
it by the Go toolchain itself (D-6 defence text to quote verbatim in the PR description).

## Objective

Offline gaps filled in the default run; the same parity scenarios re-run against `postgres:16` in a parallel
CI job; dev loop = `make test-integration` (task 07) or the canonical docker one-liner copied **verbatim**
from #418 task 04's integration doc-comment.

## Files

| File | Op |
|---|---|
| `src/database/postgres_offline_gaps_test.go` | create (package `database_test`) |
| `src/database/parity/postgres_parity_test.go` | create (`//go:build integration`) |
| `.github/workflows/pull_request.yml` | modify — new job only (D-9/Risk-9: 05 owns this file) |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `pull_request.yml` | *(measure at task start; job adds ~35)* | ≤ +40 | ok — job appended, existing jobs byte-identical (`git diff` shows append only) |
| new test files | — | ~110 + ~130 | ok |

## Implementation spec

1. **Offline gaps** (`postgres_offline_gaps_test.go`, no server):
   - pool: `Stats().MaxOpenConnections == cfg.Pool.MaxOpenConns` on a Postgres-opened handle opened
     **lazily/dial-freely** (gorm opens on first use — assert config propagation without connecting;
     if #418's constructor pings, reuse its no-ping option; **confirm at task time in shipped `#418` code**);
   - DSN matrix: valid forms (URL + keyword/value), missing-host, `sslmode` variants ⇒ typed
     `ErrInvalidDSN`-class outcomes (mirror #418 task 04 unit table; this file adds anything task 04's file
     left at the factory level *and* re-asserts through `FromApplicationConfig` wiring — #420 seam);
   - `RedactDSN` table incl. keyword-form (`password=x`) and unparseable-string fallback;
   - string-date fidelity (#419 D-1): create+reload model with string `CreatedAt`, byte-equal — the offline
     stand-in for Postgres type-handling (comment says so).
2. **Integration file** (`//go:build integration`, `POSTGRES_TEST_DSN` skip-guard, canonical docker snippet
   verbatim from #418 task 04 in the doc-comment):
   `RunOnEngines(t, scenario, []EngineFactory{testutil.PostgresEngine()})` running **selected scenarios**
   (start: auth-core + one catalog + one error-matrix subset — full-matrix-on-postgres is a
   follow-up decision, PR notes; the D-4 constraint rows run here: SQLite already proved
   `constraint failed`, Postgres must prove `SQLSTATE 23505` maps to the **same sentinel** — the AC-3 row);
   unique schema per run: DSN param `search_path=prl_t425_<nanos>` + `DROP SCHEMA ... CASCADE` in cleanup
   (concurrent CI runs safe); pool-under-concurrency smoke: 8 goroutines CRUD + `Stats()` invariants.
3. **CI job (exact block, append to `pull_request.yml`; adapt checkout/setup steps to the file's existing
   action versions — READ the file first, mirror its step style):**

```yaml
  test-postgres-integration:
    name: Postgres integration tests
    runs-on: ubuntu-latest
    needs: []                      # parallel with the unit job — D-6 budget defence
    services:
      postgres:
        image: postgres:16
        env: { POSTGRES_PASSWORD: ci-pg, POSTGRES_DB: prl_test }
        ports: ['5432:5432']
        options: >-
          --health-cmd "pg_isready -U postgres" --health-interval 5s
          --health-retries 10 --health-timeout 5s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: src/go.mod }
      - name: Postgres integration suite
        working-directory: src
        env:
          POSTGRES_TEST_DSN: postgres://postgres:ci-pg@localhost:5432/prl_test?sslmode=disable
        run: go test -tags integration -count=1 ./database/...
```

   (`go-version-file`/action versions must match the existing jobs' style — adjust, don't modernize
   unilaterally; D-8 recorded the coverage-summary step is `fail_below_min: false` ⇒ non-blocking for the
   new package — paste that line in PR description.)

## Out of scope

Compose file (forbidden per D-6), `values*.yaml`/helm (#427 + #420 scope), running migration suite on PG
(#424's own integration tests), full-matrix-on-postgres expansion.

## Validation

- [ ] Offline file green in default `go test ./...` (wall-time delta ≈ file's compile+run ≈ <2 s — logged)
- [ ] Integration file locally: `docker run` one-liner up ⇒ `POSTGRES_TEST_DSN=… go test -tags integration ./database/... -count=1 -v` green incl. 23505-sentinel row
- [ ] Actionlint (or existing lint workflow's tool) on the modified YAML; PR notes job's wall time from first CI run
- [ ] Full `cd src && go test ./... -count=1` unchanged (integration files excluded — prove with `go list -tags ... ` or the CI unit-step wall time)

## Rollback

Delete 2 files + revert job block (append-only diff ⇒ revert is trivial). Unit path untouched by design.

## PBI acceptance mapping

**Backend coverage** (AC 3) — offline + integration both shipped, same-scenarios-both-backends by
construction; **constraint-error parity** row of D-4 finally *proven*; **+30 s budget** — the defence text
and the parallel-job topology together mean the AC is met *architecturally*, and 07 proves it *measurably*.
