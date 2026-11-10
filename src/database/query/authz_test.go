package query

import (
	"context"
	"reflect"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/models"
	log "github.com/cjlapao/common-go-logger"
)

// fakeContext is the recording/controllable basecontext.ApiContext the spec
// sanctions (~60 lines); it also serves the later domain parity tasks.
type fakeContext struct {
	auth *basecontext.AuthorizationContext
}

func (f *fakeContext) Context() context.Context { return context.Background() }
func (f *fakeContext) GetAuthorizationContext() *basecontext.AuthorizationContext {
	return f.auth
}
func (f *fakeContext) GetRequestId() string { return "test-request" }
func (f *fakeContext) GetUser() *models.ApiUser {
	if f.auth == nil {
		return nil
	}
	return f.auth.User
}
func (f *fakeContext) Verbose() bool                             { return false }
func (f *fakeContext) EnableLog()                                {}
func (f *fakeContext) DisableLog()                               {}
func (f *fakeContext) ToggleLogTimestamps(value bool)            {}
func (f *fakeContext) Logger() *log.LoggerService                { return log.Get() }
func (f *fakeContext) LogInfof(format string, a ...interface{})  {}
func (f *fakeContext) LogErrorf(format string, a ...interface{}) {}
func (f *fakeContext) LogDebugf(format string, a ...interface{}) {}
func (f *fakeContext) LogWarnf(format string, a ...interface{})  {}
func (f *fakeContext) LogTracef(format string, a ...interface{}) {}

// rec is a minimal AuthorizedRecord; recOwner adds GetUserID for OwnRecords.
type rec struct {
	roles  []string
	claims []string
}

func (r rec) GetRequiredRoles() []string  { return r.roles }
func (r rec) GetRequiredClaims() []string { return r.claims }

type recOwner struct {
	rec
	userID string
}

func (r recOwner) GetUserID() string { return r.userID }

// TestAuthzContextMatrix is the spec's 10-row matrix: every row asserts the
// golden expectation AND query-vs-data equality (both engines side-by-side).
func TestAuthzContextMatrix(t *testing.T) {
	guarded := rec{roles: []string{"admin"}, claims: []string{"read"}}
	public := rec{}

	cases := []struct {
		name   string
		auth   *basecontext.AuthorizationContext
		record rec
		want   bool
	}{
		{"1 nil auth ctx", nil, guarded, false},
		{"2 root bypass", &basecontext.AuthorizationContext{AuthorizedBy: "RootAuthorization"}, guarded, true},
		{"3 api-key microservice bypass", &basecontext.AuthorizationContext{AuthorizedBy: "ApiKeyAuthorization", IsMicroService: true}, guarded, true},
		{"4 injected suppresses bypass", &basecontext.AuthorizationContext{AuthorizedBy: "RootAuthorization", InjectedClaims: []string{"x"}}, guarded, false},
		{"5 super user flag", &basecontext.AuthorizationContext{IsAuthorized: true, IsSuperUser: true, User: &models.ApiUser{ID: "u1"}}, guarded, true},
		{"6 super user role", &basecontext.AuthorizationContext{IsAuthorized: true, User: &models.ApiUser{ID: "u1", Roles: []string{constants.SUPER_USER_ROLE}}}, guarded, true},
		{"7 not authorized", &basecontext.AuthorizationContext{IsAuthorized: false, User: &models.ApiUser{ID: "u1"}}, guarded, false},
		{"8 role but missing claim (AND-of-groups)", &basecontext.AuthorizationContext{IsAuthorized: true, User: &models.ApiUser{ID: "u1", Roles: []string{"admin"}}}, guarded, false},
		{"9 public record", &basecontext.AuthorizationContext{IsAuthorized: true, User: &models.ApiUser{ID: "u1"}}, public, true},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			ctx := &fakeContext{auth: c.auth}
			got := IsAuthorized(ctx, c.record)
			legacy := data.IsAuthorized(ctx, c.record)
			if got != c.want {
				t.Errorf("query.IsAuthorized: want %v, got %v", c.want, got)
			}
			if legacy != c.want {
				t.Errorf("data.IsAuthorized (parity anchor): want %v, got %v", c.want, legacy)
			}
		})
	}

	// Row 10 is a static-verification row: the implementation must reference
	// constants.SUPER_USER_ROLE, never the literal — enforced by the slice
	// grep gate; behaviourally covered by row 6.
}

func TestIsRootUser_Parity(t *testing.T) {
	cases := []struct {
		name string
		auth *basecontext.AuthorizationContext
		want bool
	}{
		{"nil", nil, false},
		{"root auth", &basecontext.AuthorizationContext{AuthorizedBy: "RootAuthorization"}, true},
		{"root username", &basecontext.AuthorizationContext{User: &models.ApiUser{Username: "root"}}, true},
		{"normal user", &basecontext.AuthorizationContext{User: &models.ApiUser{Username: "bob"}}, false},
	}
	for _, c := range cases {
		ctx := &fakeContext{auth: c.auth}
		if got := IsRootUser(ctx); got != c.want {
			t.Errorf("query.IsRootUser %s: want %v, got %v", c.name, c.want, got)
		}
		if got := data.IsRootUser(ctx); got != c.want {
			t.Errorf("data.IsRootUser %s: want %v, got %v", c.name, c.want, got)
		}
	}
}

func TestGetAuthorizedRecords_Parity(t *testing.T) {
	auth := &basecontext.AuthorizationContext{IsAuthorized: true, User: &models.ApiUser{ID: "u1", Roles: []string{"admin"}}}
	ctx := &fakeContext{auth: auth}
	input := []rec{{}, {roles: []string{"admin"}, claims: []string{"read"}}} // public, then claim-gated

	got := GetAuthorizedRecords(ctx, input...)
	want := data.GetAuthorizedRecords(ctx, input...)
	if len(got) != 1 || !reflect.DeepEqual(got, want) {
		t.Errorf("AND-of-groups filtering: want only public record, got %v vs %v", got, want)
	}
}

// TestOwnRecords_Pins: passthrough returns the SAME slice (not a copy,
// api_key.go:21-24) and ownership match is case-SENSITIVE == (:28).
func TestOwnRecords_Pins(t *testing.T) {
	owned := []recOwner{{userID: "abc"}, {userID: "ABC"}, {userID: "xyz"}}

	// passthrough — no auth ctx
	same := OwnRecords(&fakeContext{}, owned...)
	if len(same) != 3 || &same[0] != &owned[0] {
		t.Errorf("nil-auth passthrough must return the same slice, got %p vs %p", &same[0], &owned[0])
	}
	// passthrough — auth ctx without user
	passthru := OwnRecords(&fakeContext{auth: &basecontext.AuthorizationContext{}}, owned...)
	if &passthru[0] != &owned[0] {
		t.Errorf("nil-user passthrough must return the same slice")
	}

	// case-sensitive equality
	ctx := &fakeContext{auth: &basecontext.AuthorizationContext{User: &models.ApiUser{ID: "abc"}}}
	got := OwnRecords(ctx, owned...)
	if len(got) != 1 || got[0].userID != "abc" {
		t.Errorf("case-sensitive ownership: want [abc], got %v", got)
	}
	want := data.GetOwnRecords(ctx, owned...)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("OwnRecords drifted from legacy GetOwnRecords: %v vs %v", got, want)
	}
}
