package database

// SQL implementation of the users domain (PBI #421 task 06). Ported
// method-for-method from data/users.go including its quirks: filters apply to
// the stored shape, GetUser resolves ID/Email/Username cross-field EqualFold,
// AddRole/ClaimToUser mix an EqualFold resolve with an EXACT-match append gate
// (the case-skew ErrUserNotFound / ErrClaimNotFound fall-throughs), DeleteUser
// silently no-ops on empty id and reports root deletion with the UPDATE
// sentinel. Legacy's read-time rehydration of Roles[].Claims (users.go:44-56)
// is intrinsic here: the mandatory preloads always read live data (#419 D-7).

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/security/password"
	"gorm.io/gorm"
)

// loadUsers is THE user read shape (spec: every read path preloads Claims and
// Roles.Claims — ComputeEffectiveClaims downstream depends on it). All orders
// pinned to rowid = legacy append order.
func (g *GormDatabase) loadUsers(tx *gorm.DB) ([]models.User, error) {
	var users []models.User
	err := tx.
		Preload("Claims", func(db *gorm.DB) *gorm.DB { return db.Order("rowid") }).
		Preload("Roles", func(db *gorm.DB) *gorm.DB { return db.Order("rowid") }).
		Preload("Roles.Claims", func(db *gorm.DB) *gorm.DB { return db.Order("rowid") }).
		Order("rowid").Find(&users).Error
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return users, nil
}

// findUser mirrors GetUser's scan: one string resolves across ID / Email /
// Username (EqualFold), first user wins.
func findUser(users []models.User, idOrEmail string) (*models.User, bool) {
	for i := range users {
		if strings.EqualFold(users[i].ID, idOrEmail) ||
			strings.EqualFold(users[i].Email, idOrEmail) ||
			strings.EqualFold(users[i].Username, idOrEmail) {
			return &users[i], true
		}
	}
	return nil, false
}

func isRootShaped(u *models.User) bool {
	return u.Name == "root" || u.Username == "root" || u.Email == "root@localhost"
}

func (g *GormDatabase) GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	users, err := g.loadUsers(g.gormDB)
	if err != nil {
		return nil, err
	}
	return query.FilterByProperty(users, dbFilter)
}

func (g *GormDatabase) GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	users, err := g.loadUsers(g.gormDB)
	if err != nil {
		return nil, err
	}
	if user, ok := findUser(users, idOrEmail); ok {
		return user, nil
	}
	return nil, interfaces.ErrUserNotFound
}

