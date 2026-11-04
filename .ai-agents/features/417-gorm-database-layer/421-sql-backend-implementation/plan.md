# PBI Plan #421 — SQL backend implementation (GormDatabase)

> Reading order: [feature brief](../brief.md) → [feature plan](../plan.md) (AD-1, AD-2, AD-4, ER-1…ER-5, R-1/R-3) →
> [PBI brief](./brief.md) → this plan → your task spec.
> Contracts consumed (merged before this PBI starts, per feature plan §8): **#418** (`database.NewDatabaseConnection`,
> `serviceprovider.GetSQLDatabase`, pragmas/logger/ping — see [418/tasks/05](../418-gorm-connection/tasks/05-startup-wiring-healthcheck-logging.md)),
> **#420** (`database.DatabaseConfig.FromApplicationConfig()` filling the seam of 418/task 02), **#419**
> (tagged models + **its mapping-table doc** `../419-gorm-model-tags/plan.md` — the *normative schema* for every
> domain task here; this plan references it and never re-decides table/join choices).

## Scope

**In:**
- `DatabaseService` interface capture: 157 methods, per-domain interface files under `src/data/interfaces/`,
  composed aggregate, compile-time assertions (AD-1, R-1).
- `src/database/query/`: exact-semantics port of the Mongo-flavoured query language
  (`filter_layer.go` / `order_layer.go` / `index_layer.go` / `authorize_layer.go` + `GetOwnRecords`
  `src/data/api_key.go:19`) and the **differential fixture harness** (R-3 mitigation).
- `src/database/`: `GormDatabase` (hybrid delegate) + one implementation file per domain, GORM transactions,
  error mapping to existing `Err*` sentinels, SQL-mode stale-job watcher.
- Provider switch mechanism (3 steps: field → API → selection) and the **final flip**: `DB_TYPE` selection,
  escape-hatch audit, panic-snapshot policy, WAL verification, postgres enablement.

**Out (with owner):**
- Table/column/relationship design, AutoMigrate model list, soft-delete choices → **#419** (we consume its mapping doc).
- `DB_*` env/config parsing → **#420** (we only call `FromApplicationConfig()`).
- `data.json → SQL` data copy, completion marker, dry-run → **#424** (we guarantee its write target is a complete
  `DatabaseService` before it runs; #424 starts after this PBI).
- `// Deprecated:` annotations, lint guard, startup warn on legacy selection → **#426**.
- Suite-wide coverage/edge-case completion → **#425** (we ship per-domain parity tests it can extend).
- Splitting mega-files the feature plan declares untouchable (`parallelsdesktop/main.go` 3,079,
  `system/main.go` 1,076; controllers 4,471/2,599/2,148/1,399) — see **D-10**.
- MySQL stub in `src/sql/` — untouched (feature brief §8).

## Design decisions

### D-1 — Hybrid delegate `GormDatabase` (adopts the recommended architecture, refined)

```go
// src/database/gorm_database.go  (package database, stays ≤ ~120 lines)
type GormDatabase struct {
    gormDB *gorm.DB            // handle from #418 factory (config via #420 FromApplicationConfig)
    legacy *data.JsonDatabase  // hybrid delegate: unported domains route here; nil after task 18
    onJobTimeout func(models.Job)
    timeoutMu  sync.RWMutex
    connected  bool
    closeOnce  sync.Once
    stopWatch  chan struct{}    // stale-job watcher shutdown (D-8)
}
func NewGormDatabase(ctx basecontext.ApiContext, gormDB *gorm.DB, legacy *data.JsonDatabase) *GormDatabase
func (g *GormDatabase) DB() *gorm.DB          // test/audit accessor (WAL pragma checks, #424 writes)
```

- Every domain task replaces that domain's delegate stubs with GORM code and nothing else (D-2 keeps stubs
  compiling per-domain). Task 18 sets `legacy = nil` and deletes the field (grep-enforced: no `.legacy` remains).
