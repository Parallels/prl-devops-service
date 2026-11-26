# config (+ constants)

Process configuration: `config.Get()` singleton, loaded from env vars / config file / CLI flags. `constants/` holds every env var name, claim name and module name.

- Modules: `modules.go` exposes `IsHost()`, `IsOrchestrator()`, `IsCatalog()`, `IsCatalogManager()`, `IsApi()`, `IsModuleEnabled()`. Controllers and startup gate features on these.
- Files by concern: `database.go` (DB/backup/engine settings), `network.go`, `platform.go`, `placement.go` (host-admission limits), `models.go` (config structs), `main.go` (loader).
- Add a setting: define the env var name in `constants/`, add a getter in the matching file here, document it in `docs/docs/devops/getting-started/configuration.md`, add a test (`*_test.go` shows the override pattern).
- Never log or echo secret values (DB passwords, keys); see the capped-echo guard in `database.go`.
- `shape_security_test.go` guards that config structs do not leak secrets; keep it passing.
