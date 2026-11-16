package models

import "strings"

// Settings domain (centralized DB-backed configuration overrides with
// orchestrator → host synchronization). All five tables are AutoMigrated on
// every SQL-mode node; the node role decides which ones it writes:
//   - orchestrator nodes write setting_values / setting_sync_states
//     (source of truth) and read setting_definitions;
//   - every node applies through node_settings (its resolver layer) and
//     tracks its epoch in the setting_meta singleton.

// Value types a setting can declare in the definitions catalog.
type SettingValueType string

const (
	SettingValueTypeString SettingValueType = "string"
	SettingValueTypeBool   SettingValueType = "bool"
	SettingValueTypeInt    SettingValueType = "int"
)

// SettingValue scopes (orchestrator-side source of truth). Scopes are plain
// strings: "orchestrator", "hosts" (global for all hosts) or "host:<hostId>"
// (per-host override, wins over the hosts-global value).
const (
	SettingScopeOrchestrator = "orchestrator"
	SettingScopeHosts        = "hosts"
	SettingScopeHostPrefix   = "host:"
)

// HostSettingScope builds the per-host scope key for a host id.
func HostSettingScope(hostID string) string {
	return SettingScopeHostPrefix + hostID
}

// IsHostSettingScope reports whether a scope addresses a single host.
func IsHostSettingScope(scope string) bool {
	return strings.HasPrefix(scope, SettingScopeHostPrefix)
}

// HostIDFromSettingScope extracts the host id from a "host:<id>" scope
// ("" when the scope is not host-scoped).
func HostIDFromSettingScope(scope string) string {
	if !IsHostSettingScope(scope) {
		return ""
	}
	return strings.TrimPrefix(scope, SettingScopeHostPrefix)
}

// NodeSetting sources — which authority owns an applied resolver row on a
// node: rows the orchestrator sync pushed (managed) vs. rows edited directly
// on this node (local). Managed rows always win over local ones.
const (
	NodeSettingSourceLocal        = "local"
	NodeSettingSourceOrchestrator = "orchestrator"
)

// SettingSyncState status values (orchestrator-side per-host drift view).
const (
	SettingSyncStatusNever   = "never"
	SettingSyncStatusPending = "pending"
	SettingSyncStatusSynced  = "synced"
	SettingSyncStatusError   = "error"
)

// SettingDefinition.AppliesTo values — who may legitimately receive a key.
const (
	SettingAppliesToAll         = "all"
	SettingAppliesToOrchestrator = "orchestrator"
	SettingAppliesToHost        = "host"
)

// SettingDefinition is the catalog of configurable keys (the whitelist).
// Only keys present here can ever be stored in setting_values /
// node_settings; the catalog is seeded idempotently (insert-if-absent) from
// the compiled-in list in the settings package, so it grows with releases.
// Boot-critical and secret keys are permanently excluded from the catalog.
type SettingDefinition struct {
	Key             string           `json:"key" gorm:"primaryKey"`
	Type            SettingValueType `json:"type"`
	DefaultValue    string           `json:"default_value"`
	// ValidValues is a comma-separated allow-list for enum-like string keys
	// ("" = free-form string). Enforced by the settings service validator on
	// every write and on sync-apply.
	ValidValues     string           `json:"valid_values"`
	Description     string           `json:"description"`
	AppliesTo       string           `json:"applies_to"`
	RequiresRestart bool             `json:"requires_restart"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	*DbRecord      `json:"db_record" gorm:"-:all"`
}

// SettingValue is one override row on the orchestrator: a (scope, key) pair.
// UniqueIndex keeps the (scope,key) pair authoritative; deletes are hard
// (removal = fall back to the next layer, no tombstones).
type SettingValue struct {
	ID        string      `json:"id" gorm:"primaryKey"`
	Scope     string      `json:"scope" gorm:"index;uniqueIndex:ux_settingvalue_scope_key"`
	Key       string      `json:"key" gorm:"uniqueIndex:ux_settingvalue_scope_key"`
	Value     string      `json:"value"`
	UpdatedBy string      `json:"updated_by"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	*DbRecord `json:"db_record" gorm:"-:all"`
}

// SettingSyncState is the orchestrator's per-host synchronization bookkeeping
// (drift display). ContentHash is the sha256 of the last bundle the host
// acknowledged; a recompute of the desired bundle that no longer matches it
// means the host is pending a sync.
type SettingSyncState struct {
	HostID      string `json:"host_id" gorm:"primaryKey"`
	Revision    int64  `json:"revision"`
	ContentHash string `json:"content_hash"`
	Status      string `json:"status"`
	LastError   string `json:"last_error"`
	LastSyncedAt string `json:"last_synced_at"`
	*DbRecord   `json:"db_record" gorm:"-:all"`
}

// NodeSetting is the applied resolver layer on a node. Two sources coexist:
// the orchestrator-managed bundle (replaced wholesale by each sync-apply) and
// host-local edits (written through the host-side API, never touched by
// sync). Lookup order on a node: orchestrator source first, then local.
type NodeSetting struct {
	Key       string `json:"key" gorm:"primaryKey"`
	Source    string `json:"source" gorm:"primaryKey"`
	Value     string `json:"value"`
	Revision  int64  `json:"revision"`
	UpdatedAt string `json:"updated_at"`
	*DbRecord `json:"db_record" gorm:"-:all"`
}

// SettingMeta is the per-node singleton carrying the settings epoch and the
// management marker. Revision bumps on every local write/apply; Managed +
// ManagedHash are written by the sync-apply path and tell the node (and its
// UI) that an orchestrator owns the managed layer.
type SettingMeta struct {
	ID          int64  `json:"id" gorm:"primaryKey"`
	Revision    int64  `json:"revision"`
	Managed     bool   `json:"managed"`
	ManagedHash string `json:"managed_hash"`
	*DbRecord   `json:"db_record" gorm:"-:all"`
}
