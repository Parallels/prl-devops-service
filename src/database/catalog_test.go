package database

import (
	"errors"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func seedMani(t *testing.T, g *GormDatabase, id, catalogId, version, arch string, extra ...func(*models.CatalogManifest)) {
	t.Helper()
	m := &models.CatalogManifest{ID: id, CatalogId: catalogId, Name: catalogId + "-" + version,
		Version: version, Architecture: arch, UpdatedAt: "2026-01-01T00:00:00Z", CreatedAt: "2026-01-01T00:00:00Z",
		Tags: []string{"v" + version}}
	for _, e := range extra {
		e(m)
	}
	if err := g.gormDB.Table("catalog_manifests").Create(m).Error; err != nil {
		t.Fatal(err)
	}
}

func maniRootCtx() basecontext.ApiContext { return tplRootCtx() } // shared root ctx from packer tests

func TestSQLCatalogManifests_OrderingAuthz(t *testing.T) {
	g := newSQLDB(t)
	// UpdatedAt ties: rowid decides (catches any stray SQL ORDER BY).
	seedMani(t, g, "m-a", "LIN", "1.0", "amd64")                                   // rowid 1
	seedMani(t, g, "m-b", "WIN", "1.0", "amd64", func(m *models.CatalogManifest) { // rowid 2
		m.UpdatedAt = "2026-05-05T00:00:00Z" // newest
	})
	seedMani(t, g, "m-c", "LIN", "2.0", "amd64", func(m *models.CatalogManifest) {
		m.UpdatedAt = "2026-03-03T00:00:00Z"
		m.RequiredRoles = []string{"operator"} // authz-guarded
	})
	root := maniRootCtx()

	all, err := g.GetCatalogManifests(root, "")
	if err != nil || len(all) != 3 || all[0].ID != "m-b" || all[1].ID != "m-c" || all[2].ID != "m-a" {
		t.Fatalf("UpdatedAt desc: %v %+v", err, all)
	}
	// nil-auth deny-all (matrix row 1 again).
	if l, _ := g.GetCatalogManifests(basecontext.NewBaseContext(), ""); len(l) != 0 {
		t.Fatalf("nil-auth must see nothing: %+v", l)
	}
	// Array filters are Go-side over the serializer:json columns: D-4 renders
	// []string as "[a b]" — match via regex on that rendering.
	l, err := g.GetCatalogManifests(root, "tags=.*prod.*")
	if err != nil || len(l) != 1 || l[0].ID != "m-a" {
		// m-a tags = [v1.0]; add prod below and re-check via filter on it.
		t.Logf("tags filter preliminary: %v %+v", err, l)
	}
	if err := g.AddCatalogManifestTags(root, "m-a", "prod"); err != nil {
		t.Fatal(err)
	}
	l, err = g.GetCatalogManifests(root, "tags=.*prod.*")
	if err != nil || len(l) != 1 || l[0].ID != "m-a" {
		t.Fatalf("array filter over serialized column: %v %+v", err, l)
	}
	// Ordering corpus: empty result keeps non-nil shape via nil-safe layers;
	// filter errors propagate.
	if _, err := g.GetCatalogManifests(root, "bad~filter"); err == nil {
		t.Fatal("malformed filter must error")
	}
}

func TestSQLCatalogManifests_ReadPipelines(t *testing.T) {
	g := newSQLDB(t)
	root := maniRootCtx()
	seedMani(t, g, "m-1", "UBUNTU SERVER", "1.0", "amd64") // Name = "UBUNTU SERVER-1.0"

	// ByName: ID arm takes the NORMALISED needle; Name arm the RAW one
	// (space vs underscore matters: the raw arm never normalises).
	if _, err := g.GetCatalogManifestByName(root, "ubuntu server-1.0"); err != nil {
		t.Fatalf("by name (raw needle, folded): %v", err)
	}
	if m, err := g.GetCatalogManifestByName(root, "M-1"); err != nil || m.ID != "m-1" {
		t.Fatalf("by id via normalised needle: %v %+v", err, m)
	}
	// ByCatalogId: empty is a non-nil EMPTY slice, no error.
	empty, err := g.GetCatalogManifestsByCatalogId(root, "nope")
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("by catalog id empty shape: %v %#v", err, empty)
	}
	// ByCatalogIdAndVersion: empty IS an error (:128).
	if _, err := g.GetCatalogManifestsByCatalogIdAndVersion(root, "nope", "1.0"); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("by id+version empty must sentinel: %v", err)
	}
	// catalog_id normalises: "ubuntu server" → "ubuntu_server", rows store
	// whatever seed wrote — match needs the SAME normalisation both sides.
	seedMani(t, g, "m-2", "UBUNTU_SERVER", "2.0", "amd64")
	hit, err := g.GetCatalogManifestsByCatalogIdAndVersion(root, "ubuntu server", "2.0")
	if err != nil || len(hit) != 1 || hit[0].ID != "m-2" {
		t.Fatalf("normalised catalog id match: %v %+v", err, hit)
	}
	// ByTag: catalogId normalised, tag folded.
	if err := g.AddCatalogManifestTags(root, "m-2", "LTS"); err != nil {
		t.Fatal(err)
	}
	if m, err := g.GetCatalogManifestByTag(root, "ubuntu server", "lts"); err != nil || m.ID != "m-2" {
		t.Fatalf("by tag: %v %+v", err, m)
	}
	if _, err := g.GetCatalogManifestByTag(root, "ubuntu server", "ghost"); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("by tag missing: %v", err)
	}
	// ByCatalogIdVersionAndArch.
	if _, err := g.GetCatalogManifestsByCatalogIdVersionAndArch(root, "UBUNTU_SERVER", "2.0", "ARM64"); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("arch miss must sentinel: %v", err)
	}
}

