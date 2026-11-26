# placement

Capacity reservations that close the read-decide-dispatch window in orchestrator VM placement (two concurrent creates both seeing "N CPUs free"). Design is locked in `.ai-agents/features/vm-placement-reservations/` (`brief.md`, `plan.md`, `tasks/`): read `plan.md` before changing behavior, it records rejected alternatives.

- `interfaces.go` — `Store` seam, `CapacityView`, outcomes. `store_sql.go` is the SQL implementation (also covered on Postgres); a Redis store is expected to slot in behind `Store` without caller changes.
- `service.go` — the `Service` the orchestrator uses. It alone owns: whether enforcement is on, release ordering (release only after a host refresh), and fail-closed behavior. Callers must not re-implement these.
- `sweeper.go` — expires stale reservations.
- Import rules: `jobs` and this package must not import `orchestrator`; the host-refresh function is injected at startup. Job completion reaches placement via the hook in `../jobs/hooks.go`.
- Tests are SQL-backed (`store_sql_test.go`, `store_postgres_test.go`); concurrency tests should run with `-race`.
- Config knobs: `../config/placement.go`, `../constants/placement.go`.
