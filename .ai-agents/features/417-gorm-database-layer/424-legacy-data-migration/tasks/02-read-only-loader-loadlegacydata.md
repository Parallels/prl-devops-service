# Task 02 — Read-only legacy loader `data.LoadLegacyData`

**PBI** [#424](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 01 · **PR size:** M
**Plan refs:** [D-2](../plan.md#d-2--read-only-loader-additive-dateloadlegacydata-never-jsondatabase), [D-10](../plan.md#d-10--fixtures-committed-tiny-json-everything-generated-at-test-time), locked **AD-5**

## Context

AD-5 forbids hand-parsing `data.json` and AD-5's practical form — "reuse the legacy loader" — cannot mean
`JsonDatabase`: `NewJsonDatabase` (`src/data/main.go:75-125`) binds the package singleton, calls `Load`
(may write), and starts the save-interval (`:102`), backup (`:107`), and ghost-job (`:110-122`) goroutines.
Even `loadFromFile` itself mutates: it invokes `j.Backup(ctx)` *before* reading (`:535`) and
`RecoverOngoingJobs` after (`:578`), and `Load`'s residual recovery calls `SaveNow` + deletes the `.save`
files (`:476-484`). This task adds the one **additive** exported helper the brief pre-approves
("legacy code is frozen except additive helpers approved in this PBI's tasks"): decode with the *same
rules* as `loadFromFile`, side effects zero.

## Objective

`data.LoadLegacyData(ctx, filename) (Data, error)` returns the full in-memory `Data` struct
(`src/data/main.go:33-51`, 16 collections incl. pointer `Configuration`/`ReverseProxy`) for
plaintext **and** encrypted files, falls back read-only to residual `*.save`/`*.save_bak`/`*.panic`
files, never creates/modifies/deletes anything, and never touches `memoryDatabase`.

## Files

| File | Op |
|---|---|
| `src/data/legacy_load.go` | create (~120 lines, package `data`) |
| `src/data/legacy_load_test.go` | create |
| `src/database/legacy_migrator/testdata/golden.json` | create (hand-crafted synthetic) |
| `src/database/legacy_migrator/testdata/partial.json` | create |
| `src/database/legacy_migrator/testdata/corrupt.json` | create |

`src/data/main.go` is **not edited** (686 lines stays 686; pure addition of a new same-package file).

## File size guard

Real `wc -l` (2026-11-03): `src/data/main.go` **686** → untouched, **ok**. New `legacy_load.go` ≈ 120,
test ≈ 220. Fixtures < 100 lines JSON each. Verdict: **ok** everywhere; no ER-3 trigger.

## Implementation spec

### `src/data/legacy_load.go`

```go
package data

// LoadLegacyData reads a legacy data file into the Data struct WITHOUT side effects:
// no file creation, no backup (contrast loadFromFile's j.Backup at main.go:535), no save,
// no singleton, no goroutines, no RecoverOngoingJobs (main.go:578). Read-only residual
// recovery (*.save → *.save_bak → *.panic, Load's order at main.go:135-157) decodes the
// newest glob match as a FALLBACK ONLY — never SaveNow, never removeAllBackupSavedFiles
// (contrast recoverFromResidualSaveFiles, main.go:476-484): the legacy store is frozen (OQ-1).
// Decoding mirrors loadFromFile (main.go:550-570) EXACTLY — keep the two in sync if the
// legacy format ever changes; migration is a one-time consumer, legacy stays canonical.
//
// #424 additive helper — approved in PBI #424 tasks. Do not call from legacy paths.
func LoadLegacyData(ctx basecontext.ApiContext, filename string) (Data, error) {
        var empty Data
        info, err := os.Stat(filename)
        if err != nil || info.Size() == 0 {
                if os.IsNotExist(err) { return empty, ErrNoLegacyDataFile }
                if err != nil { return empty, errors.NewFromError(err) }
                return empty, ErrLegacyDataFileEmpty
        }
        raw, err := helper.ReadFromFile(filename)          // same reader as loadFromFile:539
        if err != nil { return empty, errors.NewFromError(err) }
        data, err := decodeLegacyBytes(ctx, raw, filename) // plaintext → encrypted (decrypt), main.go:550-570
        if err == nil { return data, nil }
        for _, glob := range []string{"*.save", "*.save_bak", "*.panic"} {   // read-only recovery
                matches, _ := filepath.Glob(filename + glob)
                if len(matches) == 0 { continue }
                r2, rErr := helper.ReadFromFile(matches[len(matches)-1])     // newest, legacy semantics
                if rErr != nil || len(r2) == 0 { continue }
                if d2, dErr := decodeLegacyBytes(ctx, r2, matches[len(matches)-1]); dErr == nil {
                        ctx.LogWarnf("[LegacyMigration] main file unreadable; decoded residual %s (read-only)", matches[len(matches)-1])
                        return d2, nil
                }
        }
        return empty, err
}

// decodeLegacyBytes = the exact main.go:550-570 sequence, extracted so loadFromFile could
// call it later — but this task does NOT refactor loadFromFile (frozen; no behaviour risk).
func decodeLegacyBytes(ctx basecontext.ApiContext, content []byte, origin string) (Data, error) {
        var data Data
        if err := json.Unmarshal(content, &data); err == nil { return data, nil }
        cfg := config.Get()
        if cfg.EncryptionPrivateKey() == "" {
                return data, errors.NewWithCodeAndDescription(
                        fmt.Sprintf("failed to parse legacy data file %s: %v (encrypted file? set ENCRYPTION_PRIVATE_KEY)", origin, err), 500,
                        "the legacy data store may be encrypted; the migration needs the encryption key")
        }
        plain, err := security.DecryptString(cfg.EncryptionPrivateKey(), content)
        if err != nil { return data, errors.NewFromError(err) }
        if err := json.Unmarshal([]byte(plain), &data); err != nil { return data, errors.NewFromError(err) }
        return data, nil
}
```

Plus two sentinels appended to `src/data/main.go`'s `var (…)`? **No** — main.go frozen; declare
`ErrNoLegacyDataFile` / `ErrLegacyDataFileEmpty` in the new file (same package, same idiom:
`errors.NewWithCode(msg, 404|404)`). `config.EncryptionPrivateKey()` returns the **PEM content**
(base64-decoded env, `src/config/main.go:261-274`), `security.DecryptString(privateKey string,
cipherText []byte)` is the exact call used at `main.go:559`.

### Fixtures (D-10)

`golden.json`: one `schema`, a `configuration` object, 3 users (one with 2 roles + 2 embedded claims),
2 roles each with claims, top-level `claims`, 2 api_keys, 1 packer template, 1 catalog manifest with
`virtual_machine_contents`, 1 orchestrator host, 1 snapshot record, `reverse_proxy` object, 1 rp host,
1 catalog manager, 2 jobs with steps, 1 vm_snapshots record, 1 enrollment token, 1 user_config. Hand-
crafted values; string dates like the models produce (`"created_at":"2024-01-02 15:04:05"`).

### Tests (`legacy_load_test.go`, package `data`, testify)

| Test | Assert |
|---|---|
| plaintext golden | all 16 collections round-trip (`len`, pointer non-nil); **no new files**: dir listing before == after |
| encrypted | `security.GenPrivateRsaKey(t.TempDir()+"/k.pem", 2048)` → read PEM → `security.EncryptString(pem, string(goldenBytes))` → write raw bytes as the "data.json" → `t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, base64.StdEncoding.EncodeToString(pemPEMBytes))` → load ⇒ equals golden. **Generated at test time, never committed** — `EncryptString` embeds a fresh random AES key + PKCS1v15 randomness, so golden ciphertext bytes are impossible to pin; the *decrypt* direction is what matters and it's proven end-to-end (D-10 rationale in PR description) |
| encrypted, key missing | `t.Setenv(key, "")` ⇒ error message contains `ENCRYPTION_PRIVATE_KEY`; no panic |
| corrupt JSON | `corrupt.json` (random non-JSON bytes), no key ⇒ error, zero-value `Data` |
| absent file | `ErrNoLegacyDataFile` (404 via `errors.GetSystemErrorCode`) |
| zero-byte file | `ErrLegacyDataFileEmpty`; **file not created** when path absent |
| residual `.save` | main file corrupt + `data.json.save` = valid JSON ⇒ decoded + warn log captured; `.save` file still on disk afterwards, main file bytes unchanged (read-only!) |
| no globals | `memoryDatabase == nil` before/after (same-package test can assert it directly) |

## Out of scope

Any change to `loadFromFile`/`Load`/`IsDataFileEmpty` bodies; wiring into the migrator (task 03);
`.save.bak.*` backup-chain recovery (legacy `recoverFromBackupFile` is a *repair* path — frozen-legacy
decision OQ-D says no).

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./data/... -count=1` — new tests + untouched legacy suite green (`memoryDatabase` assertions prove no singleton pollution)
- [ ] `cd src && go test ./data/ -run 'TestLoadLegacyData' -count=1 -v` — table above visible, < 2 s
- [ ] Mutation audit: test that writes the encrypted fixture asserts dir-snapshot equality pre/post (in-suite, not manual)
- [ ] `cd src && go test ./... -count=1` regression green

## Rollback

Revert PR removes `legacy_load.go` + tests + 3 fixtures; no existing code path referenced them.

## PBI acceptance mapping

Enables **AC-3/AC-7** (the migration and dry-run read their source through this helper — AD-5) and
keeps the *"data.json is never modified"* guarantee testable. DoD "unit tests written" for the loader.
