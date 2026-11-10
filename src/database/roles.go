package database

// SQL implementation of the roles domain (PBI #421 task 05). Semantics are
// ported method-for-method from data/roles.go, including guard order and the
// legacy match quirks. Filters run Go-side through database/query (D-4
// evaluation-first); row order is insertion order (sqlite rowid), matching
// the legacy append-ordered slices — no ORDER BY semantics beyond that.

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

// loadRoles returns roles in insertion order. Association order is pinned to
// rowid so join-loaded Claims/Users come back in the same order the legacy
// engine built them (append order).
func (g *GormDatabase) loadRoles(tx *gorm.DB, preloads ...string) ([]models.Role, error) {
	var roles []models.Role
	q := tx
	for _, p := range preloads {
		q = q.Preload(p, func(db *gorm.DB) *gorm.DB { return db.Order("rowid") })
	}
	if err := q.Order("rowid").Find(&roles).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return roles, nil
}

// findRole mirrors the legacy EqualFold ID-or-Name scan over the full list,
// first match wins.
func findRole(roles []models.Role, idOrName string) (*models.Role, bool) {
	for i := range roles {
		if strings.EqualFold(roles[i].ID, idOrName) || strings.EqualFold(roles[i].Name, idOrName) {
			return &roles[i], true
		}
	}
	return nil, false
}

func (g *GormDatabase) GetRoles(ctx basecontext.ApiContext, filter string) ([]models.Role, error) {
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	roles, err := g.loadRoles(g.gormDB, "Claims", "Users")
	if err != nil {
		return nil, err
	}
	return query.FilterByProperty(roles, dbFilter)
}

func (g *GormDatabase) GetRole(ctx basecontext.ApiContext, idOrName string) (*models.Role, error) {
	roles, err := g.loadRoles(g.gormDB, "Claims", "Users")
	if err != nil {
		return nil, err
	}
	if role, ok := findRole(roles, idOrName); ok {
		return role, nil
	}
	return nil, interfaces.ErrRoleNotFound
}

func (g *GormDatabase) CreateRole(ctx basecontext.ApiContext, role models.Role) (*models.Role, error) {
	if role.Name == "" {
		return nil, interfaces.ErrRoleEmptyName
	}

	role.Name = strings.ToUpper(helpers.NormalizeString(role.Name))
	role.ID = role.Name

	roles, err := g.loadRoles(g.gormDB)
	if err != nil {
		return nil, err
	}
	if existing, ok := findRole(roles, role.ID); ok && existing != nil {
		return nil, errors.NewWithCodef(400, "role %s already exists with ID %s", role.Name, role.ID)
	}

	// Resolve any claims provided against the DB to ensure they exist
	// (legacy roles.go:109-120, error text pinned).
	if len(role.Claims) > 0 {
		resolved := make([]models.Claim, 0, len(role.Claims))
		for _, c := range role.Claims {
			dbClaim, err := g.GetClaim(ctx, c.ID)
			if err != nil {
				return nil, errors.NewWithCodef(400, "claim %s not found", c.ID)
			}
			resolved = append(resolved, *dbClaim)
		}
		role.Claims = resolved
	}

	if err := g.gormDB.Create(&role).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return &role, nil
}

// AddClaimToRole ports roles.go:127-153 INCLUDING the check-order quirk: the
// claim is resolved first, so a garbage role id with a missing claim reports
// ErrClaimNotFound, not ErrRoleNotFound.
func (g *GormDatabase) AddClaimToRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error {
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		claims, err := g.loadClaims(tx)
		if err != nil {
			return err
		}
		dbClaim, ok := findClaim(claims, claimIdOrName)
		if !ok {
			return interfaces.ErrClaimNotFound
		}

		roles, err := g.loadRoles(tx, "Claims")
		if err != nil {
			return err
		}
		role, ok := findRole(roles, roleIdOrName)
		if !ok {
			return interfaces.ErrRoleNotFound
		}
		if role.Internal && !query.IsRootUser(ctx) {
			return interfaces.ErrUpdateInternalRole
		}
		for _, existing := range role.Claims {
			if strings.EqualFold(existing.ID, dbClaim.ID) {
				return interfaces.ErrRoleAlreadyContainsClaim
			}
		}
		if err := tx.Create(&models.RoleClaims{RoleID: role.ID, ClaimID: dbClaim.ID}).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		return nil
	})
}

