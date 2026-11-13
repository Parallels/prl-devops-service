# Feature Brief — Show the per-host VM count limit in Control Center

| Field | Value |
|---|---|
| Issue | not opened yet — assign a PBI number at triage |
| Feature folder | `.ai-agents/features/max-vms-per-host-ui/` |
| Status | Ready for UI design; backend implemented in the working tree, not yet committed |
| Audience | Control Center (Web UI) front-end |
| Backend surface | `Parallels/prl-devops-service`, hosts + orchestrator, sources under `src/` |
| Depends on | `MAX_VMS_PER_HOST` / `MAX_VMS_PER_HOST_MODE` (host admission + orchestrator `vm_count` check) |
| Last verified against code | 2026-11-10 |

---

## 1. Why

The service can now cap how many virtual machines a single host may hold, independently of
free CPU, memory and disk. Two things became visible to users that the UI cannot currently
explain:

- a host is **skipped during placement** while it still looks healthy and mostly idle,
  because it reached its VM count cap;
- a create, clone or start is **refused with HTTP 409** by the host itself.

Both are reported as plain text today. The ask is to surface the counter where operators
already read CPU / memory / disk: the host card in the list and the resources panel in the
host detail.

## 2. The rule being displayed

Set on the host (env var or `prldevops_config.yaml` `environment` map), never on the
orchestrator:

| Key | Values | Meaning |
|---|---|---|
| `MAX_VMS_PER_HOST` | int, `0`/absent/invalid = unlimited | hard cap on VM count, applied even with free CPU/RAM/disk |
| `MAX_VMS_PER_HOST_MODE` | `any` (default) \| `running` | `any` = every VM registered on the host, stopped ones included; `running` = VMs holding a hypervisor slot (running, paused or mid-transition) |

There is **no API to change it**: it is host-local configuration, not an orchestrator host
field. Read-only in the UI.

## 3. Where the numbers come from

No new endpoint is needed. Both host endpoints already carry everything, and both use the
same mapper, so the list and the detail payload are identical:

- `GET /v1/orchestrator/hosts`
- `GET /v1/orchestrator/hosts/{id}`

| JSON path | Meaning | Notes |
|---|---|---|
| `detailed_resources.max_vms_per_host` | the cap | **omitted when 0** (`omitempty`) — absent means no cap |
| `detailed_resources.max_vms_per_host_mode` | `"any"` or `"running"` | omitted when the host did not report it → treat as `"any"` |
| `detailed_resources.total.total_vms` | every VM registered on the host | the `any` counter |
| `detailed_resources.total_in_use.total_vms` | VMs holding a hypervisor slot | the `running` counter |
| `detailed_resources.total_reserved.total_vms` | operations in flight that will consume a counted slot | creates in `any` mode, starts in `running` mode |
| `resources.total_vms` | same as `total.total_vms`, on the headline item | convenient for the compact list card |

Real payload, produced by the mapper (trimmed to the relevant fields):

```json
{
  "resources": { "total_vms": 26, "logical_cpu_count": 16, "memory_size": 65536 },
  "detailed_resources": {
    "max_vms_per_host": 10,
    "max_vms_per_host_mode": "running",
    "total":          { "total_vms": 26, "logical_cpu_count": 16, "memory_size": 65536 },
    "total_in_use":   { "total_vms": 7 },
    "total_reserved": { "total_vms": 2 },
    "total_available": { "logical_cpu_count": 8, "memory_size": 32768 },
    "system_reserved": { "logical_cpu_count": 0, "memory_size": 0 }
  }
}
```

Gotchas, both verified against the generated payload:

- every counter is `omitempty`, so **`0` is absent, not `0`** — default missing numbers to
  `0` rather than rendering `undefined`;
- the item objects (`total`, `total_in_use`, `total_reserved`, `system_reserved`) are always
  present; only the numbers inside them can be missing;
- the mode is reported even when no cap is configured, so an uncapped host can send
  `"max_vms_per_host_mode": "any"` with no `max_vms_per_host`. Gate any mode chip on
  `hasLimit`, or it will advertise a mode on hosts that enforce nothing.

### Deriving what to show

```ts
const dr = host.detailed_resources;
const limit   = dr?.max_vms_per_host ?? 0;                    // 0 = no cap
const mode    = dr?.max_vms_per_host_mode ?? 'any';
const used    = mode === 'running'
  ? dr?.total_in_use?.total_vms ?? 0
  : dr?.total?.total_vms ?? 0;
const pending = dr?.total_reserved?.total_vms ?? 0;
const hasLimit = limit > 0;
const ratio   = hasLimit ? Math.min(1, (used + pending) / limit) : null;
```

`used + pending` is the number the enforcement compares against, so it is the number to
compare against `limit` — showing `used` alone would understate a busy host.

## 4. What to build

1. **Host list card** — compact counter, only when `hasLimit`:
   `VMs 7 / 10` plus a `+2` pending badge while `pending > 0`. Reuse the thresholds the
   existing CPU / memory bars use, with the full state at `used + pending >= limit`. No cap
   → render nothing, exactly as today; do not render `0 / 0` or "unlimited" on every host.
