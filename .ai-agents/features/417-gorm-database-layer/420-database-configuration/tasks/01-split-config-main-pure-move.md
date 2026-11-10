# Task 01 — Pure-move split of `src/config/main.go` (ER-3 unblocker)

**PBI** [#420](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** #418 merged (delivery order only — this task touches no #418 file) · **PR size:** S (mechanical; wide diff)
**Plan refs:** [D-6](./plan.md#d-6--er-3-split-of-srcconfigmaingofour-cohesive-files-maingo-core-measured) · feature plan ER-3 (trigger table lists this exact file)

## Context

`src/config/main.go` measures **1,013 lines** (re-measured 2026-11-03: `wc -l src/config/main.go`)
— already over the ER-2 cap of 1,000. Every later #420 task edits this package, and ER-3 forbids
adding feature code to an over-cap file before a split. This task performs that split **first**, as
a pure code move with zero behavior change, so tasks 02+ never mix a refactor with feature diff.
It also creates `database.go`, the cohesive home where task 02 adds the new DB getters.

## Objective

Same package, same exported API, identical behavior; `main.go` drops to ~370 lines; four new
same-package files hold the moved clusters; `go build ./...` and the full test suite green at the
merge boundary (ER-5). No new functionality of any kind.

## Files

| File | Op |
|---|---|
| `src/config/main.go` | modify — function-group removals only (1,013 → ~370) |
| `src/config/database.go` | create — 6 DB-related getters (moved) |
| `src/config/modules.go` | create — module/command-mode getters (moved) |
| `src/config/network.go` | create — API/TLS/CORS/reverse-proxy/Localhost getters (moved) |
| `src/config/platform.go` | create — root/cache/system/orchestrator/misc feature flags (moved) |

Untouched on purpose: `src/config/models.go` (72 lines — verified; no split needed, no edits),
`main_test.go`, `modules_test.go`, `job_timeout_test.go` (same-package tests compile unchanged).

## File size guard

| File | wc -l now (measured 2026-11-03) | After split | Verdict |
|---|---|---|---|
| `src/config/main.go` | **1,013** | ~370 | **refactor-first** → executed by this task |
| `src/config/database.go` | n/a (new) | ~70 | ok |
| `src/config/modules.go` | n/a (new) | ~215 | ok |
| `src/config/network.go` | n/a (new) | ~196 | ok |
| `src/config/platform.go` | n/a (new) | ~220 | ok |
| `src/config/models.go` | 72 | 72 | ok (not touched) |

Body-line measurements come from function extents in the current file (`GetEnabledModules` alone is
137 lines, main.go:760; totals in [plan D-6](./plan.md#d-6--er-3-split-of-srcconfigmaingofour-cohesive-files-maingo-core-measured)).

## Implementation spec

**Pure move rules:** copy function bodies verbatim (no renames, no comment edits, no signature
changes, no reordering within a file beyond grouping); move each function to exactly one new file;
redistribute the shared import block (`src/config/main.go:1-27`) — each new file imports only what
its functions use (run `goimports` or compile-fix; e.g. `crypto/tls`, `crypto/x509`, `encoding/pem`
land in `network.go`; `runtime` in `platform.go` for `RootFolder`; `time` in several). File header
comment on each new file: `// Cohesive getter groups moved from main.go (pure move, no behaviour
change) — see .ai-agents/features/417-gorm-database-layer/420-database-configuration/plan.md D-6`
(one line, aids blame navigation; required by the ER-3 "review sees the move cleanly" intent).

### `database.go` — move exactly these functions (current line anchors, re-grep before editing)

`DbNumberBackupFiles` (:401) · `DbBackupInterval` (:410) · `DbSaveInterval` (:419) ·
`DbGhostJobTimeoutMinutes` (:432) · `IsDatabaseAutoRecover` (:525) · `DatabaseFolder` (:598)

### `modules.go` — move

`IsOrchestrator` (:614) · `IsCatalog` (:618) · `IsCatalogManager` (:622) · `IsHost` (:626) ·
`IsApi` (:630) · `UseOrchestratorResources` (:634) · `GetEnabledModules` (:760, 137 lines) ·
`IsModuleEnabled` (:896) · `DisableModule` (:907) · `EnableModule` (:921)

### `network.go` — move

`ApiPort` (:209) · `ApiPrefix` (:219) · `IsCorsEnabled` (:255) · `TlsCertificate` (:284) ·
`TlsPrivateKey` (:295) · `TlsPort` (:307) · `TlsEnabled` (:325) · `ValidateTlsConfiguration` (:338) ·
`ShouldDisableHttpWhenTls` (:383) · `DisableTlsValidation` (:654) · `IsReverseProxyEnable` (:660) ·
`IsReverseProxyEnabled` (:669) · `ReverseProxyHost` (:673) · `ReverseProxyPort` (:682) ·
`GetReverseProxyConfig` (:691) · `Localhost` (:602)

### `platform.go` — move

`RootFolder` (:440) · `CatalogCacheFolder` (:471) · `IsCatalogCachingEnable` (:490) ·
`CacheKeepFreeDiskSpace` (:499) · `CacheMaxSize` (:508) · `AllowCacheAboveFreeDiskSpace` (:521) ·
`Mode` (:540) · `SystemReservedCpu` (:549) · `SystemReservedMemory` (:558) ·
`SystemReservedDisk` (:567) · `OrchestratorPullFrequency` (:576) · `OrchestratorPublicUrl` (:590) ·
`IsRemoteProviderStreamEnabled` (:695) · `IsCanaryEnabled` (:704) · `CatalogCompressVM` (:711) ·
`CatalogCompressRatio` (:716) · `IsBetaEnabled` (:728) · `IsCacheRefreshEnabled` (:735) ·
`ForceCacheRefreshInterval` (:743) · `EnableCredentialsObfuscation` (:751)

### `main.go` keeps (core)

`Config` struct (:76) + `globalConfig`/`extensions` vars · `findConfigFile`/`resolveConfigFile` ·
`New`/`Get` · `Load`/`Save`/`Refresh`/`CleanReverseProxyDataFromConfig` · `LogLevel` ·
`EncryptionPrivateKey`/`RootPassword`/`Source` · `SetRunningCommand`/`GetRunningCommand` ·
`GetKey` (:933) / `GetIntKey` / `GetBoolKey` / `SetKey` / `SetIntKey` / `SetBoolKey` ·
`GetConfig`/`GetFilename`/`GetConfigContent`.

**Commit discipline (ER-3):** one commit, message `refactor(config): split main.go into cohesive
same-package files (pure move) [#420]`. `git show --stat` should show 5 files, ~equal +/− line
counts. No second commit — this PR contains no feature work.

## Out of scope

Any new getter, constant, `DB_*` name, behavior tweak, comment fix, import tidy beyond what the
move forces, test changes, `models.go` split (it is 72 lines — the feature plan's pre-seeded 1,085
figure is stale; see plan OQ-420-5).

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go vet ./config/...`
- [ ] `cd src && go test ./config/... -count=1 -v` — existing `main_test.go`/`modules_test.go`/`job_timeout_test.go` untouched and green
- [ ] `cd src && go test ./... -count=1` — repo-wide, no regressions (ER-5)
- [ ] `wc -l src/config/*.go` → `main.go` ≤ 400; every file < 1,000 (ER-2)
- [ ] `git diff --stat HEAD~1` → only the 5 `src/config/` files; `git diff HEAD~1 -- src/config/*_test.go src/config/models.go` → **empty**
- [ ] Behavioral spot-check: `git diff HEAD~1 -U0 src/config/main.go` shows deletions matching additions in new files line-for-line (pure move; `git show -w --stat` aids spotting whitespace-only drift)

## Rollback

`git revert` the single commit — the package is byte-equivalent to pre-task state; no consumers,
no config-format change, zero runtime risk. If anything else lands on top, revert the split alone
still applies (pure moves in one package).

## PBI acceptance mapping

Moves no AC checkbox directly; **unblocks all of them** (ER-3 gate for tasks 02–04, which are the
AC-bearing PRs). Feature-level: satisfies the ER-3 trigger the feature plan assigns to this PBI's
task chain.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `420-01`.** Cut `slice/420-01` from `sql-migration` HEAD
(`.worktrees/slice-420-01`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