func TestSQLCatalogManifests_CreateUpsert(t *testing.T) {
	g := newSQLDB(t)
	root := maniRootCtx()

	// Fresh create: caller id discarded, version defaults, names normalised,
	// "latest" tag force-added, timestamps stamped.
	created, err := g.CreateCatalogManifest(root, models.CatalogManifest{Name: "My Image!", CatalogId: "My Image!", Architecture: "amd64"})
	if err != nil || created.ID == "" || created.Version != "latest" ||
		created.CatalogId != "MY_IMAGE_" || !created.HasTag("latest") ||
		created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("create normalisation: %v %+v", err, created)
	}
	// Upsert arm: same (catalogId, version, arch) replaces in place —
	// identity fields inherited, id unchanged.
	again, err := g.CreateCatalogManifest(root, models.CatalogManifest{Name: "whatever", CatalogId: "my image_", Version: "latest", Architecture: "amd64", Size: 42})
	if err != nil || again.Size != 42 {
		t.Fatalf("upsert arm: %v %+v", err, again)
	}
	if again.ID != created.ID || again.Name != created.Name || again.CatalogId != created.CatalogId {
		t.Fatalf("upsert must inherit identity: %+v vs %+v", again, created)
	}
	var n int64
	g.gormDB.Table("catalog_manifests").Count(&n)
	if n != 1 {
		t.Fatalf("upsert created a second row: %d", n)
	}
	// Sibling "latest" stripping: new VERSION strips latest from siblings,
	// and the NEWER version keeps latest.
	sib, err := g.CreateCatalogManifest(root, models.CatalogManifest{Name: "whatever", CatalogId: "my image_", Version: "2.0", Architecture: "amd64"})
	if err != nil || !sib.HasTag("latest") {
		t.Fatalf("new version must own latest: %+v", sib)
	}
	old, err := g.GetCatalogManifestByName(root, created.ID)
	if err != nil || old.HasTag("latest") {
		t.Fatalf("sibling latest must be stripped: %v %+v", err, old)
	}
}

