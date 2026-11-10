package database

// SQL implementation of the catalog-manifest domain (PBI #421 task 11).
// The biggest port in the chain: the only domain exercising all four query
// layers at once — GetCatalogManifests is filter -> authz wrap -> ORDER
// UpdatedAt desc (legacy :36, the D-4 ordering showcase: Go-side via the
// ported order layer, load order rowid, NO SQL ORDER BY; the opt-in
// push-down experiment was deliberately NOT shipped with this PR).
// Quirks mirrored (spec lines corrected against code at task start):
//   - ByName compares ID against NormalizeString(idOrName) but NAME against
//     the RAW needle (:55 asymmetry).
//   - ByCatalogIdAndVersion answers the sentinel on EMPTY results (:128);
//     ByCatalogId answers an empty non-nil slice instead.
//   - Create always mints a new id, defaults version to "latest",
//     normalises Name (lower) / CatalogId (upper), force-adds the latest
//     tag, strips "latest" from siblings (authz-visible ones only — the
//     sibling loop rides the authz-wrapped list), then UPSERTS: a visible
//     (catalogId, version, arch) hit replaces the stored row via
//     UpdateCatalogManifest; otherwise insert with helper timestamps.
//   - The deletes are authz-WRAPPED (they scan GetCatalogManifests):
//     an invisible row cannot be deleted, matching legacy exactly.
//     DeleteCatalogManifestVersion removes only the FIRST match in
//     UpdatedAt-desc order (:276-286 returns inside the first hit — the
//     spec's "delete all rows of that version" claim is stale).
//   - UpdateCatalogManifest: raw scan (no authz), version/architecture are
//     immutable (verbatim Newf messages), and the assignment set at
//     :336-351 leaves CatalogId/Version/Description/Provider/DownloadCount/
//     IsCompressed/PackRelativePath/PackSize/MinimumSpecRequirements/
//     Tainted*/Revoked* untouched; CreatedAt keeps the stored value.
//   - The counter/taint/revoke family addresses rows by (catalog_id,
//     version) FoldEqual over the RAW list, first hit only, and stamps
//     ctx.GetAuthorizationContext().User.Username or "unknown".
// 3 tx sites = the DeleteCatalogManifest* cascades (D-9 catalog.go:216-320).
// Serializer:json columns are written targeted-style with encoding/json
// (same encoder and bytes as GORM's serializer — see packer_template.go).

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

func (g *GormDatabase) loadCatalogManifests() ([]models.CatalogManifest, error) {
	var rows []models.CatalogManifest
	err := g.gormDB.Table("catalog_manifests").Order("rowid").Find(&rows).Error
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

// catalogUsername mirrors the downloadUser/taintUser preamble shared by
// legacy :576-580, :600-604, :624-628, :649-653.
func catalogUsername(ctx basecontext.ApiContext) string {
	auth := ctx.GetAuthorizationContext()
	if auth != nil && auth.User != nil && auth.User.Username != "" {
		return auth.User.Username
	}
	return "unknown"
}

// catalogJSON renders one serializer:json column value — encoding/json is
// exactly what GORM's serializer:json writes on Create/Update, so targeted
// column updates stay byte-identical to full-row writes.
func catalogJSON(col string, v interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return map[string]interface{}{col: string(b)}, nil
}

// catalogAuthVisible is the legacy read pattern: authz wrap then ORDER
// UpdatedAt desc over a loaded row slice (order AFTER authz, :34-36).
func catalogAuthVisible(ctx basecontext.ApiContext, rows []models.CatalogManifest) ([]models.CatalogManifest, error) {
	authorized := query.GetAuthorizedRecords(ctx, rows...)
	return query.OrderByProperty(authorized, &query.Order{Property: "UpdatedAt", Direction: query.OrderDirectionDesc})
}

func (g *GormDatabase) GetCatalogManifests(ctx basecontext.ApiContext, filter string) ([]models.CatalogManifest, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	rows, err := g.loadCatalogManifests()
	if err != nil {
		return nil, err
	}
	filtered, err := query.FilterByProperty(rows, dbFilter)
	if err != nil {
		return nil, err
	}
	return catalogAuthVisible(ctx, filtered)
}

// GetCatalogManifestByName goes through the authz+ordered list like every
// legacy read (:49). ID matches the NORMALISED needle, Name the raw one.
func (g *GormDatabase) GetCatalogManifestByName(ctx basecontext.ApiContext, idOrName string) (*models.CatalogManifest, error) {
	manifests, err := g.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}
	for i := range manifests {
		if strings.EqualFold(manifests[i].ID, helpers.NormalizeString(idOrName)) ||
			strings.EqualFold(manifests[i].Name, idOrName) {
			return &manifests[i], nil
		}
	}
	return nil, interfaces.ErrCatalogManifestNotFound
}

