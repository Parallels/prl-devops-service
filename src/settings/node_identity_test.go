package settings

import (
	"errors"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	str "github.com/stretchr/testify/require"
)

// The per-replica identity is the one part of the settings domain that can break
// the placement feature from the inside: replicas share this table, so an identity
// that lands in a shared scope, or that two machines end up sharing, merges their
// ledger and their job ownership — the over-commit bug arriving through the UI.
// These tests pin the four things that stop that.

const (
	machineA = "AA-AA-AA-AA"
	machineB = "BB-BB-BB-BB"
)

// identityService returns an orchestrator service whose own machine is machineA,
// which is what makes the node layer addressable at all.
func identityService(t *testing.T) *Service {
	t.Helper()

	svc := newTestService(t, true, false)
	svc.nodeScope = models.NodeSettingScope(machineA)

	return svc
}

func writeIdentityHook(t *testing.T, hook IdentityApplyHook) {
	t.Helper()

	SetIdentityApplyHook(hook)
	t.Cleanup(func() { SetIdentityApplyHook(nil) })
}

func TestNodeScopedValueBeatsTheSharedScope(t *testing.T) {
	svc := identityService(t)

	_, err := svc.SetScopeValue(models.SettingScopeOrchestrator,
		constants.PlacementReservationTTLEnvVar, "30", "admin")
	str.NoError(t, err)
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.PlacementReservationTTLEnvVar, "90", "admin")
	str.NoError(t, err)

	// The replica's own row wins over the deployment-wide one, which is the whole
	// reason the scope exists.
	value, ok := svc.lookup(constants.PlacementReservationTTLEnvVar)
	str.True(t, ok)
	str.Equal(t, "90", value)

	// A peer replica's row must not be visible here at all.
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineB),
		constants.PlacementReservationTTLEnvVar, "10", "admin")
	str.NoError(t, err)
	value, _ = svc.lookup(constants.PlacementReservationTTLEnvVar)
	str.Equal(t, "90", value)

	own, ok := svc.NodeScopedValue(constants.PlacementReservationTTLEnvVar)
	str.True(t, ok)
	str.Equal(t, "90", own)
}

func TestNodeScopedValuesNeverReachHosts(t *testing.T) {
	// Regression guard for the worst failure this feature could ship with: a
	// per-replica row travelling to hosts would hand one replica's identity (or
	// its tuning) to every machine in the fleet. The bundle builder selects
	// scopes by allow-list, so node rows are excluded — and this test is what
	// keeps it that way when someone adds a layer.
	svc := identityService(t)

	_, err := svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.PlacementReservationTTLEnvVar, "77", "admin")
	str.NoError(t, err)
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.LOG_LEVEL_ENV_VAR, "debug", "admin")
	str.NoError(t, err)

	bundle, _, _ := svc.DesiredBundleForHost("host-1")
	str.NotContains(t, bundle, constants.PlacementReservationTTLEnvVar)
	str.NotContains(t, bundle, constants.LOG_LEVEL_ENV_VAR)

	// A host-scope row with the same key does reach it: the exclusion is about
	// the node layer, not about the keys.
	_, err = svc.SetScopeValue(models.HostSettingScope("host-1"),
		constants.LOG_LEVEL_ENV_VAR, "error", "admin")
	str.NoError(t, err)

	bundle, _, _ = svc.DesiredBundleForHost("host-1")
	str.Equal(t, "error", bundle[constants.LOG_LEVEL_ENV_VAR])
}

func TestIdentityIsRefusedOutsideItsOwnMachineScope(t *testing.T) {
	svc := identityService(t)

	// The trap this closes: one shared row read by every replica attached to this
	// database. The error has to do more than refuse — it has to say where the
	// value belongs and what this machine is called, or the operator retries it
	// somewhere else that is also wrong.
	_, err := svc.SetScopeValue(models.SettingScopeOrchestrator,
		constants.OrchestratorNodeIDEnvVar, "orchestrator-1", "admin")
	str.Error(t, err)
	str.Contains(t, err.Error(), models.SettingScopeNodePrefix)
	str.Contains(t, err.Error(), constants.OrchestratorNodeIDEnvVar)

	_, err = svc.SetScopeValue(models.HostSettingScope("host-1"),
		constants.OrchestratorNodeIDEnvVar, "orchestrator-1", "admin")
	str.Error(t, err)

	// Its proper home works.
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.OrchestratorNodeIDEnvVar, "orchestrator-1", "admin")
	str.NoError(t, err)
	value, ok := svc.NodeScopedValue(constants.OrchestratorNodeIDEnvVar)
	str.True(t, ok)
	str.Equal(t, "orchestrator-1", value)
}

