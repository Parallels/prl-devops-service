package models

// API surface models for the settings domain (centralized configuration).
// Pointer layers in the effective entries distinguish "not set at this layer"
// (nil) from "deliberately set to the empty string" (pointer to "").

// SettingDefinitionResponse is one catalogued key.
type SettingDefinitionResponse struct {
	Key             string `json:"key"`
	Type            string `json:"type"`
	DefaultValue    string `json:"default_value"`
	ValidValues     string `json:"valid_values,omitempty"`
	Description     string `json:"description"`
	AppliesTo       string `json:"applies_to"`
	RequiresRestart bool   `json:"requires_restart"`
}

// SettingValueResponse is one stored override row (orchestrator side).
type SettingValueResponse struct {
	ID        string `json:"id"`
	Scope     string `json:"scope"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedBy string `json:"updated_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SettingSyncStateResponse is the per-host drift row (orchestrator side).
type SettingSyncStateResponse struct {
	HostID       string `json:"host_id"`
	HostName     string `json:"host_name,omitempty"`
	Revision     int64  `json:"revision"`
	ContentHash  string `json:"content_hash"`
	Status       string `json:"status"`
	LastError    string `json:"last_error,omitempty"`
	LastSyncedAt string `json:"last_synced_at,omitempty"`
}

// SettingEffectiveEntry resolves one key against every layer that could
// supply it. Layer fields are pointers: nil means the layer does not carry
// the key at all. Source is where the effective value actually comes from:
// managed_db | orchestrator_db | hosts_db | host_db | local_db | env |
// config_file | default. ShadowedBy is set when a lower settings layer is
// hidden by a higher one (e.g. a host-local edit shadowed by the
// orchestrator-managed value).
type SettingEffectiveEntry struct {
	Key             string  `json:"key"`
	Type            string  `json:"type"`
	DefaultValue    string  `json:"default_value"`
	ValidValues     string  `json:"valid_values,omitempty"`
	Description     string  `json:"description"`
	AppliesTo       string  `json:"applies_to"`
	RequiresRestart bool    `json:"requires_restart"`
	ManagedValue    *string `json:"managed_value,omitempty"`
	OrchestratorValue *string `json:"orchestrator_value,omitempty"`
	HostsValue      *string `json:"hosts_value,omitempty"`
	HostValue       *string `json:"host_value,omitempty"`
	LocalValue      *string `json:"local_value,omitempty"`
	Effective       string  `json:"effective"`
	Source          string  `json:"source"`
	ShadowedBy      string  `json:"shadowed_by,omitempty"`
}

// SettingsNodeInfo describes this node's settings posture for the UI.
type SettingsNodeInfo struct {
	Managed     bool   `json:"managed"`
	Revision    int64  `json:"revision"`
	ManagedHash string `json:"managed_hash,omitempty"`
}

// SettingsResponse is the full settings listing — definitions, stored
// overrides, resolved effective view with sources, and drift (orchestrator).
type SettingsResponse struct {
	Definitions []SettingDefinitionResponse  `json:"definitions"`
	Values      []SettingValueResponse       `json:"values"`
	Effective   []SettingEffectiveEntry      `json:"effective"`
	Sync        []SettingSyncStateResponse   `json:"sync,omitempty"`
	Node        SettingsNodeInfo             `json:"node"`
}

// SettingsHostEffectiveResponse is the orchestrator's per-host resolved
// bundle plus its sync status (remote env values are only visible on the
// host itself — this view is the managed bundle the host will hold).
type SettingsHostEffectiveResponse struct {
	HostID       string                    `json:"host_id"`
	HostName     string                    `json:"host_name,omitempty"`
	Bundle       map[string]string         `json:"bundle"`
	Effective    []SettingEffectiveEntry   `json:"effective"`
	Revision     int64                     `json:"revision"`
	ContentHash  string                    `json:"content_hash"`
	Status       string                    `json:"status"`
	LastError    string                    `json:"last_error,omitempty"`
	LastSyncedAt string                    `json:"last_synced_at,omitempty"`
}

// SettingValueRequest carries a single override write.
type SettingValueRequest struct {
	Value string `json:"value"`
}

// SettingsSyncTargetRequest selects the manual sync fan-out: "all" or
// "host:<hostId>".
type SettingsSyncTargetRequest struct {
	Target string `json:"target"`
}

// SettingsSyncRequest is the full managed bundle pushed by the orchestrator
// to a host (POST /settings/apply on the host).
type SettingsSyncRequest struct {
	Revision int64             `json:"revision"`
	Hash     string            `json:"hash"`
	Values   map[string]string `json:"values"`
}

// SettingsSyncResponse is the host's acknowledgement.
type SettingsSyncResponse struct {
	AppliedRevision int64  `json:"applied_revision"`
	Hash            string `json:"hash"`
	Unchanged       bool   `json:"unchanged"`
	AppliedCount    int    `json:"applied_count"`
}

// SettingsSyncResult reports one host's outcome of a fan-out sync.
type SettingsSyncResult struct {
	HostID    string `json:"host_id"`
	HostName  string `json:"host_name,omitempty"`
	Status    string `json:"status"`
	Skipped   bool   `json:"skipped"`
	Unchanged bool   `json:"unchanged"`
	Revision  int64  `json:"revision"`
	Error     string `json:"error,omitempty"`
}
