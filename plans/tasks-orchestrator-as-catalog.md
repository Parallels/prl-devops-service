# Task Breakdown: Orchestrator-as-Catalog VM Creation Support

Reference: `feature-orchestrator-as-catalog.md` (feature description)

---

## Task 1: Add Type Field to ApiKey Model

**Files:** `src/data/models/api_key.go`

### What to do

Add a `Type` field to the `ApiKey` struct:

```go
type ApiKey struct {
    // ... existing fields ...
    Type      string `json:"type,omitempty"` // "internal" or "external". Default: "external"
    // ...
}
```

### Validation

- [ ] `go build ./...` passes
- [ ] Existing code that creates ApiKeys still compiles (new field has zero-value default)
- [ ] JSON serialization works: `{"type": "internal"}` and `{}` (omitted when empty) both parse correctly

---

## Task 2: Normalize and Filter API Keys in DB Layer

**Files:** `src/data/api_key.go`

### What to do

Add two helper functions and modify `GetApiKeys`:

```go
// normalizeApiKeyType converts empty Type to "external" for backward compat
func normalizeApiKeyType(key *models.ApiKey) {
    if key.Type == "" {
        key.Type = "external"
    }
}

// filterInternalKeys removes internal keys from a list
func filterInternalKeys(keys []models.ApiKey) []models.ApiKey {
    result := make([]models.ApiKey, 0, len(keys))
    for _, k := range keys {
        if k.Type != "internal" {
            result = append(result, k)
        }
    }
    return result
}
```

In `GetApiKeys()`, after the existing filter logic:
```go
// Normalize types on all returned keys
for i := range filteredData {
    normalizeApiKeyType(&filteredData[i])
}
// Filter internal keys from list
filteredData = filterInternalKeys(filteredData)
```

In `GetApiKey()`, normalize before returning:
```go
normalizeApiKeyType(apiKey)
return apiKey, nil
```

In `CreateApiKey()`, set default type before hashing:
```go
if apiKey.Type == "" {
    apiKey.Type = "external"
}
```

### Validation

- [ ] `go build ./...` passes
- [ ] Existing API keys (without Type field) are treated as "external" — visible in UI
- [ ] New keys with `Type="internal"` are hidden from `GET /api/v1/auth/api_keys` response
- [ ] Single key lookup (`GetApiKey`) returns internal keys by name — needed for auth validation
- [ ] New key creation defaults to `"external"` type

---

## Task 3: Add ORCHESTRATOR_PUBLIC_URL Constant

**File:** `src/constants/main.go`

### What to do

Add constant:

```go
const ORCHESTRATOR_PUBLIC_URL = "ORCHESTRATOR_PUBLIC_URL"
```

### Validation

- [ ] `go build ./...` passes
- [ ] Constant is exported and accessible from other packages

---

## Task 4: Create buildLocalCatalogConnection Function

**File:** `src/controllers/machines.go`

### What to do

Add the function at the end of the file:

```go
func buildLocalCatalogConnection(ctx basecontext.ApiContext, callerID string, jobID string) (string, string, error) {
    cfg := config.Get()
    apiPort := cfg.ApiPort()
    if apiPort == "" {
        return "", "", fmt.Errorf("API port not configured")
    }

    schema := "http"
    host := os.Getenv(constants.ORCHESTRATOR_PUBLIC_URL)
    if host == "" {
        host = "localhost"
    }

    keyName := "temp-vm-" + jobID

    db := serviceprovider.Get().JsonDatabase
    _ = db.Connect(ctx)

    tempKey := models.ApiKey{
        Name:      keyName,
        Key:       keyName,
        Secret:    helpers.GenerateId(),
        Type:      "internal",
        UserID:    callerID,
        ExpiresAt: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
    }

    createdKey, err := db.CreateApiKey(ctx, tempKey)
    if err != nil {
        return "", "", fmt.Errorf("failed to create temp API key: %w", err)
    }

    connStr := fmt.Sprintf("host=%s@%s://%s:%s", createdKey.Secret, schema, host, apiPort)
    return connStr, keyName, nil
}
```

### Validation

- [ ] `go build ./...` passes
- [ ] Function creates an API key with correct structure (name, secret, type="internal", expires=+2h)
- [ ] Connection string format matches `CatalogManifestProvider.Parse()` expectations
- [ ] Returns plaintext secret in connection string (stored hashed in DB)

