# startup

Boot sequence. `Init(ctx)` loads config, sets up logging, the progress tracker, password/JWT/brute-force services. `Start(ctx)` opens the database, runs schema migrations, seeds defaults, wires hooks (e.g. job terminal hook for placement), then the API starts (`api.go`).

- `migrations/` — versioned data migrations (`Version0_6_0`, `AddApiKeyOwnClaims`...) implementing `migrations.Migration`; registered in `startup/main.go`. Add new ones there, with a test.
- `seeds/` — default claims, roles, users, catalog managers; idempotent.
- `legacy_migration.go` — one-way import of the old JSON store into the SQL database (uses `database/legacy_migrator`).
- `host_module_sync.go` — syncs host-module state at boot.
- Order matters: DB before seeds, seeds before API, hooks before any job can run. Do not reorder without reading the placement and jobs guides.
- Docs: `docs/docs/devops/getting-started/database-upgrade.md`, `configuration.md`.
