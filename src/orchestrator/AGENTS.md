# orchestrator

`OrchestratorService` runs when the orchestrator module is enabled. It keeps a registry of remote hosts, mirrors their state (resources, VMs, catalog cache, reverse proxy, logs) and forwards VM/host operations to them. Files are named by operation: `create_host_virtual_machine.go`, `start_host_virtual_machine.go`, `get_hosts.go`, `register_host.go`, `*_host_reverse_proxy*.go`, ...

Key pieces:
- `check_host_requirements.go` — decides whether a host can take a create request (CPU, memory, VM count, Apple VM limits, tags). `create_host_virtual_machine.go` picks among valid hosts (`getValidHostsForCreate`). Algorithm doc: `docs/docs/devops/orchestrator/host-selection.md` — update it when you change selection.
- `placement_gate.go` — claims/releases capacity in `../placement` around a create, so concurrent creates cannot overcommit. See the placement guide before touching admission.
- `handlers/` — event handlers for host websocket events (health, stats, jobs, logs, PD events, catalog cache, reverse proxy), implementing `interfaces/host_event_handler.go`. `host_websocket_manager.go` / `websocket_client.go` own the connections.
- `registry/job_link_registry.go` — links orchestrator job ids to host job ids. `job_links.go` uses it.
- `vm_refetch_loop.go`, `hardware_update_queue.go` — background refresh of host state.

Rules: never import this package from `jobs` or `placement` (import cycle; use injected hooks). Host calls go through the API client in `serviceprovider/apiclient`. Mutating host data goes via `DatabaseService`. Add tests next to the file (`resource_failure_test.go` and `check_host_requirements_test.go` show the fake-host style).
Docs: `docs/docs/devops/orchestrator/`, `events-hub/`.
