# Task 03 — depguard rule + dedicated CI guard job (the *live* no-new-callers lock)

**PBI** [#426](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01 (parallel with 02) · **PR size:** M
**Plan refs:** D-4 (verified YAML, dead-linter finding, `--enable-only` posture) · D-5 (nolint audit table) · D-1 (package-doc SA1019 double-lock)

## Context

D-4's field investigation reshaped this task: the repo's `.golangci.yml` is **v2 dialect** (107 lines,
`version: "2"`) but `.github/workflows/linter.yml` is **dead** (`if: false`, super-linter with
`VALIDATE_GO: false`, Go 1.21, wrong `go.sum` path) — so "linter rule **or CI check**" cannot ride existing
infrastructure. The solution splits: the rule lands in canonical `.golangci.yml` (devs/IDEs/future lint
inherit it; devs must run `golangci-lint v2.14.0`), and a **new minimal workflow** runs exactly that one
linter with `--enable-only=depguard` so CI fails **only** on new `data` imports — never on pre-existing lint
debt. The deny pattern is the empirically-verified `github.com/Parallels/prl-devops-service/data$` — the `$`
anchors to the engine package and provably ignores `/data/models` and `/data/interfaces` (live imports).

## Objective

New import of `…/data` ⇒ red CI in both directions (verified with a scratch importer); today's importers
pass via a **generated** grandfather exclusion list, each entry carrying a `#426 grandfather` comment; the
D-5 nolint audit closes any IDE-side leak outside the grandfather paths.

## Files

| File | Op |
|---|---|
| `.golangci.yml` | modify — depguard rule + exclusions block |
| `.github/workflows/go-legacy-guard.yml` | create (~30 lines; Risk-9 ownership: 03 owns CI files in this PBI) |
| *(per D-5)* scattered `//nolint:staticcheck` one-liners | only for non-grandfathered deprecated-symbol references the flip leaves behind |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `.golangci.yml` | 107 | ~130 | ok (non-Go file; ER-2's spirit respected) |
| `go-legacy-guard.yml` | new | ~30 | ok |

## Implementation spec

1. **Rule (D-4, append at the v2-dialect-correct position — read the file's structure first; `linters.`**
   section already exists):**

```yaml
  depguard:
    settings:
      main:
        desc: "PBI #426 — no new importers of the deprecated legacy data engine (removal issue links here)"
        files: ["**/*"]
        deny:
          - pkg: github.com/Parallels/prl-devops-service/data$
            desc: "legacy JSON data layer is deprecated (#426): use serviceprovider.GetDatabaseService()/interfaces.DatabaseService; see removal issue"
```

2. **Grandfather exclusions (generated):** audit command in PR body:

```bash
grep -rln '"github.com/Parallels/prl-devops-service/data"' src --include='*.go' \
  | grep -v '^src/data/' | sed 's|^src/||; s|/[^/]*\.go$|/**|' | sort -u
```

   → one `exclusions.rules` entry per surviving path pattern (each `#426 grandfather — shrink via removal
   issue`), **plus** the always-list: `src/data/**` self-exclusion. Post-#421/#425 the list should be much
   shorter than D-4's today-baseline (`serviceprovider/**`, `main.go`, `catalog/`, `cmd/`, `jobs/`,
   `orchestrator/`, `reverse_proxy/`, `startup/migrations/**`, `startup/seeds/**`, legacy test dirs,
   `database/legacy_migrator/**`) — ship what the command says, not what D-4 predicted; PR notes the delta.
3. **CI workflow (v2.14.0 pinned):**

```yaml
name: legacy-import-guard
on: { pull_request: { paths: ['src/**', '.golangci.yml', '.github/workflows/go-legacy-guard.yml'] } }
jobs:
  guard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: src/go.mod }
      - uses: golangci/golangci-lint-action@v6
        with: { version: v2.14.0, working-directory: src, args: >-
            -c ../.golangci.yml --enable-only=depguard ./... }
```

   (Action v6 vs v7/v8: match what repo pins elsewhere if any — none exists; v6 supports v2 config, verify
   at task start and bump consciously.)
4. **D-5 nolint audit:** run its enumerate command; every file **outside** the grandfather list that still
   references a deprecated symbol gets the targeted line-level
   `//nolint:staticcheck // #426: legacy escape hatch — remove with the legacy layer (removal issue #<n>)`
   (issue number from task 04 if merged first, else `TBD-04` fixed by 04's PR — coordinate; never
   file-level nolint); parallelsdesktop explicitly gets **no** sprinkle (3,079-line ER-3 exclusion).

## Out of scope

Touching `linter.yml` (dead file — removal is repo hygiene, not this PR), full-lint enablement (its own
battle), shrinking grandfathers (removal issue's job), `// Deprecated:` text (01), warns (02).

## Validation

- [ ] **Gate 1 (clean):** `cd src && golangci-lint run -c ../.golangci.yml --enable-only=depguard ./...` → exit 0, zero issues (v2.14.0 — the version D-4 verified; record `golangci-lint --version` in PR)
- [ ] **Gate 2 (tripwire):** scratch file `src/tmp_guard_check.go` importing `…/data` + one live symbol → same command → exit 1, message contains the rule `desc`; file deleted after (PR pastes both runs)
- [ ] Negative space: a scratch file importing `…/data/models` passes (the `$` anchor proof)
- [ ] CI green on this PR itself; workflow file linted (actionlint or yamllint per repo convention)
- [ ] `cd src && go build ./... && go test ./... -count=1` (ER-5 — nolint comments only, zero logic)
- [ ] Grandfather list in PR == the generated command output (reviewer can re-run it)

## Rollback

Revert: rule + workflow + nolints vanish; nothing behavioural; grandfather list regenerates on next attempt.

## PBI acceptance mapping

**No new callers** AC — CI half (the CI-check option the AC explicitly allows, on infrastructure that
actually runs); IDE half via SA1019 from 01 + the D-1 import-line flag forms the double-lock; every
exclusion entry is a dated, linked debt item the removal issue tracks to zero.
