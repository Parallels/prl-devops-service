package settings

// Per-replica identity through the settings service.
//
// A replica's identity is the one setting that cannot simply be a row in the
// settings table, because replicas share that table: a single orchestrator-scope
// ORCHESTRATOR_NODE_ID would be read by every process attached to the database,
// and they would all report the same owner. That is not cosmetic — node-scoped
// job recovery would sweep a live replica's running creates, and the reservation
// ledger could not tell the replicas apart. So the value is addressed per
// machine ("node:<machine-id>", see models.NodeSettingScope) and read only by
// the process whose hardware matches.
//
// Three things guard the obvious ways to get that wrong:
//
//   - the scope is accepted only for keys that may legitimately vary per
//     replica (checkScopeAllowed in main.go);
//   - two machines may not be mapped onto the same identity (rejectDuplicate),
//     because the collision then happens through the UI instead of through a
//     shared row;
//   - the bundle pushed to hosts selects scopes by allow-list, so a node row is
//     never delivered to a host (covered by TestNodeScopedValuesNeverReachHosts).

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
)

// IdentityApplyHook is called after a per-replica identity has been stored, so
// the process can adopt it instead of waiting for a restart. It receives the
// identity being left behind and the one taking its place ("" when the override
// was deleted and the identity falls back to what the machine derives for
// itself).
//
// It is a hook rather than an import because the settings service cannot reach
// into the placement ledger or the jobs table: those are separate bounded
// contexts and the dependency points the other way. Same shape as
// config.SetOverrideResolver.
type IdentityApplyHook func(previous, current string) error

var (
	identityHookMu sync.RWMutex
	identityHook   IdentityApplyHook
)

// SetIdentityApplyHook registers (or clears, with nil) the live-apply hook.
// Called once from startup in orchestrator mode; last writer wins, which is the
// same contract the config resolver hook offers.
func SetIdentityApplyHook(hook IdentityApplyHook) {
	identityHookMu.Lock()
	defer identityHookMu.Unlock()

	identityHook = hook
}

func registeredIdentityHook() IdentityApplyHook {
	identityHookMu.RLock()
	defer identityHookMu.RUnlock()

	return identityHook
}

// IsNodeIdentityKey reports whether a setting key addresses this replica's
// identity, which is the key the guard and the live-apply hook act on.
func IsNodeIdentityKey(key string) bool {
	return normKey(key) == normKey(constants.OrchestratorNodeIDEnvVar)
}

// nodeScopeFor builds the scope this process reads its own per-replica values
// from. Empty when the machine cannot be identified — in which case per-machine
// settings are simply unavailable, which is honest, rather than falling back to
// a shared scope and merging replicas.
func nodeScopeFor() string {
	machine := helpers.MachineID()
	if machine == "" {
		return ""
	}

	return models.NodeSettingScope(machine)
}

// NodeScopedValue returns the value stored for this machine alone, ignoring the
// shared orchestrator scope. It is what the identity resolver asks first, so
// that the settings layer outranks the environment pin without this package
// having to import config (config imports helpers, helpers is below us).
func (s *Service) NodeScopedValue(key string) (string, bool) {
	if s == nil || !s.isOrchestrator {
		return "", false
	}

	scope := s.nodeScope
	if scope == "" {
		return "", false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.values[scope][normKey(key)]

	return value, ok
}

// nodeScopeOfThisMachine is used by the write paths to decide whether a write
// targets this process's own identity.
func (s *Service) nodeScopeOfThisMachine(scope string) bool {
	return s.nodeScope != "" && scope == s.nodeScope
}

// nodeIdentityHolder returns the machine id that already claims this identity
// under another replica's scope, or "" when nobody does. The current scope is
// excluded so a rewrite of the same value is not a collision with itself.
func (s *Service) nodeIdentityHolder(identity, exceptScope string) string {
	k := normKey(constants.OrchestratorNodeIDEnvVar)

	s.mu.RLock()
	defer s.mu.RUnlock()

	for scope, keys := range s.values {
		if !models.IsNodeSettingScope(scope) || scope == exceptScope {
			continue
		}

		if strings.EqualFold(strings.TrimSpace(keys[k]), identity) {
			return models.MachineIDFromNodeSettingScope(scope)
		}
	}

	return ""
}

// guardNodeIdentityWrite validates a per-replica identity write and returns the
// identity being replaced, which the caller hands to the apply hook. Only the
// identity key reaches it: the callers check IsNodeIdentityKey first, because
// every other key under a node scope is an ordinary override with no ledger
// behind it.
func (s *Service) guardNodeIdentityWrite(scope, value string) (string, error) {
	if !models.IsNodeSettingScope(scope) {
		return "", nil
	}

	previous := ""
	s.mu.RLock()
	previous = s.values[scope][normKey(constants.OrchestratorNodeIDEnvVar)]
	s.mu.RUnlock()

	if holder := s.nodeIdentityHolder(value, scope); holder != "" {
		// The message names both machines: "another replica already answers to
		// this name" is only actionable if the operator can tell which one.
		return "", fmt.Errorf("identity %q is already used by node %s; identities must be unique per machine",
			value, holder)
	}

	return previous, nil
}

// applyNodeIdentity runs the live-apply hook after a store. It returns the
// hook's failure so the caller can report it: the value *is* stored, and the
// next boot will adopt it, so this is a partial success and must be told as one
// rather than either hidden or dressed up as a rollback.
func applyNodeIdentity(previous, current string) error {
	hook := registeredIdentityHook()
	if hook == nil {
		return nil
	}

	return hook(previous, current)
}