func TestSQLCatalogManifests_DeleteCascades(t *testing.T) {
	g := newSQLDB(t)
	root := maniRootCtx()
	seedMani(t, g, "d-1", "LIN", "1.0", "amd64")
	seedMani(t, g, "d-2", "LIN", "1.0", "arm64")
	seedMani(t, g, "d-3", "LIN", "2.0", "amd64")

	if err := g.DeleteCatalogManifest(root, ""); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("empty delete must sentinel: %v", err)
	}
	// Version delete removes ONLY the first UpdatedAt-desc match (legacy
	// returns inside the first hit): rows 1 and 2 tie on UpdatedAt → the
	// rowid-later d-2 comes first in desc order with ties keeping input
	// order reversed — the ported order layer decides; assert exactly one
	// row of that version died.
	if err := g.DeleteCatalogManifestVersion(root, "LIN", "1.0"); err != nil {
		t.Fatalf("version delete: %v", err)
	}
	var n int64
	g.gormDB.Table("catalog_manifests").Where("version = ?", "1.0").Count(&n)
	if n != 1 {
		t.Fatalf("version delete must kill exactly one row: %d", n)
	}
	// VersionArch delete is precise.
	if err := g.DeleteCatalogManifestVersionArch(root, "lin", "1.0", "arm64"); err != nil {
		t.Fatalf("version-arch delete (folded catalog id): %v", err)
	}
	g.gormDB.Table("catalog_manifests").Where("version = ?", "1.0").Count(&n)
	if n != 0 {
		t.Fatalf("version-arch delete drift: %d", n)
	}
	// Whole-catalog cascade: deletes every visible match.
	if err := g.DeleteCatalogManifest(root, "LIN"); err != nil {
		t.Fatalf("cascade: %v", err)
	}
	g.gormDB.Table("catalog_manifests").Count(&n)
	if n != 0 {
		t.Fatalf("cascade left rows: %d", n)
	}
	if err := g.DeleteCatalogManifest(root, "LIN"); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("cascade with no match: %v", err)
	}
}

func TestSQLCatalogManifests_UpdateQuirks(t *testing.T) {
	g := newSQLDB(t)
	root := maniRootCtx()
	seedMani(t, g, "u-1", "LIN", "1.0", "amd64", func(m *models.CatalogManifest) {
		m.Description = "keep me"
		m.DownloadCount = 7
		m.Tags = []string{"keep"}
	})

	// Immutable version/architecture with verbatim messages (the errors.Newf
	// Error() rendering carries the "error: " prefix on BOTH engines).
	_, err := g.UpdateCatalogManifest(root, models.CatalogManifest{ID: "U-1", Version: "2.0", Architecture: "amd64"})
	if err == nil || err.Error() != "error: cannot update version of catalog manifest U-1, it is trying to change the version 1.0 to version 2.0, not allowed" {
		t.Fatalf("version guard message drift: %v", err)
	}
	if _, err := g.UpdateCatalogManifest(root, models.CatalogManifest{ID: "u-1", Version: "1.0", Architecture: "arm64"}); err == nil {
		t.Fatal("arch guard must fire")
	}

	// Assignment set: Name/Size/Tags land, Description/DownloadCount do NOT.
	got, err := g.UpdateCatalogManifest(root, models.CatalogManifest{ID: "u-1", Version: "1.0", Architecture: "amd64",
		Name: "renamed", Size: 99, Tags: []string{"fresh"}})
	if err != nil || got.Name != "renamed" || got.Size != 99 || len(got.Tags) != 1 || got.Tags[0] != "fresh" {
		t.Fatalf("assignment set: %v %+v", err, got)
	}
	if got.Description != "keep me" || got.DownloadCount != 7 {
		t.Fatalf("update leaked untouched fields: %+v", got)
	}
	if got.CreatedAt == "" || got.UpdatedAt == "" {
		t.Fatalf("timestamps: %+v", got)
	}
	// UpdateCatalogManifestTags ALSO matches CatalogId+Version (:366).
	if err := g.UpdateCatalogManifestTags(root, models.CatalogManifest{CatalogId: "LIN", Version: "1.0", Tags: []string{"via-cv"}}); err != nil {
		t.Fatalf("tags by catalog+version arm: %v", err)
	}
	if _, err := g.UpdateCatalogManifest(root, models.CatalogManifest{ID: "ghost"}); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("update missing: %v", err)
	}
}

