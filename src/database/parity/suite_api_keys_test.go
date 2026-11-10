package parity

// DomainSuite "api_keys" (PBI #421 task 07): differential over the auth hot
// path — internal/external split, the own-list claim masking (listings,
// single fetches, create-probes, revoke lookups), GetUserByApiKey's
// revoked/ownerless/root fallbacks, hash-verification of secrets.

import (
	"context"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/security/password"
	log "github.com/cjlapao/common-go-logger"
)

// parityCtx — the sanctioned controllable ApiContext (see query/authz_test.go).
type parityCtx struct {
	auth *basecontext.AuthorizationContext
}

func (f *parityCtx) Context() context.Context { return context.Background() }
func (f *parityCtx) GetAuthorizationContext() *basecontext.AuthorizationContext {
	return f.auth
}
func (f *parityCtx) GetRequestId() string { return "test-request" }
func (f *parityCtx) GetUser() *apimodels.ApiUser {
	if f.auth == nil {
		return nil
	}
	return f.auth.User
}
func (f *parityCtx) Verbose() bool                             { return false }
func (f *parityCtx) EnableLog()                                {}
func (f *parityCtx) DisableLog()                               {}
func (f *parityCtx) ToggleLogTimestamps(value bool)            {}
func (f *parityCtx) Logger() *log.LoggerService                { return log.Get() }
func (f *parityCtx) LogInfof(format string, a ...interface{})  {}
func (f *parityCtx) LogErrorf(format string, a ...interface{}) {}
func (f *parityCtx) LogDebugf(format string, a ...interface{}) {}
func (f *parityCtx) LogWarnf(format string, a ...interface{})  {}
func (f *parityCtx) LogTracef(format string, a ...interface{}) {}

func keyAuthCtx(userID string, claims ...string) basecontext.ApiContext {
	return &parityCtx{auth: &basecontext.AuthorizationContext{
		User:           &apimodels.ApiUser{ID: userID},
		InjectedClaims: claims,
	}}
}

func normaliseKeys(in []models.ApiKey) []models.ApiKey {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.ApiKey(nil), in...)
	for i := range out {
		if out[i].Secret != "" {
			out[i].Secret = "***"
		}
		for _, ts := range []*string{&out[i].CreatedAt, &out[i].UpdatedAt, &out[i].RevokedAt} {
			if *ts != "" {
				*ts = "SET"
			}
		}
	}
	return out
}

func seedApiKeys(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, r := range constants.DefaultRoles {
		if _, err := db.CreateRole(ctx, models.Role{Name: r}); err != nil {
			t.Fatalf("seed default role: %v", err)
		}
	}
	for _, c := range constants.DefaultClaims {
		if _, err := db.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed default claim: %v", err)
		}
	}
	for _, u := range []models.User{
		{ID: "root", Username: "root", Email: "root@localhost", Name: "root", Password: "Seed-Pass-1!"},
		{ID: "u1", Username: "user1", Email: "u1@x.io", Name: "User One", Password: "Seed-Pass-1!"},
		{ID: "u2", Username: "user2", Email: "u2@x.io", Name: "User Two", Password: "Seed-Pass-1!"},
	} {
		if _, err := db.CreateUser(ctx, u); err != nil {
			t.Fatalf("seed user %s: %v", u.ID, err)
		}
	}
	k := func(id, name, key, owner, typ string) {
		if _, err := db.CreateApiKey(ctx, models.ApiKey{ID: id, Name: name, Key: key,
			Secret: "Seed-Secret-1!", UserID: owner, Type: typ}); err != nil {
			t.Fatalf("seed key %s: %v", id, err)
		}
	}
	k("k1", "primary key", "key-one", "u1", "")
	k("k2", "legacy key", "key-two", "", "")
	k("k3", "service key", "key-three", "u1", "internal")
	k("k4", "stale key", "key-four", "u1", "")
	k("k5", "ghost key", "key-five", "ghost-owner", "")
	if err := db.RevokeKey(ctx, "k4"); err != nil {
		t.Fatalf("seed revoke: %v", err)
	}
}

func TestParityDomain_ApiKeys(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "api_keys",
		Seed: seedApiKeys,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkKeyListings,
			checkKeyAuthzMatrix,
			checkKeyLookups,
			checkKeyUserResolution,
			checkKeyCreate,
			checkKeyRevoke,
			checkKeyUpdateDelete,
		},
	})
}