---

## Task 5: Modify AsyncCreateOrchestratorVirtualMachineHandler Goroutine

**File:** `src/controllers/orchestrator.go`

### What to do

Modify the goroutine starting around line 4232. Add temp key creation and cleanup:

```go
go func(jobID string, req models.CreateVirtualMachineRequest) {
    asyncCtx := basecontext.NewRootBaseContext()
    
    var cleanupFn func()
    defer func() {
        if cleanupFn != nil {
            cleanupFn()
        }
    }()
    
    defer func() {
        if rec := recover(); rec != nil {
            asyncCtx.LogErrorf("[Orchestrator] Panic in async create goroutine for job %s: %v", jobID, rec)
            _ = jobManager.MarkJobError(jobID, fmt.Errorf("internal error: %v", rec))
        }
    }()
    
    // NEW: Build connection string if orchestrator is the catalog
    if req.CatalogManifest != nil && req.CatalogManifest.Connection == "" {
        connStr, keyName, err := buildLocalCatalogConnection(asyncCtx, callerID, jobID)
        if err != nil {
            _ = jobManager.MarkJobError(jobID, fmt.Errorf("failed to resolve catalog connection: %w", err))
            return
        }
        req.CatalogManifest.Connection = connStr
        cleanupFn = func() {
            _ = db.DeleteApiKey(asyncCtx, keyName)
        }
    }
    
    // ... rest unchanged
}(job.ID, request)
```

Also ensure `db` variable is accessible in the goroutine closure (may need to capture it earlier).

### Validation

- [ ] `go build ./...` passes
- [ ] When `catalog_manifest.connection` is empty AND orchestrator is catalog → temp key created, connection populated
- [ ] When `catalog_manifest.connection` is provided → no temp key created (existing behavior)
- [ ] On goroutine panic → deferred cleanup runs, temp key deleted
- [ ] On goroutine success → deferred cleanup runs, temp key deleted
- [ ] On goroutine error → deferred cleanup runs, temp key deleted
- [ ] Job progress/completion/error marking works correctly with/without temp key path

---

## Task 6: Add Periodic Temp Key Cleanup Goroutine

**File:** `src/orchestrator/main.go` or startup sequence

### What to do

Add a background goroutine that cleans up orphaned temp keys every hour:

```go
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        cleanupTempKeys()
    }
}()

func cleanupTempKeys() {
    db := serviceprovider.Get().JsonDatabase
    _ = db.Connect(context.Background())
    keys, err := db.GetApiKeys(context.Background(), "")
    if err != nil {
        return
    }
    cutoff := time.Now().Add(-3 * time.Hour)
    for _, k := range keys {
        if strings.HasPrefix(k.Name, "temp-vm-") {
            if expiresAt, err := time.Parse(time.RFC3339, k.ExpiresAt); err == nil && expiresAt.Before(cutoff) {
                _ = db.DeleteApiKey(context.Background(), k.ID)
            }
        }
    }
}
```

Note: This cleanup goroutine calls `GetApiKeys()` which filters internal keys. To catch orphaned internal keys, either:
1. Call the raw database query directly (bypassing the filter), OR
2. Use `GetApiKey` with known temp key names pattern

Option 1 is preferred — query directly without filtering.

### Validation

- [ ] `go build ./...` passes
- [ ] Background goroutine starts on orchestrator initialization
- [ ] Orphaned temp keys (expired > 3 hours ago) are deleted on each tick
- [ ] Non-temp keys are never affected
- [ ] No errors logged when no orphaned keys exist

---

## Integration Test Plan

Run these tests after implementing all tasks:

| # | Scenario | Expected Result |
|---|---|---|
| 1 | Orchestrator IS catalog, client provides no connection | VM created successfully, temp key cleaned up |
| 2 | Orchestrator IS catalog, client provides own connection | Uses client's connection, no temp key |
| 3 | Orchestrator NOT catalog | Falls through to existing error path |
| 4 | Multiple concurrent VM creations | Each gets unique temp key, all succeed |
| 5 | Force panic in goroutine | Temp key cleaned up via defer |
| 6 | Manually insert expired temp key | Background cleanup deletes it |
| 7 | Existing API keys (no Type field) | Still visible in UI/API list |
| 8 | New API key with type="internal" | Hidden from API key list endpoint |