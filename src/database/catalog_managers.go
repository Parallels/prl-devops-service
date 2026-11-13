package database

// SQL implementation of the catalog-managers domain (PBI #421 task 11,
// legacy src/data/catalog_managers.go). Quirks mirrored:
//   - GetCatalogManagers/GetCatalogManager carry NO ctx and NO connection
//     guard (legacy bodies only take the mutex) — mirrored, the SQL handles
//     answer from the pool regardless.
//   - Lookup is EXACT id equality (mgr.ID == id, :28 — no EqualFold; the
//     spec's "EqualFold fallback" line is stale).
//   - Add has NO duplicate pre-check (:36-48 appends blindly): legacy would
//     store the same id twice; SQL's primary key refuses — STRICTER,
//     deviation documented and pinned by test.
//   - Update is a WHOLE-ROW replacement (j.data[i] = mgr, :60 — payload
//     wins every field incl. Internal/Active/OwnerID) but is blocked for
//     internal rows for EVERYONE (:56-58, root included).
//   - Delete blocks internal rows only for non-root (:80-82 — the one
//     guard in the file with an IsRootUser escape).

import (
	"encoding/json"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
)

func (g *GormDatabase) GetCatalogManagers() ([]models.CatalogManager, error) {
	var rows []models.CatalogManager
	if err := g.gormDB.Table("catalog_managers").Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

func (g *GormDatabase) GetCatalogManager(id string) (*models.CatalogManager, error) {
	var rows []models.CatalogManager
	if err := g.gormDB.Table("catalog_managers").Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	for i := range rows {
		if rows[i].ID == id { // exact equality, first hit (:27-31)
			return &rows[i], nil
		}
	}
	return nil, interfaces.ErrCatalogManagerNotFound
}

func (g *GormDatabase) AddCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	// No dup pre-check mirrored from legacy — the primary key is the only
	// gate, so a repeated id surfaces as a mapped error (deviation: legacy
	// would double-store).
	return mapGormError(g.gormDB.Table("catalog_managers").Create(&mgr).Error,
		nil, nil)
}

func (g *GormDatabase) UpdateCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	found, err := g.GetCatalogManager(mgr.ID)
	if err != nil {
		return err
	}
	if found.Internal {
		return interfaces.ErrUpdateInternalCatalogManager // everyone, root included
	}
	set := map[string]interface{}{
		"id": mgr.ID, "name": mgr.Name, "url": mgr.URL,
		"internal": mgr.Internal, "active": mgr.Active, "global": mgr.Global,
		"authentication_method": mgr.AuthenticationMethod,
		"username":              mgr.Username, "password": mgr.Password, "api_key": mgr.ApiKey,
		"owner_id": mgr.OwnerID, "created_at": mgr.CreatedAt, "updated_at": mgr.UpdatedAt,
	}
	b, err := json.Marshal(mgr.RequiredClaims)
	if err != nil {
		return mapGormError(err, nil, nil)
	}
	set["required_claims"] = string(b)
	return mapGormError(g.gormDB.Table("catalog_managers").Where("id = ?", found.ID).Updates(set).Error,
		nil, nil)
}

func (g *GormDatabase) DeleteCatalogManager(ctx basecontext.ApiContext, id string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	found, err := g.GetCatalogManager(id)
	if err != nil {
		return err
	}
	if found.Internal && !query.IsRootUser(ctx) {
		return interfaces.ErrRemoveInternalCatalogManager
	}
	return mapGormError(g.gormDB.Table("catalog_managers").Where("id = ?", found.ID).
		Delete(&models.CatalogManager{}).Error, nil, nil)
}
