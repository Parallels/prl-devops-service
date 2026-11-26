# jobs

Async job manager. Long operations (VM create, catalog push/pull, ...) create a job record, run in the background, and clients poll `/jobs` or watch events.

- `main.go` — `JobManagerService` (`jobs.Get(ctx)`): persists job state through `DatabaseService`, runs timeouts and cleanup.
- `tracker/` — `JobProgressService` (`NewProgressService`): progress messages and percentages emitted while a job runs; used by catalog and download code.
- `hooks.go` — injected terminal-state hook (`SetJobTerminalHook`), called when a job finishes. This is how placement releases capacity without `jobs` importing `orchestrator` or `placement` (would create a cycle). Set at startup before any job can run.
- Do not add dependencies from this package to `orchestrator`, `placement` or `controllers`. Extend via hooks.
- Changing job states or fields needs updates in `../data/models/job.go`, `../mappers/job.go`, `../models/job.go` and the API docs.
