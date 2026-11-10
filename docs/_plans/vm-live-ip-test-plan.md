# Test Plan — Live VM IP Reporting (`noCache` endpoint, refetch cadence, IP poke worker)

Covers the three shipped features plus the root-cause fix they wrap. Target components: **host** (Parallels service), **orchestrator**, and the API responses that expose them.

**Feature refs**

| Area | Code |
|---|---|
| Live endpoint | `GET /v1/machines?noCache=true`, `GET /v1/machines/{id}?noCache=true` (`src/controllers/machines.go`) |
| Refetch cadence | `VM_CACHE_REFRESH_INTERVAL_SECONDS` (host, default 300) · `ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS` (orchestrator, default 300) |
| Poke worker | `src/serviceprovider/parallelsdesktop/ip_worker.go` — attempt at t=0, then 10s→20s→30s→40s→50s→60s→60s (8 attempts), **give-up at 280s** → `ip_status: failed` |
| New VM fields | `ip_status` (`pending`/`resolving`/`resolved`/`failed`), `ip_status_detail` — host API, WS events, orchestrator DB |
| Clobber fix | `preserveResolvedIp` (host), `mergePreservedIp` (orchestrator) |

**Auth**: all `curl` samples assume `-H "X-Auth/User/API-key: $KEY"` (or bearer) against the host/orchestrator base URL; omit for brevity.

---

## 0. Automated baseline (must be green before manual pass)

```bash
cd src
go build ./...
go test ./config/ ./mappers/ ./orchestrator/ ./serviceprovider/parallelsdesktop/ -count=1
go test -race ./serviceprovider/parallelsdesktop/ -run 'TestIpPoke|TestNotifyMissing' -count=1
```

Expected: all `ok`. (Known-unrelated flakes: `TestAsyncCreateOrchestratorVirtualMachineHandler_NoUser`, `TestRealProcessLauncher` under sandbox.)

---

## 1. Environment & fixtures

- **H1 host**: macOS with Parallels Desktop, service in `host` role.
- **O1 orchestrator**: `orchestrator` role, H1 registered and `healthy`.
- **VM-MAC** — macOS guest, guest tools installed, boots to a normal login (gets DHCP lease).
- **VM-MAC-BAD** — macOS guest, tools installed but guest network broken/unconfigured (never gets a lease).
- **VM-LIN** — Linux guest, tools installed.
- **VM-NOTOOLS** — any guest with guest tools **not** installed.

Baseline record for every case: `GET /v1/machines` (no `noCache`) and `?noCache=true` on both H1 and O1; note `internal_ip_address`, `ip_status`, `ip_status_detail`.

Useful log markers (service log):

```
[IP] Arming IP poke worker for VM        → worker started
[IP] Probe en0 for VM                    → macOS exec probe fired
[IP] Resolved internal IP for VM         → success (VM_UPDATED broadcast)
[IP] Giving up on IP poke for VM         → 280s deadline hit
Starting VM refetch loop (interval: …)   → orchestrator loop up
Live VM refetch complete for host        → per-tick result
```

---

## 2. Live endpoint — `?noCache=true`

| # | Case | Steps | Expected | Priority |
|---|---|---|---|---|
| A1 | List live vs cached | `curl '/v1/machines'` then `'/v1/machines?noCache=true'` on H1 | Both 200; same VM set; live version slower (~prlctl time, check with `time`/`-w '%{time_total}'`); live rows carry current `state` | P0 |
| A2 | Single VM live | `curl '/v1/machines/<VM-MAC-id>?noCache=true'` | 200; live prlctl fields; `ip_status` reflects current resolution state | P0 |
| A3 | Live read kick-starts resolution | Stop VM-MAC, start it; before first probe lands, call `?noCache` | Response may show `pending`/empty; within ~5s log shows `Arming IP poke worker` for that VM (live read armed it) | P1 |
| A4 | Cache not poisoned by live path | Call `?noCache=true` twice quickly (inside 5s attempt cooldown) | Second call must not fire extra `prlctl exec` probes (attempt cooldown) and must not change cache negatively | P2 |
| A5 | Param hygiene | `?noCache=false`, `?noCache=1`, no param | Served from cache path (no live prlctl latency); param is strictly `"true"`-gated | P2 |
| A6 | Swagger | Open host swagger UI/json | `noCache` query param documented on both endpoints | P3 |

