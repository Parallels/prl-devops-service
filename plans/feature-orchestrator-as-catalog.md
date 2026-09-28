# Feature: Orchestrator-as-Catalog VM Creation Support

## Overview

When the devops-service runs as both **orchestrator** and **catalog** (single process with both `orchestrator` and `catalog` modules enabled), VM creation via the orchestrator API fails when dispatching to remote hosts. The catalog connection string is empty, so remote hosts cannot reach back to pull pack files from the local catalog.

This feature adds support for the orchestrator to generate temporary, self-referencing credentials that allow remote hosts to authenticate against the orchestrator's built-in catalog service.

## The Problem

### Current Behavior (Broken)

```
Client → POST /api/v1/orchestrator/machines/async
         body: { catalog_manifest: { catalog_id: "ubuntu-22.04" } }
         (no connection or catalog_manager_id provided)
         │
         ▼
    Orchestrator resolves connection → ""  ← EMPTY STRING
         │
         ▼
    Dispatches to remote host:
    POST {host}/machines/async
    body: { catalog_manifest: { connection: "", ... } }
         │
         ▼
    Remote host receives connection=""
    → Fails: "missing connection or catalog_manager_id;
             local catalog is not enabled"
```

### Why It Happens

The `resolveCatalogMachineConnection()` function at `src/controllers/machines.go:2051-2054` returns an empty string when:
- No `connection` or `catalog_manager_id` is provided by the client
- The orchestrator IS the catalog (`config.Get().IsCatalog() == true`)

An empty connection works fine for **local processing** — the orchestrator uses the catalog inline within the same process. But when the orchestrator **dispatches to a remote host**, that host needs a resolvable URL with credentials to reach the catalog over HTTP.

### Why We Can't Use Existing Credentials

| Credential Type | Storage Format | Recoverable? |
|---|---|---|
| Catalog manager secrets | Encrypted (AES) | Yes, but requires knowing which manager record to use |
| API keys | Hashed (bcrypt/scrypt) | **No — one-way hash, impossible to recover** |

There is no catalog manager record when the orchestrator itself IS the catalog. And API key secrets are hashed before storage, so they cannot be retrieved.

## The Solution

Generate a **temporary internal API key** per VM creation job, embed it in the connection string, and clean it up after the job completes or fails.

### What Changes

1. **API Key Type Field** — Add `Type` field to `ApiKey` model with values `"internal"` or `"external"`. Internal keys are hidden from the UI/API list.
2. **Temp Key Generation** — When the orchestrator needs to provide a connection string for a remote host, create a short-lived internal API key and build a connection string like:
   ```
   host=<plaintext_secret>@http://orchestrator-address:{port}
   ```
3. **Automatic Cleanup** — Defer deletion on success/failure/panic, plus a background cleanup goroutine for orphaned keys.

### What Doesn't Change

- Client-provided `connection` or `catalog_manager_id` — used directly, no temp key created
- Orchestrator NOT running as catalog — existing error path unchanged
- Direct VM creation endpoint (`POST /api/v1/machines/async`) — unaffected
- Packer/Vagrant template creation — unaffected

## Who This Affects

| Role | Impact |
|---|---|
| End users | Transparent — no UI changes, no new configuration needed |
| Operators deploying orchestrator+catalog on same machine | Fix required — VM creation was broken |
| Operators deploying separate orchestrator/catalog | Unaffected — they always provide `connection` or `catalog_manager_id` |
| Host operators | Unaffected — this is purely an orchestrator-side fix |

## Success Criteria

- [ ] VM creation via `POST /api/v1/orchestrator/machines/async` succeeds when orchestrator is also the catalog
- [ ] Temp API keys are never visible in the UI or API key list
- [ ] Temp API keys are cleaned up after job completion (success or failure)
- [ ] Orphaned temp keys (from panics/hung jobs) are cleaned up by background process
- [ ] Existing API keys continue to work without modification
- [ ] Client-provided connections/catalog managers are unaffected