package database

// SQL implementation of the api-keys domain (PBI #421 task 07) — the auth
// hot path. Ported method-for-method from data/api_key.go including the
// internal/external split, the own-list claim masking (which also masks
// CREATE-collision probes and REVOKE lookups because every single fetch
// funnels through GetApiKey), the revoked-key/owner-missing fallbacks of
// GetUserByApiKey, and the silent empty-id no-op of Delete/Revoke.
// GetUserByApiKey resolves its user through the GORM users path (task 06),
// never the delegate.

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/security/password"
	"gorm.io/gorm"
)

func loadApiKeys(tx *gorm.DB) ([]models.ApiKey, error) {
	var keys []models.ApiKey
	if err := tx.Order("rowid").Find(&keys).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return keys, nil
}

// normalizeApiKeyType mirrors api_key.go:36-40 (rows migrated from pre-Type
// data may carry the empty string; the read path renders them "external"
// without rewriting the store).
func normalizeApiKeyType(key *models.ApiKey) {
	if key.Type == "" {
		key.Type = "external"
	}
}

// filterInternalKeys mirrors api_key.go:44-54 verbatim (normalizes the copy
// it keeps; the store stays untouched).
func filterInternalKeys(keys []models.ApiKey) []models.ApiKey {
	result := make([]models.ApiKey, 0, len(keys))
	for _, k := range keys {
		normalizeApiKeyType(&k)
		if k.Type != "internal" {
			result = append(result, k)
		}
	}
	return result
}

func (g *GormDatabase) GetApiKeys(ctx basecontext.ApiContext, filter string) ([]models.ApiKey, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	keys, err := loadApiKeys(g.gormDB)
	if err != nil {
		return nil, err
	}
	filtered, err := query.FilterByProperty(keys, dbFilter)
	if err != nil {
		return nil, err
	}

	// Internal keys (function also normalizes types) BEFORE the auth check — legacy :70.
	filtered = filterInternalKeys(filtered)

	authContext := ctx.GetAuthorizationContext()
	if authContext == nil || authContext.User == nil {
		return filtered, nil
	}
	hasFullListClaim := authContext.HasEffectiveClaim(constants.LIST_API_KEY_CLAIM)
	hasOwnListClaim := authContext.HasEffectiveClaim(constants.LIST_OWN_API_KEY_CLAIM)
	if !hasFullListClaim && hasOwnListClaim {
		// GetOwnRecords port (task 03): query.OwnRecords — verbatim semantics,
		// including the pass-through when no user is attached.
		filtered = query.OwnRecords(ctx, filtered...)
	}
	return filtered, nil
}

// GetAllApiKeysIncludingInternal mirrors :90-108 — admin listing, deliberately
// NOT authz-wrapped and NOT internal-filtered (confirmed at task start per the
// spec's open question; :95-107 shows a plain copy + normalize, no
// GetAuthorizedRecords call).
func (g *GormDatabase) GetAllApiKeysIncludingInternal(ctx basecontext.ApiContext) ([]models.ApiKey, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	keys, err := loadApiKeys(g.gormDB)
	if err != nil {
		return nil, err
	}
	for i := range keys {
		normalizeApiKeyType(&keys[i])
	}
	return keys, nil
}

// GetApiKey mirrors :110-133: ID EXACT or EqualFold on Name or KEY value,
// first hit in store order; the own-list claim masks OTHER users' keys —
// internal keys are NOT filtered on this path (quirk kept). No empty-input
// sentinel exists in legacy (the scan simply matches nothing) — mirrored.
func (g *GormDatabase) GetApiKey(ctx basecontext.ApiContext, idOrName string) (*models.ApiKey, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	keys, err := loadApiKeys(g.gormDB)
	if err != nil {
		return nil, err
	}
	return scanApiKey(ctx, keys, idOrName)
}

func scanApiKey(ctx basecontext.ApiContext, keys []models.ApiKey, idOrName string) (*models.ApiKey, error) {
	for _, apiKey := range keys {
		if apiKey.ID == idOrName || strings.EqualFold(apiKey.Name, idOrName) || strings.EqualFold(apiKey.Key, idOrName) {
			authContext := ctx.GetAuthorizationContext()
			if authContext != nil && authContext.User != nil {
				hasFullListClaim := authContext.HasEffectiveClaim(constants.LIST_API_KEY_CLAIM)
				hasOwnListClaim := authContext.HasEffectiveClaim(constants.LIST_OWN_API_KEY_CLAIM)
				if !hasFullListClaim && hasOwnListClaim && apiKey.UserID != authContext.User.ID {
					return nil, interfaces.ErrApiKeyNotFound
				}
			}
			normalizeApiKeyType(&apiKey)
			return &apiKey, nil
		}
	}
	return nil, interfaces.ErrApiKeyNotFound
}

