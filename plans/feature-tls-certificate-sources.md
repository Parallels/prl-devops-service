# Feature: Pluggable TLS Certificate Sources and Hot Reload

## Overview

Today the DevOps Service only accepts TLS material as base64-encoded PEM strings in
`TLS_CERTIFICATE` and `TLS_PRIVATE_KEY`. A customer wants to use host certificates
issued by their PKI (Venafi) and enrolled/renewed by Jamf over SCEP. Those
certificates live in the macOS **System Keychain** with **non-exportable private
keys**, so they can never be turned into base64 PEM values.

This feature:

1. Introduces a **certificate source** abstraction so the TLS listener no longer
   needs the private key bytes. It only needs a `crypto.Signer`.
2. Adds new sources: **PEM files on disk** (all OSs), **macOS Keychain** (darwin),
   **Windows Certificate Store** (windows). The existing base64 env-var source stays
   the default and keeps working unchanged.
3. Makes the HTTPS listener **reload a renewed certificate without restarting** the
   service, for every source.
4. Documents and tightens restart behaviour for other config changes and upgrades.

The customer questions this answers:

| Question | Answer today | Answer after this feature |
|---|---|---|
| Can the service use a System Keychain identity with a non-exportable key, without user interaction? | No. Only base64 PEM is supported. | Yes, on macOS, provided the key's access control allows the service binary (see [Key access control](#key-access-control-macos)). |
| Can a renewed certificate be picked up without a restart? | No. The cert is parsed once at startup. | Yes. New TLS handshakes use the new certificate; open connections are not dropped. |
| Does a restart (config change, upgrade) affect running VMs? | No. See [Restart impact](#restart-impact-today). | No. |
| Does a restart affect jobs in progress? | Yes. They are marked failed. | Yes, still. Restart is still required for non-TLS config and upgrades. Graceful drain is improved (Phase 4). |

## Current State (findings)

### TLS loading

- Config keys: `TLS_ENABLED`, `TLS_PORT`, `TLS_CERTIFICATE`, `TLS_PRIVATE_KEY`,
  `DISABLE_HTTP_WHEN_TLS` (`src/constants/main.go:98-101`).
- Listener options read the base64 values once
  (`src/restapi/http_listener.go:597-603`).
- `HttpListener.Start` builds a static `tls.Config{Certificates: []tls.Certificate{cert}}`
  from `tls.X509KeyPair` (`src/restapi/http_listener.go:428-436`). On failure it
  logs and **falls back to HTTP-only**.
- `config.TlsEnabled()` returns false unless both base64 values decode
  (`src/config/network.go:76-85`). `ValidateTlsConfiguration` also assumes PEM
  key bytes (`src/config/network.go:89-110`).
- Other users of `TLS_ENABLED`: orchestrator registration and deployer
  (`src/cmd/register_with_orchestrator.go:313`, `src/orchestrator/deployer.go:99`),
  installer (`src/install/main.go:150-156,484-493`).

### Config reload

- Config is read once at startup. No file watcher, no `SIGHUP` handler.
- Only `SIGINT`/`SIGTERM` are handled, in two places:
  - `src/main.go:67-100`: runs DB cleanup, then calls `os.Exit(0)`.
  - `src/restapi/http_listener.go:550-592`: calls `Server.Shutdown` with a context.
  - Both race on the same signal. `main.go` can exit before HTTP servers drain.

### Build constraints

- Every release build uses `CGO_ENABLED=0`, including darwin
  (`.github/workflows/release*.yml`, `Makefile:132-154`).
- Darwin binaries are built and codesigned on a macOS job with an Apple
  Developer ID certificate. Linux/Windows build on `ubuntu-latest`.
- `golang.org/x/sys v0.49.0` is already a dependency.

### Service definition

- macOS: LaunchDaemon, `UserName root`, `KeepAlive.SuccessfulExit=false`
  (`src/install/mac_plist_daemon.go:17-45`). A clean exit (code 0) is **not**
  relaunched by launchd.
- Linux: systemd, `User=root`, `Restart=on-failure`
  (`src/install/linux_systemd_unit.go:18-32`).

### Restart impact today

- **VMs:** not affected. VMs are owned by Parallels Desktop's dispatcher, not by
  this process. The service only runs `prlctl` commands and a `prlctl` event
  monitor (`src/serviceprovider/parallelsdesktop/event_monitor.go:69`). Stopping
  the service stops monitoring, not VMs.
- **Jobs in progress:** at startup `RecoverOngoingJobs` marks every `running` or
  `pending` job as `failed` with "Service crashed or was restarted"
  (`src/database/jobs.go:221`, `src/startup/main.go:176`). Long operations
  (catalog pull, clone, create) do not resume.
  - Open question to verify: a `prlctl` child started via `exec.CommandContext`
    may keep running after `os.Exit`, because the context is never cancelled.
    The VM operation could then finish while the job record says `failed`.
- **Orchestrator:** the host goes unhealthy until it is back. WebSocket
  connections drop and reconnect.
- **API clients:** in-flight HTTP requests can be cut, because `main.go` may call
  `os.Exit(0)` before `Server.Shutdown` finishes.

## Design

### Principles

- Keep `TLS_CERTIFICATE` / `TLS_PRIVATE_KEY` as the default. No behaviour change
  for existing installs.
- Same config surface on every OS. A source that is not available on the
  current OS fails validation with a clear error. It is never silently ignored.
- Never require private key bytes in memory. The listener uses `crypto.Signer`.
- Keep `CGO_ENABLED=0` for every target (see [Decision D1](#d1-how-to-call-platform-apis-without-cgo)).
- No secrets in logs. Log fingerprints, subject, issuer and expiry only.

### New configuration keys

| Key | Values | Default | Notes |
|---|---|---|---|
| `TLS_CERTIFICATE_SOURCE` | `env`, `file`, `keychain`, `windows_store` | `env` | Selects the source. `env` = current base64 behaviour. |
| `TLS_CERTIFICATE_FILE` | path | | `file` source. PEM, may contain the full chain. |
| `TLS_PRIVATE_KEY_FILE` | path | | `file` source. PEM, PKCS#1/PKCS#8/EC. |
| `TLS_CERTIFICATE_SELECTOR` | see below | | `keychain` and `windows_store` sources. |
| `TLS_KEYCHAIN` | `system`, `login`, absolute path | `system` | `keychain` source only. |
| `TLS_WINDOWS_STORE` | `LocalMachine\My`, `CurrentUser\My` | `LocalMachine\My` | `windows_store` source only. |
| `TLS_CHAIN_FILE` | path | | Optional intermediates for store sources that do not return a chain. |
| `TLS_RELOAD_INTERVAL` | duration | `5m` | How often the source is re-checked. `0` disables polling. |

`TLS_CERTIFICATE_SELECTOR` format, one of:

- `sha1:<hex>` or `sha256:<hex>`: certificate thumbprint. Exact, but changes on
  every renewal, so it is a poor fit for SCEP.
- `subject:<substring>`: matches the subject common name. Fits SCEP renewal,
  because the renewed cert keeps the same subject.
- `issuer:<substring>`: optional extra filter, combined with `subject:` using `;`.
  Example: `subject:mac-host-01.corp.example;issuer:Venafi Issuing CA`.

When several identities match, pick the one that is currently valid with the
latest `NotAfter`. That rule makes renewal work: the new cert wins as soon as it
is in the store, and the old one is ignored even before it is deleted.

All keys can be set in `prldevops_config.yaml` or as environment variables, like
the existing keys. The installer (`src/install/main.go`) must accept and write
them.

### Component layout

New package `src/security/certsource`:

```
certsource/
  source.go            // interface, selector parsing, selection rule
  env.go               // base64 env vars (current behaviour)
  file.go              // PEM files, all OSs
  keychain_darwin.go   // macOS Security.framework
  keychain_other.go    // !darwin: returns ErrSourceUnsupported
  winstore_windows.go  // CryptoAPI / CNG
  winstore_other.go    // !windows: returns ErrSourceUnsupported
  reloader.go          // caching, polling, GetCertificate
  signer_darwin.go     // crypto.Signer backed by SecKeyRef
  signer_windows.go    // crypto.Signer backed by NCRYPT_KEY_HANDLE
```

Interface:

```go
type Source interface {
    // Name is used in logs and errors, e.g. "keychain".
    Name() string
    // Load returns the current certificate. It must not prompt the user.
    Load(ctx context.Context) (*tls.Certificate, error)
    // Close releases OS handles held by previously loaded certificates.
    Close() error
}

var ErrSourceUnsupported = errors.New("certificate source is not supported on this operating system")
```

`tls.Certificate.PrivateKey` is a `crypto.Signer`. For `env` and `file` it is the
parsed Go key. For store sources it is an OS-backed signer.

### Reloader and listener change

`Reloader` wraps a `Source`:

- Loads once at startup. Startup failure keeps today's behaviour: log the error
  and fall back to HTTP-only (unless `DISABLE_HTTP_WHEN_TLS` is set, in which
  case the service should exit non-zero. Today it silently starts plain HTTP
  anyway, because the flag only applies when TLS started).
- Re-loads every `TLS_RELOAD_INTERVAL`. For `file`, also re-load when the file
  modification time changes (stat on each tick, no fsnotify dependency).
- Swaps the cached certificate atomically (`atomic.Pointer[tls.Certificate]`)
  only when the leaf fingerprint changes and the new cert parses and passes a
  test signature.
- On reload failure keeps serving the last good certificate and logs a warning.
  A new certificate that cannot sign (for example a renewed key whose ACL does
  not allow the service) counts as a reload failure.
- Logs a warning when the active certificate expires within 14 days. While a
  newer certificate exists that the service cannot use, logs an error on every
  tick, naming both certificates and their expiry dates.
- If the active certificate stops signing (handshake signature errors), logs an
  error and re-runs the source immediately instead of waiting for the next tick.

Limits of "keep the last good certificate" for store sources:

- Store sources never hold the private key in memory. They hold an OS handle
  (`SecKeyRef`, `NCRYPT_KEY_HANDLE`), and every signature still goes through the
  OS. Keeping the old certificate only works while the old key still exists and
  the old certificate has not expired. It buys time until expiry. It is not a
  fix.
- Two cases are probably not covered, both unverified (see the spike):
  - The old identity is deleted from the keychain at renewal. The retained
    handle may stop working.
  - The ACL of the key in use is changed. The keychain likely checks the ACL on
    each use, so signing would fail.
- When no usable certificate is left, the listener applies the startup rule:
  fall back to HTTP-only, or exit non-zero when `DISABLE_HTTP_WHEN_TLS=true`.
  The log must say why.

Listener change in `src/restapi/http_listener.go`:

```go
tlsConfig := &tls.Config{
    MinVersion:     tls.VersionTLS12,
    GetCertificate: reloader.GetCertificate,
}
```

`GetCertificate` runs per handshake, so a new certificate applies to new
connections immediately. Existing connections keep their session and are not
dropped. This answers the "reload without restart" question for all sources.

Also remove `InsecureSkipVerify: true` from the server `tls.Config`. It has no
effect on a server and misleads readers.

`config.TlsEnabled()` and `ValidateTlsConfiguration()` must become
source-aware. They should check that the source is configured, not that PEM
strings exist. Callers in the orchestrator registration and deployer only need
the boolean.

### macOS Keychain source

API flow (Security.framework):

1. Build a query: `kSecClass = kSecClassIdentity`, `kSecMatchLimit = kSecMatchLimitAll`,
   `kSecReturnRef = true`, `kSecMatchSearchList = [System keychain]`.
2. `SecItemCopyMatching` returns `SecIdentityRef`s.
3. For each identity: `SecIdentityCopyCertificate` then `SecCertificateCopyData`
   gives DER. Parse with `crypto/x509`, apply the selector and selection rule.
4. `SecIdentityCopyPrivateKey` gives a `SecKeyRef`. The key is never exported.
5. Chain: build with `SecTrustCreateWithCertificates` + `SecTrustEvaluateWithError`,
   then `SecTrustCopyCertificateChain` (macOS 12+; `SecTrustGetCertificateAtIndex`
   before that). Drop the root. If `TLS_CHAIN_FILE` is set, use it instead.
6. Signer: `SecKeyCreateSignature` with the algorithm matched to Go's request:

| Key | `crypto.SignerOpts` | `SecKeyAlgorithm` |
|---|---|---|
| RSA | PKCS#1 v1.5, SHA-256/384/512 | `kSecKeyAlgorithmRSASignatureDigestPKCS1v15SHA256` (…384/512) |
| RSA | `*rsa.PSSOptions`, SHA-256/384/512 | `kSecKeyAlgorithmRSASignatureDigestPSSSHA256` (…384/512) |
| ECDSA | SHA-256/384/512 | `kSecKeyAlgorithmECDSASignatureDigestX962SHA256` (…384/512) |

TLS 1.3 requires RSA-PSS for RSA keys. If a key refuses PSS, the signer must
return an error, and the spec requires a test for that case. A fallback is to
set `MaxVersion: tls.VersionTLS12` for that source. That fallback is a config
option, not automatic.

The signer holds a retained `SecKeyRef`. Release it with `CFRelease` when the
reloader replaces the certificate, after a grace period, so in-flight handshakes
that already picked the old certificate can finish.

#### Key access control (macOS)

This is the main deployment risk and must be in the customer docs.

**Access is the customer's responsibility.** The service never grants itself
access to a key. It uses an identity the customer's Identity team made
accessible, and fails with a clear message when it is not.

How macOS decides access:

- Reading a **certificate** needs no ACL entry and no prompt. Detecting that a
  certificate changed is therefore always possible.
- Using a **private key** to sign is controlled by that key's access control
  list (ACL). There is no global "this executable may use every key" grant.
  Root does not bypass the ACL. No entitlement does either:
  `keychain-access-groups` applies to the data protection keychain, not to the
  System keychain where SCEP identities are stored.
- The ACL belongs to the **key**, not the certificate. SCEP renewal normally
  generates a new key pair, so every renewal produces a key with a fresh ACL.
- The ACL of a new key is set by whoever creates the key. With SCEP that is the
  Jamf configuration profile.
- If the service binary is not allowed, `SecKeyCreateSignature` returns
  `errSecInteractionNotAllowed` in a daemon (no UI) instead of prompting. The
  service must detect that error and log an actionable message.

Supported customer setups:

- **Recommended:** the SCEP payload sets "Allow all apps access"
  (`AllowAllAppsAccess`) to true and keeps "Allow export from keychain"
  (`KeyIsExtractable`) false. The key stays non-exportable. Jamf applies the
  setting to every key it creates, so renewals need no further action.
- **Fallback:** a Jamf policy script runs after each renewal and adds the service
  binary to the new key's ACL. The ACL entry is bound to the binary's code
  signature requirement. Our binary is signed with our Developer ID, so the
  entry survives upgrades as long as the signing identity and identifier do not
  change. **We must keep the identifier stable** and document it.

Rejected: granting access from our installer.

- An ACL entry added at install time covers only the key that exists then. The
  first renewal creates a new key without it, and TLS breaks.
- Changing a key's ACL usually needs the keychain password or shows an
  authorization dialog, even as root. Installs are often headless (Jamf, CI).
- Profile-managed identities may block or revert changes made outside the
  profile (unverified, part of the spike).

Needs validation on a real Jamf-managed Mac before we commit to a timeline.
See [Validation plan](#validation-plan).

### Windows Certificate Store source

Pure syscalls through `golang.org/x/sys/windows`, no cgo:

1. `CertOpenStore` on `LocalMachine\My` (or the configured store).
2. Enumerate with `CertEnumCertificatesInStore`, parse DER, apply the selector.
3. `CryptAcquireCertificatePrivateKey` with
   `CRYPT_ACQUIRE_ONLY_NCRYPT_KEY_FLAG | CRYPT_ACQUIRE_SILENT_FLAG`. The silent
   flag guarantees no UI.
4. Signer: `NCryptSignHash` (lazy-loaded from `ncrypt.dll`) with
   `BCRYPT_PAD_PKCS1` or `BCRYPT_PAD_PSS` for RSA, raw for ECDSA. Convert the
   ECDSA `r||s` output to ASN.1 DER, which Go's TLS stack expects.
5. Chain: `CertGetCertificateChain`.

The service account needs read access to the private key. LocalSystem has it by
default. A custom account needs it granted through the certificate's
"Manage Private Keys" permissions.

### Linux

No standard system certificate store holds private keys. Supported sources:
`env` and `file`. `file` with hot reload covers certbot, cert-manager, Vault
Agent and similar tools, which all write PEM files and rotate them.
PKCS#11 / TPM is out of scope (see [Out of scope](#out-of-scope)).

### OS support matrix

| Source | macOS | Windows | Linux |
|---|---|---|---|
| `env` (default) | yes | yes | yes |
| `file` | yes | yes | yes |
| `keychain` | yes | unsupported error | unsupported error |
| `windows_store` | unsupported error | yes | unsupported error |
| Hot reload | all sources | all sources | all sources |

## Decisions

### D1: How to call platform APIs without cgo

All builds use `CGO_ENABLED=0`. The Windows source does not need cgo. The macOS
source needs Security.framework and CoreFoundation.

| Option | Pros | Cons |
|---|---|---|
| **A. `github.com/ebitengine/purego` (recommended)** | Keeps `CGO_ENABLED=0`. No pipeline change. Cross-compiles. Mature, widely used. | CoreFoundation dictionaries built by hand through `dlsym`. More code, needs careful `CFRelease`. |
| B. cgo for darwin only | Simplest Security.framework code. | Darwin job must switch to `CGO_ENABLED=1` and build each arch with clang. Local `make` targets change. Mixed build modes across OSs. |
| C. Shell out to `security` CLI | Trivial. | Cannot sign with a non-exportable key. Not viable. |

Recommendation: A. Revisit B only if purego proves unstable in the validation spike.

### D2: Reload mechanism

Polling with `TLS_RELOAD_INTERVAL` (default 5 minutes), plus `file` mtime checks
on the same tick. No fsnotify and no Keychain notification APIs. SCEP renewal
happens days before expiry, so a 5-minute delay is fine. Optionally add `SIGHUP`
to force an immediate TLS reload (Phase 3).

### D3: Config reload beyond TLS

Out of scope for hot reload. Most keys (ports, modules, database, auth, CORS)
are wired at startup into singletons, routers and listeners. Making them
hot-reloadable is a large refactor with a high regression risk. This feature
keeps "restart required" for non-TLS changes and makes restarts cleaner instead
(Phase 4).

## Phases

### Phase 1: Abstraction, `env` and `file` sources, hot reload (all OSs)

- `certsource` package with `Source`, selector parsing, selection rule.
- `env` source wrapping today's behaviour. Must produce identical results.
- `file` source.
- `Reloader` + `GetCertificate` in `http_listener.go`.
- Source-aware `TlsEnabled` / `ValidateTlsConfiguration`.
- Installer and `cmd/api.go` help text for new keys.
- Tests (London-school, mock `Source`): reload swaps on fingerprint change,
  keeps last good cert on failure, selection rule, expiry warning, `env` parity.
- Integration test: start listener with `file`, rotate files, assert the new
  leaf is served on a new connection while an old connection stays open.

### Phase 2: macOS Keychain source

- Validation spike first (see below). Go/no-go before building the full source.
- purego bindings, identity lookup, chain, `SecKey` signer, handle lifetime.
- Clear errors for: no match, `errSecInteractionNotAllowed`, PSS unsupported.
  The `errSecInteractionNotAllowed` message must point to the Jamf SCEP
  `AllowAllAppsAccess` setting.
- Installer check (`src/install/main.go`): when `TLS_CERTIFICATE_SOURCE=keychain`,
  find the configured identity and do one test signature. On failure, report
  the same actionable message. The installer checks access. It never changes
  the ACL.
- Tests: unit tests on selector/selection with fake identities; darwin-only
  integration test that creates a temporary keychain with a non-exportable key.
- Docs: Jamf SCEP payload requirements, ACL, stable code-signing identifier.

### Phase 3: Windows Certificate Store source, `SIGHUP` reload

- CryptoAPI / CNG bindings and signer.
- Windows-only integration test with a non-exportable key in `CurrentUser\My`.
- `SIGHUP` triggers an immediate TLS reload on darwin/linux. Windows has no
  `SIGHUP`, so polling is the only path there.

### Phase 4: Graceful restart

Independent of the certificate work, but answers the restart questions.

- Single shutdown path: one signal handler drains HTTP servers
  (`Server.Shutdown` with timeout), stops the event monitor, cancels job
  contexts, saves the DB, then exits. Remove the race between `main.go` and
  `http_listener.go`.
- Cancel the context of running `prlctl` children on shutdown, or decide to let
  them finish and reconcile on startup. Needs the open question above answered.
- Document restart commands:
  - macOS: `sudo launchctl kickstart -k system/com.parallels.devops-service`.
    Plain `launchctl stop` gives exit 0, which launchd does not relaunch because
    of `SuccessfulExit=false`.
  - Linux: `sudo systemctl restart prl-devops-service`.
- Document what a restart affects: VMs no, in-flight jobs yes, orchestrator
  health briefly.

## Validation plan

Before committing to a Phase 2 timeline, run a spike on a Jamf-managed test Mac:

1. Enroll a SCEP certificate into the System keychain with
   `KeyIsExtractable=false`, once with `AllowAllAppsAccess=true` and once false.
2. From a signed test binary running as a root LaunchDaemon, find the identity
   and call `SecKeyCreateSignature` with RSA PKCS#1, RSA-PSS and ECDSA.
3. Record which combinations work without a prompt.
4. Complete a TLS 1.2 and TLS 1.3 handshake against `curl` and the orchestrator.
5. Trigger a SCEP renewal and confirm the reloader picks up the new identity.
6. Replace the binary with a newer signed build and confirm ACL access survives.
7. Check whether Jamf deletes the old identity at renewal. If it does, check
   whether a retained `SecKeyRef` to the deleted key can still sign.
8. Remove the service binary from the ACL of the key in use, and check whether
   a retained `SecKeyRef` can still sign.
9. Change the ACL of a profile-installed key from outside the profile, and check
   whether the change is blocked or reverted on the next profile push.

## Estimate

Rough, for one engineer, before the spike:

| Phase | Effort |
|---|---|
| Spike (macOS) | 2-3 days |
| Phase 1 | 4-5 days |
| Phase 2 | 6-8 days |
| Phase 3 | 4-5 days |
| Phase 4 | 3-4 days |

Customer-facing timeline should only be given after the spike, by the product
owner. **Draft: review before sharing externally.**

## Out of scope

- Hot reload of non-TLS configuration.
- Zero-downtime binary upgrades (socket handoff between processes).
- PKCS#11 / HSM / TPM-backed keys on Linux.
- Client certificate (mTLS) authentication of the orchestrator. The same
  `certsource` package could later provide the client cert for the
  orchestrator-to-host WebSocket and health checks.
- Changing the orchestrator's `InsecureSkipVerify` client settings
  (`src/controllers/orchestrator.go:3856`, `src/orchestrator/get_hosts_health_check.go:46`).
  Worth a separate ticket: a properly issued host cert enables real verification.

## Open questions

1. Will the customer's Identity team set `AllowAllAppsAccess=true` in the SCEP
   payload, or run a Jamf script after each renewal to update the key ACL?
   Either way the setup is theirs. The service only checks access.
2. Key type and size issued by Venafi (RSA 2048/3072 or ECDSA P-256/P-384)?
3. Does the renewed certificate keep the same subject CN? The selection rule
   assumes it does.
4. Do `prlctl` children keep running after the service exits? Decides the
   Phase 4 job handling.
5. Should `DISABLE_HTTP_WHEN_TLS=true` with a broken TLS source exit non-zero
   instead of silently falling back to plain HTTP
   (`src/restapi/http_listener.go:486`)? This spec proposes yes.
