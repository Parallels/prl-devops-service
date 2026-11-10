package database

// SQL implementation of the claims domain (PBI #421 task 05). Ported
// method-for-method from data/claims.go; cascade semantics (DeleteClaim) and
// the rename cascade (UpdateClaim) follow the legacy in-place mutations.

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

// loadClaims returns claims in insertion order with Users ordered by rowid
// (same pinning rationale as loadRoles).
func (g *GormDatabase) loadClaims(tx *gorm.DB, preloads ...string) ([]models.Claim, error) {
	var claims []models.Claim
	q := tx
	for _, p := range preloads {
		q = q.Preload(p, func(db *gorm.DB) *gorm.DB { return db.Order("rowid") })
	}
	if err := q.Order("rowid").Find(&claims).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return claims, nil
}

func findClaim(claims []models.Claim, idOrName string) (*models.Claim, bool) {
	for i := range claims {
		if strings.EqualFold(claims[i].ID, idOrName) || strings.EqualFold(claims[i].Name, idOrName) {
			return &claims[i], true
		}
	}
	return nil, false
}

func (g *GormDatabase) GetClaims(ctx basecontext.ApiContext, filter string) ([]models.Claim, error) {
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	claims, err := g.loadClaims(g.gormDB, "Users")
	if err != nil {
		return nil, err
	}
	return query.FilterByProperty(claims, dbFilter)
}

func (g *GormDatabase) GetClaim(ctx basecontext.ApiContext, idOrName string) (*models.Claim, error) {
	claims, err := g.loadClaims(g.gormDB, "Users")
	if err != nil {
		return nil, err
	}
	if claim, ok := findClaim(claims, idOrName); ok {
		return claim, nil
	}
	return nil, interfaces.ErrClaimNotFound
}

func (g *GormDatabase) CreateClaim(ctx basecontext.ApiContext, claim models.Claim) (*models.Claim, error) {
	if claim.Name == "" {
		return nil, interfaces.ErrClaimEmptyName
	}

	claim.Name = strings.ToUpper(helpers.NormalizeString(claim.Name))
	claim.ID = claim.Name

	// Non-atomic in legacy (claims.go:104-108); D-9 keeps it so.
	claims, err := g.loadClaims(g.gormDB)
	if err != nil {
		return nil, err
	}
	if existing, ok := findClaim(claims, claim.ID); ok && existing != nil {
		return nil, errors.NewWithCodef(400, "claim %s already exists with ID %s", claim.Name, claim.ID)
	}

	if err := g.gormDB.Create(&claim).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return &claim, nil
}

// DeleteClaim ports claims.go:113-161: empty guard → internal guard →
// user-usage scan → delete claim + cascade user_claims + role_claims, all in
// one transaction (D-9, three tables).
func (g *GormDatabase) DeleteClaim(ctx basecontext.ApiContext, idOrName string) error {
	if idOrName == "" {
		return interfaces.ErrClaimEmptyNameOrId
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		claims, err := g.loadClaims(tx)
		if err != nil {
			return err
		}
		claim, ok := findClaim(claims, idOrName)
		if !ok {
			return interfaces.ErrClaimNotFound
		}
		if claim.Internal && !query.IsRootUser(ctx) {
			return interfaces.ErrRemoveInternalClaim
		}
		var used int64
		if err := tx.Table("user_claims").Where("claim_id = ?", claim.ID).Count(&used).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if used > 0 {
			return interfaces.ErrClaimHasUsers
		}
		if err := tx.Where("id = ?", claim.ID).Delete(&models.Claim{}).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if err := tx.Exec("DELETE FROM user_claims WHERE claim_id = ?", claim.ID).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if err := tx.Where("claim_id = ?", claim.ID).Delete(&models.RoleClaims{}).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		return nil
	})
}

// UpdateClaimMetadata is the guard-free targeted update used by seed
// backfills (claims.go:166-183). Legacy overwrites all four fields
// unconditionally, empty included — map-based update mirrors that.
func (g *GormDatabase) UpdateClaimMetadata(ctx basecontext.ApiContext, idOrName, description, group, resource, action string) error {
	if idOrName == "" {
		return interfaces.ErrClaimEmptyNameOrId
	}
	claims, err := g.loadClaims(g.gormDB)
	if err != nil {
		return err
	}
	claim, ok := findClaim(claims, idOrName)
	if !ok {
		return interfaces.ErrClaimNotFound
	}
	if err := g.gormDB.Model(&models.Claim{}).Where("id = ?", claim.ID).
		Select("description", "group", "resource", "action").
		Updates(map[string]interface{}{
			"description": description, "group": group, "resource": resource, "action": action,
		}).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	return nil
}

// UpdateClaim ports claims.go:185-226 with its match quirk (ID matches ID,
// else NAME matches NAME), the RENAME side effect (ID is recomputed from the
// input Name), and the in-place replacement + cascade: legacy rewrites every
// stored copy; SQL rewrites the row and the two join tables' references.
// Non-transactional per D-9's inventory (legacy rename is single-lock but the
// D-9 table does not list it as a tx flow).
func (g *GormDatabase) UpdateClaim(ctx basecontext.ApiContext, claim *models.Claim) (*models.Claim, error) {
	if claim.ID == "" {
		return nil, interfaces.ErrClaimEmptyNameOrId
	}
	claims, err := g.loadClaims(g.gormDB)
	if err != nil {
		return nil, err
	}
	found, ok := func() (*models.Claim, bool) {
		for i := range claims {
			if strings.EqualFold(claims[i].ID, claim.ID) || strings.EqualFold(claims[i].Name, claim.Name) {
				return &claims[i], true
			}
		}
		return nil, false
	}()
	if !ok {
		return nil, interfaces.ErrClaimNotFound
	}
	if found.Internal && !query.IsRootUser(ctx) {
		return nil, interfaces.ErrUpdateInternalClaim
	}

	oldID := found.ID
	claim.ID = strings.ToUpper(helpers.NormalizeString(claim.Name))

	// Single in-place UPDATE (id column included): sqlite keeps the rowid, so
	// a rename preserves list POSITION — legacy replaced the slice entry in
	// place, and rowid order is our position contract (D-4). Join tables get
	// the reference rewrite (legacy replaced the embedded copies).
	// Divergence note: renaming onto an existing ID collides here (500),
	// while legacy happily stored a duplicate ID — corruption-in-legacy is
	// not a behaviour to port.
	if err := g.gormDB.Table("claims").Where("id = ?", oldID).
		Updates(map[string]interface{}{
			"id": claim.ID, "name": claim.Name, "description": claim.Description,
			"internal": claim.Internal, "group": claim.Group, "resource": claim.Resource,
			"action": claim.Action,
		}).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	if oldID != claim.ID {
		if err := g.gormDB.Exec("UPDATE user_claims SET claim_id = ? WHERE claim_id = ?", claim.ID, oldID).Error; err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		if err := g.gormDB.Exec("UPDATE role_claims SET claim_id = ? WHERE claim_id = ?", claim.ID, oldID).Error; err != nil {
			return nil, mapGormError(err, nil, nil)
		}
	}
	return claim, nil
}