func (g *GormDatabase) GetCatalogManifestByTag(ctx basecontext.ApiContext, catalogId string, tag string) (*models.CatalogManifest, error) {
	manifests, err := g.GetCatalogManifests(ctx, "") // most-recently-updated wins (:73)
	if err != nil {
		return nil, err
	}
	for i := range manifests {
		if strings.EqualFold(manifests[i].CatalogId, helpers.NormalizeString(catalogId)) {
			for _, t := range manifests[i].Tags {
				if strings.EqualFold(t, tag) {
					return &manifests[i], nil
				}
			}
		}
	}
	return nil, interfaces.ErrCatalogManifestNotFound
}

func (g *GormDatabase) GetCatalogManifestsByCatalogId(ctx basecontext.ApiContext, catalogId string) ([]models.CatalogManifest, error) {
	result := make([]models.CatalogManifest, 0) // non-nil empty (:91)
	manifests, err := g.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}
	needle := helpers.NormalizeString(catalogId)
	for _, manifest := range manifests {
		if strings.EqualFold(manifest.ID, needle) ||
			strings.EqualFold(manifest.Name, needle) ||
			strings.EqualFold(manifest.CatalogId, needle) {
			result = append(result, manifest)
		}
	}
	return result, nil
}

func (g *GormDatabase) GetCatalogManifestsByCatalogIdAndVersion(ctx basecontext.ApiContext, catalogId string, version string) ([]models.CatalogManifest, error) {
	result := make([]models.CatalogManifest, 0)
	manifests, err := g.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}
	needle := helpers.NormalizeString(catalogId)
	for _, manifest := range manifests {
		if (strings.EqualFold(manifest.ID, needle) ||
			strings.EqualFold(manifest.CatalogId, needle) ||
			strings.EqualFold(manifest.Name, needle)) &&
			strings.EqualFold(manifest.Version, version) {
			result = append(result, manifest)
		}
	}
	if len(result) == 0 {
		return nil, interfaces.ErrCatalogManifestNotFound // empty is an error here (:128)
	}
	return result, nil
}

func (g *GormDatabase) GetCatalogManifestsByCatalogIdVersionAndArch(ctx basecontext.ApiContext, catalogId string, version string, arch string) (*models.CatalogManifest, error) {
	manifests, err := g.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}
	needle := helpers.NormalizeString(catalogId)
	for i := range manifests {
		if (strings.EqualFold(manifests[i].ID, needle) ||
			strings.EqualFold(manifests[i].CatalogId, needle) ||
			strings.EqualFold(manifests[i].Name, needle)) &&
			strings.EqualFold(manifests[i].Version, version) &&
			strings.EqualFold(manifests[i].Architecture, arch) {
			return &manifests[i], nil
		}
	}
	return nil, interfaces.ErrCatalogManifestNotFound
}

