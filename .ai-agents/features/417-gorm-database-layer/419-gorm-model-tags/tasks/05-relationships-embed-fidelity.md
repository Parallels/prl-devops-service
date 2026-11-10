# Task 05 — Relationships in schema + embed-fidelity parity tests (GORM read ≡ legacy read)

## Context
All tags exist after waves 02–04. This PR proves the **read shapes** survive normalisation: the
golden constraint from `users.go:29-59` — `GetUsers` re-fills every embedded `Role.Claims` from the
live `Roles` table at read time — must be reproducible from GORM with the D-7 preload chain, and the
marshalled result must be **byte-identical** to the legacy engine's output for the same fixture data.

## Objective
A parity test suite where each fixture is loaded twice — once as legacy `models.*` values, once
seeded through GORM (join rows, has-many children) — and `json.Marshal(GORM read w/ preload)` ==
`json.Marshal(legacy value)` byte-for-byte, for: user(+roles+claims+role-claims live-update),
host(+reverse proxy+proxy hosts+VM list), snapshot map regroup, api_key/manifest serialiser
nil-vs-empty matrix.

## Files
- Create: `src/data/models/relationships_parity_test.go` (`package models_test`)
- Create: `testdata/parity/` fixtures (`users_store.json`, `hosts_store.json`, `snapshots_store.json` —
  captured from legacy JSON-store shape)
- Modify (only if a gap is found): relation tag tweaks in the four wave files — any change here must
  be re-golden-checked (02–04 tests are the tripwire)

## File size guard (ER-4)
| File | Now | After (proj.) | Verdict |
|---|---|---|---|
| `relationships_parity_test.go` | new | ~260 (proj.) | ok |
| `orchestrator_host.go` / `user.go` / `role.go` / `claim.go` | ≤380/45/25/30 (post-wave) | +0..15 | ok |

## Implementation spec
Core sketch (user parity = the binding constraint):
```go
func seedGORM(t *testing.T, f parityFixture) *gorm.DB {
	db, _ := database.NewDatabaseConnection(ctx, database.DefaultDatabaseConfig(ctx)) // :memory:
	require.NoError(t, db.AutoMigrate(allAutoMigrateModels()...)) // shared helper, finalized in task 06
	return db.Create(&f.user).Error == nil && db.Create(f.roles...).Error == nil // GORM writes user_roles/user_claims/role_claims rows
}
func TestUserRead_ParityWithLegacyRehydration(t *testing.T) {
	want := legacyFixtureUsers()                    // []models.User as legacy engine returns it
	var got []models.User
	require.NoError(t, db.Preload("Claims").Preload("Roles.Claims").Find(&got).Error)
	require.Equal(t, string(mustJSON(want)), string(mustJSON(got))) // BYTE-equal
}
func TestUserRoleClaims_LiveRehydrationParity(t *testing.T) {
	// after seed: db.Create(&models.RoleClaims{RoleID: "r1", ClaimID: newClaim.ID})
	// then re-read with Preload("Roles.Claims") ⇒ got[0].Roles[0].Claims MUST contain newClaim,
	// mirroring users.go:44-56 (role claims change ⇒ every user view changes without touching user rows).
}
```
- Host parity: seed host + `ReverseProxy{HostID}` + `[]ReverseProxyHost` + `VirtualMachine{HostId}`
  rows; read `Preload("ReverseProxy").Preload("ReverseProxyHosts")` + explicit VM query assigned to
  `host.VirtualMachines`; byte-compare vs legacy fixture. Assert default `Find(&hosts)` does **not**
  load VMs (N+1 guard per D-7).
- Snapshot parity: seed `VMSnapshot` rows with `HostID`/`VMID`; regroup query
  (`WHERE host_id=?`) into `map[string][]VMSnapshot`; byte-compare against legacy
  `HostsVMSnapshotsRecord.VMSnapshots`.
- Serialiser matrix: nil slice, empty slice, empty map, populated — round-trip through GORM Create→
  Find; assert both the Go value and the marshalled parent struct bytes match the fixture (guards
  `null` vs `[]` drift on `serializer:json`).
- Seeding rule: join rows are created **only via associations** (`db.Create(&user)` with
  `Roles`/`Claims` populated) or the explicit `RoleClaims` model — never hand-written SQL — so the
  tests double as proof the many2many/join tags actually emit the join tables task 06 asserts.
- Known failure modes pinned: (a) preload missing ⇒ `roles[].claims` empty ⇒ parity fail;
  (b) `omitempty` interplay — an unclaimed user must marshal `claims` absent, exactly as legacy;
  (c) `RoleClaims` composite PK means duplicate `(role,claim)` seeding errors — treat as fixture bug.

## Out of scope
Legacy `JsonDatabase` code changes (none — legacy engine untouched), AutoMigrate schema assertions
(06), any #421 query helpers.

## Validation
1. `cd src && go build ./... && go test ./... -count=1` green.
2. `go test ./data/models/... -run 'Parity' -count=1` green (all four suites).
3. Negative check: temporarily removing `Preload("Roles.Claims")` must FAIL the parity test (proves
   the test actually pins the read shape); revert the sabotage.
4. Wave goldens re-run untouched: `git diff testdata/golden/` empty.

## Rollback
Revert PR; pure test addition (+ any tag fixes, which revert atomically).

## PBI acceptance mapping
AC2 (joins exercised end-to-end via GORM-created join rows), AC4 (embed fidelity = the AC's core
claim for embedded collections), D-7 documented by executable test.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `419-05`.** Cut `slice/419-05` from `sql-migration` HEAD
(`.worktrees/slice-419-05`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