func TestSQLCatalogArrays_Quirks(t *testing.T) {
	g := newSQLDB(t)
	root := maniRootCtx()
	seedMani(t, g, "a-1", "LIN", "1.0", "amd64", func(m *models.CatalogManifest) {
		m.Tags = []string{"keep"}
	})

	// ADD quirk: "keep" (existing, FIRST) poisons `found`, so "new" — a GENUINE
	// new value — is SKIPPED for the whole call (found never resets).
	if err := g.AddCatalogManifestTags(root, "a-1", "keep", "new"); err != nil {
		t.Fatal(err)
	}
	m, _ := g.GetCatalogManifestByName(root, "a-1")
	if len(m.Tags) != 1 || m.Tags[0] != "keep" {
		t.Fatalf("found-outside-loop quirk not mirrored: %+v", m.Tags)
	}
	// Same call with only the new value goes through.
	if err := g.AddCatalogManifestTags(root, "a-1", "fresh"); err != nil {
		t.Fatal(err)
	}
	m, _ = g.GetCatalogManifestByName(root, "a-1")
	if len(m.Tags) != 2 || m.Tags[1] != "fresh" {
		t.Fatalf("plain add: %+v", m.Tags)
	}
	// REMOVE of an absent tag is a silent no-op (:452 return nil).
	if err := g.RemoveCatalogManifestTags(root, "a-1", "ghost"); err != nil {
		t.Fatalf("remove absent tag must be nil: %v", err)
	}
	// Remove hits folded matches.
	if err := g.RemoveCatalogManifestTags(root, "a-1", "FRESH"); err != nil {
		t.Fatal(err)
	}
	m, _ = g.GetCatalogManifestByName(root, "a-1")
	if len(m.Tags) != 1 {
		t.Fatalf("folded remove: %+v", m.Tags)
	}
	// UpdateRequiredRoles is ADD-with-dedupe — but the dedupe only checks
	// the STORED array (the scanned slice never refreshes), so passing the
	// same new value twice IN ONE CALL stores it twice, exactly as legacy.
	if err := g.UpdateCatalogManifestRequiredRoles(root, "a-1", "op", "op"); err != nil {
		t.Fatal(err)
	}
	m, _ = g.GetCatalogManifestByName(root, "a-1")
	if len(m.RequiredRoles) != 2 {
		t.Fatalf("in-call dedupe must NOT happen (legacy): %+v", m.RequiredRoles)
	}
	// Claims round-trip too.
	if err := g.AddCatalogManifestRequiredClaims(root, "a-1", "read-vm"); err != nil {
		t.Fatal(err)
	}
	if err := g.RemoveCatalogManifestRequiredClaims(root, "a-1", "READ-VM"); err != nil {
		t.Fatal(err)
	}
	m, _ = g.GetCatalogManifestByName(root, "a-1")
	if len(m.RequiredClaims) != 0 {
		t.Fatalf("claims mutate: %+v", m.RequiredClaims)
	}
	// Unknown record ⇒ sentinel on every mutator.
	if err := g.AddCatalogManifestTags(root, "ghost", "x"); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("mutator sentinel: %v", err)
	}
}

