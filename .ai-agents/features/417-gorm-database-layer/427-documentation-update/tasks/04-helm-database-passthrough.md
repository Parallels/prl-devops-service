# Task 04 — Helm `database:` values passthrough (OQ-4) + values docs + Postgres-page paragraph

**PBI** [#427](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 03 (edits its page; menu conflicts rule) + #420 merged (the four settings exist) · **PR size:** S/M
**Plan refs:** D-4 (OQ-4 decision — minimal gated passthrough; verified template anatomy) · D-6 (validation toolchain) · D-2 (accuracy protocol — chart facts cited with line refs)

## Context

D-4's investigation found the chart **already stateless-ready**: env vars flow via an explicit list in
`deployment.yaml:68-259` sourced from ConfigMap/Secret, `.Values.envFrom` (`values.yaml:34-37` →
`deployment.yaml:256-259`) is a generic escape an operator can use *today*, and the PVC is gated on
`storage.databasePath` (`deployment.yaml:46-50, 282-285`) — setting it `""` drops persistence. What's
missing is **first-class values** for the four new settings. OQ-4 resolves minimal-and-additive: a
`database:` block, every template touch `{{- if }}`-gated with empty defaults, so **default renders stay
byte-identical** — that gate is the whole backward-compatibility argument and is checked mechanically.

## Objective

`helm upgrade` with a Postgres values block ships `DB_TYPE/DB_DSN` (dsn via Secret); `helm template` with
stock values is unchanged byte-for-byte; the chart's values doc gains the new block; the Postgres page gains
the k8s-secret snippet paragraph pointing here.

## Files

| File | Op |
|---|---|
| `helm/values.yaml` | modify — `database:` block (type/path/autoMigrate + `dsn: ""` + `dsnSecret: ""`) + comments |
| `helm/templates/config-map.yaml` | modify — `DB_TYPE`/`DB_PATH`/`DB_AUTO_MIGRATE` entries, each `{{- if }}`-gated |
| `helm/templates/secrets.yaml` | modify — `DB_DSN` key **only if** `.Values.database.dsn` non-empty; `dsnSecret` = operator-supplied existing-secret name |
| `helm/templates/deployment.yaml` | modify — env entries mirroring existing style (`:68-259` patterns), same gating; `dsnSecret` → `secretKeyRef` passthrough |
| `kubernetes/helm-chart.md` | modify — values-table rows for `database.*` |
| `docs/docs/devops/getting-started/postgresql.md` | append — "Kubernetes/Helm" section: values block example + note that `dsnSecret` keeps DSNs out of values files (cross-ref task 03's k8s-secret section — one owns prose, this appends the helm form) |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| all chart files | *(measure at task start; deployment.yaml large but gating adds ~20 — YAML, ER-2 is `.go`-scoped; spirit: no rewrite, append-gated entries only)* | +≤25 total per file | ok |

## Implementation spec

1. **Values block** (append-only; commented):

```yaml
database:
  type: ""            # "" | sqlite | postgres | legacy ("" = app default: sqlite) — #417/#420
  path: ""            # sqlite file path; empty ⇒ app default (AD-3). NB: distinct from storage.databasePath (legacy PVC + data.json home)
  autoMigrate: ""     # "" | true | false ("" = app default true)
  dsn: ""             # postgres DSN — LEAKS INTO CONFIGMAP IF SET; prefer dsnSecret
  dsnSecret: ""       # existing secret + key: {name: ..., key: DSN} shape per D-4; consumed via secretKeyRef
```

   (`dsn` intentionally lands ConfigMap?? NO — D-4 routes non-secret → ConfigMap, secret → Secret: `dsn`
   goes to the chart Secret template despite being in the same values block; the comment above is the
   leak-warning, the template puts it where D-4 says. Keep values-key shape, honor routing.)
2. **Gating pattern** (exact style of existing entries — read them first, mirror indentation/quoting):
   every new entry wrapped `{{- if .Values.database.X }}`; `autoMigrate` string-passthrough (no `bool` —
   empty means unset; #420 parses tolerantly).
3. **Byte-identity gate (the AC of THIS task):** `helm template helm/ > /tmp/before.txt` on main, same after
   (`helm` installed? if absent, `helm/template` rendering via docker `helm/helm:v3` — record which),
   `diff` → **empty**. Any diff = gating bug, not "minor change".
4. **Docs:** values-table rows in `kubernetes/helm-chart.md` (match its table format); postgres-page
   paragraph ≤ 15 lines with a values example (Postgres + dsnSecret) and one sentence on `storage.databasePath: ""`
   for the no-PVC posture (D-4's verified behavior — cite the deployment.yaml lines in the HTML comment, per
   D-2 evidence protocol).

## Out of scope

Compose files (D-4 forbids — user-copy guidance lives in task 03's page), chart version bump (maintainer
release flow), any `kustomize`/other deploy surface, changing `storage.databasePath` semantics (it stays
legacy-home + PVC gate; docs disambiguate only).

## Validation

- [ ] `helm lint helm/` clean; **byte-identity diff empty** (both outputs pasted in PR)
- [ ] Rendered Postgres values: `helm template helm/ --set database.type=postgres --set database.dsnSecret.name=x --set database.dsnSecret.key=y` shows DB_TYPE + secretKeyRef env, NO DSN literal anywhere in output (`grep -i dsn` on rendered yaml as evidence)
- [ ] `helm-chart.md` + postgres page: markdownlint clean **new/changed lines only** vs D-6 baseline rules (npm cache prefix export), prose ≤80 cols; new content prettier-clean (existing pages' pre-existing failures untouched — never `--write`)
- [ ] jekyll build-if-feasible (D-6 command) — page renders, no broken internal links
- [ ] `git diff --stat` → only the six files

## Rollback

Chart gating means revert = zero behavioral change for every existing release (byte-identity gate proved
stock renders never changed); docs revert trivially.

## PBI acceptance mapping

**Deployment config examples** AC — helm half (compose half = task 03's snippet per D-4); OQ-4 resolved as
**shipped**, not written; backward-compat posture proven mechanically, which is the feature's ER-5 spirit
applied to the deploy surface.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `427-04`.** Cut `slice/427-04` from `sql-migration` HEAD
(`.worktrees/slice-427-04`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