// CreateUser is D-9 row 1 — the chain's biggest transaction: 3 duplicate
// probes, role/claim validation or default materialisation, and the user +
// user_roles + user_claims inserts run against one snapshot. The bcrypt work
// happens BEFORE the tx opens (spec: keep CPU out of the lock).
func (g *GormDatabase) CreateUser(ctx basecontext.ApiContext, user models.User) (*models.User, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}

	if user.ID == "" {
		user.ID = helpers.GenerateId()
	}
	if user.Email == "" {
		return nil, interfaces.ErrUserEmailCannotBeEmpty
	}
	if user.Username == "" {
		return nil, interfaces.ErrUserUsernameCannotBeEmpty
	}
	if user.Name == "" {
		return nil, interfaces.ErrUserNameCannotBeEmpty
	}

	hashedPassword, err := password.Get().Hash(user.Password, user.ID)
	if err != nil {
		return nil, err
	}

	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		users, err := g.loadUsers(tx)
		if err != nil {
			return err
		}
		for _, probe := range []string{user.ID, user.Email, user.Username} {
			if _, found := findUser(users, probe); found {
				return interfaces.ErrUserAlreadyExists
			}
		}

		if len(user.Roles) == 0 {
			for _, role := range constants.DefaultRoles {
				roles, err := g.loadRoles(tx, "Claims", "Users")
				if err != nil {
					return err
				}
				dbRole, ok := findRole(roles, role)
				if !ok {
					return interfaces.ErrRoleNotFound
				}
				user.Roles = append(user.Roles, *dbRole)
			}
		} else {
			for _, role := range user.Roles {
				roles, err := g.loadRoles(tx)
				if err != nil {
					return err
				}
				if _, ok := findRole(roles, role.ID); !ok {
					return errors.NewWithCodef(400, "role %s does not exist", role.Name)
				}
			}
		}

		if len(user.Claims) == 0 {
			for _, claim := range constants.DefaultClaims {
				claims, err := g.loadClaims(tx, "Users")
				if err != nil {
					return err
				}
				dbClaim, ok := findClaim(claims, claim)
				if !ok {
					return interfaces.ErrClaimNotFound
				}
				user.Claims = append(user.Claims, *dbClaim)
			}
		} else {
			for _, claim := range user.Claims {
				claims, err := g.loadClaims(tx)
				if err != nil {
					return err
				}
				if _, ok := findClaim(claims, claim.ID); !ok {
					return errors.NewWithCodef(400, "claim %s does not exist", claim.Name)
				}
			}
		}

		user.Password = hashedPassword
		user.UpdatedAt = helpers.GetUtcCurrentDateTime()
		user.CreatedAt = helpers.GetUtcCurrentDateTime()

		// Associations carry primary keys → GORM writes join rows only.
		if err := tx.Create(&user).Error; err != nil {
			return mapGormError(err, nil, interfaces.ErrUserAlreadyExists)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser ports users.go:160-201: root guard keyed on the INPUT shape
// (not the stored row), exact-match ID scan, and non-empty-field assignment —
// Username and the block fields are NOT updatable here (legacy ignores them).
// Non-transactional per D-9; the Roles/Claims join replacement therefore runs
// statement-sequential (legacy held one lock; sqlite single conn — ledger
// deviation note).
func (g *GormDatabase) UpdateUser(ctx basecontext.ApiContext, key models.User) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}

	if key.Name == "root" || key.Username == "root" || key.Email == "root@localhost" {
		if ctx == nil || ctx.GetAuthorizationContext() == nil || ctx.GetAuthorizationContext().AuthorizedBy != "RootAuthorization" {
			return interfaces.ErrCannotUpdateRootUser
		}
	}

	users, err := g.loadUsers(g.gormDB)
	if err != nil {
		return err
	}
	var found *models.User
	for i := range users {
		if users[i].ID == key.ID { // EXACT match, legacy :171
			found = &users[i]
			break
		}
	}
	if found == nil {
		return interfaces.ErrUserNotFound
	}

	fields := map[string]interface{}{"updated_at": helpers.GetUtcCurrentDateTime()}
	if key.Name != "" {
		fields["name"] = key.Name
	}
	if key.Password != "" {
		hashedPassword, err := password.Get().Hash(key.Password, key.ID)
		if err != nil {
			return err
		}
		fields["password"] = hashedPassword
	}
	if key.Email != "" {
		fields["email"] = key.Email
	}
	if err := g.gormDB.Table("users").Where("id = ?", found.ID).Updates(fields).Error; err != nil {
		return mapGormError(err, nil, nil)
	}

	if len(key.Roles) > 0 {
		if err := g.gormDB.Exec("DELETE FROM user_roles WHERE user_id = ?", found.ID).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		for _, r := range key.Roles {
			if err := g.gormDB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", found.ID, r.ID).Error; err != nil {
				return mapGormError(err, nil, nil)
			}
		}
	}
	if len(key.Claims) > 0 {
		if err := g.gormDB.Exec("DELETE FROM user_claims WHERE user_id = ?", found.ID).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		for _, c := range key.Claims {
			if err := g.gormDB.Exec("INSERT INTO user_claims (user_id, claim_id) VALUES (?, ?)", found.ID, c.ID).Error; err != nil {
				return mapGormError(err, nil, nil)
			}
		}
	}
	return nil
}