func TestSQLCatalogManifests_FlagsAndCounters(t *testing.T) {
	g := newSQLDB(t)
	root := maniRootCtx()
	seedMani(t, g, "f-1", "LIN", "1.0", "amd64", func(m *models.CatalogManifest) { m.Description = "orig" })

	// Download counter: read-modify-write, user from ctx.
	if err := g.UpdateCatalogManifestDownloadCount(root, "LIN", "1.0"); err != nil {
		t.Fatal(err)
	}
	m, _ := g.GetCatalogManifestByName(root, "f-1")
	if m.DownloadCount != 1 || m.LastDownloadedUser != "root" || m.LastDownloadedAt == "" {
		t.Fatalf("counter: %+v", m)
	}
	// nil-auth ctx ⇒ "unknown" user (the ctx preamble, not an authz gate:
	// the write still happens).
	if err := g.UpdateCatalogManifestDownloadCount(basecontext.NewBaseContext(), "lin", "1.0"); err != nil {
		t.Fatalf("counter with folded catalog id: %v", err)
	}
	m, _ = g.GetCatalogManifestByName(root, "f-1")
	if m.DownloadCount != 2 || m.LastDownloadedUser != "unknown" {
		t.Fatalf("unknown user path: %+v", m)
	}

	// Taint / Untaint / Revoke flag triples + returned row agrees with store.
	tw, err := g.TaintCatalogManifestVersion(root, "LIN", "1.0")
	if err != nil || !tw.Tainted || tw.TaintedBy != "root" || tw.TaintedAt == "" {
		t.Fatalf("taint: %v %+v", err, tw)
	}
	uw, err := g.UnTaintCatalogManifestVersion(root, "LIN", "1.0")
	if err != nil || uw.Tainted || uw.UnTaintedBy != "root" || uw.TaintedBy != "" || uw.TaintedAt != "" {
		t.Fatalf("untaint: %v %+v", err, uw)
	}
	rw, err := g.RevokeCatalogManifestVersion(root, "LIN", "1.0")
	if err != nil || !rw.Revoked || rw.RevokedBy != "root" || rw.RevokedAt == "" {
		t.Fatalf("revoke: %v %+v", err, rw)
	}
	stored, _ := g.GetCatalogManifestByName(root, "f-1")
	if stored.Revoked != rw.Revoked || stored.RevokedAt != rw.RevokedAt {
		t.Fatalf("returned row must agree with stored: %+v vs %+v", stored, rw)
	}
	if _, err := g.TaintCatalogManifestVersion(root, "LIN", "9.9"); !errors.Is(err, interfaces.ErrCatalogManifestNotFound) {
		t.Fatalf("taint unknown version: %v", err)
	}

	// Metadata: nil description/nil slices DON'T touch; non-nil (even empty) DO.
	got, err := g.UpdateCatalogManifestMetadata(root, "LIN", "1.0", "amd64", nil, []string{}, nil, nil)
	if err != nil || got.Description != "orig" || len(got.Tags) != 0 {
		t.Fatalf("metadata nil-semantics: %v %+v", err, got)
	}
	desc := "touched"
	got, err = g.UpdateCatalogManifestMetadata(root, "LIN", "1.0", "amd64", &desc, nil, []string{"c1"}, []string{"r1"})
	// tags stayed nil here ⇒ untouched, still the [] the previous call set.
	if err != nil || got.Description != "touched" || len(got.Tags) != 0 || got.RequiredClaims[0] != "c1" || got.RequiredRoles[0] != "r1" {
		t.Fatalf("metadata set: %v %+v", err, got)
	}

	// Provider: value marshalled into the serializer column.
	if err := g.UpdateCatalogManifestProvider(root, "f-1", models.CatalogManifestProvider{Type: "sftp", Host: "h"}); err != nil {
		t.Fatal(err)
	}
	m2, _ := g.GetCatalogManifestByName(root, "F-1")
	if m2.Provider == nil || m2.Provider.Type != "sftp" || m2.Provider.Host != "h" {
		t.Fatalf("provider round-trip: %+v", m2.Provider)
	}
}

