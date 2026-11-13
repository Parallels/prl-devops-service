package parity

// #425 task 02 — the R-3 net: a cross-engine DIFFERENTIAL over the
// Mongo-flavoured query layer (filter/order/authorize layers, ~370 lines
// with zero legacy tests pre-#425). One shared seed; both engines run the
// identical call list; EXACT sequences (not just sets) must agree, and
// errors must agree. Deviations from the task wording, justified:
//   - The "order runs" over CLOCK properties (+name, -created style) are
//     reduced to insertion-order surfaces: per-engine UpdatedAt values
//     differ by construction and RFC3339Nano string order is the T14
//     coin-flip — cross-engine order parity over wall-clock stamps is
//     excluded per doc.go E-3 (order semantics stay covered by the
//     deterministic order-layer unit tests, shared tiebreak contract).
//     Exact-sequence assertions here therefore pin the filter+mask+authz
//     ordering (input order preserved), which IS deterministic.
//   - The authorize differential runs on the catalog surface: within this
//     wave's domains users/roles/claims/keys listings are not authz-wrapped
//     (verified data/users.go GetUsers, data/api_key.go GetApiKeys), so the
//     authorize_layer semantics only exercise through the wrapped catalog
//     listings (GetCatalogManifests → GetAuthorizedRecords both engines).

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParityFilterOrderAuthDifferential(t *testing.T) {
	legacyEng := authLegacyEngine(nil)
	gormEng := authGormEngine(nil)
	legacySvc, legacyCloser := legacyEng.New(t)
	gormSvc, gormCloser := gormEng.New(t)
	t.Cleanup(legacyCloser) // MUST disconnect: engines left connected leak
	t.Cleanup(gormCloser)   // daemons that race the NEXT test's globals

	for _, db := range []dbT{legacySvc, gormSvc} {
		ctx := tplCtxRoot()
		seedDefaultRolesClaims(ctx, db) // CreateUser resolves default roles (fidelity of every legacy body)
		_, err := db.CreateUser(ctx, models.User{ID: "seed-u1", Username: "user1", Name: "User One", Email: "u1@example.com", Password: fixtureSecret("seed-user")})
		require.NoError(t, err)
		_, err = db.CreateUser(ctx, models.User{ID: "seed-u3", Username: "User3", Name: "user three", Email: "u3@example.com", Password: fixtureSecret("seed-user")})
		require.NoError(t, err)
		_, err = db.CreateUser(ctx, models.User{ID: "seed-u2", Username: "user2", Name: "User Two", Email: "u2@other.org", Password: fixtureSecret("seed-user")})
		require.NoError(t, err)
		for _, k := range []models.ApiKey{
			{ID: "k-a", Name: "Alpha Key", Key: "KA", Secret: "s", Type: "external"},
			{ID: "k-c", Name: "Gamma Key", Key: "KC", Secret: "s", Type: "internal"},
			{ID: "k-b", Name: "beta key", Key: "KB", Secret: "s", Type: "external"},
		} {
			_, err := db.CreateApiKey(ctx, k)
			require.NoError(t, err)
		}
		for _, m := range []models.CatalogManifest{
			{CatalogId: "diff", Version: "1.0", Architecture: "amd64", RequiredRoles: []string{"operator"}},
			{CatalogId: "diff", Version: "2.0", Architecture: "amd64"},
		} {
			_, err := db.CreateCatalogManifest(ctx, m)
			require.NoError(t, err)
		}
	}

	corpus := []struct{ surface, name, filter string }{
		{"users", "all_empty_filter", ""},
		{"users", "eq", "username=user1"},
		{"users", "regex_partial", "username=user"},
		{"users", "anchored", "username=^user1$"},
		{"users", "email_substring", "email=@example.com"},
		{"users", "dot_meta_escaped", "email=u1\\.example\\.com$"},
		{"users", "ci_option", "username=USER1,i"},
		{"users", "unknown_property", "nosuchfield=x"},
		{"users", "property_case_matters", "USERNAME=user1"},
		{"users", "invalid_regex", "username=["},
		{"users", "no_equals_parse_error", "username"},
		{"users", "alternation", "username=(user1|user2)"},
		{"users", "quoted_comma_literal", `email="u1,quoted`},
		{"users", "empty_value_regex_all", "username="},
		{"keys", "all_empty_filter", ""},
		{"keys", "eq_name", "name=Alpha Key"},
		{"keys", "ci_option", "name=KEY,i"},
		{"keys", "type_filter_internal_masked", "type=internal"},
		{"keys", "id_prefix", "id=k-"},
		{"keys", "invalid_regex", "name=["},
	}
	call := func(db dbT, surface, filter string) ([]string, error) {
		ctx := tplCtxRoot()
		if surface == "users" {
			l, err := db.GetUsers(ctx, filter)
			ids := make([]string, 0, len(l))
			for i := range l {
				ids = append(ids, l[i].ID)
			}
			return ids, err
		}
		l, err := db.GetApiKeys(ctx, filter)
		ids := make([]string, 0, len(l))
		for i := range l {
			ids = append(ids, l[i].ID)
		}
		return ids, err
	}
	for _, c := range corpus {
		t.Run("filter/"+c.surface+"/"+c.name, func(t *testing.T) {
			l, lerr := call(legacySvc, c.surface, c.filter)
			g, gerr := call(gormSvc, c.surface, c.filter)
			assert.Equal(t, lerr, gerr, "error agreement for filter %q", c.filter)
			assert.Equal(t, l, g, "EXACT sequence agreement for filter %q", c.filter)
		})
	}
	// Corpus non-vacuity: the unfiltered surface must actually hold rows.
	vUsers, verr := call(legacySvc, "users", "")
	require.NoError(t, verr)
	require.Len(t, vUsers, 3, "filter corpus non-vacuous (users)")
	vKeys, verr := call(legacySvc, "keys", "")
	require.NoError(t, verr)
	require.Len(t, vKeys, 2, "filter corpus non-vacuous (keys, internal masked)")

	// RequiredRoles matches HasEffectiveRole over InjectedRoles and the
	// IsAuthorized flag gates the fall-through path (authorize_layer.go
	// :56-98) — roles, not claims, are the record pairing here.
	noRolesCtx := &parityCtx{auth: &basecontext.AuthorizationContext{IsAuthorized: true, User: &apimodels.ApiUser{ID: "ext-user"}}}
	operatorCtx := &parityCtx{auth: &basecontext.AuthorizationContext{IsAuthorized: true, User: &apimodels.ApiUser{ID: "ext-user"}, InjectedRoles: []string{"operator"}}}
	authzCases := []struct {
		name string
		ctx  basecontext.ApiContext
	}{
		{"nil_auth_all_hidden", authCtxPlain()},
		{"user_no_roles_partial", noRolesCtx},
		{"roles_operator_shown", operatorCtx},
	}
	for _, a := range authzCases {
		t.Run("authz/catalog/"+a.name, func(t *testing.T) {
			l, lerr := legacySvc.GetCatalogManifests(a.ctx, "")
			g, gerr := gormSvc.GetCatalogManifests(a.ctx, "")
			assert.Equal(t, lerr, gerr)
			lk, gk := manifestTriples(l), manifestTriples(g)
			// ElementsMatch, NOT Equal: this row pins authz VISIBILITY
			// (empty-vs-shown), not sequence — the catalog listing orders
			// UpdatedAt desc over per-engine wall-clock stamps, which is
			// the E-3 exclusion this header already declares (line 9).
			// macOS coarse clock ticks tie the back-to-back seed stamps on
			// the fast legacy engine; the T6 tiebreak is ID ascending and
			// CreateCatalogManifest DISCARDS the seed ID for a random
			// GenerateId (data/catalog.go:162, database/catalog.go:200),
			// so a tied stamp orders by per-engine random bytes — an
			// engine-local coin flip (found at the 425-06 human-gate
			// macOS run, -count=5 -shuffle=on). Cross-engine SEQUENCE
			// equality on an order surface is pinned deterministically by
			// TestParityCatalogOrderDifferential (explicit distinct
			// UpdatedAt, IDs preserved by raw-table seeds).
			assert.ElementsMatch(t, lk, gk, "authz-visible agreement (empty-vs-shown)")
		})
	}
	// Non-vacuity: the operator ctx must see strictly more than the no-roles
	// posture on BOTH engines (otherwise the authz rows prove nothing).
	lMore, _ := legacySvc.GetCatalogManifests(operatorCtx, "")
	lLess, _ := legacySvc.GetCatalogManifests(noRolesCtx, "")
	require.Greater(t, len(lMore), len(lLess), "legacy authz differential non-vacuous")
	gMore, _ := gormSvc.GetCatalogManifests(operatorCtx, "")
	gLess, _ := gormSvc.GetCatalogManifests(noRolesCtx, "")
	require.Greater(t, len(gMore), len(gLess), "gorm authz differential non-vacuous")
}

func manifestTriples(l []models.CatalogManifest) []string {
	out := make([]string, 0, len(l))
	for i := range l {
		out = append(out, l[i].CatalogId+"|"+l[i].Version+"|"+l[i].Architecture)
	}
	return out
}
