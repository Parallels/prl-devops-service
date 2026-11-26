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
//	orchestrator node:    managed > node-scope > orchestrator-scope > local
//
// orchestrator-scope covers setting_values rows with scope "orchestrator"
// (the service's own overrides); host-scope rows never touch the orchestrator
// process. node-scope is the value addressed to *this* replica by machine id
// ("node:<machine-id>") — see node_identity.go for why a per-replica key must
// not live in the shared orchestrator scope. Below the DB layers sit env >
// flag > config file > catalog default (handled by the untouched GetKey
// fallback).
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
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// Value source labels in the effective view (spec constant set).
const (
	SourceManaged      = "managed_db"
	SourceOrchestrator = "orchestrator_db"
	// SourceNode marks a value addressed to this replica alone by machine id.
	SourceNode       = "node_db"
	SourceHosts      = "hosts_db"
	SourceHost       = "host_db"
	SourceLocal      = "local_db"
	SourceEnv        = "env"
	SourceConfigFile = "config_file"
	SourceDefault    = "default"
)

const maxSettingValueBytes = 4096

type Service struct {
	ctx basecontext.ApiContext
	st  *store

	mu          sync.RWMutex
	definitions map[string]models.SettingDefinition // catalogued keys (upper)
	values      map[string]map[string]string        // scope -> key -> value
	managed     map[string]string                   // node_settings: orchestrator source
	local       map[string]string                   // node_settings: local source
	meta        models.SettingMeta

	isOrchestrator bool
	isHost         bool

	// nodeScope is the scope this process reads its own per-replica values
	// from ("node:<machine-id>"), empty on a host or on a machine that cannot
	// be identified. Computed once at init: the machine does not change under a
	// running process, and reading it per lookup would shell out on every
	// GetKey.
	nodeScope string
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
	// Only a replica addresses itself by machine: a host's settings rows live in
	// its own database, so a node scope there would be a scope nobody else can
	// ever write and nothing gains from it.
	if svc.isOrchestrator {
		svc.nodeScope = nodeScopeFor()
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
		// This replica's own value beats the shared one: the whole point of the
		// node scope is that two replicas attached to one database can disagree.
		if s.nodeScope != "" {
			if v, ok := s.values[s.nodeScope][k]; ok {
				return v, true
			}
		}
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
	case models.IsNodeSettingScope(scope):
		// A replica's own row can only carry keys that mean something on a
		// replica. A host-scoped key stored here would be read by nobody — the
		// hosts pull their values from the host scopes, not from a peer
		// replica — so it would sit in the table looking like it had taken
		// effect.
		if def.AppliesTo == models.SettingAppliesToHost {
			return fmt.Errorf("setting %s applies to hosts only", def.Key)
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

	// The replica identity has exactly one home. Accepting it in the shared
	// scopes too would store a value that reads as configured and does nothing
	// — or worse, would be honoured by every replica attached to this database.
	if IsNodeIdentityKey(k) && !models.IsNodeSettingScope(scope) {
		return 0, fmt.Errorf("setting %s is per-replica: write it to scope %s<machine-id>; this machine's id is %q",
			def.Key, models.SettingScopeNodePrefix, helpers.MachineID())
	}

	renamed := false
	previousIdentity := ""
	if IsNodeIdentityKey(k) {
		previous, err := s.guardNodeIdentityWrite(scope, normalized)
		if err != nil {
			return 0, err
		}
		previousIdentity, renamed = previous, true
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

	// Adopted only after the row is durable: a process running under an
	// identity the database does not record would stamp work nobody can
	// attribute to it after a restart.
	if renamed {
		if err := applyNodeIdentity(previousIdentity, normalized); err != nil {
			return revision, err
		}
	}

	return revision, nil
}

// DeleteScopeValue hard-removes one override. Returns existed.
func (s *Service) DeleteScopeValue(scope, key string) (bool, error) {
	scope = strings.TrimSpace(scope)
	k := normKey(key)

	renamed := IsNodeIdentityKey(k) && models.IsNodeSettingScope(scope)
	previousIdentity := ""
	if renamed {
		s.mu.RLock()
		previousIdentity = s.values[scope][k]
		s.mu.RUnlock()
	}

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
	// "" means the identity falls back to what this machine derives for itself;
	// the hook owns that chain, this package does not know it exists.
	if renamed {
		if err := applyNodeIdentity(previousIdentity, ""); err != nil {
			return true, err
		}
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