func TestTwoMachinesCannotClaimOneIdentity(t *testing.T) {
	svc := identityService(t)

	_, err := svc.SetScopeValue(models.NodeSettingScope(machineB),
		constants.OrchestratorNodeIDEnvVar, "same-name", "admin")
	str.NoError(t, err)

	// Two machines under one name is the collision again, arrived at through the
	// UI instead of through a shared row: node-scoped recovery would sweep a live
	// replica's work and the ledger could not tell them apart.
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.OrchestratorNodeIDEnvVar, "same-name", "admin")
	str.Error(t, err)
	str.Contains(t, err.Error(), machineB) // actionable only if it says who holds it

	// Saving the same value again on the same machine is not a collision with
	// itself — an operator re-saving a form must not be refused.
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineB),
		constants.OrchestratorNodeIDEnvVar, "same-name", "admin")
	str.NoError(t, err)

	// Comparison ignores surrounding space, which is what an input field leaves
	// behind.
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.OrchestratorNodeIDEnvVar, "  same-name  ", "admin")
	str.Error(t, err)
}

func TestNodeScopeRejectsHostOnlyKeys(t *testing.T) {
	svc := identityService(t)

	// A host-side gate written to a replica's scope would be stored, read by
	// nobody and look like it had taken effect.
	_, err := svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.VMCPUAdmissionEnabledEnvVar, "true", "admin")
	str.Error(t, err)
	str.Contains(t, err.Error(), "hosts only")

	// Keys that mean something on either side are fine.
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.LOG_LEVEL_ENV_VAR, "debug", "admin")
	str.NoError(t, err)
}

func TestIdentityHookSeesTheOutgoingAndIncomingNames(t *testing.T) {
	svc := identityService(t)

	type call struct{ previous, current string }
	var calls []call
	writeIdentityHook(t, func(previous, current string) error {
		calls = append(calls, call{previous: previous, current: current})

		return nil
	})

	_, err := svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.OrchestratorNodeIDEnvVar, "first", "admin")
	str.NoError(t, err)
	str.Equal(t, []call{{previous: "", current: "first"}}, calls)

	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.OrchestratorNodeIDEnvVar, "second", "admin")
	str.NoError(t, err)
	str.Equal(t, "first", calls[1].previous)
	str.Equal(t, "second", calls[1].current)

	// Deleting the override tells the hook to fall back, not to keep a name.
	_, err = svc.DeleteScopeValue(models.NodeSettingScope(machineA), constants.OrchestratorNodeIDEnvVar)
	str.NoError(t, err)
	str.Equal(t, "second", calls[2].previous)
	str.Equal(t, "", calls[2].current)

	// An ordinary key under the same scope runs no identity work.
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.PlacementReservationTTLEnvVar, "45", "admin")
	str.NoError(t, err)
	str.Len(t, calls, 3)
}

func TestAFailedRenameIsReportedAndTheValueStillStands(t *testing.T) {
	// The honest reading of a partial failure: the row is stored and will apply at
	// next boot, so the caller must be told the rows did not follow rather than
	// being shown either a silent success or a rollback that did not happen.
	svc := identityService(t)
	writeIdentityHook(t, func(_, _ string) error {
		return errBoom
	})

	revision, err := svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.OrchestratorNodeIDEnvVar, "adopted-anyway", "admin")
	str.Error(t, err)
	str.Greater(t, revision, int64(0))

	value, ok := svc.NodeScopedValue(constants.OrchestratorNodeIDEnvVar)
	str.True(t, ok)
	str.Equal(t, "adopted-anyway", value)
}

func TestNodeLayerIsInertOnAHost(t *testing.T) {
	// A host keeps its settings in its own database; a node row there would be a
	// scope nobody can write and nothing reads.
	svc := newTestService(t, false, true)
	svc.nodeScope = models.NodeSettingScope(machineA)

	_, err := svc.st.upsertValue(&models.SettingValue{
		Scope: models.NodeSettingScope(machineA),
		Key:   constants.PlacementReservationTTLEnvVar, Value: "99",
	})
	str.NoError(t, err)
	str.NoError(t, svc.load())

	_, ok := svc.NodeScopedValue(constants.PlacementReservationTTLEnvVar)
	str.False(t, ok)
	_, ok = svc.lookup(constants.PlacementReservationTTLEnvVar)
	str.False(t, ok)
}

