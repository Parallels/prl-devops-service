# controllers

HTTP handlers, one file per resource (`users.go`, `machines.go`, `orchestrator.go`, `catalog.go`, `jobs.go`, ...). `main.go` has `RegisterV1Handlers`, which registers groups conditionally on the enabled module (`config.Get().IsHost()`, `IsOrchestrator()`, ...).

Pattern for a new endpoint:
1. `registerXHandlers(ctx, version)` in the resource file; add it to `RegisterV1Handlers` under the right module gate.
2. Builder chain from `restapi` with the required claims/roles (constants in `../constants`).
3. Handler: decode and validate the request, call the orchestrator/serviceprovider/DB, map via `../mappers`, write with the helpers in `common.go`. Return API DTOs from `../models`, never `data/models`.
4. Swagger annotations on the handler, then `make generate-swagger` (and `generate-postman` if requested).
5. Test next to it; add/update the page under `docs/docs/devops/restapi/reference/`.

Rules: no business logic or SQL here; on an orchestrator, VM operations are forwarded to hosts by `../orchestrator`, on a host they go to `serviceprovider/parallelsdesktop`. Check ownership claims (`*_OWN_CLAIM`) where a resource has an owner.