func (g *GormDatabase) CreateCatalogManifest(ctx basecontext.ApiContext, manifest models.CatalogManifest) (*models.CatalogManifest, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	manifest.ID = helpers.GenerateId() // caller id always discarded (:162)
	if manifest.Version == "" {
		manifest.Version = constants.LATEST_TAG
	}
	manifest.Name = helpers.NormalizeString(manifest.Name)
	manifest.CatalogId = helpers.NormalizeStringUpper(manifest.CatalogId)
	manifest.AddTag(constants.LATEST_TAG)

	// Tag exclusivity: "latest" is stripped from authz-VISIBLE siblings of a
	// different version (:179-189 — the sibling loop rides the wrapped list).
	siblings, err := g.GetCatalogManifestsByCatalogId(ctx, manifest.CatalogId)
	if err != nil {
		return nil, err
	}
	for _, sibling := range siblings {
		if strings.EqualFold(sibling.Version, manifest.Version) {
			continue
		}
		if sibling.HasTag(constants.LATEST_TAG) {
			sibling.RemoveTag(constants.LATEST_TAG)
			if err := g.UpdateCatalogManifestTags(ctx, sibling); err != nil {
				return nil, err
			}
		}
	}

	exists, err := g.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, manifest.CatalogId, manifest.Version, manifest.Architecture)
	if err != nil {
		if errors.GetSystemErrorCode(err) != 404 {
			return nil, err
		}
	}
	if exists != nil { // upsert arm: identity fields inherited from the stored row
		manifest.ID = exists.ID
		manifest.Name = exists.Name
		manifest.CatalogId = exists.CatalogId
		manifest.Architecture = exists.Architecture
		return g.UpdateCatalogManifest(ctx, manifest)
	}

	manifest.CreatedAt = helpers.GetUtcCurrentDateTime()
	manifest.UpdatedAt = helpers.GetUtcCurrentDateTime()
	if err := mapGormError(g.gormDB.Table("catalog_manifests").Create(&manifest).Error,
		interfaces.ErrCatalogManifestNotFound, nil); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (g *GormDatabase) DeleteCatalogManifest(ctx basecontext.ApiContext, catalogIdOrId string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if catalogIdOrId == "" {
		return interfaces.ErrCatalogManifestNotFound
	}
	// Cascade tx #1 of three (D-9 catalog.go:216-320). Legacy loops
	// delete-first-then-rescan until nothing matches: the net effect is
	// "delete every authz-visible FoldEqual match" (:228-253).
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		var rows []models.CatalogManifest
		if err := tx.Table("catalog_manifests").Order("rowid").Find(&rows).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		visible, err := catalogAuthVisible(ctx, rows)
		if err != nil {
			return err
		}
		var ids []string
		for _, m := range visible {
			if strings.EqualFold(m.ID, catalogIdOrId) ||
				strings.EqualFold(m.CatalogId, catalogIdOrId) ||
				strings.EqualFold(m.Name, catalogIdOrId) {
				ids = append(ids, m.ID)
			}
		}
		if len(ids) == 0 {
			return interfaces.ErrCatalogManifestNotFound
		}
		for _, id := range ids {
			if err := tx.Table("catalog_manifests").Where("id = ?", id).Delete(&models.CatalogManifest{}).Error; err != nil {
				return mapGormError(err, interfaces.ErrCatalogManifestNotFound, nil)
			}
		}
		return nil
	})
}

func (g *GormDatabase) DeleteCatalogManifestVersion(ctx basecontext.ApiContext, catalogIdOrId string, version string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if catalogIdOrId == "" {
		return interfaces.ErrCatalogManifestNotFound
	}
	// Cascade tx #2 of three. The UpdatedAt-desc order picks WHICH row:
	// legacy deletes only the FIRST match, returning inside the hit (:285).
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		var rows []models.CatalogManifest
		if err := tx.Table("catalog_manifests").Order("rowid").Find(&rows).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		visible, err := catalogAuthVisible(ctx, rows)
		if err != nil {
			return err
		}
		for _, m := range visible {
			if (strings.EqualFold(m.ID, catalogIdOrId) || strings.EqualFold(m.CatalogId, catalogIdOrId) ||
				strings.EqualFold(m.Name, catalogIdOrId)) && strings.EqualFold(m.Version, version) {
				return mapGormError(tx.Table("catalog_manifests").Where("id = ?", m.ID).
					Delete(&models.CatalogManifest{}).Error, interfaces.ErrCatalogManifestNotFound, nil)
			}
		}
		return interfaces.ErrCatalogManifestNotFound
	})
}

func (g *GormDatabase) DeleteCatalogManifestVersionArch(ctx basecontext.ApiContext, catalogIdOrId string, version string, architecture string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if catalogIdOrId == "" {
		return interfaces.ErrCatalogManifestNotFound
	}
	// Cascade tx #3 of three: the narrowest (catalogId, version, arch) delete.
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		var rows []models.CatalogManifest
		if err := tx.Table("catalog_manifests").Order("rowid").Find(&rows).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		visible, err := catalogAuthVisible(ctx, rows)
		if err != nil {
			return err
		}
		for _, m := range visible {
			if (strings.EqualFold(m.ID, catalogIdOrId) || strings.EqualFold(m.CatalogId, catalogIdOrId) ||
				strings.EqualFold(m.Name, catalogIdOrId)) &&
				strings.EqualFold(m.Version, version) && strings.EqualFold(m.Architecture, architecture) {
				return mapGormError(tx.Table("catalog_manifests").Where("id = ?", m.ID).
					Delete(&models.CatalogManifest{}).Error, interfaces.ErrCatalogManifestNotFound, nil)
			}
		}
		return interfaces.ErrCatalogManifestNotFound
	})
}