// GetUserByApiKey mirrors :137-155 (auth hot path): ANY lookup error is
// re-served as ErrApiKeyNotFound (even ErrDatabaseNotConnected — quirk kept),
// revoked keys are indistinguishable from missing ones, and a missing owner
// (or no linked user) resolves to the root user via the SQL users path.
func (g *GormDatabase) GetUserByApiKey(ctx basecontext.ApiContext, apiKey string) (*models.User, error) {
	apiKeyRec, err := g.GetApiKey(ctx, apiKey)
	if err != nil {
		return nil, interfaces.ErrApiKeyNotFound
	}
	if apiKeyRec.Revoked {
		return nil, interfaces.ErrApiKeyNotFound
	}
	if apiKeyRec.UserID != "" {
		user, err := g.GetUser(ctx, apiKeyRec.UserID)
		if err != nil || user == nil {
			return g.GetUser(ctx, constants.ROOT_USER_ID)
		}
		return user, nil
	}
	return g.GetUser(ctx, constants.ROOT_USER_ID)
}

// CreateApiKey — D-9 row 9 (+ internal-key flows): the two GetApiKey-flavoured
// collision probes (which see through the same auth masking — a collision
// hidden from an own-claims caller inserts anyway and then hits the UNIQUE
// index; the driver error maps to ErrApiKeyAlreadyExists) and the insert share
// one transaction. bcrypt runs BEFORE the tx opens.
func (g *GormDatabase) CreateApiKey(ctx basecontext.ApiContext, apiKey models.ApiKey) (*models.ApiKey, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}

	hashSecret, err := password.Get().Hash(apiKey.Secret, apiKey.ID)
	if err != nil {
		return nil, err
	}

	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		keys, err := loadApiKeys(tx)
		if err != nil {
			return err
		}
		if found, _ := scanApiKey(ctx, keys, apiKey.ID); found != nil {
			return interfaces.ErrApiKeyAlreadyExists
		}
		if found, _ := scanApiKey(ctx, keys, apiKey.Name); found != nil {
			return interfaces.ErrApiKeyAlreadyExists
		}

		if apiKey.Type == "" {
			apiKey.Type = "external"
		}
		apiKey.Secret = hashSecret
		apiKey.UpdatedAt = helpers.GetUtcCurrentDateTime()
		apiKey.CreatedAt = helpers.GetUtcCurrentDateTime()

		if err := tx.Create(&apiKey).Error; err != nil {
			return mapGormError(err, nil, interfaces.ErrApiKeyAlreadyExists)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// DeleteApiKey mirrors :189-214: silent no-op on empty id; EqualFold scan over
// ID/Name/KEY value. The legacy record-lock spin is an artifact of the
// in-memory model (417 non-goal: concurrency model port) — the row delete is
// atomic on its own table, no transaction per D-9.
func (g *GormDatabase) DeleteApiKey(ctx basecontext.ApiContext, id string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if id == "" {
		return nil
	}
	keys, err := loadApiKeys(g.gormDB)
	if err != nil {
		return err
	}
	for i := range keys {
		k := keys[i]
		if strings.EqualFold(k.ID, id) || strings.EqualFold(k.Name, id) || strings.EqualFold(k.Key, id) {
			if err := g.gormDB.Where("id = ?", k.ID).Delete(&models.ApiKey{}).Error; err != nil {
				return mapGormError(err, nil, nil)
			}
			return nil
		}
	}
	return interfaces.ErrApiKeyNotFound
}

// UpdateKey mirrors :216-242: EXACT ID match only; writes Revoked/RevokedAt/
// UpdatedAt and nothing else. Non-transactional per D-9.
func (g *GormDatabase) UpdateKey(ctx basecontext.ApiContext, key models.ApiKey) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	var n int64
	if err := g.gormDB.Model(&models.ApiKey{}).Where("id = ?", key.ID).Count(&n).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	if n == 0 {
		return interfaces.ErrApiKeyNotFound
	}
	if err := g.gormDB.Table("api_keys").Where("id = ?", key.ID).
		Updates(map[string]interface{}{
			"revoked":    key.Revoked,
			"revoked_at": key.RevokedAt,
			"updated_at": helpers.GetUtcCurrentDateTime(),
		}).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	return nil
}

// RevokeKey mirrors :244-268: silent no-op on empty id, lookup through the
// MASKED GetApiKey (an own-claims caller cannot revoke another user's key),
// then UpdateKey.
func (g *GormDatabase) RevokeKey(ctx basecontext.ApiContext, id string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if id == "" {
		return nil
	}
	key, err := g.GetApiKey(ctx, id)
	if err != nil {
		return err
	}
	if key == nil {
		return interfaces.ErrApiKeyNotFound
	}
	key.Revoked = true
	key.RevokedAt = helpers.GetUtcCurrentDateTime()
	return g.UpdateKey(ctx, *key)
}