## 3. Refetch cadence config

| # | Case | Steps | Expected | Priority |
|---|---|---|---|---|
| B1 | Orchestrator default | Start O1 with no env var | Log: `Starting VM refetch loop (interval: 5m0s)` | P0 |
| B2 | Orchestrator override | `ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS=60` restart O1 | Log shows 1m0s; `Live VM refetch complete for host` appears every ~60s per healthy host | P0 |
| B3 | Refetch hits host live | While B2 runs, watch H1 access log | O1 polls `GET /v1/machines?noCache=true` each tick (query visible), not plain `/v1/machines` | P1 |
| B4 | Skipped hosts | Disable H1 (or force state ≠ healthy) during B2 | No refetch calls to that host | P2 |
| B5 | Host sweep interval | Restart H1 with `VM_CACHE_REFRESH_INTERVAL_SECONDS=60` | Cache-swap + missing-IP sweep log every ~60s (existing behaviour, unchanged default 300 when unset) | P1 |

## 4. Poke worker lifecycle

| # | Case | Steps | Expected | Priority |
|---|---|---|---|---|
| C1 | Boot → resolved | Start VM-MAC | `vm_state_changed→running` arms worker; within ~10–60s `internal_ip_address` set, `ip_status: resolved`, `ip_status_detail: ""`; `VM_UPDATED` logged | P0 |
| C2 | Status progression | Poll `?noCache=true` every 2s from power-on | Transitions `resolving` → `resolved`; never back to `pending` while running | P1 |
| C3 | Give-up after 280s | Start VM-MAC-BAD (no lease ever) and watch for 5 min | Attempts at ≈0,10,30,60,100,150,210,270s (count `Probe en0` lines: 8 probes/round over en0-en2); at ≈280s `Giving up`, `ip_status: failed`, detail `no IP address obtained after 8 attempts within 4m40s` | P0 |
| C4 | Re-arm on sweep | After C3, wait one host sweep (≤ default 300s, or 60s with B5) | Worker rearms: status `failed` → `resolving`; a fresh 280s window runs; if still broken → `failed` again | P0 |
| C5 | Stop clears status | With VM-MAC resolved, stop it, then read cache | Worker stops; `ip_status` cleared to empty (next boot re-verifies); last known `internal_ip_address` retained for display; running VM never shows stale `resolving` | P1 |
| C6 | Reboot gets new lease | Reboot VM-MAC, change DHCP lease (renew DHCP or reconnect NIC), wait | After boot, `ip_status` re-runs to `resolved` and IP value **updates** to the new lease | P0 |
| C7 | Tools missing → skip | Start VM-NOTOOLS (tools explicitly *not installed*) | No `Arming IP poke worker` for it, no probes, `ip_status` stays empty | P1 |
| C8 | Tools arrive late | Start VM-NOTOOLS, install/finish tools so `vm_tools_state_changed` fires | Worker arms after the event (or on next sweep) and resolves | P2 |
| C9 | Linux coverage | Start VM-LIN with tools; boot without waiting for guest network report | If prlctl shows the IP immediately → `resolved` right away; if not, worker re-reads prlctl each round until present (no macOS `exec` probes logged for it) | P1 |
| C10 | Non-mac with IP never re-poked | VM-LIN running with IP | After first resolve, later sweeps do **not** arm workers for it (`ip_status: resolved` + IP present) | P1 |
| C11 | Unregister mid-probe | Start VM-MAC-BAD, `prlctl unregister` while probing | Worker exits (no `failed` written for a gone VM); no error spam; VM gone from cache | P2 |
| C12 | Shutdown hygiene | Restart the H1 service while 2+ workers active | Clean stop (no goroutine panic in logs); after start, sweep re-arms everything | P2 |

## 5. IP preservation — the original "IP disappears" regression

