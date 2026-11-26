# Parallels DevOps Service

Go service (`github.com/Parallels/prl-devops-service`) that provisions and manages Parallels Desktop VMs. One binary, several **modules** switched on by config: `api`, `host`, `orchestrator`, `catalog`, `catalog_manager`, `reverse_proxy` (see `src/config/modules.go`, `src/constants`).

- Orchestrator: registers hosts, picks a host for a VM, forwards operations, mirrors host state.
- Host: runs Parallels Desktop locally (VMs, packer, snapshots, local catalog cache).
- Catalog / catalog manager: push/pull VM images to remote storage; the manager is the catalog index.

Go module lives in `src/` (`go.work` points at it). Everything below is relative to `src/` unless noted.

## Where things are

| Path | Purpose | Guide |
|---|---|---|
| `main.go`, `cmd/` | Entry point and CLI commands | [cmd](src/cmd/AGENTS.md) |
| `config/`, `constants/` | Config loading, module flags, env var names | [config](src/config/AGENTS.md) |
| `startup/` | Init order, seeds, schema migrations, legacy migration | [startup](src/startup/AGENTS.md) |
| `restapi/` | HTTP listener, middleware, auth, controller builder | [restapi](src/restapi/AGENTS.md) |
| `controllers/` | HTTP handlers, one file per resource | [controllers](src/controllers/AGENTS.md) |
| `models/` | API DTOs (request/response shapes) | [models](src/models/AGENTS.md) |
| `mappers/` | API model <-> data model conversion | [mappers](src/mappers/AGENTS.md) |
| `data/` | `data/models` (GORM models), `data/interfaces` (`DatabaseService`), plus the deprecated JSON engine | [data](src/data/AGENTS.md) |
| `database/` | Current GORM (SQLite/Postgres) implementation of `DatabaseService`, legacy JSON migrator, parity tests | [database](src/database/AGENTS.md) |
| `orchestrator/` | Multi-host orchestration, host selection, event handlers | [orchestrator](src/orchestrator/AGENTS.md) |
| `placement/` | Capacity reservations that prevent host overcommit | [placement](src/placement/AGENTS.md) |
| `jobs/` | Async job manager and progress tracker | [jobs](src/jobs/AGENTS.md) |
| `serviceprovider/` | Singleton registry of service wrappers (PD, packer, git, ssh, events...) | [serviceprovider](src/serviceprovider/AGENTS.md) |
| `catalog/` | Catalog push/pull/import, storage providers, cache | [catalog](src/catalog/AGENTS.md) |
| `pdfile/` | PDFile (Dockerfile-like) parser and executor | [pdfile](src/pdfile/AGENTS.md) |
| `security/` | JWT, password hashing, API keys, brute-force guard | [security](src/security/AGENTS.md) |

Small support packages, no guide: `basecontext` (request/app context), `errors`, `helpers`, `writers`, `compressor`, `telemetry`, `settings`, `reverse_proxy`, `install`, `logs`, `processlauncher`.
Generated, do not hand-edit: `docs/` (swagger, from `make generate-swagger`), `api_documentation/` (generator).
Tests that span packages: `integration_test/`, `tests/`. Unit tests sit next to the code as `*_test.go`.

## Request flow

`main.go` -> `startup.Init/Start` -> `restapi` listener -> middleware (auth, claims) -> `controllers.*Handler` -> `orchestrator` / `serviceprovider` / `catalog` -> `data/interfaces.DatabaseService` (implemented by `database/`).
Long operations return a job id; `jobs/` runs them and `jobs/tracker` reports progress.

## Documentation (read before designing)

`docs/docs/devops/` (repo root, Jekyll site, also the public docs):
`getting-started/architecture.md`, `getting-started/configuration.md`, `orchestrator/` (overview, host-selection), `catalog/` (concepts, pdfile), `restapi/`, `security/rbac.md`, `settings/`, `events-hub/`, `reverse-proxy/`, `snapshots/`.
Feature design briefs/plans live in `.ai-agents/features/<feature>/` (`brief.md`, `plan.md`, `tasks/`). `plans/` holds older plans.
When behavior or an API changes, update the matching page under `docs/docs/devops/` and `CHANGELOG.md`.

## Commands (repo root)

```bash
make build              # binary
make test               # cd src && go test -count=1 ./...
make test-race
make test-integration   # needs Postgres, see Makefile
make lint
make generate-swagger   # after changing controller annotations; also generate-postman
```

Always run `make test` (or `go test ./<pkg>/...` for the area touched) and make sure `make build` passes before finishing.

## Conventions

- Layering: controllers -> orchestrator/serviceprovider -> `DatabaseService`. Controllers never touch the DB driver. Never import `orchestrator` from `jobs` or `placement` (cycle); use injected hooks.
- Every handler/service takes a `basecontext.ApiContext`; log through it.
- DB access goes through `data/interfaces.DatabaseService` only. Schema changes mean a GORM model change in `data/models/` registered in `database/migrate.go` (`AllGormModels`); see [database](src/database/AGENTS.md).
- At API boundaries return `models.ApiErrorResponse` / `errors.Diagnostics` like neighbouring code, not bare `fmt.Errorf`.
- Validate input at the boundary (controllers/middleware); sanitize file paths.
- New REST endpoints need swagger annotations, claims/role checks, and a docs page.
- Keep files under ~500 lines; prefer table-driven tests; mock via the interfaces, not concrete types.
- Never commit secrets, `.env`, keys, or `*.pem`. Never hardcode credentials. Never weaken auth or security checks.
- Do what was asked, nothing more. Prefer editing existing files; do not add docs/README files unless asked.
- Module-specific rules are in the `AGENTS.md` next to the code; read the one for the package you are changing.
