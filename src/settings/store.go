package settings

// GORM persistence layer for the settings domain. It deliberately does NOT
// extend the frozen datainterfaces.DatabaseService surface (#421 R-1): the
// store talks to the raw handle from serviceprovider.GetSQLDatabase().
// Dialect-neutral SQL only (sqlite + postgres); timestamps follow the
// string-timestamp convention every other model uses.

import (
	"errors"

	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrStoreUnavailable is returned when the store was built without a live
// handle (legacy JSON mode or a failed SQL open).
var ErrStoreUnavailable = errors.New("settings: sql database unavailable")

// SettingMeta singleton primary key — one row, id pinned to 1.
const settingMetaID = 1

type store struct {
	db *gorm.DB
}

func newStore(db *gorm.DB) *store { return &store{db: db} }

func (s *store) available() bool { return s != nil && s.db != nil }

// ensureMeta seeds the singleton row on first boot (INSERT ... WHERE NOT
// EXISTS — race-closed, same posture as the configurations singleton).
func (s *store) ensureMeta() error {
	if !s.available() {
		return ErrStoreUnavailable
	}
	return s.db.Exec(
		"INSERT INTO setting_meta (id, revision, managed, managed_hash) SELECT ? , 0, ?, ? WHERE NOT EXISTS (SELECT 1 FROM setting_meta)",
		settingMetaID, false, "").Error
}

// meta reads the singleton (ensureMeta first — every meta read is seed-on-read).
func (s *store) meta() (models.SettingMeta, error) {
	var meta models.SettingMeta
	if !s.available() {
		return meta, ErrStoreUnavailable
	}
	if err := s.ensureMeta(); err != nil {
		return meta, err
	}
	err := s.db.Table("setting_meta").Where("id = ?", settingMetaID).Take(&meta).Error
	return meta, err
}

// bumpRevision increments the node epoch inside tx and returns the new value.
// The row is guaranteed by ensureMeta (tx-scoped seed).
func (s *store) bumpRevision(tx *gorm.DB) (int64, error) {
	if err := tx.Exec(
		"INSERT INTO setting_meta (id, revision, managed, managed_hash) SELECT ?, 0, ?, ? WHERE NOT EXISTS (SELECT 1 FROM setting_meta)",
		settingMetaID, false, "").Error; err != nil {
		return 0, err
	}
	if err := tx.Exec("UPDATE setting_meta SET revision = revision + 1 WHERE id = ?", settingMetaID).Error; err != nil {
		return 0, err
	}
	var meta models.SettingMeta
	if err := tx.Table("setting_meta").Where("id = ?", settingMetaID).Take(&meta).Error; err != nil {
		return 0, err
	}
	return meta.Revision, nil
}

// ── definitions ─────────────────────────────────────────────────────────────

// seedDefinitions inserts catalog rows that do not exist yet — insert-if-
// absent only; previously seeded rows keep their stored description/default
// (operators may have local tweaks, upgrades add new keys without churn).
func (s *store) seedDefinitions(defs []models.SettingDefinition) error {
	if !s.available() {
		return ErrStoreUnavailable
	}
	now := helpers.GetUtcCurrentDateTime()
	for _, def := range defs {
		var n int64
		if err := s.db.Table("setting_definitions").Where("key = ?", def.Key).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		def.CreatedAt = now
		def.UpdatedAt = now
		if err := s.db.Table("setting_definitions").Create(&def).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *store) loadDefinitions() ([]models.SettingDefinition, error) {
	var rows []models.SettingDefinition
	if !s.available() {
		return nil, ErrStoreUnavailable
	}
	err := s.db.Table("setting_definitions").Order("key").Find(&rows).Error
	return rows, err
}

// ── scoped values (orchestrator source of truth) ────────────────────────────

// upsertValue writes one (scope,key) row and bumps the epoch atomically.
func (s *store) upsertValue(value *models.SettingValue) (int64, error) {
	var revision int64
	if !s.available() {
		return 0, ErrStoreUnavailable
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if value.ID == "" {
			value.ID = helpers.GenerateId()
		}
		now := helpers.GetUtcCurrentDateTime()
		if value.CreatedAt == "" {
			value.CreatedAt = now
		}
		value.UpdatedAt = now
		if err := tx.Table("setting_values").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "scope"}, {Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_by", "updated_at"}),
		}).Create(value).Error; err != nil {
			return err
		}
		var err error
		revision, err = s.bumpRevision(tx)
		return err
	})
	return revision, err
}

// deleteValue hard-removes one (scope,key) row, bumping the epoch only when
// a row actually disappeared. Returns existed.
func (s *store) deleteValue(scope, key string) (bool, int64, error) {
	if !s.available() {
		return false, 0, ErrStoreUnavailable
	}
	var (
		existed  bool
		revision int64
	)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table("setting_values").Where("scope = ? AND key = ?", scope, key).Delete(&models.SettingValue{})
		if res.Error != nil {
			return res.Error
		}
		existed = res.RowsAffected > 0
		if !existed {
			return nil
		}
		var err error
		revision, err = s.bumpRevision(tx)
		return err
	})
	return existed, revision, err
}

func (s *store) loadValues() ([]models.SettingValue, error) {
	var rows []models.SettingValue
	if !s.available() {
		return nil, ErrStoreUnavailable
	}
	err := s.db.Table("setting_values").Order("scope, key").Find(&rows).Error
	return rows, err
}

// deleteHostValues drops every override of a single host scope (unregister
// cleanup). Hard delete; the per-host sync state is dropped separately.
func (s *store) deleteHostValues(hostID string) error {
	if !s.available() {
		return ErrStoreUnavailable
	}
	return s.db.Table("setting_values").
		Where("scope = ?", models.HostSettingScope(hostID)).Delete(&models.SettingValue{}).Error
}

