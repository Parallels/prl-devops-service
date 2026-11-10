package parity

// DomainSuite "catalog" (PBI #421 task 11): the ordering matrix (UpdatedAt
// desc, ties keep input order — D-4's showcase), Go-side array filters over
// the serializer:json columns, the read pipelines (ByName/ByTag/ByCatalogId
// /+Version/+Arch with their divergent empty-result verdicts), and the
// taint/revoke/counter/manager flows run as an identical call order on both
// engines. Manifest IDs are engine-generated → normalised to the
// (catalog_id, version, architecture) key before list compare; all
// clock-derived stamps normalised away.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func maniCtx() basecontext.ApiContext { return tplCtxRoot() }

func seedCatalog(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := maniCtx()
	create := func(catalog, version, arch string, extra ...func(m *models.CatalogManifest)) {
		m := models.CatalogManifest{CatalogId: catalog, Version: version, Architecture: arch, Name: catalog + " image"}
		for _, e := range extra {
			e(&m)
		}
		if _, err := db.CreateCatalogManifest(ctx, m); err != nil {
			t.Fatalf("seed manifest %s/%s/%s: %v", catalog, version, arch, err)
		}
	}
	create("win", "1.0", "amd64", func(m *models.CatalogManifest) { m.Tags = []string{"stable", "prod"} })
	create("win", "1.1", "amd64", func(m *models.CatalogManifest) { m.RequiredRoles = []string{"operator"} })
	create("lin", "2.0", "amd64")
	create("lin", "2.0", "arm64") // same version, second arch — upsert must NOT swallow it
	if err := db.AddCatalogManager(ctx, models.CatalogManager{ID: "importer-a", Name: "A", URL: "http://a"}); err != nil {
		t.Fatalf("seed manager: %v", err)
	}
	if err := db.AddCatalogManager(ctx, models.CatalogManager{ID: "importer-i", Name: "I", Internal: true}); err != nil {
		t.Fatalf("seed internal manager: %v", err)
	}
}

// normaliseManifests keys rows by their stable business triple and blanks
// per-engine clock stamps; order (UpdatedAt desc) is preserved.
func normaliseManifests(in []models.CatalogManifest) []models.CatalogManifest {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.CatalogManifest(nil), in...)
	for i := range out {
		out[i].ID = out[i].CatalogId + "|" + out[i].Version + "|" + out[i].Architecture
		out[i].CreatedAt, out[i].UpdatedAt = "", ""
		out[i].LastDownloadedAt, out[i].TaintedAt, out[i].RevokedAt = "", "", ""
	}
	return out
}

func TestParityDomain_Catalog(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "catalog",
		Seed: seedCatalog,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkCatalogFilters,
			checkCatalogReads,
			checkCatalogMutations,
			checkCatalogManagers,
		},
	})
}

func checkCatalogFilters(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := maniCtx()
	for _, f := range []string{"", "name=win.*", "tags=.*prod.*", "requiredroles=.*operator.*",
		"tainted=true", "version=2.0", "architecture=arm64", "id=nope", "name=WIN.*", "name=win.*,i"} {
		lt, lerr := legacy.GetCatalogManifests(ctx, f)
		st, serr := sql.GetCatalogManifests(ctx, f)
		assertSameError(t, "GetCatalogManifests["+f+"]", lerr, serr)
		AssertSameList(t, "GetCatalogManifests["+f+"]", normaliseManifests(lt), normaliseManifests(st))
	}
}

