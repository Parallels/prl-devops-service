package database

import (
	"context"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/security/password"
	log "github.com/cjlapao/common-go-logger"
)

// fakeCtx is the controllable basecontext.ApiContext (same sanctioned fake as
// database/query/authz_test.go; the fake cannot be shared across packages).
type fakeCtx struct {
	auth *basecontext.AuthorizationContext
}

func (f *fakeCtx) Context() context.Context { return context.Background() }
func (f *fakeCtx) GetAuthorizationContext() *basecontext.AuthorizationContext {
	return f.auth
}
func (f *fakeCtx) GetRequestId() string { return "test-request" }
func (f *fakeCtx) GetUser() *apimodels.ApiUser {
	if f.auth == nil {
		return nil
	}
	return f.auth.User
}
func (f *fakeCtx) Verbose() bool                             { return false }
func (f *fakeCtx) EnableLog()                                {}
func (f *fakeCtx) DisableLog()                               {}
func (f *fakeCtx) ToggleLogTimestamps(value bool)            {}
func (f *fakeCtx) Logger() *log.LoggerService                { return log.Get() }
func (f *fakeCtx) LogInfof(format string, a ...interface{})  {}
func (f *fakeCtx) LogErrorf(format string, a ...interface{}) {}
func (f *fakeCtx) LogDebugf(format string, a ...interface{}) {}
func (f *fakeCtx) LogWarnf(format string, a ...interface{})  {}
func (f *fakeCtx) LogTracef(format string, a ...interface{}) {}

// authCtx builds a context whose EffectiveClaims come from InjectedClaims —
// GetEffectiveClaims returns them verbatim (basecontext/authorization_context.go).
func authCtx(userID string, claims ...string) basecontext.ApiContext {
	return &fakeCtx{auth: &basecontext.AuthorizationContext{
		User:           &apimodels.ApiUser{ID: userID},
		InjectedClaims: claims,
	}}
}

func newKey(id, name, key, owner, typ string) models.ApiKey {
	return models.ApiKey{ID: id, Name: name, Key: key, Secret: fixtureSecret("api-key"), UserID: owner, Type: typ}
}

func seedKeys(t *testing.T, g *GormDatabase) basecontext.ApiContext {
	t.Helper()
	ctx := basecontext.NewBaseContext()
	seedDefaults(t, g, ctx)
	for _, u := range []models.User{
		{ID: "root", Username: "root", Email: "root@localhost", Name: "root", Password: "pw"},
		{ID: "u1", Username: "user1", Email: "u1@x.io", Name: "User One", Password: "pw"},
		{ID: "u2", Username: "user2", Email: "u2@x.io", Name: "User Two", Password: "pw"},
	} {
		if _, err := g.CreateUser(ctx, u); err != nil {
			t.Fatalf("seed user %s: %v", u.ID, err)
		}
	}
	for _, k := range []models.ApiKey{
		newKey("k1", "primary key", "key-one", "u1", ""),
		newKey("k2", "legacy key", "key-two", "", ""),
		newKey("k3", "service key", "key-three", "u1", "internal"),
		newKey("k4", "stale key", "key-four", "u1", ""),
	} {
		if _, err := g.CreateApiKey(ctx, k); err != nil {
			t.Fatalf("seed key %s: %v", k.ID, err)
		}
	}
	if err := g.RevokeKey(ctx, "k4"); err != nil {
		t.Fatalf("revoke seed: %v", err)
	}
	return ctx
}

