package settings

// Central settings service: DB-backed configuration overrides that beat
// config file / CLI flags / environment variables and apply live through the
// Config.GetKey override hook — no restart. Two layers exist on every node:
//
//	managed — rows the orchestrator pushed (source=orchestrator in
//	          node_settings); replaced wholesale on each sync-apply.
//	local   — rows edited directly on this node (UI pointed at a host, or
//	          orchestrator-scope edits on the orchestrator itself).
//
// Lookup precedence in the override resolver:
//
//	host node:            managed > local
//	orchestrator node:    managed > orchestrator-scope > local
//
// orchestrator-scope covers setting_values rows with scope "orchestrator"
// (the service's own overrides); host-scope rows never touch the orchestrator
// process. Below the DB layers sit env > flag > config file > catalog
// default (handled by the untouched GetKey fallback).
//
// The orchestrator is the source of truth for host-scoped values
// (setting_values scope "hosts" / "host:<id>"); hosts receive full bundles
// and never host-to-host fan out. Legacy JSON DB mode leaves this service
// uninitialized — the override hook is simply never installed.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// Value source labels in the effective view (spec constant set).
const (
	SourceManaged      = "managed_db"
	SourceOrchestrator = "orchestrator_db"
	SourceHosts        = "hosts_db"
	SourceHost         = "host_db"
	SourceLocal        = "local_db"
	SourceEnv          = "env"
	SourceConfigFile   = "config_file"
	SourceDefault      = "default"
)

const maxSettingValueBytes = 4096

type Service struct {
	ctx basecontext.ApiContext
	st  *store

	mu          sync.RWMutex
	definitions map[string]models.SettingDefinition   // catalogued keys (upper)
	values      map[string]map[string]string          // scope -> key -> value
	managed     map[string]string                     // node_settings: orchestrator source
	local       map[string]string                     // node_settings: local source
	meta        models.SettingMeta

	isOrchestrator bool
	isHost         bool
}

var globalService *Service

// InitService opens the settings store against the live SQL handle, seeds the
// definitions catalog and the meta singleton, loads everything into memory
// and registers the config override resolver. Safe to call once from
// startup; returns an error (not fatal) when SQL mode is unavailable so the
// boot path only logs a warning.
func InitService(ctx basecontext.ApiContext) (*Service, error) {
	db, err := serviceprovider.GetSQLDatabase(ctx)
	if err != nil {
		return nil, err
	}
	svc := &Service{
		ctx:            ctx,
		st:             newStore(db),
		definitions:    map[string]models.SettingDefinition{},
		values:         map[string]map[string]string{},
		managed:        map[string]string{},
		local:          map[string]string{},
		isOrchestrator: config.Get().IsOrchestrator(),
		isHost:         config.Get().IsHost(),
	}
	if err := svc.load(); err != nil {
		return nil, err
	}
	config.SetOverrideResolver(svc.lookup)
	globalService = svc
	return svc, nil
}

// Get returns the initialized service or nil when the settings domain is
// inert (legacy DB mode or init failure).
func Get() *Service { return globalService }

// Shutdown detaches the resolver hook (tests / graceful stop).
func Shutdown() {
	if globalService != nil {
		config.SetOverrideResolver(nil)
		globalService = nil
	}
}

func (s *Service) load() error {
	if err := s.st.ensureMeta(); err != nil {
		return err
	}
	if err := s.st.seedDefinitions(Catalog()); err != nil {
		return err
	}
	defs, err := s.st.loadDefinitions()
	if err != nil {
		return err
	}
	vals, err := s.st.loadValues()
	if err != nil {
		return err
	}
	nodes, err := s.st.loadNodeSettings()
	if err != nil {
		return err
	}
	meta, err := s.st.meta()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.definitions = map[string]models.SettingDefinition{}
	for _, d := range defs {
		s.definitions[normKey(d.Key)] = d
	}
	s.values = map[string]map[string]string{}
	for _, v := range vals {
		scope := strings.TrimSpace(v.Scope)
		if s.values[scope] == nil {
			s.values[scope] = map[string]string{}
		}
		s.values[scope][normKey(v.Key)] = v.Value
	}
	s.managed = map[string]string{}
	s.local = map[string]string{}
	for _, n := range nodes {
		switch n.Source {
		case models.NodeSettingSourceOrchestrator:
			s.managed[normKey(n.Key)] = n.Value
		case models.NodeSettingSourceLocal:
			s.local[normKey(n.Key)] = n.Value
		}
	}
	s.meta = meta
	return nil
}