// UpdateUserBlockStatus ports users.go:203-221: exact-match ID scan, block
// fields written unconditionally (zero values included).
func (g *GormDatabase) UpdateUserBlockStatus(ctx basecontext.ApiContext, key models.User) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	users, err := g.loadUsers(g.gormDB)
	if err != nil {
		return err
	}
	var found *models.User
	for i := range users {
		if users[i].ID == key.ID {
			found = &users[i]
			break
		}
	}
	if found == nil {
		return interfaces.ErrUserNotFound
	}
	if err := g.gormDB.Table("users").Where("id = ?", found.ID).
		Updates(map[string]interface{}{
			"blocked":               key.Blocked,
			"blocked_since":         key.BlockedSince,
			"blocked_reason":        key.BlockedReason,
			"failed_login_attempts": key.FailedLoginAttempts,
			"updated_at":            helpers.GetUtcCurrentDateTime(),
		}).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	return nil
}

// UpdateRootPassword ports users.go:223-243: first row whose Email is exactly
// "root@localhost"; re-hash with the same (pw, userID) salt order as create.
func (g *GormDatabase) UpdateRootPassword(ctx basecontext.ApiContext, newPassword string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	users, err := g.loadUsers(g.gormDB)
	if err != nil {
		return err
	}
	var found *models.User
	for i := range users {
		if users[i].Email == "root@localhost" {
			found = &users[i]
			break
		}
	}
	if found == nil {
		return interfaces.ErrUserNotFound
	}
	hashedPassword, err := password.Get().Hash(newPassword, found.ID)
	if err != nil {
		return err
	}
	if err := g.gormDB.Table("users").Where("id = ?", found.ID).
		Updates(map[string]interface{}{
			"password": hashedPassword, "updated_at": helpers.GetUtcCurrentDateTime(),
		}).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	return nil
}

// DeleteUser ports users.go:245-267: empty id is a SILENT no-op; EqualFold
// cross-field resolve; root deletion answers with ErrCannotUpdateRootUser
// (legacy reuses the update sentinel — kept). Unlike legacy's embedded model,
// a row-based delete must purge user_roles/user_claims too (orphan join rows
// would outlive the user) — that is ≥2 tables, so per the task-06 spec clause
// this gets one transaction and the D-9 table gains a line (recorded in the
// ledger; the "exactly 5" grep pre-dates that clause — see deviations).
func (g *GormDatabase) DeleteUser(ctx basecontext.ApiContext, id string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if id == "" {
		return nil
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		users, err := g.loadUsers(tx)
		if err != nil {
			return err
		}
		found, ok := findUser(users, id)
		if !ok {
			return interfaces.ErrUserNotFound
		}
		if isRootShaped(found) {
			return interfaces.ErrCannotUpdateRootUser
		}
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", found.ID).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if err := tx.Exec("DELETE FROM user_claims WHERE user_id = ?", found.ID).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if err := tx.Where("id = ?", found.ID).Delete(&models.User{}).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		return nil
	})
}

// AddRoleToUser — D-9 row 2. Quirk kept: membership dedupe runs on the
// EqualFold-resolved user, but the append targets the row whose ID matches
// the CALLER's casing exactly (users.go:298-306) — a case-skewed userId that
// resolves but never appends.
func (g *GormDatabase) AddRoleToUser(ctx basecontext.ApiContext, userId string, roleId string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if userId == "" {
		return interfaces.ErrUserIDCannotBeEmpty
	}
	if roleId == "" {
		return interfaces.ErrRoleIDCannotBeEmpty
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		users, err := g.loadUsers(tx)
		if err != nil {
			return err
		}
		user, ok := findUser(users, userId)
		if !ok {
			return interfaces.ErrUserNotFound
		}
		roles, err := g.loadRoles(tx)
		if err != nil {
			return err
		}
		role, ok := findRole(roles, roleId)
		if !ok {
			return interfaces.ErrRoleNotFound
		}
		for _, r := range user.Roles {
			if r.ID == role.ID { // EXACT dedupe, legacy :293
				return interfaces.ErrUserAlreadyContainsRole
			}
		}
		if user.ID != userId { // exact-append gate, legacy :299
			return interfaces.ErrUserNotFound
		}
		if err := tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID).Error; err != nil {
			return mapGormError(err, nil, interfaces.ErrUserAlreadyContainsRole)
		}
		return nil
	})
}