func (g *GormDatabase) UpdateCatalogManifest(ctx basecontext.ApiContext, record models.CatalogManifest) (*models.CatalogManifest, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadCatalogManifests() // RAW scan: no authz on this one (:327)
	if err != nil {
		return nil, err
	}
	var found *models.CatalogManifest
	for i := range rows {
		if strings.EqualFold(rows[i].ID, record.ID) {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		return nil, interfaces.ErrCatalogManifestNotFound
	}
	if !strings.EqualFold(found.Version, record.Version) {
		return nil, errors.Newf("cannot update version of catalog manifest %s, it is trying to change the version %s to version %s, not allowed", record.ID, found.Version, record.Version)
	}
	if !strings.EqualFold(found.Architecture, record.Architecture) {
		return nil, errors.Newf("cannot update architecture of catalog manifest %s, it is trying to change the version %s to version %s, not allowed", record.ID, found.Architecture, record.Architecture)
	}
	updates, err := catalogUpdateSet(record)
	if err != nil {
		return nil, err
	}
	if err := g.gormDB.Table("catalog_manifests").Where("id = ?", found.ID).Updates(updates).Error; err != nil {
		return nil, mapGormError(err, interfaces.ErrCatalogManifestNotFound, nil)
	}
	return g.reloadCatalogManifest(found.ID)
}

// catalogUpdateSet is the :336-351 assignment set verbatim: CreatedAt is
// deliberately NOT in the set (legacy re-stamps the OLD value, a no-op) and
// CatalogId/Version/Description/Provider/DownloadCount/IsCompressed/
// PackRelativePath/PackSize/MinimumSpecRequirements/Tainted*/Revoked* are
// never written here.
func catalogUpdateSet(record models.CatalogManifest) (map[string]interface{}, error) {
	set := map[string]interface{}{
		"name": record.Name, "architecture": record.Architecture,
		"updated_at":         helpers.GetUtcCurrentDateTime(),
		"last_downloaded_at": record.LastDownloadedAt, "last_downloaded_user": record.LastDownloadedUser,
		"size": record.Size, "path": record.Path, "metadata_file": record.MetadataFile,
		"pack_file": record.PackFile, "type": record.Type,
	}
	for col, v := range map[string]interface{}{
		"virtual_machine_contents": record.VirtualMachineContents, "pack_contents": record.PackContents,
		"tags": record.Tags, "required_claims": record.RequiredClaims, "required_roles": record.RequiredRoles,
	} {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		set[col] = string(b)
	}
	return set, nil
}

func (g *GormDatabase) reloadCatalogManifest(id string) (*models.CatalogManifest, error) {
	var row models.CatalogManifest
	err := g.gormDB.Table("catalog_manifests").Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, mapGormError(err, interfaces.ErrCatalogManifestNotFound, nil)
	}
	return &row, nil
}

// catalogRawByID loads the raw table and returns the first FoldEqual-ID
// row — the preamble of every mutator that skips the authz wrap.
func (g *GormDatabase) catalogRawByID(recordId string) (*models.CatalogManifest, error) {
	rows, err := g.loadCatalogManifests()
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, recordId) {
			return &rows[i], nil
		}
	}
	return nil, interfaces.ErrCatalogManifestNotFound
}

// catalogSetColumn performs one targeted serializer-column write plus reload
// when the caller must return the stored row (legacy returns &data[i]).
// catalogSetColumn performs one targeted column write.
func (g *GormDatabase) catalogSetColumn(id string, set map[string]interface{}) error {
	if err := g.gormDB.Table("catalog_manifests").Where("id = ?", id).Updates(set).Error; err != nil {
		return mapGormError(err, interfaces.ErrCatalogManifestNotFound, nil)
	}
	return nil
}

func (g *GormDatabase) UpdateCatalogManifestTags(ctx basecontext.ApiContext, record models.CatalogManifest) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	// Broader match than the ID-only mutators: ID OR (CatalogId AND
	// Version), all FoldEqual over the raw list (:366).
	rows, err := g.loadCatalogManifests()
	if err != nil {
		return err
	}
	var found *models.CatalogManifest
	for i := range rows {
		if strings.EqualFold(rows[i].ID, record.ID) ||
			(strings.EqualFold(rows[i].CatalogId, record.CatalogId) && strings.EqualFold(rows[i].Version, record.Version)) {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		return interfaces.ErrCatalogManifestNotFound
	}
	set, err := catalogJSON("tags", record.Tags)
	if err != nil {
		return err
	}
	return g.catalogSetColumn(found.ID, set)
}