// ── resolver hook ───────────────────────────────────────────────────────────

// lookup is the function handed to config.SetOverrideResolver. Thread-safe:
// callers hit it on every GetKey invocation.
func (s *Service) lookup(key string) (string, bool) {
	k := normKey(key)
	if k == "" {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.managed[k]; ok {
		return v, true
	}
	if s.isOrchestrator {
		if v, ok := s.values[models.SettingScopeOrchestrator][k]; ok {
			return v, true
		}
	}
	if v, ok := s.local[k]; ok {
		return v, true
	}
	return "", false
}

// ── catalog + validation ────────────────────────────────────────────────────

// normKey canonicalizes a setting key (env-var style upper-case).
func normKey(key string) string { return strings.ToUpper(strings.TrimSpace(key)) }

// Definition returns the catalogued definition for a key (nil = not
// catalogued — the whitelist reject).
func (s *Service) Definition(key string) *models.SettingDefinition {
	k := normKey(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if d, ok := s.definitions[k]; ok {
		return &d
	}
	return nil
}

// validateValue checks a candidate value against the definition: type parse,
// size ceiling, control chars and the optional allow-list. Returns the
// normalized (trimmed) value.
func validateValue(def *models.SettingDefinition, value string) (string, error) {
	if len(value) > maxSettingValueBytes {
		return "", fmt.Errorf("value for %s exceeds %d bytes", def.Key, maxSettingValueBytes)
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("value for %s contains control characters", def.Key)
		}
	}
	switch def.Type {
	case models.SettingValueTypeBool:
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return "", fmt.Errorf("value %q for %s is not a boolean", value, def.Key)
		}
		return strconv.FormatBool(parsed), nil
	case models.SettingValueTypeInt:
		if _, err := strconv.Atoi(strings.TrimSpace(value)); err != nil {
			return "", fmt.Errorf("value %q for %s is not an integer", value, def.Key)
		}
		return strings.TrimSpace(value), nil
	default:
		trimmed := strings.TrimSpace(value)
		if def.ValidValues != "" {
			allowed := false
			for _, candidate := range strings.Split(def.ValidValues, ",") {
				if strings.EqualFold(strings.TrimSpace(candidate), trimmed) {
					trimmed = strings.TrimSpace(candidate)
					allowed = true
					break
				}
			}
			if !allowed {
				return "", fmt.Errorf("value %q for %s must be one of %s", value, def.Key, def.ValidValues)
			}
		}
		return trimmed, nil
	}
}

// checkScopeAllowed enforces AppliesTo against the target scope: host-scope
// writes may not target orchestrator-only keys and vice versa.
func checkScopeAllowed(def *models.SettingDefinition, scope string) error {
	switch {
	case scope == models.SettingScopeOrchestrator:
		if def.AppliesTo == models.SettingAppliesToHost {
			return fmt.Errorf("setting %s applies to hosts only", def.Key)
		}
	case scope == models.SettingScopeHosts || models.IsHostSettingScope(scope):
		if def.AppliesTo == models.SettingAppliesToOrchestrator {
			return fmt.Errorf("setting %s applies to the orchestrator only", def.Key)
		}
	default:
		return fmt.Errorf("invalid settings scope %q", scope)
	}
	return nil
}

// ── orchestrator writes (source of truth for host scopes) ──────────────────

// SetScopeValue upserts one (scope,key) override and bumps the epoch. Host
// scopes additionally mark affected sync states pending so the UI shows
// drift without waiting for the push.
func (s *Service) SetScopeValue(scope, key, value, updatedBy string) (int64, error) {
	scope = strings.TrimSpace(scope)
	k := normKey(key)
	def := s.Definition(k)
	if def == nil {
		return 0, fmt.Errorf("setting %s is not configurable through the settings service", k)
	}
	if err := checkScopeAllowed(def, scope); err != nil {
		return 0, err
	}
	normalized, err := validateValue(def, value)
	if err != nil {
		return 0, err
	}
	revision, err := s.st.upsertValue(&models.SettingValue{
		Scope: scope, Key: k, Value: normalized, UpdatedBy: updatedBy,
	})
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	if s.values[scope] == nil {
		s.values[scope] = map[string]string{}
	}
	s.values[scope][k] = normalized
	s.meta.Revision = revision
	s.mu.Unlock()

	if scope != models.SettingScopeOrchestrator {
		s.markPendingForScope(scope)
	}
	return revision, nil
}

