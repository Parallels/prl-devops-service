# Task 12 — Reverse proxy domain on GORM (reorder invariant)

**PBI** [#421](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 06 · **PR size:** M
**Plan refs:** D-9 tx inventory (`reverse_proxy.go:312-394`) · D-11 error mapping · D-12 schema ownership · [06 §parity](./06-domain-users.md) harness usage

## Context

Sixteen methods over a config singleton + hosts + embedded routes (legacy `src/data/reverse_proxy.go`, 424
lines, **untouched** by this PR). Per #419's D-3 walk, `ReverseProxyHost.HttpRoutes` is a **serialised-JSON
column** — route mutations are read-modify-write of the host row, and the plan's D-12 rule applies: route
filters are Go-side. The one real hazard is the **reorder invariant** (`ReorderReverseProxyHostHttpRoute`,
`reverse_proxy.go:342`): legacy rewrites *every* route's order in one `dataMutex.Lock()` critical section —
the SQL port must reproduce all-or-nothing ordering under a single transaction or concurrent updates can
produce duplicate/missing order values.

## Objective

All 16 methods served by GORM with parity vs legacy on the task-03 harness, including concurrent reorder
stress; the route-order invariant provably survives two simultaneous reorder calls.

## Files

| File | Op |
|---|---|
| `src/database/reverse_proxy.go` | create (config + host methods) |
| `src/database/reverse_proxy_routes.go` | create only if `reverse_proxy.go` projects >200 lines (plan index) |
| `src/database/delegate_reverse_proxy.go` | delete (task 04 stubs) |
| `src/database/reverse_proxy_parity_test.go` | create |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/data/reverse_proxy.go` | 424 (measured) | 424 (untouched) | ok |
| `src/database/reverse_proxy.go` | new | ~180 | ok — split routes file if growth >200 (pre-authorised above) |

## Implementation spec

**Method inventory** (measured via `grep -n "func (j \*JsonDatabase)" src/data/reverse_proxy.go`):

| Method | Legacy | SQL shape |
|---|---|---|
| `GetReverseProxyConfig` | :28 | `First` on single-row `reverse_proxy_config`; seed-on-read? — **implementer to confirm legacy seed semantics (:28-36)**; if none, missing row ⇒ `interfaces.ErrReverseProxyConfigNotFound`-equivalent via D-11 |
| `EnableProxyConfig` / `DisableProxyConfig` | :36/:51 | field-scoped `Update("enabled", true/false)` — NOT `Save` (whole-row write races the hosts' view readers) |
| `UpdateReverseProxy` | :66 | read-modify-write single row |
| `GetReverseProxyHosts` | :86 | `Find` all hosts; route arrays already JSON columns |
| `GetReverseProxyHost` | :104 | `First` by id ⇒ D-11 not-found |
| `CreateReverseProxyHost` | :124 | dup-name pre-check (mirrors legacy check-first, D-11 row 2) + `Create` |
| `DeleteReverseProxyHost` | :154 | existence check + `Delete` by id |
| `UpdateReverseProxyHost` | :174 | read-modify-write whole host row |
| `ConfigureReverseProxyHostTls` / `…Cors` | :215/:234 | field-scoped updates on host row (`Select("tls_config")` / `Select("cors_config")`) |
| `CreateReverseProxyHostHttpRoute` | :253 | **tx**: read host → append route (order = len) → `Select("http_routes")` update |
| `DeleteReverseProxyHostHttpRoute` | :277 | **tx**: read → remove + compact orders → update (compaction inside same tx — legacy behaviour, **implementer to confirm compaction (:277-312)**) |
| `UpdateReverseProxyHostHttpRoute` | :312 | tx: read-modify-write |
| `ReorderReverseProxyHostHttpRoute` | :342 | **tx**: read-modify-write of ALL orders — one `UPDATE` of the serialised column inside `Transaction`; Postgres: `clause.Locking{Strength: "UPDATE"}` on the read (SQLite is writer-serialised anyway) |
| `UpdateReverseProxyHostTcpRoute` | :395 | field-scoped update, same discipline |

1. Authz wraps where legacy had them — **implementer to confirm wraps per method (:86-124 read paths)**;
    apply task-03 `query` authz exactly where legacy applied it, no more.
2. `mapGormError(err, notFound, dup)` per D-11; sentinel names pinned from the task-01 interfaces file.
3. Empty-collection fidelity: no hosts ⇒ `[]`, never nil into wire (legacy slices are non-nil — pin in test).

**Parity tests** (`database/parity`): fixture with 1 config + 3 hosts × mixed routes (tls/cors/tcp present,
one host with 0 routes); per-method legacy-vs-GORM deep-equal incl. ordering after each mutation class; plus
the invariant test:

```go
func TestReorderInvariant_Concurrent(t *testing.T) {
    // 2 goroutines reorder the same host differently on the GORM engine;
    // after both: orders for that host == 0..n-1 each exactly once (sorted slice compare).
    // Same test on the legacy engine must yield the same post-state (serialised by dataMutex).
}
```

## Out of scope

Orchestrator's RP **views** (`GetOrchestratorReverseProxy*` — read models over hosts; task 14), provider
flip (17/18), schema changes (#419 owns columns; missing field ⇒ #419 amendment note per D-12), legacy file
refactors.

## Validation

- [ ] `cd src && go build ./... && go vet ./database/...`
- [ ] `cd src && go test ./database/... -run ReverseProxy -count=1 -race` (`-race` mandatory: the concurrency test is the point)
- [ ] Parity suite green on in-memory SQLite (both engines side-by-side per harness)
- [ ] `cd src && go test ./... -count=1` full suite (ER-5)
- [ ] `wc -l src/database/reverse_proxy*.go` ≤ guard table

## Rollback

Restore `delegate_reverse_proxy.go` (git revert), delete the three new files. Delegate path returns the
domain to legacy; zero schema risk (AutoMigrate columns are additive and unused when delegated).

## PBI acceptance mapping

**CRUD via GORM (SQLite parity)** — 16/16 methods; **Transactions** — reorder/compaction tx (D-9
`reverse_proxy.go:312-394` row); **Error handling** — D-11 table instantiated; **Concurrent access** —
invariant stress pre-proves what task 18 verifies end-to-end on WAL.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-12`.** Cut `slice/421-12` from `sql-migration` HEAD
(`.worktrees/slice-421-12`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
