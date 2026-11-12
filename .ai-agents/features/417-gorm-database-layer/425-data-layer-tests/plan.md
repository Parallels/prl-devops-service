# PBI Plan #425 — Update data layer unit tests

> Reading order: [feature brief](../brief.md) → [feature plan](../plan.md) (esp. §12 test strategy, ER-1…ER-5, risk R-4/R-8) →
> [PBI brief](./brief.md) → this plan → your task spec.
>
> All repo anchors below were re-verified against HEAD `6e0d5af` on 2026-11-03 (`wc -l`, `grep`, timed runs).

## Scope

**In:** three-tier test strategy (D-1) · exported `src/database/testutil` harness (unique in-memory DSNs,
recording `ApiContext`, engine factories) · engine-parity suite over the 157-method surface in four
domain waves (auth → infra → orchestrator → jobs/config) with a finalized LEGACY-ONLY exclusion list ·
error-sentinel mapping matrix · Postgres offline + `//go:build integration` suite + CI job ·
legacy-migration edge-case suite against #424's package/fixtures · flake/perf sweeps · `Makefile` test
targets + CI green audit.
**Out:** any change to legacy engine behaviour or to the legacy tests themselves (Tier 1, D-1) · any
`src/data/*.go` non-test edit · deleting legacy tests (that call belongs to #426's removal round) ·
user-facing docs (that's #427; harness usage lives in package doc-comments) · `docker-compose.yml`
(product-facing file — D-6) · linter/coverage-threshold policy (D-8 audit records findings only) ·
changes to `src/database/` non-test files beyond one additive guard in #418's `sqlite.go` if needed (D-2).

**Merge position:** #418 → #420 → #419 → #421 → #424 all merged before this PBI's first task PR.
Tasks 01–07 merge in numeric order; 05 may parallel 03/04 (no file overlap: 05 owns
`.github/workflows/pull_request.yml`, 07 owns `Makefile`).

## Design decisions

### D-1 — “Existing tests pass” ⇒ THREE-TIER strategy (resolves the AC ambiguity)

AC 1 reads literally: *“All tests that previously tested the legacy data layer have been updated and
pass against the GORM implementation.”* Taken literally this converts-and-destroys the legacy suite —
which contradicts feature brief §2 item 4 (“legacy layer **stays**… deprecated”), AD-1's runtime
fallback (a fallback you stopped testing is not a fallback), and ER-5 (every merged PR must keep the
suite green, including the engine that is still default until #421 flips). The AC was written when
the plan assumed cutover; the locked architecture says coexistence. **Decision — three tiers:**

| Tier | What | Disposition |
|---|---|---|
| **1** | The 81 legacy test functions in `src/data/*_test.go` (3,983 lines, measured) + `src/startup/migrations/common_test.go` | **Stay as-is.** They still prove the deprecated engine works. Not edited, not deleted, not "converted". #425 guarantees they keep passing unchanged. Deletion of any subsumed legacy test is a #426 decision, not ours. |
| **2** | **ENGINE-PARITY SUITE** (new, `src/database/parity/`): shared scenario functions — same fixtures, same assertions — executed against **both** engines via `EngineFactory` parameters typed on `interfaces.DatabaseService` (AD-1). Every legacy test function becomes either a parity scenario (run under `legacy/` and `gorm/` subtests — coverage preserved **and** exceeded) or a documented LEGACY-ONLY exclusion (below). | The literal answer to AC 1: every scenario that "previously tested the legacy data layer" now also passes against GORM, and CI log output shows both legs. |
| **3** | New-backend tests: migration edge cases (fixtures from #424), error-path matrix, Postgres offline + integration | AC 3/4/5 directly. |

**Conversion waves by domain** (function counts verified 2026-11-03 via `grep -n "^func Test" src/data/*_test.go`):

| Wave | Domains (legacy files) | Scenarios | Task |
|---|---|---|---|
| auth | users 6, roles 4, claims 4, roles_claims 7, api_key 15 (+ filter/order/authorize differentials — **new** coverage: `filter_layer.go`/`order_layer.go`/`authorize_layer.go` have **no** legacy tests today) | ~35 | 02 |
| infra | catalog 5, catalog_managers 5, packer_template 4, reverse_proxy 7, vm_snapshots 2 (+ enrollment-token scenarios — new, none exist) | ~27 | 03 |
| orchestrator | orchestrator 38 (1 capability-gated, see exclusions) | ~37 | 04 |
| jobs/config | jobs 2 (1 excluded), config 3 (+ user_config scenarios — new) | ~7 | 04 |

**LEGACY-ONLY EXCLUSION LIST** (engine features with no SQL-engine counterpart; each stays Tier 1
forever within #425 — rationale per entry, mirrored as a doc-comment table in
`src/database/parity/doc.go` by task 04):

| Test(s) | Rationale |
|---|---|
| `main_test.go`: `TestNewJsonDatabase` `TestConnectDisconnect` `TestLoadSave` `TestLocking` `TestSaveAsync` `TestSaveAs` `TestLoadFromEmpty` `TestLoadWithEncryption` `TestIsDataFileEmpty` `TestRecoveryFromResidualFiles` `TestProcessSaveQueue` (11) | Whole-file save/backup engine (`src/data/main.go`): atomic `data.json` rewrites, save-interval goroutine + `ProcessSaveQueue`, `SaveAs` (AD-1 escape-hatch API, not on `DatabaseService`), cross-process file locks, encryption-at-rest of the file, residual `.save`/`.panic` recovery, empty-file detection with file-creation side effect (`main.go:645-655`). The SQL backend has no save interval (DB engine owns durability), no SaveAs concept, and encryption is a backend/deployment concern. |
| `concurrency_test.go`: `TestConcurrentSaveOperations` (and the `SaveNow` interleaving inside the other 3) | `SaveNow` + file-lock concurrency — file-engine only. The pure concurrent-CRUD portions of `TestConcurrentReadsWrites` / `TestDeadlockPrevention` / `TestRaceConditions` **do** convert (parity scenarios; SQLite leg exercises WAL + `busy_timeout` from #418). |
| `jobs_timeout_test.go`: `TestDetectStaleJobsNotifiesWithTimeoutReason` | Stale/ghost-job watcher is a legacy-engine goroutine (`src/data/main.go:109`, `GhostJobTimeoutMinutesEnvVar`). SQL-side retention/notification is open question OQ-2 (#421 territory) — not testable here. `TestUpdateJobPersistsForwardedFailureMessage` converts (jobs CRUD). |
| `orchestrator_test.go`: `TestOrchestratorHostNotConnected` | "Not connected" is engine semantics (`ErrDatabaseNotConnected` when the file store isn't loaded). A live `GormDatabase` handle is always connected. Capability-gated skip (`Engine.HasNotConnectedSemantics`) with rationale; see OQ-425-1. |
| `startup/migrations/common_test.go`: `TestCompareVersion` | Legacy JSON-store schema versioning — **different concept** from GORM AutoMigrate (AD-4). Untouched, out of #425 scope entirely. |
| api_key pure-function tests `TestNormalizeApiKeyType` `TestFilterInternalKeys` | Pure domain helpers, engine-agnostic; already Tier 1. Their storage-dependent siblings (`TestGetApiKey*Normalizes*`) convert. |

### D-2 — `src/database/testutil` as an **exported** package; unique-DSN scheme; `#418` URI guard

The brief floated `src/database/testutil_test.go` (package-internal). **Decision: exported package
`src/database/testutil/`** because `legacy_migrator` tests (task 06) and later #426 tests must import
the harness, and `_test.go` files are not importable. Import graph stays acyclic:
`testutil → {database, data, data/interfaces, data/models, basecontext, config}`; nothing under test imports `testutil` except tests.

Verified facts driving the harness:
- shared-cache `file::memory:?cache=shared` is **one logical DB per process** (feature-plan §12 names
  it; brief AC 2) — unshared, every pool connection of a bare `:memory:` DSN gets its **own** empty DB.
  Neither default is test-safe as-is.
- **Scheme:** DSN path `file:t425-<sanitized-test-name>-<atomic-seq>?mode=memory&cache=shared` —
  named ⇒ per-test isolation; `mode=memory` ⇒ RAM only; `cache=shared` ⇒ visible across the
  `database/sql` pool's connections (GORM opens >1 conn). Sanitize = `[^A-Za-z0-9]→-`, lowercase,
  length-capped at 48; seq from `atomic.Uint64` — unique under `-count=5`/`-shuffle=on` too.
- **`t.Cleanup(close)` is mandatory, not cosmetic:** a named shared-cache memory DB survives until
  *all* conns close — forgetting `Close` leaks schema across tests in the same binary. The harness
  registers cleanup itself; raw-DSN footguns become impossible.
- **No truncate-per-test:** deleting by DSN name is complete and free; TRUNCATE-per-table would list
  every #419 table, fight autoincrement/unique residue, and race under `t.Parallel`.
- **Contract seam (risk → D-9/R-3):** #418's `sqlite.go` special-cases the literal `":memory:"`; a
  `file:` URI with an existing `?` would double-append its `?_pragma=…` query. Task 01's first step is
  a 5-line additive guard in `src/database/sqlite.go`: if `strings.HasPrefix(cfg.Path, "file:")` use
  the path **verbatim** as DSN (memory DSNs skip WAL pragma — journal default for memory; busy_timeout
  harmless). Test: opening a `file:` URI twice sees the same schema.
- Schema bring-up: `testutil.OpenMemoryDB(t)` returns `(*gorm.DB, close)` via #418's
  `database.NewDatabaseConnection`; `testutil.MigrateSchema(t, db)` calls #419's exported migrator
  (`database.AutoMigrate(db)` — confirm name at task-01 time, grep `func AutoMigrate` in
  `src/database/`; fallback: local model registry declared in testutil listing every `data/models` type).
- Engine factories (`testutil/engines.go`):
  ```go
  type Engine struct {
      Name string                       // "legacy" | "gorm" | "postgres"
      New  func(t *testing.T) (interfaces.DatabaseService, func()) // svc, cleanup (t.Cleanup-registered)
      Capabilities EngineCapabilities   // HasNotConnectedSemantics, IsFileEngine, …
  }
  func LegacyEngine() Engine  // t.TempDir()/test_db.json + NewJsonDatabase — mirrors setupTestDB (D-3)
  func GormEngine() Engine    // OpenMemoryDB + MigrateSchema + database.NewGormDatabase (name per #421; confirm)
  func RunOnEngines(t *testing.T, scens []Scenario, engines ...Engine)
  // → t.Run(scen.Name+"/legacy"), t.Run(scen.Name+"/gorm"); capability skips print the exclusion rationale.
  type Scenario struct {
      Name string
      Run  func(t *testing.T, ctx basecontext.ApiContext, db interfaces.DatabaseService)
  }
  ```
- **Recording `ApiContext`** (`testutil/context.go`): fake implementing all **14** methods of
  `basecontext.ApiContext` (`src/basecontext/api_context.go:10-25`; #418 task-05 spec says "15" —
  recount at task time) and capturing `Log*` lines as `"level: formatted"`. This **is** the
  recording-fake pattern of
  [`418 tasks/05`](../418-gorm-connection/tasks/05-startup-wiring-healthcheck-logging.md) —
  #418's copy lives in `database`'s `_test.go` and is not importable, so testutil re-implements it
  once, exported. Helpers: `RecordingContext() (*RecordingApiContext, basecontext.ApiContext)`,
  `(r) .Containing(substr) []string`, `.CountLevel("warn", s) bool`. Task 06 (warn-not-fail
  validation AC) and every log-assertion in the parity suite ride on this.
- `testutil.WithCleanConfig(t)`: `t.Setenv("DATABASE_FOLDER", t.TempDir())` + `config.New(ctx)` —
  the global-reset pattern from `src/data/main_test.go:16-36`, safe because nothing here uses `t.Parallel` (D-3).

### D-3 — Global-state hazards → isolation rules (documented as package doc-comment, not a TESTING.md)

Verified hazards: `data.memoryDatabase` singleton (`src/data/main.go:27`, set at `:76-101`, plus the
save goroutine's `cancel` chan `:62` and the ghost-job goroutine `:109`); `config` global
(reset idiom `config.New(ctx)`, `src/data/main_test.go:25-26`); `serviceprovider` singleton
(`Get()`/`globalProvider`). Also: `t.Setenv` forbids `t.Parallel`; `testing_reset.go`'s
`ResetForTesting` is `//go:build integration`-gated while `main_test.go` resets `memoryDatabase`
directly (same package) — **document, do not "fix"** (brief instruction; it's Tier-1-internal anyway).

Isolation rules, enforced by review checklist + the sweeps in task 07, recorded in
`testutil/doc.go` (docs-as-code; #427 may *copy* them later into user docs — TESTING.md itself is
out of scope per the brief/prompt):
1. **One DB per test.** Every test opens its own uniquely-named memory DB (D-2); never share handles.
2. **No cross-test globals in new code.** `src/database/**` and `parity/**` never touch
   `serviceprovider.Get()`, never assign `data.memoryDatabase` (can't — unexported, wrong package);
   the parity suite's *legacy leg* constructs local `JsonDatabase`s exactly like `setupTestDB` and
   never relies on the singleton being nil.
3. **Config:** only via `WithCleanConfig`/`t.Setenv`; tests using them are sequential (no
   `t.Parallel`). New packages read config through the #420/`FromApplicationConfig` seam, never `config.Get()` at package-init.
4. **Order-independence gates:** new suites must pass `go test -count=5 ./database/...` and
   `go test -shuffle=on ./database/...` (supported: `go.mod` is `go 1.25.0`, `-shuffle` landed in 1.24).
   The **legacy** suite gets the same runs *informationally* (AC 6 scopes "no flaky" to new tests);
   any legacy order-dependency is recorded, not fixed (Tier 1).
5. **No disk** for the GORM legs (memory DSN) — with the one honest exception in AC 2's own wording
   discussed in task 06 (migration fixtures are `data.json` files in `t.TempDir()`; they are *inputs*, not test-written DB state).

### D-4 — Error-sentinel mapping matrix (AC 5)

Legacy pattern (`src/data/users.go:14-26` archetype): `errors.NewWithCode(msg, httpCode)` sentinels;
controllers branch on `SystemError.Code()` (`src/errors/main.go:70,114`). Verified sentinel inventory
(2026-11-03 grep; the mapping in `GormDatabase` must **mirror legacy exactly, quirks included** —
fixing codes is an API behaviour change, out of #417):

| Class | Sentinels (domain → sentinel) | Assertion |
|---|---|---|
| not-found (404) | `ErrUserNotFound`, `ErrRoleNotFound`, `ErrClaimNotFound`, `ErrApiKeyNotFound`, `ErrCatalogManifestNotFound`, `ErrOrchestratorHostNotFound`, `ErrOrchestratorReverseProxyHostNotFound`, `ErrOrchestratorHostVirtualMachineNotFound`, `ErrOrchestratorHostsVMSnapshotsNotFound`, `ErrPackerTemplateNotFound`, `ErrJobNotFound`, `ErrEnrollmentTokenNotFound`, `ErrUserConfigNotFound`, `ErrorReverseProxyHostNotFound`, `ErrorReverseProxyHttpRouteNotFound`, `ErrorReverseProxyTcpRouteNotFound` | `Get*(missing)` on the gorm leg returns the same sentinel; `errors.Is` **or** `==` per legacy compare-style; `Code() == 404`. |
| not-found (no code!) | `ErrCatalogManagerNotFound = errors.New(...)` (`catalog_managers.go:11`) | gorm leg returns the *same sentinel identity*; assert no code panic — mirror the quirk. |
| duplicate → 400 | `ErrUserAlreadyExists`(400), `ErrPackerTemplateAlreadyExists`(400), `ErrRoleAlreadyContainsClaim`(400), `ErrUserAlreadyContainsRole/Claim`(400), codef'd 400s in claims/roles/orchestrator/reverse_proxy create paths | Insert existing ID/name → assert same sentinel/code, **not** raw unique-violation text. SQLite leg forces `constraint failed`; Postgres leg (integration) forces `SQLSTATE 23505` — same assertion ⇒ the `TranslateError`/mapping in #421 is what's under test. |
| quirk duplicates | `ErrApiKeyAlreadyExists`(500!), `ErrCatalogAlreadyExists`(404!), `ErrUserConfigAlreadyExists`(500!) | Mirror as-is + `t.Log` the quirk; comment links this row so nobody "fixes" it. |
| unreachable | `database.ErrDatabaseUnreachable` (wraps `#418` factory/ping errors, D-5 of #418) | `NewDatabaseConnection` against `postgres://u:p@127.0.0.1:1/db` → `errors.Is(err, ErrDatabaseUnreachable)` true; message contains redacted DSN (no `p` leak). Offline, no network beyond loopback. |

Lives in `src/database/parity/errors_parity_test.go` (task 04) + wave-embedded per-domain assertions
(each scenario asserts legacy-identical error by construction — that's what "same scenario, both
engines" buys).

### D-5 — Parity suite mechanics: differentials promoted, not duplicated

Feature plan §12's "differential tests: same dataset in legacy vs GORM asserting identical list/filter
output" ships in #421's foundation task against captured fixtures; the brief says #425 **owns the
promotion** of that harness into permanent CI-runnable suites. Mechanic: `Scenario.Run` executes
identically on both engines and asserts **behavioural equality against expectations**, plus
`filter_differential_test.go` (task 02) additionally runs a filter/order table **cross-engine**
(both engines seeded from one fixture dataset ⇒ assert `Get*` ID sequences equal per expression —
expressions harvested from `src/data/filter_layer.go` operators + controller usage). This is the
R-3 (ParseFilter parity) safety net at suite level.

### D-6 — Postgres: offline vs integration split; CI as **service container in a separate job**, not compose

AC 3 ("behaviour may differ — type handling, constraint errors") is only fully provable against a
real server; the default `go test ./...` step must stay server-free.
- **Offline (default run):** everything #418 shipped (pool `Stats().MaxOpenConnections`, DSN
  validation, `RedactDSN` matrix) **plus** a small gap-fill file: string-date fidelity (#419 models keep
  string dates — assert round-trip identical on the SQLite leg as the offline stand-in for the
  Postgres type-handling row) and the offline rows of D-4.
- **Integration:** `//go:build integration` + `POSTGRES_TEST_DSN` skip-guard (same convention as
  #418's `postgres_integration_test.go`, whose planned file carries the **canonical docker snippet**:
  `docker run --rm -e POSTGRES_PASSWORD=… -p 5432:5432 postgres:16` — grep hit in
  [`418 tasks/04`](../418-gorm-connection/tasks/04-connection-factory-postgres-and-pool.md) §integration;
  #425 copies it verbatim into its own integration file's doc-comment and CI uses `postgres:16` to match).
  Contents: `RunOnEngines(t, smoke, engines)` with a `PostgresEngine()` (`t.Setenv` DSN, unique schema
  per package via `search_path=prl_t425_<seq>`) — the **same** scenarios re-run against Postgres;
  constraint-error parity row from D-4; pool sanity under concurrency (`Stats()`).
- **CI:** new top-level job `test-postgres-integration` in `pull_request.yml` (exact YAML in task 05):
  `services: postgres:16` container + health-cmd + `go test -tags integration -count=1 ./database/...`
  with `POSTGRES_TEST_DSN` env. **`docker-compose.yml` untouched** — it is the user-facing product
  compose; a postgres service there ships half-implemented DB support to users. Dev loop = the
  one-liner doc-comment + `make test-integration` (task 07).
- **Budget defence (+30 s AC):** the constraint reads "Test execution time for **the unit test
  suite** must not increase by more than 30 seconds overall" — it constrains the default `go test ./...`
  step (CI `pull_request.yml:66`, VERIFIED: `go test -coverprofile=../coverage.out -covermode=count ./...`).
  The integration job is a *separate, parallel* CI job: build-tagged files are **excluded from the
  default run** by the Go build system itself, so its wall time (container up ~10–20 s + tests < 60 s)
  never enters the unit step; overall CI latency unchanged (parallel jobs). Cost that *does* hit the
  unit budget — the parity suite's per-test memory-DB open + AutoMigrate — is estimated in Risks with
  a measured gate (task 07).

### D-7 — Migration suite consumes #424's fixtures; constructs state where fixtures can't express it

Fixtures under `src/database/legacy_migrator/testdata/` (owned by #424's tasks — this PBI **reuses**;
adds only genuine gaps, e.g. a corrupt-JSON file if absent): empty store, partial collections,
encrypted store (+ test-only key fixture — synthetic, never a real secret), marker-present,
corrupt-JSON. State machines the migration runs through can't all be static JSON, so the suite also
constructs: **crash-mid-run** = migrate into a target DB then delete the completion-marker row (or
commit a subset of collection batches via #424's exported seams) and re-run ⇒ upsert re-entry, zero
duplicates; **marker short-circuit** = full migrate, run again ⇒ assert the second run never opens
collections (cheap assertion via RecordingContext: no per-collection info lines). Required behaviour
matrix (each maps to a #424 AC checkbox, asserted through the exported package API only — white-box
inside `legacy_migrator_test` package):

| Scenario | Asserts |
|---|---|
| empty store / missing file | silent skip: `err == nil`, no marker, no rows. **Trap:** must use #424's detector, never `IsDataFileEmpty` (creates the file — `src/data/main.go:645-655`). |
| happy path | all 16 collections incl. pointers (`Configuration`, `ReverseProxy`) and embedded slices (`User.Roles`/`Claims`) land; validation logs no warn. |
| encrypted | loads via AD-5 loader path with key fixture; rows land. |
| corrupt JSON | error returned, no panic, target DB untouched (count 0). |
| re-run ×2 (marker present) | second run no-op; counts unchanged. |
| crash-mid-run re-entry | marker absent + partial rows ⇒ completion, **no duplicates** (upsert-by-ID per #424 brief), validation consistent. |
| validation mismatch | warn-only: `err == nil`, `RecordingContext` contains a warn line naming the collection; migration still committed (warn-not-fail per #424 AC 6). |
| dry-run | zero rows in target, zero marker, `data.json` byte-identical (sha256 before/after), dry-run report lines present at info. |
| never-modify-source | shared across scenarios: hash of `data.json` invariant (OQ-1 frozen-source stance). |

### D-8 — CI/Makefile audit findings (verified — the record differs from the common assumption)

- `pull_request.yml` (109 lines): `Test` job, step at lines 62–67 runs `go test -coverprofile=../coverage.out -covermode=count ./...`; coverage thresholds comment-style `40 80` with `fail_below_min: false` (non-blocking). **This is the step the +30 s budget applies to.** `setup-go` uses `go-version-file: src/go.mod` (step *name* says "1.21.x" — stale label, actual toolchain 1.25).
- `linter.yml`: **the whole job is `if: false` (disabled)** and had `VALIDATE_GO: false` — so *no Go lint gate runs in CI at all today*. The assumption "linter.yml runs golangci" is **false as verified**; do not promise lint enforcement in PRs. `golangci-lint` exists only as a Makefile install helper (`Makefile:287-289`) with root `.golangci.yml` config; `make lint` runs super-linter with `VALIDATE_GO=false` (Makefile:303-308). Task 07 records this, recommends a follow-up issue, and adds a *local* `make test-race`/lint discipline to PR instructions instead.
- `Makefile` (323 lines): `test:` runs `@scripts/test -d ./src` — **`scripts/test` does not exist in the tree** (verified `ls scripts/`) ⇒ `make test` is broken today; `coverage:` runs `go test -coverprofile … ./...` directly (works). Decision (feeds #427's promise that documented `make` behaviour stays honest): task 07 retargets `test:` to mirror the CI step (`cd src && go test -count=1 ./...`) and adds `test-race` / `test-integration` — 3 small targets, additive, no other Makefile edits.

**Task-07 appended record (2026-11-12, executed on `slice/425-07`):**

- `test:` retarget + `test-race` + `test-integration` landed (Makefile 323→336 lines, +13; help line updated inline). `test-integration` defaults `POSTGRES_TEST_DSN` per spec and carries the docker one-liner (`POSTGRES_DB=prl_test` included, matching the CI service env). `make test` verified live (40 ok pkgs); `make test-race` verified live; `make test-integration` verified at `make -n` + DSN-expansion level — live run needs docker (human gate / CI).
- **CI runs no `-race` step ever** — `make test-race` on the whole module therefore surfaces *pre-existing* legacy races for the first time: `jobs/tracker` map write (service.go:482/709 worker) vs cleanup reads (:1375–79), surfaced via `compressor` and `telemetry` tests; plus failing-under-race `restapi` (1), `serviceprovider/parallelsdesktop` (1), `tests/controllers` (4). All outside `database/` — ledger → #426 (Risk-6/Tier-1 territory). **`./database/...` itself is race-clean: gate = `-count=5 -shuffle=on -race` × two consecutive runs, zero failures, zero races (504 s / 502 s).**
- Tier-1 legacy sweep (`./data/... -count=3 -shuffle=on`, informational per AC 6): `TestLoadSave` failed 1/3 — `unlinkat /tmp/prl-devops-data-test-*: directory not empty` at main_test.go:145 cleanup, racing the async save queue's residual files. Ledger → #426.
- Wall-time (this machine, warm cache, `go test ./... -count=1`): pre-#417 `cc7d3d4` = **29–33 s** (2 warm runs); pre-#425 `2538a7a` = 43 s; post-feature branch = **42 s**. Feature delta ≈ **+10 s vs true baseline — D-9 ≤ +30 s budget holds**; AutoMigrate cost model (D-9) vindicated, fallback ladder untouched.
- T12-class flake frequency observed 3/5 full-suite runs (`TestHostSelectionPreservesResourceFailuresInJobs`, isolated rerun green each time) — strengthened argument for #426's flake ledger; also 425-06's macOS shuffle flake class (clock-tick-tied seeds + random-ID tiebreak) documented in that entry.

### D-9 — Baseline numbers (measured 2026-11-03, warm build cache)

| Measurement | Value |
|---|---|
| `go test ./data/... -count=1` | **30.4 s** package time (32.7 s wall) — the single heaviest suite; the +30 s budget is real against this baseline |
| `go build ./data/...` | 2.5 s |
| Legacy data tests | 3,983 lines / 81 top-level `Test*` funcs across 15 files (`wc -l` above) |
| `JsonDatabase` public methods | 157 (grep-verified) — matches AD-1 |

Parity cost model (estimate — task 07 proves it): ~100 scenarios × gorm-leg (open unique mem DSN +
AutoMigrate ~10–40 ms glebarez + scenario body) + legacy leg (~2–5 ms, file in `t.TempDir()` +
`config.New` — same cost the legacy suite already pays per test) ⇒ **+4…8 s expected**, comfortably
inside +30 s *provided AutoMigrate stays sub-40 ms*; that assumption is task 01's early-warning
measurement (see task 01 Validation). Fallbacks if measured over budget (ordered): (a) memoize
`gorm.Schema` parse via `sync.Once` on model set (parse is DSN-independent; migration itself stays per-DSN),
(b) group scenarios per DSN with precise cleanup — only if (a) insufficient, violates rule D-3.1 so it's a human decision,
(c) shrink waves. Numbers pasted before/after in every wave PR (task 07 owns the final proof).

## Target files & layout (end state of this PBI)

```
src/database/
├── sqlite.go                      # D-2 additive URI guard (task 01, ≤5 lines, if #418 lacks it)
├── testutil/                      # exported harness package (task 01)
│   ├── doc.go                     #   isolation rules + DSN scheme + usage examples (the "TESTING.md" per D-3)
│   ├── harness.go                 #   OpenMemoryDB / MigrateSchema / WithCleanConfig / sanitize+seq
│   ├── context.go                 #   RecordingContext fake (14 methods) + matchers
│   ├── engines.go                 #   Engine/Scenario/RunOnEngines + LegacyEngine/GormEngine
│   └── harness_test.go            #   leak tests (two opens same-DSN; distinct DSNs isolated) + sample scenario
├── parity/                        # engine-parity suite (tasks 01–05)
│   ├── doc.go                     #   exclusion ledger table (finalized task 04) + how-to-add-scenario
│   ├── parity_test.go             #   bootstrap + users smoke scenario (task 01)
│   ├── users_parity_test.go / roles_claims_parity_test.go / api_key_parity_test.go
│   ├── filter_differential_test.go            # cross-engine filter/order/authorize (task 02)
│   ├── catalog_parity_test.go / reverse_proxy_parity_test.go / snapshots_parity_test.go  (task 03)
│   ├── orchestrator_host_parity_test.go / orchestrator_vm_parity_test.go /
│   │   orchestrator_resources_parity_test.go / jobs_config_parity_test.go
│   ├── errors_parity_test.go                  # D-4 matrix (task 04)
│   └── postgres_parity_test.go                # //go:build integration (task 05)
├── postgres_offline_gaps_test.go              # D-6 gap-fill, default run (task 05)
└── legacy_migrator/
    ├── detector_test.go / migrator_edge_test.go / validation_dryrun_test.go   (task 06)
    └── testdata/ …                            # #424-owned; extend only on real gaps (task 06)
.github/workflows/pull_request.yml             # + test-postgres-integration job (task 05)
Makefile                                       # test/test-race/test-integration (task 07)
```

Every new file < ~300 lines (ER-4 guard per task); no existing file grows past 1,000 (only touched
existing files: `sqlite.go` +≤5, `pull_request.yml` +~30, `Makefile` +~12).

## Risks

| # | Risk | Mitigation |
|---|---|---|
| 1 | **+30 s budget blown by AutoMigrate per test** (measured baseline: `./data` alone = 30.4 s) | D-9 cost model; task 01 measures open+migrate ms in CI *before* waves land; fallbacks (a)–(c); task 07 pastes `go test ./... -count=1` wall before/after in PR. |
| 2 | Shared-cache memory DB leaks across tests (`file::memory:?cache=shared` = 1 DB/process) | D-2: unique named DSNs + harness-owned `t.Cleanup` close + explicit leak tests in `harness_test.go` + `-count=5`/`-shuffle=on` gates (D-3.4). |
| 3 | **Flakiness** from globals (`data.memoryDatabase`, config, provider) or goroutine timing | D-3 rules 1–3; parity never touches `serviceprovider`; goroutine-y legacy behaviours stay excluded (D-1 list) instead of being re-implemented in tests; task 07 `-race -count=5` sweep, zero-flake gate; CI retry count = 0 discipline (flakes fixed or quarantined, never retried away). |
| 4 | #421 contract drift (constructor/migrator names, capability semantics) between planning and merge | All references behind `testutil` seams with "confirm at task time, grep the shipped source" steps written into tasks 01/04; compile break at merge boundary is the detection mechanism, fix is one line in testutil. |
| 5 | SQLite↔Postgres behaviour deltas bite in unexpected places (types, unique-violation text) | Scenarios assert **domain sentinels**, never driver error text (D-4); Postgres leg runs the *same* scenarios (D-6), so drift surfaces in the integration job, not in production. |
| 6 | Legacy order-dependency discovered by `-count=5/-shuffle` | Informational (AC 6 = new tests only); recorded in PR; fixes, if trivial and Tier-1-safe, get their own micro-PR — otherwise a #426 input. |
| 7 | `postgres:16` service timing out in CI / runner load | `pg_isready` health-cmd with 10×5 s retries; job is non-blocking on the unit path; if infra flakes repeatedly, gate job on `src/**` paths — decision recorded in task 05, not silently. |
| 8 | Encrypted fixture becomes a "committed secret" red flag | Synthetic test-only key, file header comment, security-rules compliant (`grep` for real key material in review). |
| 9 | Two tasks editing one file | 05 owns the workflow file, 07 owns Makefile; testutil edited only in 01 (later waves read-only — additive files only). |
| 10 | Coverage-summary step (`thresholds: "40 80"`) reacts to new package | Non-blocking (`fail_below_min: false` — verified); note in task 05 PR. |

## Open questions

| ID | Question | Needs answer by |
|---|---|---|
| OQ-425-1 | Does `GormDatabase` emulate `ErrDatabaseNotConnected` for a deliberately disconnected handle, or is the capability permanently absent (→ skip stays permanent)? Confirm with #421's shipped code in task 04. | task 04 |
| OQ-425-2 | Final names: `interfaces.DatabaseService` aggregate + `database.NewGormDatabase` + AutoMigrate entrypoint. Plan writes against these; confirm by grep at task-01 time (feature plan AD-1/§7 names them). | task 01 |
| OQ-425-3 | #424 fixture inventory/naming (which of the 5 fixture kinds they ship vs. what task 06 adds). Coordinate at #424 implementation kickoff; task 06 lists its expected filenames. | task 06 |
| OQ-425-4 | Jobs retention/timeout semantics on SQL (feature-plan OQ-2) — blocks any SQL-side stale-job test forever or only temporarily. | task 04 docs it as excluded |
| OQ-425-5 | Do subsumed Tier-1 tests get deleted in #426? #425's answer: **never within #425**; parity suite makes the #426 conversation quantitative (scenario ledger in `parity/doc.go`). | #426 round |

## Task index (each task = one PR, merge in order)

| Task | Goal | Depends on | Touches | Size |
|---|---|---|---|---|
| [01](./tasks/01-testutil-harness-dsn-isolation.md) | `testutil` package (DSN scheme, isolation rules doc, recording ctx, engine factories, leak tests) + users-smoke parity scenario + early AutoMigrate cost measurement | #418/#419/#420/#421 merged | `testutil/*` new, `parity/{doc,parity_test}.go` new, maybe `sqlite.go` +5 | M |
| [02](./tasks/02-parity-wave-auth.md) | Wave **auth**: users/roles/claims/roles_claims/api_key scenarios + cross-engine filter/order/authorize differentials | 01 | `parity/{users,roles_claims,api_key,filter}_*.go` new | M |
| [03](./tasks/03-parity-wave-infra.md) | Wave **infra**: catalog/catalog_managers/packer_template/reverse_proxy/vm_snapshots + enrollment-token (new) | 02 | `parity/{catalog,reverse_proxy,snapshots}_parity_test.go` new | M |
| [04](./tasks/04-parity-wave-orchestrator-jobs-config.md) | Wave **orchestrator + jobs/config**; D-4 error matrix; exclusion ledger finalized in `parity/doc.go` | 03 | `parity/orchestrator_*, jobs_config_*, errors_parity_test.go`, `parity/doc.go` (edit) | M |
| [05](./tasks/05-postgres-offline-integration-ci-job.md) | Postgres offline gap-fill + `integration` parity file + `test-postgres-integration` CI job (exact YAML, D-6) | 01 (04 for full parity) | `postgres_offline_gaps_test.go`, `parity/postgres_parity_test.go`, `.github/workflows/pull_request.yml` | M |
| [06](./tasks/06-migration-edge-case-suite.md) | Migration suite per D-7 matrix against #424's API/fixtures | 01 + #424 merged | `legacy_migrator/*_test.go` new, `testdata/` gap-adds | M |
| [07](./tasks/07-flake-perf-sweep-make-ci-audit.md) | `-race -count=5` + shuffle sweeps; `go test ./...` before/after wall times in PR; Makefile `test`/`test-race`/`test-integration`; CI audit record (D-8) | 01–06 | `Makefile` only (+PR description artifacts) | S |

## AC-coverage table (every AC checkbox)

| AC | Where proven |
|---|---|
| **Existing tests pass** | Tier 1 untouched + green in every wave PR (`go test ./...`); Tier 2: legacy function → parity scenario ×2 engines, ledger in `parity/doc.go` (tasks 01–04); full-suite green at every merge boundary (ER-5). |
| **In-memory SQLite, no disk** | All GORM legs via `testutil.OpenMemoryDB` unique `?mode=memory&cache=shared` DSNs (D-2); leak + isolation tests in `harness_test.go`; disk exception documented for migration *input* fixtures (task 06 Validation checklist item). |
| **Backend coverage** | Task 05: offline (pool/redact/DSN/typing) + integration parity runs vs `postgres:16` in CI; D-4 constraint-error rows run on both backends. |
| **Migration tests** | Task 06 D-7 matrix: empty, missing, encrypted, corrupt, partial, re-run (marker + crash-reentry), validation warn, dry-run, source-hash invariant. |
| **Error path coverage** | D-4 matrix (task 04) + per-wave scenario error assertions + unreachable-wrapping offline test. |
| **No flaky tests** | D-3 rules + `-count=5`, `-shuffle=on`, `-race` gates in tasks 01/07; zero-flake definition in task 07 Validation. |
| *Constraint: ≤ +30 s* | D-9 model + task 01 early measurement + task 07 before/after wall-time evidence. |
