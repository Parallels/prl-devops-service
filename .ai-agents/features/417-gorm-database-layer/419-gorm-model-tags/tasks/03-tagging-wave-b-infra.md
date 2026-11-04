# Task 03 — Tagging wave B: infra domain (catalog, packer, config, user_config, job, reverse-proxy, enrollment token)

## Context
Plan mapping rows 6–13, 16–19. Runs strictly after task 02 merges (same-package files — waves must
not run in parallel). Introduces the **`serializer:json`** pattern used by every nested blob in this
PBI (D-3): GORM v2 stores canonical `json.Marshal` of the field value in a TEXT column, so the Go
value round-trips and the wire JSON is byte-stable (deterministic field order / sorted map keys).

## Objective
Every wave-B struct carries tags exactly as the mapping table says; per-struct golden-JSON tests
green; a smoke `AutoMigrate` of wave-B models against `:memory:` passes; build+tests green.

## Files
- Modify: `catalog_manifest.go` `catalog_manager.go` `packer_template.go` `config.go`
  `user_config.go` `job.go` `reverse_proxy.go` `orchestrator_enrollment_token.go`
- Touch (doc comment only): `host_resources.go` — note "aggregate/DTO types; never an AutoMigrate
  root; persist indirectly via `orchestrator_hosts.resources` blob (D-3 row 16)". No tags.
- Modify: `models_gorm_test.go` + `testdata/golden/` (`catalog_manifest.json`, `catalog_manager.json`,
  `packer_template.json`, `configuration.json`, `user_config.json`, `job.json`, `reverse_proxy.json`,
  `reverse_proxy_host.json`, `enrollment_token.json`)

## File size guard (ER-4)
| File | Now | After (proj.) | Verdict |
|---|---|---|---|
| `catalog_manifest.go` | 144 | ~200 | ok |
| `catalog_manager.go` | 18 | ~45 | ok |
| `packer_template.go` | 54 | ~85 | ok |
| `config.go` | 5 | ~15 | ok |
| `user_config.go` | 22 | ~45 | ok |
| `job.go` | 43 | ~70 | ok |
| `reverse_proxy.go` | 299 | ~340 | ok |
| `orchestrator_enrollment_token.go` | 11 | ~30 | ok |
| `host_resources.go` | 144 | 145 (doc line) | ok |

## Implementation spec
Exemplar — `catalog_manifest.go` (all list/map/struct fields SERIALISE per mapping row 7):
```go
type CatalogManifest struct {
	ID                     string                       `json:"id" gorm:"primaryKey"`
	CatalogId              string                       `json:"catalog_id" gorm:"index"`
	Name                   string                       `json:"name"`
	// …plain scalars keep identity mapping (no tag beyond PK/index)…
	Provider               *CatalogManifestProvider     `json:"provider" gorm:"serializer:json"`
	RequiredRoles          []string                     `json:"required_roles" gorm:"serializer:json"`
	RequiredClaims         []string                     `json:"required_claims" gorm:"serializer:json"`
	Tags                   []string                     `json:"tags" gorm:"serializer:json"`
	VirtualMachineContents []CatalogManifestContentItem `json:"virtual_machine_contents" gorm:"serializer:json"`
	PackContents           []CatalogManifestContentItem `json:"pack_contents" gorm:"serializer:json"`
	MinimumSpecRequirements *MinimumSpecRequirement     `json:"minimum_requirements,omitempty" gorm:"serializer:json"`
	// Revoked/RevokedBy/RevokedAt + Tainted* stay plain columns — statuses, not soft delete (D-5)
}
```
- `job.go`: `Steps []JobStep gorm:"serializer:json"`; `ID` PK; `Owner gorm:"index"`;
  `ResultRecordID gorm:"column:result_record_id;index"`; `*DbRecord` → `gorm:"-:all"`.
- `user_config.go`: `ID` PK; `UserID` `gorm:"index"`; composite `gorm:"uniqueIndex:ux_userconfig_user_slug"`
  marker on `UserID`+`Slug` pair (mapping row 12).
- `reverse_proxy.go`: `ReverseProxy.ID` PK + `HostID gorm:"column:host_id;index"`;
  `ReverseProxyHost.ID` PK + `HostID gorm:"column:host_id;index"` +
  `Tls/Cors/TcpRoute gorm:"serializer:json"`, `HttpRoutes []*ReverseProxyHostHttpRoute
  gorm:"serializer:json"` (order preserved by JSON array; `RegexpPattern` stays `json:"-"` ⇒ never
  serialised); 4 child structs gain **no** tags (not roots).
- `orchestrator_enrollment_token.go`: `ID` PK, `Token gorm:"uniqueIndex"`, `*DbRecord → gorm:"-:all"`.
- `config.go`: `Configuration.ID` `gorm:"primaryKey"` (single-row legacy meta, mapping row 25).
- Golden tests reuse the task-02 harness (capture-first in separate pre-tag commit).
- Smoke AutoMigrate sketch (full battery lands in 06):
```go
db, err := database.NewDatabaseConnection(ctx, database.DefaultDatabaseConfig(ctx)) // :memory: ok (D-4 of #420 plan)
require.NoError(t, err)
require.NoError(t, db.AutoMigrate(&models.CatalogManager{}, &models.CatalogManifest{}, &models.Job{},
	&models.PackerTemplate{}, &models.UserConfig{}, &models.ReverseProxy{}, &models.ReverseProxyHost{},
	&models.OrchestratorEnrollmentToken{}, &models.Configuration{}))
require.True(t, db.Migrator().HasIndex(&models.UserConfig{}, "ux_userconfig_user_slug"))
var typ string // sqlite_master type check: serializer columns must be TEXT
db.Raw("SELECT type FROM sqlite_master WHERE name='catalog_manifests' AND sql LIKE '%tags%'").Scan(&typ)
```

## Out of scope
`orchestrator_host.go`, snapshot models, `virtual_machine*` (task 04). No AutoMigrate test matrix
(task 06). No `gorm.DeletedAt` (D-5).

## Validation
1. `cd src && go build ./... && go test ./... -count=1` green.
2. `go test ./data/models/... -run 'JSONGolden|SmokeAutoMigrate' -count=1` green.
3. `git diff --stat testdata/` empty in the tag commit (capture-first evidence).
4. Serializer coverage: `grep -c 'serializer:json' <each file>` ≥ mapping-table column count per row.

## Rollback
Revert PR — tags are additive-only; no data touched.

## PBI acceptance mapping
AC1 (rows 6–13, 16–19), AC4 (golden bytes incl. nil-vs-empty serializer edge), AC5 evidence
reinforced (Revoked columns kept as statuses), ER-2 table above.