func checkKeyListings(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, f := range []string{"", "user_id=u1", "revoked=true"} {
		lk, lerr := legacy.GetApiKeys(ctx, f)
		sk, serr := sql.GetApiKeys(ctx, f)
		if lerr != nil || serr != nil {
			t.Fatalf("filter %q: %v / %v", f, lerr, serr)
		}
		AssertSameList(t, "GetApiKeys("+f+")", normaliseKeys(lk), normaliseKeys(sk))
	}
	la, lerr := legacy.GetAllApiKeysIncludingInternal(ctx)
	sa, serr := sql.GetAllApiKeysIncludingInternal(ctx)
	if lerr != nil || serr != nil {
		t.Fatalf("admin listing: %v / %v", lerr, serr)
	}
	AssertSameList(t, "GetAllApiKeysIncludingInternal", normaliseKeys(la), normaliseKeys(sa))
}

func checkKeyAuthzMatrix(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	cases := map[string]basecontext.ApiContext{
		"own-u1":  keyAuthCtx("u1", constants.LIST_OWN_API_KEY_CLAIM),
		"own-u2":  keyAuthCtx("u2", constants.LIST_OWN_API_KEY_CLAIM),
		"full-u2": keyAuthCtx("u2", constants.LIST_API_KEY_CLAIM),
		"bare-u2": keyAuthCtx("u2"), // quirk: claim-less caller is NOT own-filtered
	}
	for name, ctx := range cases {
		lk, lerr := legacy.GetApiKeys(ctx, "")
		sk, serr := sql.GetApiKeys(ctx, "")
		if lerr != nil || serr != nil {
			t.Fatalf("%s: %v / %v", name, lerr, serr)
		}
		AssertSameList(t, "GetApiKeys["+name+"]", normaliseKeys(lk), normaliseKeys(sk))
	}
}

func checkKeyLookups(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, probe := range []string{"k1", "K1", "PRIMARY KEY", "KEY-THREE", "ghost"} {
		lk, lerr := legacy.GetApiKey(ctx, probe)
		sk, serr := sql.GetApiKey(ctx, probe)
		assertSameError(t, "GetApiKey("+probe+")", lerr, serr)
		if lerr == nil && (lk == nil || sk == nil || lk.ID != sk.ID) {
			t.Errorf("GetApiKey(%s): %+v vs %+v", probe, lk, sk)
		}
	}
	masked := keyAuthCtx("u2", constants.LIST_OWN_API_KEY_CLAIM)
	for _, probe := range []string{"k1", "k2"} { // other's key AND ownerless key
		lk, lerr := legacy.GetApiKey(masked, probe)
		sk, serr := sql.GetApiKey(masked, probe)
		assertSameError(t, "masked GetApiKey("+probe+")", lerr, serr)
		if lk != nil || sk != nil {
			t.Errorf("masked probe leaked: %+v %+v", lk, sk)
		}
	}
}

func checkKeyUserResolution(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, tc := range []struct {
		key  string
		want string
	}{
		{"key-one", "u1"},
		{"key-two", "root"},  // no owner ⇒ root
		{"key-five", "root"}, // ghost owner ⇒ root
	} {
		lu, lerr := legacy.GetUserByApiKey(ctx, tc.key)
		su, serr := sql.GetUserByApiKey(ctx, tc.key)
		if lerr != nil || serr != nil {
			t.Fatalf("GetUserByApiKey(%s): %v / %v", tc.key, lerr, serr)
		}
		if lu.ID != tc.want || su.ID != tc.want {
			t.Errorf("GetUserByApiKey(%s): %s vs %s (want %s)", tc.key, lu.ID, su.ID, tc.want)
		}
	}
	for _, bad := range []string{"key-four", "no-such-key"} { // revoked, absent
		_, lerr := legacy.GetUserByApiKey(ctx, bad)
		_, serr := sql.GetUserByApiKey(ctx, bad)
		assertSameError(t, "GetUserByApiKey("+bad+")", lerr, serr)
	}
	masked := keyAuthCtx("u2", constants.LIST_OWN_API_KEY_CLAIM)
	_, lerr := legacy.GetUserByApiKey(masked, "key-one")
	_, serr := sql.GetUserByApiKey(masked, "key-one")
	assertSameError(t, "GetUserByApiKey masked", lerr, serr)
}

