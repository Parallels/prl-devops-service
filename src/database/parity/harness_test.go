package parity

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func seedCorpus(t testing.TB) []models.User {
	t.Helper()
	raw, err := os.ReadFile("testdata/seed_users.json")
	if err != nil {
		t.Fatalf("read seed corpus: %v", err)
	}
	var users []models.User
	if err := json.Unmarshal(raw, &users); err != nil {
		t.Fatalf("parse seed corpus: %v", err)
	}
	return users
}

// recordingTB intercepts failures so the self-proof can assert that
// AssertSameList actually CATCHES drift, not just tolerates equality.
type recordingTB struct {
	testing.TB
	failed  bool
	lastMsg string
}

func (r *recordingTB) Helper()           {}
func (r *recordingTB) Cleanup(func())    {}
func (r *recordingTB) Error(args ...any) { r.failed = true }
func (r *recordingTB) Errorf(format string, args ...any) {
	r.failed = true
	r.lastMsg = fmt.Sprintf(format, args...)
}
func (r *recordingTB) Fail()             { r.failed = true }
func (r *recordingTB) FailNow()          { r.failed = true }
func (r *recordingTB) Fatal(args ...any) { r.failed = true }
func (r *recordingTB) Fatalf(format string, args ...any) {
	r.failed = true
	r.lastMsg = fmt.Sprintf(format, args...)
}
func (r *recordingTB) Name() string                     { return "recordingTB" }
func (r *recordingTB) Skip(args ...any)                 {}
func (r *recordingTB) Skipf(format string, args ...any) {}
func (r *recordingTB) Skipped() bool                    { return false }
func (r *recordingTB) TempDir() string                  { return "" }
func (r *recordingTB) Setenv(key, value string)         {}

// TestHarness_SelfProof proves the harness plumbing end-to-end: a DomainSuite
// seeds three users through the exported DatabaseService surface and a Check
// compares both engines' GetUsers output. Task 04 has not landed yet, so the
// SQL side is the legacy fallback (identity pair) — the mutation case then
// proves AssertSameList actually detects drift.
func TestHarness_SelfProof(t *testing.T) {
	corpus := seedCorpus(t)

	RunMatrix(t, DomainSuite{
		Name: "users-seed",
		Seed: func(t testing.TB, db datainterfaces.DatabaseService) {
			ctx := basecontext.NewBaseContext()
			// CreateUser backfills DefaultRoles/DefaultClaims — they must
			// exist first; seed them through the exported surface too.
			for _, roleName := range constants.DefaultRoles {
				if _, err := db.CreateRole(ctx, models.Role{Name: roleName}); err != nil {
					t.Fatalf("seed role %s: %v", roleName, err)
				}
			}
			for _, claimName := range constants.DefaultClaims {
				if _, err := db.CreateClaim(ctx, models.Claim{Name: claimName}); err != nil {
					t.Fatalf("seed claim %s: %v", claimName, err)
				}
			}
			for _, u := range corpus {
				if _, err := db.CreateUser(ctx, u); err != nil {
					t.Fatalf("seed user %s: %v", u.ID, err)
				}
			}
		},
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			func(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
				ctx := basecontext.NewBaseContext()
				lu, err := legacy.GetUsers(ctx, "")
				if err != nil {
					t.Fatalf("legacy GetUsers: %v", err)
				}
				su, err := sql.GetUsers(ctx, "")
				if err != nil {
					t.Fatalf("sql GetUsers: %v", err)
				}
				if len(lu) != len(corpus) {
					t.Fatalf("seed did not land: want %d users, got %d", len(corpus), len(lu))
				}
				AssertSameList(t, "GetUsers(all)", lu, su)
			},
		},
	})

	// Mutation case: the assertion machinery must catch drift. Same inputs
	// minus one record must trip recordingTB with a readable diff.
	rec := &recordingTB{TB: t}
	AssertSameList(rec, "mutation", corpus, corpus[:2])
	if !rec.failed {
		t.Fatal("AssertSameList failed to detect drift — harness self-proof broken")
	}
	rec2 := &recordingTB{TB: t}
	AssertSameList(rec2, "identity", corpus, corpus)
	if rec2.failed {
		t.Fatal("AssertSameList tripped on identical lists")
	}
}
