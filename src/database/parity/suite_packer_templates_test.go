package parity

// DomainSuite "packer_templates" (PBI #421 task 10): D-4 showcase — the
// FULL task-02 golden filter corpus (database/query/testdata/fixtures.json
// filters[], not a hand-picked subset) replayed through GetPackerTemplates
// on both engines, plus the authz matrix rows the domain exercises (nil-auth
// deny-all, root bypass, required-role guard), mutation parity in identical
// call order, and the unupdatable/deletable internal-template guards.

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

func tplCtxRoot() basecontext.ApiContext {
	return &parityCtx{auth: &basecontext.AuthorizationContext{
		AuthorizedBy: "RootAuthorization", User: &apimodels.ApiUser{ID: "root", Username: "root"}}}
}

func tplCtxUser(userID string) basecontext.ApiContext {
	return &parityCtx{auth: &basecontext.AuthorizationContext{
		User: &apimodels.ApiUser{ID: userID}, InjectedClaims: []string{}}}
}

func goldenFilterCorpus(t testing.TB) []string {
	t.Helper()
	raw, err := os.ReadFile("../query/testdata/fixtures.json")
	if err != nil {
		t.Fatalf("read golden fixtures: %v", err)
	}
	var doc struct {
		Filters []string `json:"filters"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse golden fixtures: %v", err)
	}
	if len(doc.Filters) == 0 {
		t.Fatal("golden corpus empty")
	}
	return doc.Filters
}

func seedPackerTemplates(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, r := range constants.DefaultRoles {
		if _, err := db.CreateRole(ctx, models.Role{Name: r}); err != nil {
			t.Fatalf("seed default role %s: %v", r, err)
		}
	}
	for _, c := range constants.DefaultClaims {
		if _, err := db.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed default claim %s: %v", c, err)
		}
	}
	// DefaultRoles is only [user]; the guarded template references operator.
	if _, err := db.CreateRole(ctx, models.Role{Name: "operator"}); err != nil {
		t.Fatalf("seed operator role: %v", err)
	}
	// Required claim/role references must resolve ⇒ seeded first; note Add
	// stamps its own timestamps (normalise before compare).
	add := func(tpl *models.PackerTemplate) {
		if _, err := db.AddPackerTemplate(ctx, tpl); err != nil {
			t.Fatalf("seed tpl %s: %v", tpl.ID, err)
		}
	}
	add(&models.PackerTemplate{ID: "ubuntu-2204", Name: "ubuntu-2204", Owner: "team-a",
		Specs: map[string]string{"memory": "2048"}, Addons: []string{"guest"},
		RequiredRoles: []string{"operator"}, RequiredClaims: []string{constants.DefaultClaims[0]}})
	add(&models.PackerTemplate{ID: "win-2022", Name: "Windows-2022", Owner: "team-b",
		Variables: map[string]string{"user": "winadmin"}})
	add(&models.PackerTemplate{ID: "fedora-40", Name: "fedora-40", Internal: true})
	add(&models.PackerTemplate{ID: "ubu-like", Name: "Ubuntu Minimal", Hostname: "tpl-host",
		Description: "second", Defaults: map[string]string{"k": "v"}})
}

func normaliseTpls(in []models.PackerTemplate) []models.PackerTemplate {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.PackerTemplate(nil), in...)
	for i := range out {
		out[i].CreatedAt, out[i].UpdatedAt = "", "" // Add stamps per-engine clocks
	}
	return out
}

func TestParityDomain_PackerTemplates(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "packer_templates",
		Seed: seedPackerTemplates,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkGoldenCorpus,
			checkTplAuthzMatrix,
			checkTplLookups,
			checkTplMutations,
		},
	})
}

func checkGoldenCorpus(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := tplCtxRoot() // root passes authz so the corpus exercises filters, not authz
	for _, f := range goldenFilterCorpus(t) {
		lt, lerr := legacy.GetPackerTemplates(ctx, f)
		st, serr := sql.GetPackerTemplates(ctx, f)
		assertSameError(t, "golden filter ["+f+"]", lerr, serr)
		AssertSameList(t, "GetPackerTemplates["+f+"]", normaliseTpls(lt), normaliseTpls(st))
	}
}

func checkTplAuthzMatrix(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	// nil-auth: deny-all on both engines (matrix row 1 — public included).
	bare := basecontext.NewBaseContext()
	AssertSameList(t, "nil-auth list",
		normaliseTpls(mustTpls(legacy, bare)), normaliseTpls(mustTpls(sql, bare)))
	// Authenticated non-root: operator-required tpl hidden, public visible.
	lt, lerr := legacy.GetPackerTemplates(tplCtxUser("u9"), "")
	st, serr := sql.GetPackerTemplates(tplCtxUser("u9"), "")
	assertSameError(t, "user-ctx list", lerr, serr)
	AssertSameList(t, "user-ctx list", normaliseTpls(lt), normaliseTpls(st))
	// Root sees the whole table.
	AssertSameList(t, "root list",
		normaliseTpls(mustTpls(legacy, tplCtxRoot())), normaliseTpls(mustTpls(sql, tplCtxRoot())))
}

func checkTplLookups(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	probes := []struct {
		ctx basecontext.ApiContext
		arg string
	}{
		{tplCtxRoot(), "ubuntu-2204"}, {tplCtxRoot(), "UBUNTU-2204"},
		{tplCtxRoot(), "Windows-2022"}, {tplCtxRoot(), "nope"}, {tplCtxRoot(), ""},
		{basecontext.NewBaseContext(), "win-2022"}, // nil-auth: invisible through the wrapped lookup
		{tplCtxUser("u9"), "ubuntu-2024"},          // wrong id + non-root
	}
	for _, p := range probes {
		lt, lerr := legacy.GetPackerTemplate(p.ctx, p.arg)
		st, serr := sql.GetPackerTemplate(p.ctx, p.arg)
		assertSameError(t, "GetPackerTemplate["+p.arg+"]", lerr, serr)
		if lerr == nil && (lt == nil || st == nil || lt.ID != st.ID || lt.Name != st.Name) {
			t.Errorf("lookup drift %q: %+v vs %+v", p.arg, lt, st)
		}
	}
}

func checkTplMutations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	// Identical call order on both engines, re-list after each step.
	for _, db := range []datainterfaces.DatabaseService{legacy, sql} {
		ctx := tplCtxRoot()
		_, _ = db.UpdatePackerTemplate(ctx, &models.PackerTemplate{ID: "ubu-like", Name: "Renamed Full",
			Owner: "team-c", CreatedAt: "2020-01-01T00:00:00Z", Specs: map[string]string{"cpu": "4"}})
		_, _ = db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "ubuntu-2204", Name: "dup-id"})
		_, _ = db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "new-tpl", Name: "Windows-2022"}) // dup name allowed
		_, _ = db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "ghost-req", Name: "g", RequiredClaims: []string{"ghost"}})
		_ = db.DeletePackerTemplate(ctx, "fedora-40")
		_, _ = db.UpdatePackerTemplate(ctx, &models.PackerTemplate{ID: "win-2022", Name: "x"})
		_ = db.DeletePackerTemplate(ctx, "")
	}
	AssertSameList(t, "after-mutations (root)",
		normaliseTpls(mustTpls(legacy, tplCtxRoot())), normaliseTpls(mustTpls(sql, tplCtxRoot())))
	// The renamed row's full replacement must match byte-wise.
	lu, lerr := legacy.GetPackerTemplate(tplCtxRoot(), "ubu-like")
	su, serr := sql.GetPackerTemplate(tplCtxRoot(), "ubu-like")
	assertSameError(t, "post-mutation lookup", lerr, serr)
	if lerr == nil && (lu.Name != su.Name || lu.CreatedAt != su.CreatedAt || lu.Variables == nil && su.Variables != nil || lu.Variables != nil && su.Variables == nil) {
		t.Errorf("full-replacement drift: %+v vs %+v", lu, su)
	}
	if lerr == nil && lu.CreatedAt != "2020-01-01T00:00:00Z" {
		t.Errorf("payload CreatedAt must survive (full replacement): %q", lu.CreatedAt)
	}
}

func mustTpls(db datainterfaces.DatabaseService, ctx basecontext.ApiContext) []models.PackerTemplate {
	out, _ := db.GetPackerTemplates(ctx, "")
	return out
}