func (g *GormDatabase) UpdateCatalogManifestDownloadCount(ctx basecontext.ApiContext, catalogId, version string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	user := catalogUsername(ctx)
	rows, err := g.loadCatalogManifests()
	if err != nil {
		return err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].CatalogId, catalogId) && strings.EqualFold(rows[i].Version, version) {
			// Read-modify-write like legacy (:586): the lost-update race on
			// the counter is legacy behaviour and stays un-invented.
			set := map[string]interface{}{
				"last_downloaded_at": helpers.GetUtcCurrentDateTime(), "last_downloaded_user": user,
				"download_count": rows[i].DownloadCount + 1,
			}
			return g.catalogSetColumn(rows[i].ID, set)
		}
	}
	return interfaces.ErrCatalogManifestNotFound
}

func (g *GormDatabase) catalogFirstByCatalogVersion(catalogId, version string) (*models.CatalogManifest, error) {
	rows, err := g.loadCatalogManifests()
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].CatalogId, catalogId) && strings.EqualFold(rows[i].Version, version) {
			return &rows[i], nil
		}
	}
	return nil, interfaces.ErrCatalogManifestNotFound
}

func (g *GormDatabase) catalogFlagWrite(set map[string]interface{}, catalogId, version string) (*models.CatalogManifest, error) {
	found, err := g.catalogFirstByCatalogVersion(catalogId, version)
	if err != nil {
		return nil, err
	}
	if err := g.catalogSetColumn(found.ID, set); err != nil {
		return nil, err
	}
	return g.reloadCatalogManifest(found.ID)
}

func (g *GormDatabase) TaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	return g.catalogFlagWrite(map[string]interface{}{
		"tainted_at": helpers.GetUtcCurrentDateTime(), "tainted": true, "tainted_by": catalogUsername(ctx),
	}, catalogId, version)
}

func (g *GormDatabase) UnTaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	return g.catalogFlagWrite(map[string]interface{}{
		"tainted_at": "", "tainted": false, "un_tainted_by": catalogUsername(ctx), "tainted_by": "",
	}, catalogId, version)
}

func (g *GormDatabase) RevokeCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	return g.catalogFlagWrite(map[string]interface{}{
		"revoked_at": helpers.GetUtcCurrentDateTime(), "revoked": true, "revoked_by": catalogUsername(ctx),
	}, catalogId, version)
}

func (g *GormDatabase) UpdateCatalogManifestMetadata(ctx basecontext.ApiContext, catalogId, version, architecture string, description *string, tags, claims, roles []string) (*models.CatalogManifest, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadCatalogManifests()
	if err != nil {
		return nil, err
	}
	needle := helpers.NormalizeString(catalogId)
	var found *models.CatalogManifest
	for i := range rows {
		if (strings.EqualFold(rows[i].ID, needle) || strings.EqualFold(rows[i].CatalogId, needle) ||
			strings.EqualFold(rows[i].Name, needle)) &&
			strings.EqualFold(rows[i].Version, version) && strings.EqualFold(rows[i].Architecture, architecture) {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		return nil, interfaces.ErrCatalogManifestNotFound
	}
	set := map[string]interface{}{"updated_at": helpers.GetUtcCurrentDateTime()}
	if description != nil { // nil-means-don't-touch (:680-691); empty non-nil slices DO replace
		set["description"] = *description
	}
	for col, v := range map[string]interface{}{"tags": tags, "required_claims": claims, "required_roles": roles} {
		switch s := v.(type) {
		case []string:
			if s != nil {
				b, err := json.Marshal(s)
				if err != nil {
					return nil, mapGormError(err, nil, nil)
				}
				set[col] = string(b)
			}
		}
	}
	if err := g.catalogSetColumn(found.ID, set); err != nil {
		return nil, err
	}
	return g.reloadCatalogManifest(found.ID)
}

func (g *GormDatabase) UpdateCatalogManifestProvider(ctx basecontext.ApiContext, recordId string, provider models.CatalogManifestProvider) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	found, err := g.catalogRawByID(recordId)
	if err != nil {
		return err
	}
	set, err := catalogJSON("provider", provider)
	if err != nil {
		return err
	}
	return g.catalogSetColumn(found.ID, set)
}