// RemoveRoleFromUser ports users.go:309-345 including the ErrRoleNotFound
// flavour of every failure mode (missing join row OR case-skewed id).
func (g *GormDatabase) RemoveRoleFromUser(ctx basecontext.ApiContext, userId string, roleId string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if userId == "" {
		return interfaces.ErrUserIDCannotBeEmpty
	}
	if roleId == "" {
		return interfaces.ErrRoleIDCannotBeEmpty
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		users, err := g.loadUsers(tx)
		if err != nil {
			return err
		}
		user, ok := findUser(users, userId)
		if !ok {
			return interfaces.ErrUserNotFound
		}
		roles, err := g.loadRoles(tx)
		if err != nil {
			return err
		}
		role, ok := findRole(roles, roleId)
		if !ok {
			return interfaces.ErrRoleNotFound
		}
		if user.ID == userId { // exact-append gate, legacy :333
			for _, r := range user.Roles {
				if r.ID == role.ID {
					if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", user.ID, role.ID).Error; err != nil {
						return mapGormError(err, nil, nil)
					}
					return nil
				}
			}
		}
		return interfaces.ErrRoleNotFound
	})
}

// AddClaimToUser mirrors AddRoleToUser on user_claims; the case-skew
// fall-through answers with ErrClaimNotFound (legacy :384 — kept).
func (g *GormDatabase) AddClaimToUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if userId == "" {
		return interfaces.ErrUserIDCannotBeEmpty
	}
	if claimId == "" {
		return interfaces.ErrClaimIDCannotBeEmpty
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		users, err := g.loadUsers(tx)
		if err != nil {
			return err
		}
		user, ok := findUser(users, userId)
		if !ok {
			return interfaces.ErrUserNotFound
		}
		claims, err := g.loadClaims(tx)
		if err != nil {
			return err
		}
		claim, ok := findClaim(claims, claimId)
		if !ok {
			return interfaces.ErrClaimNotFound
		}
		for _, c := range user.Claims {
			if c.ID == claim.ID { // EXACT dedupe, legacy :371
				return interfaces.ErrUserAlreadyContainsClaim
			}
		}
		if user.ID != userId { // exact-append gate, legacy :377
			return interfaces.ErrClaimNotFound
		}
		if err := tx.Exec("INSERT INTO user_claims (user_id, claim_id) VALUES (?, ?)", user.ID, claim.ID).Error; err != nil {
			return mapGormError(err, nil, interfaces.ErrUserAlreadyContainsClaim)
		}
		return nil
	})
}

// RemoveClaimFromUser mirrors RemoveRoleFromUser on user_claims; every
// failure mode answers with ErrClaimNotFound (legacy :422 — kept).
func (g *GormDatabase) RemoveClaimFromUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if userId == "" {
		return interfaces.ErrUserIDCannotBeEmpty
	}
	if claimId == "" {
		return interfaces.ErrClaimIDCannotBeEmpty
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		users, err := g.loadUsers(tx)
		if err != nil {
			return err
		}
		user, ok := findUser(users, userId)
		if !ok {
			return interfaces.ErrUserNotFound
		}
		claims, err := g.loadClaims(tx)
		if err != nil {
			return err
		}
		claim, ok := findClaim(claims, claimId)
		if !ok {
			return interfaces.ErrClaimNotFound
		}
		if user.ID == userId { // exact-append gate, legacy :411
			for _, c := range user.Claims {
				if c.ID == claim.ID {
					if err := tx.Exec("DELETE FROM user_claims WHERE user_id = ? AND claim_id = ?", user.ID, claim.ID).Error; err != nil {
						return mapGormError(err, nil, nil)
					}
					return nil
				}
			}
		}
		return interfaces.ErrClaimNotFound
	})
}