- **Runtime honesty:** the provider keeps returning `JsonDatabase` until task 18 (ER-5, brief: "feature ships
  dark"), so a partially-ported `GormDatabase` never serves production traffic; mixed-engine behaviour exists
  only inside tests, which is exactly what the hybrid delegate is for — it makes *compilation and interface
  completeness* incremental while parity is proven per domain by the D-5 harness.
- **Method bodies mirror legacy method-for-method.** Executing domain tasks require the implementer to read the
  legacy file (it's in the same repo); the spec pins the quirks that tests must lock (see each task's "Quirks
  to pin" list). This is deliberate: the AC is behaviour-preserving persistence swap (feature brief §8 non-goal
  "no API/behaviour changes").

### D-2 — Interface capture: all 157 methods, one file per domain, exact signatures

Inventory verified 2026-11-03: `grep -rn "func (j \*JsonDatabase) [A-Z]" src/data/*.go | grep -v _test | wc -l` → **157**.
Signatures are **not uniform** — the interface mirrors them *verbatim*, including the legacy oddities (no `ctx`
on `catalog_managers.go:16/:23`, `vm_snapshots.go:7/:25`; no-`error` on `GetOrchestratorAvailableResources`
`orchestrator.go:398` etc.; `GetJobs(ctx)` without filter, `GetOrchestratorReverseProxyHosts` returning
`[]*models.ReverseProxyHost` `orchestrator.go:630`).

`src/data/interfaces/` files (compose into `database_service.go`):

| Interface file | Members (→ source file) |
|---|---|
| `users.go` | 11 → `data/users.go` |
| `roles.go` / `claims.go` | 9 → `data/roles.go` / 6 → `data/claims.go` |
| `api_keys.go` | 8 → `data/api_key.go` |
| `catalog.go` / `catalog_managers.go` | 25 → `data/catalog.go` / 5 → `data/catalog_managers.go` |
| `orchestrator.go` | 31 → `data/orchestrator.go` |
| `snapshots.go` | 2 → `data/vm_snapshots.go` |
| `jobs.go` | 10 → `data/jobs.go` |
| `reverse_proxy.go` | 16 → `data/reverse_proxy.go` |
| `user_config.go` | 6 → `data/user_config.go` |
| `config.go` / `schema.go` | 4 → `data/config.go` / 2 → `data/schema.go` |
| `packer_templates.go` | 5 → `data/packer_template.go` |
| `enrollment_tokens.go` | 5 → `data/orchestrator_enrollment_token.go` |
| `lifecycle.go` | 12 → `data/main.go` (Connect :127, Load :133, Disconnect :182, Filename :188, IsConnected :192, SaveAs :196, SaveAsync :219, SaveNow :252, ProcessSaveQueue :261, IsDataFileEmpty :617) + `data/backup.go` (RunBackup :15, Backup :29) |

Total 157. The 12 lifecycle members are in the interface because providers/callers invoke them
(`src/main.go:58-59` `SaveNow/SaveAs`, `src/catalog/import.go:204`, seeds) — see D-7 for their SQL semantics.
`interfaces/database_context.go` stays untouched (verified: zero references outside its own file).

Compile-time assertions (R-1), added in task 01, `src/data/interfaces_compilation_test.go`:

```go
var _ interfaces.DatabaseService = (*data.JsonDatabase)(nil)   // task 01
// task 04 adds (in src/database): var _ interfaces.DatabaseService = (*GormDatabase)(nil)
```

### D-3 — Error sentinels move to `interfaces`, aliases stay in `data` (identity is load-bearing)

Controllers branch by HTTP code (`rsp.Code` from `*errors.SystemError`, e.g. `src/controllers/users.go:258`), but
**pointer identity** is relied upon in startup code: `src/startup/migrations/add_api_key_own_claims.go:61`
(`err != data.ErrUserAlreadyContainsClaim`), `src/startup/seeds/role_claims.go:39,43`
(`err == data.ErrRoleAlreadyContainsClaim`, `err == data.ErrClaimNotFound`),
`src/startup/seeds/packertemplates/*.go:44-47` (`.Error()` string compare). `GormDatabase` must therefore return
**the identical sentinel values** — but per AD-1 `database` must not import `data`.

**Decision:** relocate every `Err*` sentinel var block from `src/data/<domain>.go` into
`src/data/interfaces/errors_<domain>.go` (same `errors.NewWithCode(msg, code)` calls), and in `data` keep
one-line aliases (`var ErrUserNotFound = interfaces.ErrUserNotFound`, `data/users.go:14-27` → alias block).
`data → interfaces` is acyclic (interfaces imports only `basecontext`, `models`, `errors`). Zero caller changes;
identity preserved across engines. GORM error → sentinel mapping per domain lives in each domain task (D-11).

### D-4 — Query language port: exact-semantics Go evaluation first, gated SQL push-down later

The four layer files (read fully, verified) implement semantics that SQL `WHERE` cannot reproduce wholesale:

| Quirk (verified) | Where |
|---|---|
| Filter **value is an unanchored Go regexp**, not equality: `regexp.Compile(value)` → `MatchString` | `filter_layer.go:81-86` |
| `i` option wraps the *pattern* in `(?i)` (repeated options double-prefix `(?i)(?i)` — legal, must not "fix") | `filter_layer.go:73-78` |
| Invalid filter strings error: exactly 2 `=`-split parts; options after `,` other than `i` silently ignored | `filter_layer.go:27-56` |
| Property resolution: case-insensitive Go field name **or json-tag name**; dotted nested paths recurse; `map[string]string` direct index; other maps scan keys EqualFold | `filter_layer.go:109-161` |
| Non-string scalars compared via `fmt.Sprintf("%v", …)` (so `[]string` matches against `"[a b]"`, pointers against `<nil>`/address text); struct-kind properties never match | `filter_layer.go:86,163-166` |
| Ordering: `sort.Slice` on `.FieldByName(prop).String()` — non-string fields sort as `""`, ties keep input order (unstable sort) | `order_layer.go:32-55` |
| Order is applied to only 3 call sites, all `UpdatedAt desc`: `catalog.go:35`, `orchestrator.go:50,85` | grep-verified |
| `GetRecordIndex`: string → `EqualFold` first-match; **bool/numeric types always return -1** (typed-switch dead code) | `index_layer.go:16-31` |
| Authorization: root/api-key/microservice bypass **unless** injected claims/roles present; super-user bypass; empty-required-roles ⇒ public; role AND claim groups must each have ≥1 match | `authorize_layer.go:39-100` |
| `GetOwnRecords` filters by `GetUserID() == authContext.User.ID` (skips when no auth ctx) | `api_key.go:19-32` |

**Baseline (all domain tasks):** load through GORM (`Find` in rowid/insertion order), then evaluate the *ported*
filter/order/authz functions in Go — legacy itself scans the whole in-memory slice per call (`users.go:39`), so
complexity parity holds and deep-equal parity with the legacy engine is achieved **by construction**, which is
R-3's mitigation taken literally. The port is a transcription, not a rewrite.
**Push-down (per-domain opt-in, only after parity tests are green):** SQL `WHERE`/`ORDER BY`/`LIMIT` allowed only
for shapes proven equivalent by a dedicated differential test in the same PR (e.g. id/unique-key lookups
`WHERE id = ?` replacing `GetRecordIndex` scans — safe because generated ids are lowercase, `helpers.GenerateId`;
`UpdatedAt DESC` secondary-sort by rowid to mimic legacy tie order). Never push down regexp semantics into
SQLite/Postgres (no REGEXP UDF decision is taken here; Postgres `~` differs from RE2).

### D-5 — Differential parity harness (`src/database/parity/`)

```go
// Legacy and SQL engine side-by-side over identical seeds, deep-equal list outputs.
func NewLegacyEngine(t testing.TB) *data.JsonDatabase          // temp dir + data.ResetForTesting() (data/testing_reset.go:7),
                                                               // NewJsonDatabase pattern from data/main_test.go setupTestDB
func NewSQLEngine(t testing.TB, legacyFallback bool) *database.GormDatabase  // :memory: via real factory
func SeedBoth(t testing.TB, seed func(db interfaces.DatabaseService))        // seed via exported API only
func AssertSameList[E any](t testing.TB, name string, want, got []E)         // reflect.DeepEqual + typed diff print
func RunMatrix(t testing.TB, domains []DomainSuite)              // per-domain: Get*(filters) × order/authz cases
```

- Test-only exception to AD-1's import rule, documented in the package doc: `parity` imports `data` + `database`;
  nothing in production code imports it.
- Each domain task adds one `DomainSuite` and MUST run green against `//go:build` default (in-memory SQLite, no
  external services) inside the ≤30 s budget (#425 owns budget enforcement).
- Seeds are byte-identical because both engines are driven through the same `DatabaseService` calls.

### D-6 — Three-step provider cutover (each step independently shippable, ER-5)

`src/serviceprovider/data.go:8` currently returns `(*data.JsonDatabase, error)` and calls `Connect` each time;
~143 non-test `GetDatabaseService` call sites plus ~24 direct `serviceprovider.Get().JsonDatabase` reads.

| Step | Task | Change | Consumer ripple |
|---|---|---|---|
| S1 field | 01 | Provider struct: `JsonDatabase` **retyped** to `interfaces.DatabaseService`; new field keeps concrete handle in `s.legacyDatabase`; new accessors `Legacy() *data.JsonDatabase` and `GetSQLDatabase` already from #418; `GetDatabaseService` keeps its concrete signature, returning `s.legacyDatabase` | **zero** (assertion inside provider; locals like `db := serviceprovider.Get().JsonDatabase` now hold the interface and keep compiling because every method they call is on it) |
| S2 API | 17 | `GetDatabaseService` returns `interfaces.DatabaseService`; the handful of struct fields/params typed `*data.JsonDatabase` are retyped (list in task 17); `ParallelsService.SetDatabaseService(globalProvider.legacyDatabase)` injection keeps concrete type via provider-side assertion | local-variable call sites unchanged; ~10 declarations retyped |
| S3 select | 18 | Selection by `FromApplicationConfig().Type`: empty/`legacy` → `JsonDatabase` (today's default until this task, per ER-5); `sqlite`/`postgres` → `GormDatabase` built on the #418 connection. Legacy goroutines exist **only** in legacy mode | see D-7/D-8/audit |

### D-7 — Lifecycle semantics in SQL mode (escape-hatch table)

| Method | Legacy | GormDatabase (task 04 stub → task 18 final) |
|---|---|---|
| `Connect`/`IsConnected` | bool flag | ping-backed (`g.gormDB` non-nil + `Ping()` cached); keeps `GetDatabaseService`'s per-call `Connect` cheap |
| `Disconnect` | no-op | stop watcher, `database.Close` via `closeOnce` |
| `Load` | file load | no-op `nil` (SQL is the store; #424 migrates) |
| `SaveNow`/`ProcessSaveQueue`/`SaveAsync` | whole-file rewrite | **no-op returning nil** — SQL commits per statement; all existing `SaveNow` callers (`main.go:58,128`, seeds ×3, `catalog/import.go:204`, `import_vm.go:132`, `register_with_orchestrator.go:171`, `orchestrator/register_host.go:59`, `controllers/performance.go:63`) keep compiling and become harmless |
| `SaveAs(ctx, name)` | snapshot copy | `VACUUM INTO <abs-path>` (pure-Go modernc supports it; task 18 proves it) — keeps **`src/main.go:59` panic-snapshot policy intact on SQL** |
| `Filename` | data.json path | the SQLite path (`:memory:` → `":memory:"`) |
| `IsDataFileEmpty` | file probe | `false, nil` (SQL emptiness is #424's detection) |
| `Backup`/`RunBackup` | rotation of data.json | log-once warn "SQL backups are operated-managed (pg_dump / file copy); legacy rotation inactive" + nil. Documented gap → #427 docs |
| `SchemaVersion`/`UpdateSchemaVersion` | JSON `Data.Schema` | single-row `schema_metadata` table (mirrors #419's OQ-3 default naming; #424's completion marker is a *separate* table — AD-4 respected). Legacy schema migrators still run **only in legacy mode** (task 18 gate in `src/startup/main.go:128` loop + `migrations/service.go` nil-guard) |
| Seeds (`startup/seeds/*`) | mutate + SaveNow | retyped to `DatabaseService` (task 17); in SQL mode they insert via GORM, SaveNow no-ops, existence checks via `Get*` make them no-ops after first run (same idempotence contract as today) |

**ParallelsDesktop exception:** `parallelsdesktop/main.go:103,168` holds a concrete `*data.JsonDatabase` solely
for `Set/GetListVMSnapshotsByVMId`; the file is 3,079 lines and feature-plan-declared untouchable. In SQL mode the
provider-side assertion yields nil and the service's existing nil-guards (`:690,705,722`) engage. Recorded as a
known gap + open question OQ-421-2 (retyping it belongs with its own ER-3 split, follow-up issue — not folded here).

### D-8 — Stale-job detection in SQL mode (jobs domain special case)

Legacy hosts the goroutine in the engine singleton (`main.go:109-122`, ticker
`constants.GhostJobCheckIntervalSeconds` = 5 s, `src/constants/jobs.go:45`; timeout from
`config.DbGhostJobTimeoutMinutes()`, `jobs.go:221-222`). Decision: **detection logic is a domain method**
(`GormDatabase.DetectStaleJobs` — SELECT pending/running, RFC3339Nano parse of `UpdatedAt`, batch UPDATE +
handler callbacks mirroring `jobs.go:216-253`), and the **ticker lives in `src/database/jobs_watcher.go`** —
`func (g *GormDatabase) StartStaleJobWatcher(ctx)` with its own `stopWatch` channel, started by the provider in
SQL mode (task 18), stopped on `Disconnect`. `SetJobTimeoutHandler` stored under RWMutex on `GormDatabase`
(caller today: `src/jobs/main.go:63`). `RecoverOngoingJobs` mirrors `jobs.go:183-204`; provider calls it once at
SQL-mode bootstrap (legacy calls it only from `loadFromEmpty` — `main.go:578` — i.e. first-run; task 15 spec
pins this and task 18 documents the one deliberate behavioural delta: fail-ongoing-jobs also after restart with
pre-existing rows, which is strictly safer).

### D-9 — Transaction inventory (AC "Transactions")

`db.Transaction(func(tx *gorm.DB) error)` sites, from the multi-write flows found by grep/read:

| Flow | Legacy evidence | Why atomic |
|---|---|---|
| `CreateUser` default-role/claim materialisation | `users.go:111-154` | user + join rows all-or-nothing |
| `AddRoleToUser`/`RemoveRoleFromUser`, `AddClaimToUser`/`RemoveClaimFromUser` | `users.go:268-422` | dedupe-check + join write |
| `AddClaimToRole`/`RemoveClaimFromRole` | `roles.go:126-176` | role_claims write |
| `DeleteClaim` cascade (claims row + user_claims + role_claims) | `claims.go:112-161` (in-place cascade) | 3 tables |
| `DeleteRole` guard-scan + delete | `roles.go:177-208` | read-of-usage + delete race window |
| `CreateOrchestratorHost` + nested VMs / `ReplaceOrchestratorHostVMs` | `orchestrator.go:121-146, 871-888` | host + vm rows (full-sync semantics) |
| `SetHostVMSnapshots` upsert (host row + snapshot map) | `orchestrator.go:907-933` | insert-or-update |
| `UpdateReverseProxyHostHttpRoute` + `ReorderReverseProxyHostHttpRoute` | `reverse_proxy.go:312-394` (route reorder rewrites all orders) | ordering invariant |
| `CreateApiKey` (+ internal-key flows), `DeleteCatalogManifest*` cascades | `api_key.go:157-188`; `catalog.go:216-320` | multi-row deletes |
| `DetectStaleJobs` batch fail | `jobs.go:229-247` | consistent timeout sweep |
| Non-atomic-in-legacy, stays non-atomic (parity): `CreateClaim`, `UpdateJob`, seed paths | — | record in PR description |

Rule: one transaction per public method that legacy performs under one `dataMutex.Lock()` (grep the
`dataMutex.Lock()` sites — they are the atomicity contract of the legacy engine).

### D-10 — ER-3 exception request (needs human approval, flagged in task 17)

Task 17 must retype declarations living in files already ≥1,000 lines: `src/reverse_proxy/main.go` (1,302 —
field `db` :79 + constructor param). ER-3 would first split it. Because the edits are 2–3 purely-typal lines and
the feature plan already excludes adjacent mega-files from this feature (`plan.md §6`), we propose: **exception
for type-level-only edits in over-cap files** (mirrors the "not our debt" precedent), or — if the reviewer
prefers strict ER-3 — task 17 adds a same-PR pure-move split of `reverse_proxy/main.go`. Controller files
(4,471/2,599/2,148/1,399) need **no edits**: their `db := serviceprovider.Get().JsonDatabase` locals adopt the
interface automatically (verified locals-only at `controllers/orchestrator.go:835,3934,3968,4456`,
`machines.go:2116`).

### D-11 — Error mapping (AC "Error handling")

| GORM outcome | Translation | Pattern |
|---|---|---|
| `gorm.ErrRecordNotFound` | the domain's `interfaces.Err*NotFound` (identity, D-3) | `errors.Is(err, gorm.ErrRecordNotFound)` first |
| `ErrDuplicatedKey` / driver constraint error (`UNIQUE constraint failed`, pg 23505) | domain `Err*AlreadyExists` where legacy has one, else `errors.NewWithCode(msg, 400)` | pre-check lookups kept (legacy checks first — `users.go:99-109`), constraints are the race-window backstop |
| everything else | `errors.NewWithCode(fmt.Sprintf("…: %v", err), 500)` (never `%w` into response text; message shape mirrors `data/main.go` idioms) | |
| not-connected | `interfaces.ErrDatabaseNotConnected` | every method's first guard, mirroring legacy's `if !j.IsConnected()` |

Mapping table is per-domain (each domain task carries the exact rows); shared helper
`src/database/errors.go: mapGormError(err error, notFound, dup error) error` lands in task 04.

### D-12 — Schema ownership (reference, don't re-decide)

Every domain task lists *its* tables **as defined by #419's mapping doc** (`../419-gorm-model-tags/plan.md`,
OQ-3 default: GORM snake_case plural + join tables for `User.Roles/Claims`, `Role.Claims`; `[]string` fields like
`Tags`/`RequiredRoles` per #419's per-field decision — serialised JSON column or join table; nested structures
(`OrchestratorHost.VirtualMachines`, `ReverseProxyHost.HttpRoutes`, `CatalogManifest` contents, `Job.Steps`,
`HostResources`, `map[string][]VMSnapshot`) likewise resolved there). Where #419 chose serialised-JSON columns,
filters/orderings on such paths are **evaluated Go-side only** (D-4 has no SQL dependency on the choice). If a
needed table/field is missing from the mapping doc, that is a #419 amendment, blocking note in the PR — not a
local decision.

## Task index (each task = one PR; merge in dependency order)

| Task | Goal | Depends | Touches (key) | Size |
|---|---|---|---|---|
| [01](./tasks/01-interface-capture.md) | `DatabaseService` (157, per-domain files) + sentinel relocation + assertions + provider field/`Legacy()` (S1) | 418/420/419 merged | `src/data/interfaces/*`, `src/data/*.go` alias blocks, `serviceprovider/data.go`, `serviceprovider/main.go` | M |
| [02](./tasks/02-query-language-filter-order-index.md) | Port filter/order/index semantics into `database/query` + golden tests | 01 | `database/query/{filter,order,index}.go`, tests | M |
| [03](./tasks/03-query-language-authorize-harness.md) | Port authorize layer + `GetOwnRecords`; differential harness `database/parity` | 02 | `query/authz.go`, `database/parity/*`, tests | M |
| [04](./tasks/04-gormdatabase-bootstrap.md) | `GormDatabase` type, all-domain delegate stubs, compile assertion, `mapGormError` | 03 | `database/gorm_database.go`, `database/errors.go`, `database/delegate_*.go` | M |
| [05](./tasks/05-domain-roles-claims.md) | Roles+claims on GORM (internal guards, cascade delete, tx) | 04 | `database/roles.go`, `database/claims.go`, parity suite | M |
| [06](./tasks/06-domain-users.md) | Users on GORM (join tables, role/claim assign tx, rehydrate claims) | 05 | `database/users.go`, parity suite | M |
| [07](./tasks/07-domain-api-keys.md) | API keys on GORM (`GetUserByApiKey`, internal/external) | 06 | `database/api_key.go` | S |
| [08](./tasks/08-domain-config-schema.md) | Configuration singleton + Id seeding + schema-version row | 04 | `database/config.go`, `database/schema.go` | S |
| [09](./tasks/09-domain-user-configs.md) | User configs on GORM | 06 (authz patterns) | `database/user_config.go` | S |
| [10](./tasks/10-domain-packer-templates.md) | Packer templates on GORM | 03 | `database/packer_template.go` | S |
| [11](./tasks/11-domain-catalog.md) | Catalog manifests + catalog managers (30 methods, tags/claims/roles arrays, taint/revoke) | 06 | `database/catalog.go`, `database/catalog_managers.go` | L |
| [12](./tasks/12-domain-reverse-proxy.md) | Reverse proxy config/hosts/routes (reorder invariant) | 06 | `database/reverse_proxy.go` (+`_routes.go` if >200 lines est.) | M |
| [13](./tasks/13-domain-orchestrator-hosts.md) | Orchestrator hosts CRUD + resource aggregations + timestamp/ws status | 11 (VM dep order noted) | `database/orchestrator.go` (hosts part) | M |
| [14](./tasks/14-domain-orchestrator-vms-snapshots.md) | Host VMs, host RP views, `HostsVMSnapshotsRecord`, `VMSnapshots` collection | 13 | `database/orchestrator_vm.go`, `database/snapshots.go` | M |
| [15](./tasks/15-domain-jobs.md) | Jobs on GORM + stale-job watcher function + timeout handler | 06 | `database/jobs.go`, `database/jobs_watcher.go` | M |
| [16](./tasks/16-domain-enrollment-tokens.md) | Enrollment tokens on GORM (TTL expiry) | 03 | `database/enrollment_token.go` | S |
| [17](./tasks/17-call-site-conversion.md) | S2: `GetDatabaseService` → `DatabaseService`; retype concrete fields/params; lifecycle no-ops proven | 04..16 **all ported** | `serviceprovider/data.go`, `jobs/main.go`, `reverse_proxy/main.go`, `orchestrator/main.go`, `startup/*`, `catalog/import_vm.go`, `src/main.go` | M (D-10) |
| [18](./tasks/18-final-flip.md) | S3: `DB_TYPE` selection, legacy goroutines off in SQL mode, panic-snapshot `VACUUM INTO`, watcher + RecoverOngoingJobs wiring, WAL check, postgres enable, bypass audit | 17 | `serviceprovider/main.go`, `serviceprovider/database.go`, `src/main.go`, `startup/main.go`, docs-in-code | L |

**Dependency notes:** 05→06→(07,09,11,12,13,15) is a hard chain (users depends on roles/claims seeding reads;
api-keys/jobs depend on users lookups being GORM-side before their parity suites run clean). 08/10/16 are
independent leaves (parallelizable with 05/06). 13→14 (VM tables live with host tables from #419). No two
unfinished tasks touch the same new file (domain files are 1:1). Task 17/18 are deliberately last and serial.

## AC coverage map (PBI brief checkboxes)

| AC | Moved by |
|---|---|
| All CRUD operations … via GORM verified against SQLite | 05–16 (every method; parity harness per domain + CRUD matrix tests per task) |
| File path respected (config path, legacy default) | 18 (uses #420 `FromApplicationConfig` + #418 AD-3 default; provider opens `GormDatabase` on that path; test asserts file materialises at `DATABASE_FOLDER`/`DB_PATH`) |
| Transactions | D-9; concretely in 05, 06, 07, 11, 12, 13, 14, 15 (each spec's tx inventory) |
| Concurrent access (WAL) | #418 task 03 applied pragmas; verified end-to-end in 18 (`PRAGMA journal_mode` over the *selected* provider handle) + 04 constructor test |
| Error handling (wrapped, consistent) | D-3/D-11; 04 (helper) + each domain task's mapping table |
| Legacy layer bypassed (SQL configured) | 17 (no concrete-typed holders), 18 (grep audit in PR: `serviceprovider.Get().JsonDatabase` outside provider, `.legacy`, `SaveNow` reachability) + D-7/D-8 goroutine shutdown + D-10/OQ-2 exception documented |

## Verification commands (every task)

```bash
cd src
go build ./... && go vet ./data/... ./database/... ./serviceprovider/...
go test ./database/... -count=1 -race      # new package tests
go test ./... -count=1                     # full suite green at every merge boundary (ER-5)
```

## Risks specific to #421

| Risk | Mitigation |
|---|---|
| R-3 interface/query drift | 01 compiles the full surface first; D-4 port-by-transcription; 03 harness gates every domain flip |
| Filter quirks accidentally "fixed" (e.g. regexp→LIKE) | golden tests in 02 pin each listed quirk line-by-line; PR checklist: "no semantic improvement" |
| Sort tie order divergence | SQL loads in rowid order; Go-side sort port keeps input-order ties; harness compares whole slices |
| Delegate stubs rot between creation and domain flip | task 04 stubs are literal `return g.legacy.X(...)` one-liners — a domain task deletes, never edits them |
| `dataMutex`-based atomicity missed | D-9 rule: grep `dataMutex.Lock()` before porting each domain (listed per task) |
| Parity suite slowing CI | in-memory SQLite only; #425 owns budget; suites are table-driven |
| Flip-day regressions | 17/18 separate; 18 keeps `DB_TYPE=legacy` escape hatch so rollback = env var, not revert |

## Open questions (human review, none blocking task 01–16)

| ID | Question | Proposal carried by this plan |
|---|---|---|
| OQ-421-1 | Task-17 ER-3 exception vs. splitting `reverse_proxy/main.go` (D-10) | exception for type-only edits |
| OQ-421-2 | ParallelsDesktop vm-snapshot cache stays legacy-concrete (D-7) until that 3,079-line file's own refactor | follow-up issue; nil-guard gap documented |
| OQ-421-3 | SQL-mode backups (rotation replaced by operator guidance) | log-once warn here, docs in #427 |
| OQ-421-4 | `RecoverOngoingJobs` on every SQL start vs legacy first-run-only (D-8) | always-on (safer); confirm with PO |
| OQ-421-5 | #419 mapping doc still to be authored (round 3 precedes this round per loop §8 — must land before tasks 05+) | normative reference D-12; if a field lacks a mapping, task PR is blocked with a #419 amendment note |

## Handover

- **→ #424**: `GormDatabase.DB()` + complete `DatabaseService` after task 16; marker table separate from `schema_metadata` (AD-4); never call legacy SaveNow in SQL mode (no-op anyway).
- **→ #425**: `parity.RunMatrix` + `AssertSameList` are the extension points; per-domain golden fixtures live next to `database/parity/testdata`.
- **→ #426**: after task 18, `Legacy()` has exactly the callers enumerated in 18's audit list — annotate those call sites' comments.
- **→ #427**: config table additions (`DB_TYPE=legacy` undocumented fallback), SQL backup guidance, OQ-421-3 text.
