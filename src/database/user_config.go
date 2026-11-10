package database

// SQL implementation of the user-configs domain (PBI #421 task 09).
// Legacy data/user_config.go is owner-scoped CRUD with DUAL addressing (id
// OR slug, all comparisons strings.EqualFold — owner match included) and NO
// authz layer: ctx is accepted and ignored on every path (no
// GetAuthorizedRecords, no OwnRecords — do not add either, per spec: they
// would change admin-by-userID behaviour). The userID parameter is the scope
// and becomes a SQL owner predicate (D-4: rows loaded owner-filtered,
// rowid-pinned, then scanned Go-side with the same EqualFold semantics).
// 0 tx sites — D-9 lists no user-configs row; UpsertUserConfig is
// deliberately TWO statements: legacy is non-atomic here and parity keeps it
// (race window: concurrent upsert of the same owner+slug can double-insert;
// documented, not silently tightened — the (user_id,slug) unique index from
// #419 backstops exact-case collisions as AlreadyExists).
// Spec deviation: the spec claims an empty-input sentinel before lookup
// (:42-56); legacy has NONE — an empty idOrSlug matches a same-owner row
// with empty ID/slug through EqualFold. Legacy mirrored, not "fixed".

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"github.com/Parallels/prl-devops-service/helpers"
)

// loadUserConfigs pins storage order (rowid) like every legacy slice scan.
func (g *GormDatabase) loadUserConfigs() ([]models.UserConfig, error) {
	var rows []models.UserConfig
	err := g.gormDB.Table("user_configs").Order("rowid").Find(&rows).Error
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

// scanUserConfig is the shared owner + (id||slug) EqualFold matcher of
// user_config.go:47-52/:129-131 — first match wins. Also serves the
// CreateUserConfig duplicate probe (:66), which asks the same matcher with
// the incoming SLUG as the id-or-slug needle (this is what lets an existing
// row's ID collide with a new slug — legacy quirk, mirrored).
func scanUserConfig(rows []models.UserConfig, userID, idOrSlug string) *models.UserConfig {
	for i := range rows {
		if strings.EqualFold(rows[i].UserID, userID) &&
			(strings.EqualFold(rows[i].ID, idOrSlug) || strings.EqualFold(rows[i].Slug, idOrSlug)) {
			return &rows[i]
		}
	}
	return nil
}

func (g *GormDatabase) GetUserConfigs(ctx basecontext.ApiContext, userID string, filter string) ([]models.UserConfig, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	rows, err := g.loadUserConfigs()
	if err != nil {
		return nil, err
	}
	filtered := make([]models.UserConfig, 0) // legacy :27 — nil-vs-empty shape mirrored
	for _, cfg := range rows {
		if strings.EqualFold(cfg.UserID, userID) {
			filtered = append(filtered, cfg)
		}
	}
	return query.FilterByProperty(filtered, dbFilter)
}

func (g *GormDatabase) GetUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) (*models.UserConfig, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadUserConfigs()
	if err != nil {
		return nil, err
	}
	if cfg := scanUserConfig(rows, userID, idOrSlug); cfg != nil {
		return cfg, nil
	}
	return nil, interfaces.ErrUserConfigNotFound
}

// CreateUserConfig mirrors :57-77 including order: generate id when empty
// BEFORE the duplicate probe, probe via the same id||slug matcher with the
// slug as needle, stamp both timestamps, insert.
func (g *GormDatabase) CreateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if cfg.ID == "" {
		cfg.ID = helpers.GenerateId() // legacy :63 — same helper
	}
	rows, err := g.loadUserConfigs()
	if err != nil {
		return nil, err
	}
	if scanUserConfig(rows, cfg.UserID, cfg.Slug) != nil {
		return nil, interfaces.ErrUserConfigAlreadyExists
	}
	cfg.CreatedAt = helpers.GetUtcCurrentDateTime()
	cfg.UpdatedAt = helpers.GetUtcCurrentDateTime()
	dup := interfaces.ErrUserConfigAlreadyExists
	notFound := interfaces.ErrUserConfigNotFound
	if err := mapGormError(g.gormDB.Table("user_configs").Create(&cfg).Error, notFound, dup); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpsertUserConfig mirrors :79-91: slug probe → create branch (reusing the
// full CreateUserConfig semantics) or update branch pinning the EXISTING id.
// Two statements, no transaction (D-9; legacy parity).
func (g *GormDatabase) UpsertUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadUserConfigs()
	if err != nil {
		return nil, err
	}
	existing := scanUserConfig(rows, cfg.UserID, cfg.Slug)
	if existing == nil {
		return g.CreateUserConfig(ctx, cfg)
	}
	cfg.ID = existing.ID
	return g.UpdateUserConfig(ctx, cfg)
}

// UpdateUserConfig mirrors :93-122: match on id AND user (EqualFold), copy
// ONLY Name/Type/Value + UpdatedAt — Slug, CreatedAt and UserID of the
// payload are ignored by legacy and stay untouched here (partial-update
// drift avoided; GORM Save is deliberately not used).
func (g *GormDatabase) UpdateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadUserConfigs()
	if err != nil {
		return nil, err
	}
	var found *models.UserConfig
	for i := range rows {
		if strings.EqualFold(rows[i].ID, cfg.ID) && strings.EqualFold(rows[i].UserID, cfg.UserID) {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		return nil, interfaces.ErrUserConfigNotFound
	}
	updates := map[string]interface{}{
		"name":       cfg.Name,
		"type":       cfg.Type,
		"value":      cfg.Value,
		"updated_at": helpers.GetUtcCurrentDateTime(),
	}
	if err := g.gormDB.Table("user_configs").Where("id = ?", found.ID).Updates(updates).Error; err != nil {
		return nil, mapGormError(err, interfaces.ErrUserConfigNotFound, interfaces.ErrUserConfigAlreadyExists)
	}
	found.Name = cfg.Name
	found.Type = cfg.Type
	found.Value = cfg.Value
	found.UpdatedAt, _ = updates["updated_at"].(string)
	return found, nil
}

func (g *GormDatabase) DeleteUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadUserConfigs()
	if err != nil {
		return err
	}
	found := scanUserConfig(rows, userID, idOrSlug)
	if found == nil {
		return interfaces.ErrUserConfigNotFound
	}
	// Legacy splices the first match out (:137); single-row delete by the
	// exact primary key of that scan hit.
	return mapGormError(g.gormDB.Table("user_configs").Where("id = ?", found.ID).Delete(&models.UserConfig{}).Error,
		interfaces.ErrUserConfigNotFound, nil)
}
