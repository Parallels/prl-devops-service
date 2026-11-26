package settings

// Effective-value views: the listings show, for every catalogued key, what
// each layer holds and where the effective value actually comes from
// (source + shadowing). Also the managed-bundle machinery (hash, desired
// bundle, sync-apply) and the sync-state accessors used by the sync engine
// and the controllers.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

// BundleHash is the canonical content fingerprint of a managed bundle:
// sha256 over "key=value\n" lines sorted by key. Both sides recompute it, so
// idempotency checks never depend on clocks or revision bookkeeping.
func BundleHash(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	digest := sha256.New()
	for _, k := range keys {
		digest.Write([]byte(k))
		digest.Write([]byte("="))
		digest.Write([]byte(values[k]))
		digest.Write([]byte("\n"))
	}
	return hex.EncodeToString(digest.Sum(nil))
}

// DesiredBundleForHost computes the full managed bundle for one host:
// hosts-global values overridden by host-specific ones, dropping anything
// that is orchestrator-only or has left the catalog. Returns bundle, hash and
// the current epoch revision.
func (s *Service) DesiredBundleForHost(hostID string) (map[string]string, string, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bundle := map[string]string{}
	apply := func(scope string) {
		for k, v := range s.values[scope] {
			def, ok := s.definitions[k]
			if !ok || def.AppliesTo == models.SettingAppliesToOrchestrator {
				continue
			}
			bundle[k] = v
		}
	}
	apply(models.SettingScopeHosts)
	apply(models.HostSettingScope(hostID))
	return bundle, BundleHash(bundle), s.meta.Revision
}

// ApplyBundle installs a received managed bundle wholesale (source-scoped
// replace — node-local rows are never touched). Validates every key against
// the catalog (a host must never apply an uncatalogued key, whatever the
// transport bug), rejects hash mismatches (transport integrity) and returns
// unchanged=true for idempotent re-delivery.
func (s *Service) ApplyBundle(revision int64, hash string, values map[string]string) (apimodels.SettingsSyncResponse, error) {
	response := apimodels.SettingsSyncResponse{}
	normalized := map[string]string{}
	for rawKey, rawValue := range values {
		k := normKey(rawKey)
		def := s.Definition(k)
		if def == nil {
			return response, fmt.Errorf("setting %s is not part of the definitions catalog", k)
		}
		if def.AppliesTo == models.SettingAppliesToOrchestrator {
			return response, fmt.Errorf("setting %s must not be synced to hosts", def.Key)
		}
		value, err := validateValue(def, rawValue)
		if err != nil {
			return response, err
		}
		normalized[k] = value
	}
	computed := BundleHash(normalized)
	if hash != "" && computed != hash {
		return response, fmt.Errorf("bundle hash mismatch: received %s, computed %s", hash, computed)
	}

	changed, err := s.st.applyManagedBundle(normalized, revision, computed)
	if err != nil {
		return response, err
	}

	s.mu.Lock()
	s.managed = normalized
	s.meta.Managed = true
	s.meta.ManagedHash = computed
	if revision > s.meta.Revision {
		s.meta.Revision = revision
	}
	s.mu.Unlock()

	response.Hash = computed
	response.Unchanged = !changed
	response.AppliedRevision = revision
	response.AppliedCount = len(normalized)
	return response, nil
}

// ── sync state accessors ────────────────────────────────────────────────────

// SyncStates lists per-host drift bookkeeping rows (orchestrator).
func (s *Service) SyncStates() []models.SettingSyncState {
	rows, err := s.st.loadSyncStates()
	if err != nil {
		return nil
	}
	return rows
}

// SyncStateFor reads one host's state (zero row when never synced).
func (s *Service) SyncStateFor(hostID string) models.SettingSyncState {
	state, err := s.st.loadSyncState(hostID)
	if err != nil {
		return models.SettingSyncState{HostID: hostID, Status: models.SettingSyncStatusNever}
	}
	return state
}

// MarkHostPending flips a host to pending (write path + health-online hook).
func (s *Service) MarkHostPending(hostID string) {
	_ = s.st.markSyncPending(hostID)
}

// RecordSyncSuccess stores the acknowledged bundle fingerprint.
func (s *Service) RecordSyncSuccess(hostID string, revision int64, hash string) {
	_ = s.st.upsertSyncState(&models.SettingSyncState{
		HostID:       hostID,
		Revision:     revision,
		ContentHash:  hash,
		Status:       models.SettingSyncStatusSynced,
		LastSyncedAt: helpers.GetUtcCurrentDateTime(),
	})
}

// RecordSyncFailure keeps the last known fingerprint, flags the error.
func (s *Service) RecordSyncFailure(hostID string, cause string) {
	state, err := s.st.loadSyncState(hostID)
	if err != nil {
		state = models.SettingSyncState{HostID: hostID}
	}
	state.Status = models.SettingSyncStatusError
	state.LastError = cause
	_ = s.st.upsertSyncState(&state)
}

// HostIsCurrent reports whether a host already holds the desired bundle.
func (s *Service) HostIsCurrent(hostID string) bool {
	bundle, hash, _ := s.DesiredBundleForHost(hostID)
	_ = bundle
	state := s.SyncStateFor(hostID)
	return state.Status == models.SettingSyncStatusSynced && state.ContentHash == hash
}

// ── effective views ─────────────────────────────────────────────────────────

func strPtr(v string) *string { return &v }