// ── node settings (applied resolver layer on any node) ─────────────────────

func (s *store) loadNodeSettings() ([]models.NodeSetting, error) {
	var rows []models.NodeSetting
	if !s.available() {
		return nil, ErrStoreUnavailable
	}
	err := s.db.Table("node_settings").Order("key, source").Find(&rows).Error
	return rows, err
}

// upsertLocalSetting writes/updates one local source row and bumps the epoch.
func (s *store) upsertLocalSetting(key, value string) (int64, error) {
	var revision int64
	if !s.available() {
		return 0, ErrStoreUnavailable
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := helpers.GetUtcCurrentDateTime()
		if err := tx.Table("node_settings").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}, {Name: "source"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).Create(&models.NodeSetting{
			Key: key, Source: models.NodeSettingSourceLocal, Value: value, UpdatedAt: now,
		}).Error; err != nil {
			return err
		}
		var err error
		revision, err = s.bumpRevision(tx)
		if err != nil {
			return err
		}
		return tx.Table("node_settings").Where("key = ? AND source = ?", key, models.NodeSettingSourceLocal).
			Update("revision", revision).Error
	})
	return revision, err
}

// deleteLocalSetting removes one local source row, bumping on actual removal.
func (s *store) deleteLocalSetting(key string) (bool, int64, error) {
	if !s.available() {
		return false, 0, ErrStoreUnavailable
	}
	var (
		existed  bool
		revision int64
	)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table("node_settings").
			Where("key = ? AND source = ?", key, models.NodeSettingSourceLocal).
			Delete(&models.NodeSetting{})
		if res.Error != nil {
			return res.Error
		}
		existed = res.RowsAffected > 0
		if !existed {
			return nil
		}
		var err error
		revision, err = s.bumpRevision(tx)
		return err
	})
	return existed, revision, err
}

// applyManagedBundle replaces the managed layer wholesale inside one
// transaction: local rows are NEVER touched, the meta row learns the bundle
// hash + revision and the managed marker. Returns changed=false (and writes
// nothing) when the received hash equals the stored managed hash.
func (s *store) applyManagedBundle(values map[string]string, revision int64, hash string) (bool, error) {
	if !s.available() {
		return false, ErrStoreUnavailable
	}
	changed := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"INSERT INTO setting_meta (id, revision, managed, managed_hash) SELECT ?, 0, ?, ? WHERE NOT EXISTS (SELECT 1 FROM setting_meta)",
			settingMetaID, false, "").Error; err != nil {
			return err
		}
		var meta models.SettingMeta
		if err := tx.Table("setting_meta").Where("id = ?", settingMetaID).Take(&meta).Error; err != nil {
			return err
		}
		if meta.ManagedHash == hash && meta.Managed {
			// Idempotent re-delivery: content identical — stamp revision only.
			if revision > meta.Revision {
				if err := tx.Table("setting_meta").Where("id = ?", settingMetaID).
					Update("revision", revision).Error; err != nil {
					return err
				}
			}
			return nil
		}
		if err := tx.Where("source = ?", models.NodeSettingSourceOrchestrator).
			Delete(&models.NodeSetting{}).Error; err != nil {
			return err
		}
		now := helpers.GetUtcCurrentDateTime()
		for key, value := range values {
			row := models.NodeSetting{
				Key: key, Source: models.NodeSettingSourceOrchestrator, Value: value,
				Revision: revision, UpdatedAt: now,
			}
			if err := tx.Table("node_settings").Create(&row).Error; err != nil {
				return err
			}
		}
		newRevision := revision
		if meta.Revision > newRevision {
			newRevision = meta.Revision
		}
		if err := tx.Table("setting_meta").Where("id = ?", settingMetaID).
			Updates(map[string]interface{}{
				"revision":     newRevision,
				"managed":      true,
				"managed_hash": hash,
			}).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	return changed, err
}

// ── sync states (orchestrator drift bookkeeping) ────────────────────────────

func (s *store) loadSyncStates() ([]models.SettingSyncState, error) {
	var rows []models.SettingSyncState
	if !s.available() {
		return nil, ErrStoreUnavailable
	}
	err := s.db.Table("setting_sync_states").Order("host_id").Find(&rows).Error
	return rows, err
}

func (s *store) loadSyncState(hostID string) (models.SettingSyncState, error) {
	var state models.SettingSyncState
	if !s.available() {
		return state, ErrStoreUnavailable
	}
	err := s.db.Table("setting_sync_states").Where("host_id = ?", hostID).Take(&state).Error
	return state, err
}

// upsertSyncState writes the full per-host state row (hash/revision/status).
func (s *store) upsertSyncState(state *models.SettingSyncState) error {
	if !s.available() {
		return ErrStoreUnavailable
	}
	return s.db.Table("setting_sync_states").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "host_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"revision", "content_hash", "status", "last_error", "last_synced_at"}),
	}).Create(state).Error
}

// markSyncPending flips status to pending for one host while keeping the last
// acknowledged hash/revision (drift display stays meaningful). A fresh row
// starts from empty-hash pending.
func (s *store) markSyncPending(hostID string) error {
	if !s.available() {
		return ErrStoreUnavailable
	}
	return s.db.Table("setting_sync_states").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "host_id"}},
		DoNothing: false,
		DoUpdates: clause.Assignments(map[string]interface{}{"status": models.SettingSyncStatusPending}),
	}).Create(&models.SettingSyncState{HostID: hostID, Status: models.SettingSyncStatusPending}).Error
}

func (s *store) deleteSyncState(hostID string) error {
	if !s.available() {
		return ErrStoreUnavailable
	}
	return s.db.Table("setting_sync_states").Where("host_id = ?", hostID).
		Delete(&models.SettingSyncState{}).Error
}