2. **Host detail → resources panel** — a `Virtual machines` row next to CPU / memory / disk:
   value `used + pending` / `limit`, progress bar, mode label, and `N pending` when
   `pending > 0`. Tooltip: `Mode: running — counts VMs holding a hypervisor slot` or
   `Mode: any — counts every VM registered on the host, stopped ones included`.
3. **Independent of the cap** — where the panel shows host inventory, `total_vms` is a
   useful line ("VMs on this host: 26") even with no cap configured. Show it whenever the
   field is present; that is also the cheapest way for an operator to see that a host build
   is too old to report it (field absent while the host clearly has VMs).
4. **Pre-flight result** — `POST /v1/orchestrator/machines/check` returns per-host
   `hosts[].checks[]`; the new entry is the same shape as the existing `macos_capacity`
   check, so reuse that renderer:

   ```json
   { "name": "vm_count", "passed": false, "required": 10, "available": 10,
     "unit": "vms", "message": "VM limit: current 8, pending 2, maximum 10" }
   ```

   The check is **only present when the host reports a cap**, and a successful pre-flight
   returns no host details at all (`{"passed":true}`), so the check appears only in failure
   responses.
5. **Rejections** — a create, clone or start refused by the host carries the text:
   - `409 host VM limit reached: current 10, pending 0, maximum 10 (mode any)`
   - `503 cannot verify host VM capacity: live host inventory failed` (a cap is configured
     but the host could not enumerate its VMs; fail-closed by design)

   For an async create the same text lands in the job `error` / `message`, not in the HTTP
   response of the submit call. A dedicated, human-readable variant of the 409 is worth
   adding if the UI has room for it.

## 5. States to handle

| State | Payload | Render |
|---|---|---|
| No cap | `max_vms_per_host` absent | nothing in the list; "No limit" or the row hidden in detail |
| Cap, room | `used + pending < limit` | normal bar |
| Cap, near | approaching `limit` | warning — reuse the existing resource-bar thresholds |
| Cap reached | `used + pending >= limit` | critical; placement will skip this host |
| Ops in flight | `pending > 0` | `+N` badge / `N pending` — this is why the counter can look ahead of the VM list. `pending` covers creates in `any` mode and starts in `running` mode; the other operation reserves nothing by design, so `pending` being `0` during a bulk start in `any` mode is correct |
| Mode `running` with many stopped VMs | `total.total_vms` ≫ `total_in_use.total_vms` | expected, not a bug: stopped VMs are free capacity in that mode |
| Host build too old, or Linux/Windows host | no `total_vms`, no `max_vms_per_host` | fall back to today's display; absence is ambiguous (no cap vs not reported), so avoid claiming "unsupported" from this payload alone |

## 6. Copy

- Row label: `Virtual machines`
- Mode labels: `any` → "All VMs"; `running` → "Running VMs"
- Pending: `{count} pending` / `+{count}`
- Limit reached hint: `Host VM limit reached — new VMs will be placed on another host`

## 7. Freshness and cost

No extra polling. The orchestrator refreshes host resources every
`ORCHESTRATOR_PULL_FREQUENCY_SECONDS` (default 30) and on host stats events over the
existing WebSocket, so the counter can lag reality by up to that interval — `pending`
exists precisely to cover the gap. All values come free with the host payloads already
fetched.

## 8. Acceptance criteria

1. With `MAX_VMS_PER_HOST=10` and `MAX_VMS_PER_HOST_MODE=running` on a macOS host, the
   detail panel shows `used + pending / 10` where `used` is `total_in_use.total_vms`.
2. Same host with mode `any` shows `total.total_vms` instead; the mode label changes.
3. A host with no cap is pixel-identical to the current build (no row, no zero).
4. In `running` mode, starting 2 VMs on a host with 2 free slots shows `+2 pending` before
   the VM list refreshes, and the bar reflects `used + pending`. In `any` mode the same
   bulk start shows no pending, which is correct.
5. When every host is at its cap, `POST /v1/orchestrator/machines/check` failure details
   render the `vm_count` check message verbatim.
6. A create refused by the host surfaces the 409 text; an async create shows it on the job.

## 9. Out of scope

- **Editing the limit from the UI.** It is host-local configuration by design; no
  orchestrator host field and no endpoint exist to write it.
- Per-host override stored on the orchestrator, and cluster-wide `Σ VMs / Σ limits`
  aggregation on the dashboard — neither is reported today.
- Migration or eviction: lowering a limit never stops or moves existing VMs, it only blocks
  new ones. Worth a hint in the tooltip if operators ask.

## 10. How to reproduce locally

```bash
# on the host, then restart the host service
export MAX_VMS_PER_HOST=2 MAX_VMS_PER_HOST_MODE=any

curl -s -H "Authorization: Bearer $TOKEN" $ORCH/v1/orchestrator/hosts | \
  jq '.[] | {host, limit: .detailed_resources.max_vms_per_host,
             mode: .detailed_resources.max_vms_per_host_mode,
             total: .detailed_resources.total.total_vms,
             in_use: .detailed_resources.total_in_use.total_vms,
             pending: .detailed_resources.total_reserved.total_vms}'
```

With the cap set, `GET /v1/orchestrator/hosts/{id}` should show the fields immediately
(after one refresh cycle), and a third create attempt should be refused with the 409.