| # | Case | Steps | Expected | Priority |
|---|---|---|---|---|
| D1 | Host ticker no longer wipes | Get VM-MAC `resolved`, then wait ≥2 host sweeps | `internal_ip_address` + `ip_status: resolved` still present on every read, even when `prlctl list` reports no NIC data | **P0 — original bug** |
| D2 | Live read doesn't regress cache | Same, call `?noCache=true` | Live response keeps the resolved IP (patch-engine merge), does not return empty | P0 |
| D3 | Orchestrator refetch preserves | Resolve on H1, then wait one O1 refetch tick (B2 setup) | O1 `GET /v1/orchestrator/machines` still shows the IP; DB entry unchanged or fresher | P0 |
| D4 | Orchestrator full refresh preserves | Trigger a full refresh (wait `ORCHESTRATOR_PULL_FREQUENCY_SECONDS`×10 or restart O1) with H1 serving IP-less prlctl data (`prlctl set <vm> --net_off` scenario or VM-MAC-BAD already resolved by exec then net disabled) | Stored IP survives the bulk replace | **P0 — original bug (orchestrator side)** |
| D5 | PDFM VM_UPDATED preserves | Force a host `VM_UPDATED` with silent network info (e.g. config change event on VM-MAC) | O1 keeps stored IP/status (never blanks between refetches) | P1 |
| D6 | Stopped VM drops IP | Stop VM-MAC, read both APIs | Empty IP + empty status on running-filtered and full reads (no resurrection of IP for stopped VMs) | P1 |
| D7 | Real prlctl data wins | `prlctl set VM-MAC --net0 … new net` producing a genuinely new prlctl-reported IP | The new IP replaces the preserved one | P2 |

## 6. Multi-VM / flood safety (host)

| # | Case | Steps | Expected | Priority |
|---|---|---|---|---|
| E1 | Staggered startup sweep | Boot H1 service with ≥6 running macOS VMs missing IPs | First probes start ~3s apart (`Arming` lines stagger); no burst of simultaneous `prlctl exec` | P1 |
| E2 | SSH logon concurrency cap | During E1, sample `ps aux | grep -c 'prlctl exec'` | Never more than **2** concurrent exec probes (sshSem) | P1 |
| E3 | Dedupe under multiple triggers | Fire boot event + live read + sweep for the same VM within seconds | Only one `Arming …` per worker lifecycle; single probe stream per VM | P1 |
| E4 | Fleet at scale (optional soak) | 10–20 VMs running for 1h with default 300s settings | CPU of service stays low; no orphaned `prlctl exec` zombies; log volume sane | P2 |

## 7. API contract regression

| # | Case | Steps | Expected | Priority |
|---|---|---|---|---|
| F1 | Back-compat | Read `/v1/machines` with an old client model | New fields are `omitempty` — payloads without them (e.g. Linux resolved) are absent, never `null`-breaking | P2 |
| F2 | X-No-Cache header path unchanged | Orchestrator `GET /v1/orchestrator/machines` with `X-No-Cache: true` | Old semantics intact (full refresh still available for that path); no crash; results sane | P2 |
| F3 | Orchestrator targeted noCache | `GET /v1/orchestrator/hosts/{id}/machines` noCache variant / `GetHostVirtualMachines(noCache)` | Only that host is contacted live (H1 access log), other hosts untouched | P1 |
| F4 | WS event shape | Subscribe to WS during a resolve | `VM_UPDATED` message contains `new_vm.ip_status: "resolved"` | P2 |

---

## Sign-off criteria

- All **P0** pass, **P1** pass or have a documented waiver.
- D1/D4 (the original bug repro) explicitly verified and signed — this is the headline fix.
- C3 timing table observed within ±2s per attempt boundary and 8 attempts exactly.
- No unowned goroutines/probes after C12.

## Suggested execution order

1. §0 automated → 2. §2 (endpoint basics) → 3. §4 C1–C5 (worker happy/timeout paths) → 4. §5 D1/D4 (bug repro) → 5. §3 (cadence) → 6. remaining P1 → 7. soak E4 if time allows.