func checkKeyCreate(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	_, d1 := legacy.CreateApiKey(ctx, models.ApiKey{ID: "k1", Name: "fresh", Key: "key-x", Secret: "s"})
	_, d2 := sql.CreateApiKey(ctx, models.ApiKey{ID: "k1", Name: "fresh", Key: "key-x", Secret: "s"})
	assertSameError(t, "create dup id", d1, d2)
	_, d3 := legacy.CreateApiKey(ctx, models.ApiKey{ID: "k9", Name: "Legacy KEY", Key: "key-x", Secret: "s"})
	_, d4 := sql.CreateApiKey(ctx, models.ApiKey{ID: "k9", Name: "Legacy KEY", Key: "key-x", Secret: "s"})
	assertSameError(t, "create dup name", d3, d4)

	lc, lerr := legacy.CreateApiKey(ctx, models.ApiKey{ID: "k6", Name: "sixth", Key: "key-six", Secret: "Rotated-Secret-1!", UserID: "u2", Type: "internal"})
	sc, serr := sql.CreateApiKey(ctx, models.ApiKey{ID: "k6", Name: "sixth", Key: "key-six", Secret: "Rotated-Secret-1!", UserID: "u2", Type: "internal"})
	if lerr != nil || serr != nil {
		t.Fatalf("create k6: %v / %v", lerr, serr)
	}
	if lc.Secret == "Rotated-Secret-1!" || sc.Secret == "Rotated-Secret-1!" {
		t.Error("secret stored unhashed")
	}
	if err := password.Get().Compare("Rotated-Secret-1!", "k6", lc.Secret); err != nil {
		t.Errorf("legacy secret hash fails verification: %v", err)
	}
	if err := password.Get().Compare("Rotated-Secret-1!", "k6", sc.Secret); err != nil {
		t.Errorf("sql secret hash fails verification: %v", err)
	}
	lk, _ := legacy.GetApiKeys(ctx, "")
	sk, _ := sql.GetApiKeys(ctx, "")
	AssertSameList(t, "GetApiKeys after create", normaliseKeys(lk), normaliseKeys(sk))
	la, _ := legacy.GetAllApiKeysIncludingInternal(ctx)
	sa, _ := sql.GetAllApiKeysIncludingInternal(ctx)
	AssertSameList(t, "admin listing after create", normaliseKeys(la), normaliseKeys(sa))
}

func checkKeyRevoke(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	masked := keyAuthCtx("u2", constants.LIST_OWN_API_KEY_CLAIM)
	assertSameError(t, "masked revoke",
		errOf(legacy.RevokeKey(masked, "k1")),
		errOf(sql.RevokeKey(masked, "k1")))
	assertSameError(t, "revoke empty (silent)",
		errOf(legacy.RevokeKey(ctx, "")),
		errOf(sql.RevokeKey(ctx, "")))
	if err := legacy.RevokeKey(ctx, "primary key"); err != nil {
		t.Fatal(err)
	}
	if err := sql.RevokeKey(ctx, "primary key"); err != nil {
		t.Fatal(err)
	}
	assertSameError(t, "revoke absent",
		errOf(legacy.RevokeKey(ctx, "ghost")),
		errOf(sql.RevokeKey(ctx, "ghost")))
	lk, _ := legacy.GetApiKeys(ctx, "")
	sk, _ := sql.GetApiKeys(ctx, "")
	AssertSameList(t, "GetApiKeys after revoke", normaliseKeys(lk), normaliseKeys(sk))
	if _, err := legacy.GetUserByApiKey(ctx, "key-one"); err == nil {
		t.Error("legacy revoked key still resolves")
	}
	if _, err := sql.GetUserByApiKey(ctx, "key-one"); err == nil {
		t.Error("sql revoked key still resolves")
	}
}

func checkKeyUpdateDelete(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	assertSameError(t, "update exact-id",
		errOf(legacy.UpdateKey(ctx, models.ApiKey{ID: "K1", Revoked: true})),
		errOf(sql.UpdateKey(ctx, models.ApiKey{ID: "K1", Revoked: true})))
	if err := legacy.UpdateKey(ctx, models.ApiKey{ID: "k5", Revoked: false, RevokedAt: ""}); err != nil {
		t.Fatal(err)
	}
	if err := sql.UpdateKey(ctx, models.ApiKey{ID: "k5", Revoked: false, RevokedAt: ""}); err != nil {
		t.Fatal(err)
	}
	assertSameError(t, "update absent",
		errOf(legacy.UpdateKey(ctx, models.ApiKey{ID: "ghost", Revoked: true})),
		errOf(sql.UpdateKey(ctx, models.ApiKey{ID: "ghost", Revoked: true})))

	assertSameError(t, "delete empty (silent)",
		errOf(legacy.DeleteApiKey(ctx, "")),
		errOf(sql.DeleteApiKey(ctx, "")))
	if err := legacy.DeleteApiKey(ctx, "KEY-SIX"); err != nil {
		t.Fatal(err)
	}
	if err := sql.DeleteApiKey(ctx, "KEY-SIX"); err != nil {
		t.Fatal(err)
	}
	assertSameError(t, "delete absent",
		errOf(legacy.DeleteApiKey(ctx, "ghost")),
		errOf(sql.DeleteApiKey(ctx, "ghost")))
	la, _ := legacy.GetAllApiKeysIncludingInternal(ctx)
	sa, _ := sql.GetAllApiKeysIncludingInternal(ctx)
	AssertSameList(t, "admin listing after delete", normaliseKeys(la), normaliseKeys(sa))
}