// DeleteScopeValue hard-removes one override. Returns existed.
func (s *Service) DeleteScopeValue(scope, key string) (bool, error) {
	scope = strings.TrimSpace(scope)
	k := normKey(key)
	existed, revision, err := s.st.deleteValue(scope, k)
	if err != nil {
		return false, err
	}
	if !existed {
		return false, nil
	}
	s.mu.Lock()
	if s.values[scope] != nil {
		delete(s.values[scope], k)
	}
	s.meta.Revision = revision
	s.mu.Unlock()
	if scope != models.SettingScopeOrchestrator {
		s.markPendingForScope(scope)
	}
	return true, nil
}

// markPendingForScope flips drift status to pending: hosts scope touches all
// known states, a host scope touches just that host. Best-effort bookkeeping
// — sync correctness never depends on it (hash compare is authoritative).
func (s *Service) markPendingForScope(scope string) {
	if scope == models.SettingScopeHosts {
		states, err := s.st.loadSyncStates()
		if err != nil {
			return
		}
		for _, st := range states {
			_ = s.st.markSyncPending(st.HostID)
		}
		return
	}
	if models.IsHostSettingScope(scope) {
		_ = s.st.markSyncPending(models.HostIDFromSettingScope(scope))
	}
}

// ClearHost drops all overrides and the sync state of a host (unregister).
func (s *Service) ClearHost(hostID string) error {
	if err := s.st.deleteHostValues(hostID); err != nil {
		return err
	}
	scope := models.HostSettingScope(hostID)
	s.mu.Lock()
	delete(s.values, scope)
	s.mu.Unlock()
	return s.st.deleteSyncState(hostID)
}

// ── host-local writes (direct UI against a host) ───────────────────────────

// SetLocalValue upserts a node-local override (source=local). Orchestrator-
// only keys are rejected; the value stays valid (shadowed, not deleted) even
// when a managed value hides it.
func (s *Service) SetLocalValue(key, value string) (int64, error) {
	k := normKey(key)
	def := s.Definition(k)
	if def == nil {
		return 0, fmt.Errorf("setting %s is not configurable through the settings service", k)
	}
	if def.AppliesTo == models.SettingAppliesToOrchestrator {
		return 0, fmt.Errorf("setting %s applies to the orchestrator only", def.Key)
	}
	normalized, err := validateValue(def, value)
	if err != nil {
		return 0, err
	}
	revision, err := s.st.upsertLocalSetting(k, normalized)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	s.local[k] = normalized
	s.meta.Revision = revision
	s.mu.Unlock()
	return revision, nil
}

// DeleteLocalValue removes a node-local override. Returns existed.
func (s *Service) DeleteLocalValue(key string) (bool, error) {
	k := normKey(key)
	existed, revision, err := s.st.deleteLocalSetting(k)
	if err != nil {
		return false, err
	}
	if !existed {
		return false, nil
	}
	s.mu.Lock()
	delete(s.local, k)
	s.meta.Revision = revision
	s.mu.Unlock()
	return true, nil
}

// ── reads ───────────────────────────────────────────────────────────────────

// Definitions lists the catalog (sorted by key).
func (s *Service) Definitions() []models.SettingDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.SettingDefinition, 0, len(s.definitions))
	for _, d := range s.definitions {
		out = append(out, d)
	}
	sortDefinitions(out)
	return out
}

// Values lists every stored scoped override (orchestrator view).
func (s *Service) Values() []models.SettingValue {
	rows, err := s.st.loadValues()
	if err != nil {
		return nil
	}
	return rows
}

// NodeInfo reports this node's settings posture (managed marker + epoch).
func (s *Service) NodeInfo() models.SettingMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.meta
}

func sortDefinitions(defs []models.SettingDefinition) {
	for i := 1; i < len(defs); i++ {
		for j := i; j > 0 && defs[j-1].Key > defs[j].Key; j-- {
			defs[j-1], defs[j] = defs[j], defs[j-1]
		}
	}
}

// osLookupEnv indirection keeps source probing testable.
var osLookupEnv = os.LookupEnv
