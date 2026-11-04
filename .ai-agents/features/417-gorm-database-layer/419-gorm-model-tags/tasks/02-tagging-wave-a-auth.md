# Task 02 — Tagging wave A: auth domain (user, role, claim, role_claims, api_key)

## Context
Plan mapping rows 1–5. Auth is the highest-risk domain because `users.go:44-56` re-hydrates each
user's embedded `Role.Claims` from the live `Roles` table at read time — the tag design must make
that read shape expressible (`Preload("Roles.Claims")`). Also lands the **golden-JSON harness**
reused by waves B/C.

## Objective
Auth structs carry GORM tags per the mapping table; M2M joins declared; golden-JSON byte tests
prove `json.Marshal` output is unchanged for every tagged struct; build+tests green.

## Files
- Modify: `src/data/models/user.go` `role.go` `claim.go` `role_claims.go` `api_key.go`
- Create: `src/data/models/models_gorm_test.go` (external `package models_test` — free to import
  `src/database` later without cycle risk), `src/data/models/testdata/golden/` fixtures
  (`user.json`, `user_minimal.json`, `role.json`, `claim.json`, `api_key.json`, `role_claims.json`)

## File size guard (ER-4)
| File | Now | After (proj.) | Verdict |
|---|---|---|---|
| `user.go` | 17 | ~45 | ok |
| `role.go` | 10 | ~25 | ok |
| `claim.go` | 12 | ~30 | ok |
| `role_claims.go` | 10 | ~25 | ok |
| `api_key.go` | 20 | ~50 | ok |
| `models_gorm_test.go` | new | ~160 (proj.) | ok |

## Implementation spec
Exemplar (exact tags; `user.go`):
```go
type User struct {
	ID                  string  `json:"id,omitempty" gorm:"primaryKey"`
	Username            string  `json:"username" gorm:"uniqueIndex"`
	Name                string  `json:"name"`
	Email               string  `json:"email" gorm:"index"` // NOT unique — empty-string legacy rows (plan risk list)
	Password            string  `json:"password,omitempty"`
	CreatedAt           string  `json:"created_at,omitempty"` // D-1: string dates
	UpdatedAt           string  `json:"updated_at,omitempty"`
	Roles               []Role  `json:"roles,omitempty" gorm:"many2many:user_roles;"`
	Claims              []Claim `json:"claims,omitempty" gorm:"many2many:user_claims;"`
	FailedLoginAttempts int     `json:"failed_login_attempts,omitempty"`
	Blocked             bool    `json:"blocked,omitempty"`
	BlockedSince        string  `json:"blocked_since,omitempty"`
	BlockedReason       string  `json:"blocked_reason,omitempty"`
}
```
- `role.go`: `ID` `gorm:"primaryKey"`; `Claims []Claim` →
  `gorm:"many2many:role_claims;joinForeignKey:RoleID;joinReferences:ClaimID"`; `Users []User` →
  `gorm:"many2many:user_roles;foreignkey:ID;associationForeignKey:ID"` (inverse; stays `json:"-"`).
- `claim.go`: `ID` PK; `Users []User` inverse M2M via `user_claims` / `role_claims` (same pattern).
- `role_claims.go`: join model — `RoleID string gorm:"column:role_id;primaryKey"` +
  `ClaimID string gorm:"column:claim_id;primaryKey"` (composite PK = the uniqueness #424 needs).
- `api_key.go`: `ID` PK; `Key` `gorm:"uniqueIndex"`; `UserID` `gorm:"index"`; embedded
  `*DbRecord` gains `gorm:"-:all"` (D-4 — JSON `db_record` key must still marshal).
- **Golden-JSON skeleton** (capture-first protocol — golden files are generated and committed in
  the FIRST commit of this PR, *before* the tag commit; the tag commit must not touch `testdata/`):
```go
//go:generate go run ./testdata/capture   // one-off, pre-tag commit only
func TestUser_JSONGolden(t *testing.T) {
	u := fixtureUser()                       // shared fixtures helper, full + minimal variants
	got, err := json.Marshal(u)
	require.NoError(t, err)
	want, err := os.ReadFile("testdata/golden/user.json")
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(string(want)), string(got)) // BYTE-equal, not JSONEq (key order + typos matter)
}
```
- Test files get golden tests for all 5 root structs + a `TestRoleClaims_RoundTrip` (marshal→unmarshal→marshal stable).

## Out of scope
No AutoMigrate (task 06), no query code (#421), no changes to `users.go` data-access, no
`gorm.DeletedAt` anywhere (D-5 evidence).

## Validation
1. `cd src && go build ./... && go test ./... -count=1` green.
2. `go test ./data/models/... -run JSONGolden -count=1` → all golden tests pass post-tag.
3. `git diff --stat testdata/` **empty in the tag commit** (proof tags didn't move bytes).
4. `grep -c 'gorm:"' src/data/models/{user,role,claim,role_claims,api_key}.go` ≥ one per persisted field.

## Rollback
Revert PR; golden fixtures pre-date tags so they can stay (harmless) or revert atomically.

## PBI acceptance mapping
AC1 (auth rows), AC2 (`user_roles`/`user_claims`/`role_claims` declared), AC4 (byte-level wire
compat for auth domain), ER-2 verified via guard table.