func TestSQLApiKeys_Listing(t *testing.T) {
	g := newSQLDB(t)
	ctx := seedKeys(t, g)

	plain, err := g.GetApiKeys(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(plain) != 3 { // k3 internal removed, order preserved
		t.Fatalf("want 3 external keys, got %d", len(plain))
	}
	if plain[0].ID != "k1" || plain[1].ID != "k2" || plain[2].ID != "k4" {
		t.Errorf("order drift: %v %v %v", plain[0].ID, plain[1].ID, plain[2].ID)
	}
	for _, k := range plain {
		if k.Type != "external" {
			t.Errorf("type not normalized: %+v", k)
		}
	}

	all, err := g.GetAllApiKeysIncludingInternal(ctx)
	if err != nil || len(all) != 4 {
		t.Fatalf("admin listing: %v %d", err, len(all))
	}

	own := authCtx("u1", constants.LIST_OWN_API_KEY_CLAIM)
	ownKeys, err := g.GetApiKeys(own, "")
	if err != nil || len(ownKeys) != 2 {
		t.Fatalf("own u1: %v %d", err, len(ownKeys)) // k1,k4 — k2 ownerless, k3 internal
	}
	own2 := authCtx("u2", constants.LIST_OWN_API_KEY_CLAIM)
	own2Keys, _ := g.GetApiKeys(own2, "")
	if len(own2Keys) != 0 {
		t.Errorf("u2 owns no keys, got %d", len(own2Keys))
	}
	full := authCtx("u2", constants.LIST_API_KEY_CLAIM)
	fullKeys, _ := g.GetApiKeys(full, "")
	if len(fullKeys) != 3 {
		t.Errorf("full claim must see all external: %d", len(fullKeys))
	}
	// QUIRK: no listing claim at all ⇒ NO own-filter (upstream authz blocks this).
	bare := authCtx("u2")
	bareKeys, _ := g.GetApiKeys(bare, "")
	if len(bareKeys) != 3 {
		t.Errorf("claim-less caller must NOT be own-filtered (quirk): %d", len(bareKeys))
	}
}

func TestSQLApiKeys_GetAndMasking(t *testing.T) {
	g := newSQLDB(t)
	ctx := seedKeys(t, g)

	k, err := g.GetApiKey(ctx, "k1") // ID exact
	if err != nil || k.ID != "k1" {
		t.Fatalf("id probe: %v %+v", err, k)
	}
	if _, err := g.GetApiKey(ctx, "K1"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("ID match is EXACT; K1 must miss Name too? got %v", err) // name is "primary key"
	}
	if k, err := g.GetApiKey(ctx, "Primary KEY"); err != nil || k.ID != "k1" { // Name EqualFold
		t.Errorf("name fold: %v %+v", err, k)
	}
	if k, err := g.GetApiKey(ctx, "KEY-THREE"); err != nil || k.ID != "k3" { // KEY value; internal NOT filtered
		t.Errorf("key-value probe must reach internal key (quirk): %v %+v", err, k)
	}

	masked := authCtx("u2", constants.LIST_OWN_API_KEY_CLAIM)
	if _, err := g.GetApiKey(masked, "k1"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("own-claim caller must get NotFound mask, got %v", err)
	}
	if _, err := g.GetApiKey(masked, "k2"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("ownerless key is not own (quirk): %v", err)
	}
	if k, err := g.GetApiKey(authCtx("u2", constants.LIST_API_KEY_CLAIM), "k1"); err != nil || k == nil {
		t.Errorf("full claim unmasked: %v", err)
	}
}

func TestSQLApiKeys_UserByApiKey(t *testing.T) {
	g := newSQLDB(t)
	ctx := seedKeys(t, g)
	if _, err := g.CreateApiKey(ctx, newKey("k5", "ghost key", "key-five", "ghost-owner", "")); err != nil {
		t.Fatal(err)
	}

	if u, err := g.GetUserByApiKey(ctx, "key-one"); err != nil || u.ID != "u1" {
		t.Errorf("owner resolve: %v %+v", err, u)
	}
	if u, err := g.GetUserByApiKey(ctx, "key-two"); err != nil || u.ID != "root" {
		t.Errorf("no-owner must resolve root: %v %+v", err, u)
	}
	if u, err := g.GetUserByApiKey(ctx, "key-five"); err != nil || u.ID != "root" {
		t.Errorf("ghost owner must fall back to root: %v %+v", err, u)
	}
	if _, err := g.GetUserByApiKey(ctx, "key-four"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("revoked key must be indistinguishable: %v", err)
	}
	if _, err := g.GetUserByApiKey(ctx, "no-such-key"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("absent key: %v", err)
	}
	masked := authCtx("u2", constants.LIST_OWN_API_KEY_CLAIM)
	if _, err := g.GetUserByApiKey(masked, "key-one"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("mask must deny hot path too: %v", err)
	}
}

func TestSQLApiKeys_Mutations(t *testing.T) {
	g := newSQLDB(t)
	ctx := seedKeys(t, g)

	if _, err := g.CreateApiKey(ctx, newKey("k1", "fresh", "key-x", "u1", "")); err != interfaces.ErrApiKeyAlreadyExists {
		t.Errorf("dup id: want ErrApiKeyAlreadyExists, got %v", err)
	}
	if _, err := g.CreateApiKey(ctx, newKey("k9", "Primary Key", "key-x", "u1", "")); err != interfaces.ErrApiKeyAlreadyExists {
		t.Errorf("dup name (folded): got %v", err)
	}
	created, err := g.CreateApiKey(ctx, newKey("k6", "sixth", "key-six", "u1", "internal"))
	if err != nil {
		t.Fatal(err)
	}
	if created.Secret == fixtureSecret("api-key") || created.Secret == "" {
		t.Errorf("secret must be hashed: %q", created.Secret)
	}
	if err := password.Get().Compare(fixtureSecret("api-key"), "k6", created.Secret); err != nil {
		t.Errorf("hash must verify with (secret, keyID) salt: %v", err)
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Errorf("timestamps missing: %+v", created)
	}

	// UpdateKey: EXACT id, writes revoke fields only
	if err := g.UpdateKey(ctx, models.ApiKey{ID: "K2", Revoked: true}); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("UpdateKey id match must be exact: %v", err)
	}
	if err := g.UpdateKey(ctx, models.ApiKey{ID: "k2", Revoked: true, RevokedAt: "2026-02-02T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	k2, _ := g.GetApiKey(ctx, "k2")
	if !k2.Revoked || k2.RevokedAt != "2026-02-02T00:00:00Z" {
		t.Errorf("revoke fields not written: %+v", k2)
	}

	// RevokeKey flows incl. silent no-op and masking
	if err := g.RevokeKey(ctx, ""); err != nil {
		t.Errorf("empty id must be silent no-op: %v", err)
	}
	if err := g.RevokeKey(ctx, "primary key"); err != nil { // EqualFold name
		t.Errorf("revoke by folded name: %v", err)
	}
	k1, _ := g.GetApiKey(ctx, "k1")
	if !k1.Revoked || k1.RevokedAt == "" {
		t.Errorf("revoked/revoked_at not set: %+v", k1)
	}
	masked := authCtx("u2", constants.LIST_OWN_API_KEY_CLAIM)
	if err := g.RevokeKey(masked, "k2"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("masked revoke must NotFound: %v", err)
	}
	if err := g.RevokeKey(ctx, "ghost"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("absent revoke: %v", err)
	}

	// DeleteApiKey
	if err := g.DeleteApiKey(ctx, ""); err != nil {
		t.Errorf("silent no-op broken: %v", err)
	}
	if err := g.DeleteApiKey(ctx, "KEY-SIX"); err != nil { // KEY value EqualFold
		t.Errorf("delete by key value: %v", err)
	}
	if _, err := g.GetApiKey(ctx, "k6"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("k6 survived delete: %v", err)
	}
	if err := g.DeleteApiKey(ctx, "ghost"); err != interfaces.ErrApiKeyNotFound {
		t.Errorf("absent delete: %v", err)
	}

	cold := NewGormDatabase(ctx, nil)
	if _, err := cold.GetApiKeys(ctx, ""); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("guard: %v", err)
	}
	if _, err := cold.CreateApiKey(ctx, newKey("z", "z", "z", "", "")); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("create guard: %v", err)
	}
}