func TestSQLCatalogManagers(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	if _, err := g.GetCatalogManager("nope"); !errors.Is(err, interfaces.ErrCatalogManagerNotFound) {
		t.Fatalf("missing manager: %v", err)
	}
	// EXACT id equality: folded lookup misses (:28).
	if err := g.AddCatalogManager(ctx, models.CatalogManager{ID: "mgr-a", Name: "A"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.GetCatalogManager("MGR-A"); !errors.Is(err, interfaces.ErrCatalogManagerNotFound) {
		t.Fatalf("case-folded get must miss: %v", err)
	}
	// No dup pre-check mirrored; SQL PK refuses (stricter deviation).
	if err := g.AddCatalogManager(ctx, models.CatalogManager{ID: "mgr-a", Name: "dup"}); err == nil {
		t.Fatal("duplicate id must surface (SQL stricter than legacy's double-store)")
	}
	// Whole-row replacement + internal blocks EVERYONE.
	if err := g.AddCatalogManager(ctx, models.CatalogManager{ID: "mgr-i", Name: "I", Internal: true}); err != nil {
		t.Fatal(err)
	}
	if err := g.UpdateCatalogManager(tplRootCtx(), models.CatalogManager{ID: "mgr-i", Name: "nope"}); !errors.Is(err, interfaces.ErrUpdateInternalCatalogManager) {
		t.Fatalf("internal update guard (root): %v", err)
	}
	if err := g.UpdateCatalogManager(ctx, models.CatalogManager{ID: "mgr-a", Name: "A2", URL: "u", Active: true,
		CreatedAt: "2020-01-01T00:00:00Z", RequiredClaims: []string{"cm"}}); err != nil {
		t.Fatal(err)
	}
	mgr, err := g.GetCatalogManager("mgr-a")
	if err != nil || mgr.Name != "A2" || mgr.URL != "u" || !mgr.Active || mgr.CreatedAt != "2020-01-01T00:00:00Z" ||
		len(mgr.RequiredClaims) != 1 {
		t.Fatalf("full replacement: %v %+v", err, mgr)
	}
	// Internal delete: non-root refused, root allowed.
	if err := g.DeleteCatalogManager(ctx, "mgr-i"); !errors.Is(err, interfaces.ErrRemoveInternalCatalogManager) {
		t.Fatalf("internal delete guard: %v", err)
	}
	if err := g.DeleteCatalogManager(tplRootCtx(), "mgr-i"); err != nil {
		t.Fatalf("internal delete by root: %v", err)
	}
	if err := g.DeleteCatalogManager(ctx, "MGR-A"); !errors.Is(err, interfaces.ErrCatalogManagerNotFound) {
		t.Fatalf("delete is exact-id too: %v", err)
	}
	if err := g.DeleteCatalogManager(ctx, "mgr-a"); err != nil {
		t.Fatal(err)
	}
	all, err := g.GetCatalogManagers()
	if err != nil || len(all) != 0 {
		t.Fatalf("managers empty: %v %+v", err, all)
	}
}

func TestSQLCatalog_NotConnected(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cold := NewGormDatabase(ctx, nil, nil)

	if _, err := cold.GetCatalogManifests(ctx, ""); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("list guard: %v", err)
	}
	if _, err := cold.GetCatalogManifestByName(ctx, "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("byname guard: %v", err)
	}
	if _, err := cold.GetCatalogManifestsByCatalogId(ctx, "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("bycatalog guard: %v", err)
	}
	if _, err := cold.CreateCatalogManifest(ctx, models.CatalogManifest{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("create guard: %v", err)
	}
	if err := cold.DeleteCatalogManifest(ctx, "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("delete guard: %v", err)
	}
	if err := cold.DeleteCatalogManifestVersion(ctx, "x", "1"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("deletev guard: %v", err)
	}
	if err := cold.DeleteCatalogManifestVersionArch(ctx, "x", "1", "a"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("deleteva guard: %v", err)
	}
	if _, err := cold.UpdateCatalogManifest(ctx, models.CatalogManifest{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("update guard: %v", err)
	}
	if err := cold.UpdateCatalogManifestTags(ctx, models.CatalogManifest{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("tags guard: %v", err)
	}
	if err := cold.AddCatalogManifestTags(ctx, "x", "y"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("array guard: %v", err)
	}
	if err := cold.UpdateCatalogManifestRequiredRoles(ctx, "x", "r"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("roles guard: %v", err)
	}
	if _, err := cold.TaintCatalogManifestVersion(ctx, "x", "1"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("taint guard: %v", err)
	}
	if _, err := cold.UnTaintCatalogManifestVersion(ctx, "x", "1"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("untaint guard: %v", err)
	}
	if _, err := cold.RevokeCatalogManifestVersion(ctx, "x", "1"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("revoke guard: %v", err)
	}
	if _, err := cold.UpdateCatalogManifestMetadata(ctx, "x", "1", "a", nil, nil, nil, nil); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("metadata guard: %v", err)
	}
	if err := cold.UpdateCatalogManifestProvider(ctx, "x", models.CatalogManifestProvider{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("provider guard: %v", err)
	}
	if err := cold.UpdateCatalogManifestDownloadCount(ctx, "x", "1"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("counter guard: %v", err)
	}
	if _, err := cold.GetCatalogManifestByTag(ctx, "x", "t"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("bytag guard: %v", err)
	}
	if _, err := cold.GetCatalogManifestsByCatalogIdAndVersion(ctx, "x", "1"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("bycv guard: %v", err)
	}
	if _, err := cold.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, "x", "1", "a"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("bycva guard: %v", err)
	}
	if _, err := cold.AddPackerTemplate(ctx, &models.PackerTemplate{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("sanity (packer still guarded): %v", err)
	}
	// Managers #26/#27 have NO guard (legacy parity) — they answer from the
	// cold handle's nil gormDB; assert only that the GUARDED manager ops error.
	if err := cold.AddCatalogManager(ctx, models.CatalogManager{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("manager add guard: %v", err)
	}
	if err := cold.DeleteCatalogManager(ctx, "x"); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("manager delete guard: %v", err)
	}
	if err := cold.UpdateCatalogManager(ctx, models.CatalogManager{}); err != interfaces.ErrDatabaseNotConnected {
		t.Errorf("manager update guard: %v", err)
	}
}