// OrchestratorNodeView resolves the effective config of THIS orchestrator
// process: managed (only when a local host agent pushed bundle exists here) >
// node-scope > orchestrator-scope > local > env > file > default.
func (s *Service) OrchestratorNodeView() []apimodels.SettingEffectiveEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	orch := s.values[models.SettingScopeOrchestrator]
	// The replica's own layer only exists when this machine can be identified;
	// resolving it against an empty scope would show a phantom layer fed by
	// whoever wrote the last "node:" row in the table.
	var node map[string]string
	if s.nodeScope != "" {
		node = s.values[s.nodeScope]
	}
	entries := make([]apimodels.SettingEffectiveEntry, 0, len(s.definitions))
	for _, def := range s.definitions {
		entry := newEntry(&def)
		if v, ok := s.managed[def.Key]; ok {
			entry.ManagedValue = strPtr(v)
		}
		if v, ok := node[def.Key]; ok {
			entry.NodeValue = strPtr(v)
		}
		if v, ok := orch[def.Key]; ok {
			entry.OrchestratorValue = strPtr(v)
		}
		if v, ok := s.local[def.Key]; ok {
			entry.LocalValue = strPtr(v)
		}
		annotateShadowing(&entry)
		entry.Effective, entry.Source = resolveLayers(def.Key, def.DefaultValue,
			[]layerSource{
				{entry.ManagedValue, SourceManaged},
				{entry.NodeValue, SourceNode},
				{entry.OrchestratorValue, SourceOrchestrator},
				{entry.LocalValue, SourceLocal},
			})
		entries = append(entries, entry)
	}
	sortEffective(entries)
	return entries
}

// HostNodeView resolves the effective config of THIS host node:
// managed > local > env > file > default.
func (s *Service) HostNodeView() []apimodels.SettingEffectiveEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]apimodels.SettingEffectiveEntry, 0, len(s.definitions))
	for _, def := range s.definitions {
		entry := newEntry(&def)
		if v, ok := s.managed[def.Key]; ok {
			entry.ManagedValue = strPtr(v)
		}
		if v, ok := s.local[def.Key]; ok {
			entry.LocalValue = strPtr(v)
		}
		annotateShadowing(&entry)
		entry.Effective, entry.Source = resolveLayers(def.Key, def.DefaultValue,
			[]layerSource{
				{entry.ManagedValue, SourceManaged},
				{entry.LocalValue, SourceLocal},
			})
		entries = append(entries, entry)
	}
	sortEffective(entries)
	return entries
}

// HostEffectiveFor is the orchestrator's view of what host hostID will hold
// after the next successful sync (bundle layers only — the host's own env
// values are only visible on the host itself).
func (s *Service) HostEffectiveFor(hostID string) []apimodels.SettingEffectiveEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hosts := s.values[models.SettingScopeHosts]
	perHost := s.values[models.HostSettingScope(hostID)]
	entries := make([]apimodels.SettingEffectiveEntry, 0, len(s.definitions))
	for _, def := range s.definitions {
		if def.AppliesTo == models.SettingAppliesToOrchestrator {
			continue
		}
		entry := newEntry(&def)
		if v, ok := hosts[def.Key]; ok {
			entry.HostsValue = strPtr(v)
		}
		if v, ok := perHost[def.Key]; ok {
			entry.HostValue = strPtr(v)
		}
		if entry.HostValue != nil && entry.HostsValue != nil {
			entry.ShadowedBy = SourceHost
		}
		// Remote view: no env/file probing against THIS process — the host's
		// own environment is only visible on the host. Below the bundle sits
		// the catalog default, nothing else.
		switch {
		case entry.HostValue != nil:
			entry.Effective, entry.Source = *entry.HostValue, SourceHost
		case entry.HostsValue != nil:
			entry.Effective, entry.Source = *entry.HostsValue, SourceHosts
		default:
			entry.Effective, entry.Source = def.DefaultValue, SourceDefault
		}
		entries = append(entries, entry)
	}
	sortEffective(entries)
	return entries
}

// ── helpers ─────────────────────────────────────────────────────────────────

func newEntry(def *models.SettingDefinition) apimodels.SettingEffectiveEntry {
	return apimodels.SettingEffectiveEntry{
		Key:             def.Key,
		Type:            string(def.Type),
		DefaultValue:    def.DefaultValue,
		ValidValues:     def.ValidValues,
		Description:     def.Description,
		AppliesTo:       def.AppliesTo,
		RequiresRestart: def.RequiresRestart,
	}
}

// annotateShadowing marks lower settings layers hidden by higher ones so the
// UI can flag host-local edits that currently have no effect.
func annotateShadowing(entry *apimodels.SettingEffectiveEntry) {
	if entry.ManagedValue != nil {
		if entry.NodeValue != nil || entry.OrchestratorValue != nil || entry.LocalValue != nil {
			entry.ShadowedBy = SourceManaged
		}
		return
	}
	if entry.NodeValue != nil {
		// A replica's own value hiding the shared one is not a mistake, but it is
		// worth showing: it is the layer that makes two replicas attached to one
		// database behave differently.
		if entry.OrchestratorValue != nil || entry.LocalValue != nil {
			entry.ShadowedBy = SourceNode
		}
		return
	}
	if entry.OrchestratorValue != nil && entry.LocalValue != nil {
		entry.ShadowedBy = SourceOrchestrator
	}
}

type layerSource struct {
	value  *string
	source string
}

// resolveLayers walks the DB layers in precedence order, falling back to
// env > config file > catalog default.
func resolveLayers(key, defaultValue string, layers []layerSource) (string, string) {
	for _, layer := range layers {
		if layer.value != nil {
			return *layer.value, layer.source
		}
	}
	if v, ok := osLookupEnv(key); ok {
		return v, SourceEnv
	}
	if v, ok := config.Get().ConfigFileValue(key); ok {
		return v, SourceConfigFile
	}
	return defaultValue, SourceDefault
}

func sortEffective(entries []apimodels.SettingEffectiveEntry) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
}