func TestPlacementKeysAreCataloguedWithTheirCompiledDefaults(t *testing.T) {
	svc := newTestService(t, true, false)

	// The catalog and the getters must not be able to drift: the defaults here
	// come from the same constants config/placement.go reads.
	expected := map[string]struct {
		defaultValue string
		appliesTo    string
		restart      bool
	}{
		constants.PlacementReservationEnabledEnvVar: {"true", models.SettingAppliesToOrchestrator, false},
		constants.PlacementReservationTTLEnvVar:     {"120", models.SettingAppliesToOrchestrator, true},
		constants.PlacementReconcileIntervalEnvVar:  {"30", models.SettingAppliesToOrchestrator, true},
		constants.OrchestratorNodeIDEnvVar:          {"", models.SettingAppliesToOrchestrator, false},
		constants.VMCPUAdmissionEnabledEnvVar:       {"false", models.SettingAppliesToHost, false},
		constants.VMCPUOvercommitRatioEnvVar:        {"1", models.SettingAppliesToHost, false},
	}

	for key, want := range expected {
		def := svc.Definition(key)
		str.NotNil(t, def, "%s must be configurable through the settings service", key)
		str.Equal(t, want.defaultValue, def.DefaultValue, key)
		str.Equal(t, want.appliesTo, def.AppliesTo, key)
		str.Equal(t, want.restart, def.RequiresRestart, key)
		str.NotEmpty(t, def.Description, key)
	}
}

func TestPlacementOverridesValidateAtTheBoundary(t *testing.T) {
	svc := identityService(t)

	// Bad values are refused where they are typed, not discovered at the sweeper.
	_, err := svc.SetScopeValue(models.SettingScopeOrchestrator,
		constants.PlacementReservationTTLEnvVar, "not-a-number", "admin")
	str.Error(t, err)

	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.OrchestratorNodeIDEnvVar, "two\nlines", "admin")
	str.Error(t, err)

	_, err = svc.SetScopeValue(models.HostSettingScope("host-1"),
		constants.VMCPUAdmissionEnabledEnvVar, "yes-please", "admin")
	str.Error(t, err)

	_, err = svc.SetScopeValue(models.HostSettingScope("host-1"),
		constants.VMCPUAdmissionEnabledEnvVar, "true", "admin")
	str.NoError(t, err)
}

func TestNodeScopeHelpersRoundTrip(t *testing.T) {
	scope := models.NodeSettingScope("  " + machineA + " ")
	str.Equal(t, "node:"+machineA, scope)
	str.True(t, models.IsNodeSettingScope(scope))
	str.Equal(t, machineA, models.MachineIDFromNodeSettingScope(scope))

	// The host scopes are not node scopes, and a bare prefix is not a machine.
	str.False(t, models.IsNodeSettingScope(models.HostSettingScope("host-1")))
	str.False(t, models.IsNodeSettingScope(models.SettingScopeHosts))
	str.Equal(t, "", models.MachineIDFromNodeSettingScope(models.SettingScopeOrchestrator))
}

func TestIdentityKeyRecognition(t *testing.T) {
	str.True(t, IsNodeIdentityKey(constants.OrchestratorNodeIDEnvVar))
	str.True(t, IsNodeIdentityKey(" orchestrator_node_id ")) // keys are case-insensitive everywhere else too
	str.False(t, IsNodeIdentityKey(constants.PlacementReservationTTLEnvVar))
}

func TestEffectiveViewShowsTheNodeLayer(t *testing.T) {
	svc := identityService(t)

	_, err := svc.SetScopeValue(models.SettingScopeOrchestrator,
		constants.PlacementReservationTTLEnvVar, "30", "admin")
	str.NoError(t, err)
	_, err = svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.PlacementReservationTTLEnvVar, "90", "admin")
	str.NoError(t, err)

	entry := findEffective(svc.OrchestratorNodeView(), constants.PlacementReservationTTLEnvVar)
	str.NotNil(t, entry)
	str.Equal(t, "90", entry.Effective)
	str.Equal(t, SourceNode, entry.Source)
	str.NotNil(t, entry.NodeValue)
	str.NotNil(t, entry.OrchestratorValue)
	// The shared value is still listed but marked as hidden, so the UI can explain
	// why editing it changed nothing.
	str.Equal(t, SourceNode, entry.ShadowedBy)
}

func TestEffectiveViewWithoutAMachineIDShowsNoNodeLayer(t *testing.T) {
	svc := newTestService(t, true, false) // nodeScope unset: unidentifiable machine

	_, err := svc.SetScopeValue(models.NodeSettingScope(machineA),
		constants.PlacementReservationTTLEnvVar, "90", "admin")
	str.NoError(t, err)

	entry := findEffective(svc.OrchestratorNodeView(), constants.PlacementReservationTTLEnvVar)
	str.NotNil(t, entry)
	str.Nil(t, entry.NodeValue)
	str.Equal(t, "120", entry.Effective) // the catalog default, not a stranger's row
}

var errBoom = errors.New("rows did not follow")