func checkCatalogReads(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := maniCtx()
	key := func(m *models.CatalogManifest) string {
		if m == nil {
			return "<nil>"
		}
		return m.CatalogId + "|" + m.Version + "|" + m.Architecture
	}
	type probe struct {
		name string
		call func(db datainterfaces.DatabaseService) (string, error)
	}
	probes := []probe{
		{"ByName/lin", func(d datainterfaces.DatabaseService) (string, error) {
			m, e := d.GetCatalogManifestByName(ctx, "LIN image")
			return key(m), e
		}},
		{"ByName/id-fold", func(d datainterfaces.DatabaseService) (string, error) {
			l, _ := d.GetCatalogManifests(ctx, "")
			if len(l) == 0 {
				return "<empty>", nil
			}
			m, e := d.GetCatalogManifestByName(ctx, l[0].ID)
			return key(m), e
		}},
		{"ByName/ghost", func(d datainterfaces.DatabaseService) (string, error) {
			m, e := d.GetCatalogManifestByName(ctx, "ghost")
			return key(m), e
		}},
		{"ByTag", func(d datainterfaces.DatabaseService) (string, error) {
			m, e := d.GetCatalogManifestByTag(ctx, "win", "PROD")
			return key(m), e
		}},
		{"ByCatalogId", func(d datainterfaces.DatabaseService) (string, error) {
			l, e := d.GetCatalogManifestsByCatalogId(ctx, "lin")
			return keyList(l), e
		}},
		{"ByCatalogId/empty", func(d datainterfaces.DatabaseService) (string, error) {
			l, e := d.GetCatalogManifestsByCatalogId(ctx, "nope")
			return keyList(l), e
		}},
		{"ByCV", func(d datainterfaces.DatabaseService) (string, error) {
			l, e := d.GetCatalogManifestsByCatalogIdAndVersion(ctx, "lin", "2.0")
			return keyList(l), e
		}},
		{"ByCV/empty-sentinel", func(d datainterfaces.DatabaseService) (string, error) {
			l, e := d.GetCatalogManifestsByCatalogIdAndVersion(ctx, "nope", "1.0")
			return keyList(l), e
		}},
		{"ByCVA", func(d datainterfaces.DatabaseService) (string, error) {
			m, e := d.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, "win", "1.1", "amd64")
			return key(m), e
		}},
		{"ByCVA/miss", func(d datainterfaces.DatabaseService) (string, error) {
			m, e := d.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, "win", "1.1", "arm64")
			return key(m), e
		}},
	}
	for _, p := range probes {
		lv, lerr := p.call(legacy)
		sv, serr := p.call(sql)
		assertSameError(t, p.name, lerr, serr)
		if lv != sv {
			t.Errorf("%s: %q vs %q", p.name, lv, sv)
		}
	}
}

func keyList(l []models.CatalogManifest) string {
	n := normaliseManifests(l)
	out := ""
	for _, m := range n {
		out += m.ID + ";"
	}
	return out
}

func checkCatalogMutations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := maniCtx()
	for _, db := range []datainterfaces.DatabaseService{legacy, sql} {
		_ = db.UpdateCatalogManifestDownloadCount(ctx, "win", "1.0")
		_ = db.UpdateCatalogManifestDownloadCount(basecontext.NewBaseContext(), "win", "1.0")
		_, _ = db.TaintCatalogManifestVersion(ctx, "win", "1.0")
		_, _ = db.UnTaintCatalogManifestVersion(ctx, "win", "1.0")
		_, _ = db.RevokeCatalogManifestVersion(ctx, "lin", "2.0")
		// Newest row id differs per engine; address through a lookup.
		l, _ := db.GetCatalogManifests(ctx, "version=1.1")
		if len(l) > 0 {
			_ = db.AddCatalogManifestTags(ctx, l[0].ID, "revoked-manually", "stable")
			_ = db.RemoveCatalogManifestRequiredRoles(ctx, l[0].ID, "OPERATOR")
		}
		desc := "meta-edited"
		_, _ = db.UpdateCatalogManifestMetadata(ctx, "win", "1.0", "amd64", &desc, []string{"meta"}, nil, []string{"r1"})
		_, _ = db.CreateCatalogManifest(ctx, models.CatalogManifest{CatalogId: "win", Version: "2.0", Architecture: "amd64"}) // latest-strip
		_ = db.DeleteCatalogManifestVersion(ctx, "lin", "2.0")                                                                // first row only
		_ = db.DeleteCatalogManifestVersionArch(ctx, "lin", "2.0", "arm64")
	}
	AssertSameList(t, "after-mutations",
		normaliseManifests(mustManifests(legacy)), normaliseManifests(mustManifests(sql)))
}

func checkCatalogManagers(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := maniCtx()
	for _, db := range []datainterfaces.DatabaseService{legacy, sql} {
		_ = db.UpdateCatalogManager(ctx, models.CatalogManager{ID: "importer-a", Name: "A-renamed", Active: true,
			RequiredClaims: []string{"catalog-read"}})
		_ = db.UpdateCatalogManager(ctx, models.CatalogManager{ID: "importer-i", Name: "nope"}) // internal guard
		_ = db.DeleteCatalogManager(ctx, "importer-i")                                          // non-root refused
	}
	lm, lerr := legacy.GetCatalogManagers()
	sm, serr := sql.GetCatalogManagers()
	assertSameError(t, "managers", lerr, serr)
	AssertSameList(t, "managers", lm, sm)
	// Root may remove the internal manager; run AFTER the list compare.
	for _, db := range []datainterfaces.DatabaseService{legacy, sql} {
		_ = db.DeleteCatalogManager(tplCtxRoot(), "importer-i")
	}
}

func mustManifests(db datainterfaces.DatabaseService) []models.CatalogManifest {
	out, _ := db.GetCatalogManifests(maniCtx(), "")
	return out
}