// RemoveClaimFromRole ports roles.go:155-176; the claim side matches by ID
// OR Name (roles.go:166 EqualFold both).
func (g *GormDatabase) RemoveClaimFromRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error {
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		roles, err := g.loadRoles(tx, "Claims")
		if err != nil {
			return err
		}
		role, ok := findRole(roles, roleIdOrName)
		if !ok {
			return interfaces.ErrRoleNotFound
		}
		if role.Internal && !query.IsRootUser(ctx) {
			return interfaces.ErrUpdateInternalRole
		}
		for _, existing := range role.Claims {
			if strings.EqualFold(existing.ID, claimIdOrName) || strings.EqualFold(existing.Name, claimIdOrName) {
				if err := tx.Where("role_id = ? AND claim_id = ?", role.ID, existing.ID).
					Delete(&models.RoleClaims{}).Error; err != nil {
					return mapGormError(err, nil, nil)
				}
				return nil
			}
		}
		return interfaces.ErrRoleDoesNotContainClaim
	})
}

// DeleteRole ports roles.go:178-207: empty guard → internal guard →
// user-usage scan → delete (+ role_claims) in one transaction.
func (g *GormDatabase) DeleteRole(ctx basecontext.ApiContext, idOrName string) error {
	if idOrName == "" {
		return interfaces.ErrRoleEmptyNameOrId
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		roles, err := g.loadRoles(tx)
		if err != nil {
			return err
		}
		role, ok := findRole(roles, idOrName)
		if !ok {
			return interfaces.ErrRoleNotFound
		}
		if role.Internal && !query.IsRootUser(ctx) {
			return interfaces.ErrRemoveInternalRole
		}
		var used int64
		if err := tx.Table("user_roles").Where("role_id = ?", role.ID).Count(&used).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if used > 0 {
			return interfaces.ErrRoleHasUsers
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&models.RoleClaims{}).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if err := tx.Where("id = ?", role.ID).Delete(&models.Role{}).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		return nil
	})
}

// UpdateRoleInternalFlag is the guard-free targeted update used by seeds
// (roles.go:211-225).
func (g *GormDatabase) UpdateRoleInternalFlag(ctx basecontext.ApiContext, idOrName string, internal bool) error {
	if idOrName == "" {
		return interfaces.ErrRoleEmptyNameOrId
	}
	roles, err := g.loadRoles(g.gormDB)
	if err != nil {
		return err
	}
	role, ok := findRole(roles, idOrName)
	if !ok {
		return interfaces.ErrRoleNotFound
	}
	if err := g.gormDB.Model(&models.Role{}).Where("id = ?", role.ID).
		Select("internal").Updates(map[string]interface{}{"internal": internal}).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	return nil
}

// UpdateRoleDescription is the guard-free targeted update used by backfills
// (roles.go:229-243).
func (g *GormDatabase) UpdateRoleDescription(ctx basecontext.ApiContext, idOrName, description string) error {
	if idOrName == "" {
		return interfaces.ErrRoleEmptyNameOrId
	}
	roles, err := g.loadRoles(g.gormDB)
	if err != nil {
		return err
	}
	role, ok := findRole(roles, idOrName)
	if !ok {
		return interfaces.ErrRoleNotFound
	}
	if err := g.gormDB.Model(&models.Role{}).Where("id = ?", role.ID).
		Select("description").Updates(map[string]interface{}{"description": description}).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	return nil
}

// UpdateRole ports roles.go:245-267 with its match quirk (ID matches ID, else
// NAME matches NAME) and whole-object replacement: legacy does
// j.data.Roles[i] = *role, so the row is replaced wholesale — including a
// possible ID change — not merged.
func (g *GormDatabase) UpdateRole(ctx basecontext.ApiContext, role *models.Role) (*models.Role, error) {
	if role.ID == "" {
		return nil, interfaces.ErrRoleEmptyNameOrId
	}
	roles, err := g.loadRoles(g.gormDB)
	if err != nil {
		return nil, err
	}
	found, ok := func() (*models.Role, bool) {
		for i := range roles {
			if strings.EqualFold(roles[i].ID, role.ID) || strings.EqualFold(roles[i].Name, role.Name) {
				return &roles[i], true
			}
		}
		return nil, false
	}()
	if !ok {
		return nil, interfaces.ErrRoleNotFound
	}
	if found.Internal && !query.IsRootUser(ctx) {
		return nil, interfaces.ErrUpdateInternalRole
	}

	// Whole-object replacement (non-atomic in legacy; D-9 keeps it so), done
	// as one in-place UPDATE (id column included) so the rowid — and with it
	// the list position — survives even an ID change, mirroring legacy's
	// roles[i] = *role slice semantics. Join rows are rewritten from the
	// input claim list.
	if err := g.gormDB.Table("roles").Where("id = ?", found.ID).
		Updates(map[string]interface{}{
			"id": role.ID, "name": role.Name,
			"description": role.Description, "internal": role.Internal,
		}).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	if err := g.gormDB.Where("role_id = ?", found.ID).Delete(&models.RoleClaims{}).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	for _, c := range role.Claims {
		if err := g.gormDB.Create(&models.RoleClaims{RoleID: role.ID, ClaimID: c.ID}).Error; err != nil {
			return nil, mapGormError(err, nil, nil)
		}
	}
	return role, nil
}
