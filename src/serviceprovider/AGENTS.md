# serviceprovider

Process-wide registry of service wrappers. `serviceprovider.Get()` returns the `ServiceProvider` holding the database service, the SQL handle, and wrappers around external tooling and subsystems. Initialised in layers: `InitCatalogServices`, `InitParallelsServices`, `InitServices`. `NewMockProvider()` is for tests.

Subpackages (one per integration): `parallelsdesktop` (prlctl/prlsrvctl wrapper, VM state, capacity, license, event monitor, IP worker), `packer`, `vagrant`, `brew`, `git`, `ssh`, `system` (OS/hardware info), `stats`, `diskSpace`, `download`, `health`, `logs`, `apiclient` (HTTP client used to call hosts and orchestrators), `eventEmitter` (events hub/websocket clients), `interfaces` (`Service` contract).

Working here:
- Get the DB with `serviceprovider.GetDatabaseService(ctx)` (don't reach into the struct fields). Check availability with `IsParallelsDesktopAvailable()`, `IsPackerAvailable()` etc. before use; these tools are absent on non-host nodes.
- A new integration = new subpackage with a `*Service`, registered in the init functions and added to `Services`; keep command execution behind a small interface so it is mockable (`parallelsdesktop/execute_command_test.go` shows how).
- External commands: build args from validated input, no shell string concatenation.
- `parallelsdesktop` has platform files (`file_owner_unix.go`/`_windows.go`); keep builds green for both.
- Docs: `docs/docs/devops/getting-started/architecture.md`, `events-hub/`.
