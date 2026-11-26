package config

import (
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	str "github.com/stretchr/testify/require"
)

// Placement knobs are the operator's whole control surface for the admission
// gate, and every one of them fails quietly: a default that leans off removes a
// safeguard without a log line, and a mis-typed value that is taken literally
// can either disable enforcement or hold capacity for months. Seeding idiom
// copied from database_getters_test.go.

func unsetPlacementEnv(t *testing.T) {
	t.Helper()

	for _, k := range []string{
		constants.PlacementReservationEnabledEnvVar,
		constants.PlacementReservationTTLEnvVar,
		constants.PlacementReconcileIntervalEnvVar,
		constants.OrchestratorNodeIDEnvVar,
		constants.VMCPUAdmissionEnabledEnvVar,
		constants.VMCPUOvercommitRatioEnvVar,
	} {
		unsetEnv(t, k)
	}
}

func TestPlacementDefaults(t *testing.T) {
	unsetPlacementEnv(t)
	cfg := seedCfg(t, nil)

	str.True(t, cfg.PlacementReservationEnabled()) // enforcement on: the bug it prevents is silent
	str.Equal(t, constants.DefaultPlacementReservationTTLEnMinutes, cfg.PlacementReservationTTLEnMinutes())
	str.Equal(t, constants.DefaultPlacementReconcileIntervalSeconds, cfg.PlacementReconcileIntervalSeconds())
	str.Equal(t, 120, cfg.PlacementReservationTTLEnMinutes()) // longer than the slowest legitimate create
	str.Equal(t, 30, cfg.PlacementReconcileIntervalSeconds()) // not the ghost-job cadence
	str.False(t, cfg.VMCPUAdmissionEnabled())                 // host gate stays opt-in: oversubscribed hosts must not break on upgrade
	str.Equal(t, 1.0, cfg.VMCPUOvercommitRatio())             // no oversubscription unless asked for
	str.Empty(t, cfg.OrchestratorNodeID())                    // derive it, do not invent one
}

func TestPlacementReservationEnabled_RollbackLever(t *testing.T) {
	for _, raw := range []string{"false", "0", "off", "no", "F", " No "} {
		t.Setenv(constants.PlacementReservationEnabledEnvVar, raw)
		str.Falsef(t, seedCfg(t, nil).PlacementReservationEnabled(), "raw=%q", raw)
	}

	for _, raw := range []string{"true", "1", "on", "yes", "Y", " Yes "} {
		t.Setenv(constants.PlacementReservationEnabledEnvVar, raw)
		str.Truef(t, seedCfg(t, nil).PlacementReservationEnabled(), "raw=%q", raw)
	}
}

func TestPlacementReservationTTL_TolerantOfNoise(t *testing.T) {
	for _, raw := range []string{"", "0", "-5", "not-a-number", "abc"} {
		t.Setenv(constants.PlacementReservationTTLEnvVar, raw)
		str.Equalf(t, constants.DefaultPlacementReservationTTLEnMinutes,
			seedCfg(t, nil).PlacementReservationTTLEnMinutes(), "raw=%q must not disable the backstop", raw)
	}
}

func TestPlacementReservationTTL_ValidOverride(t *testing.T) {
	t.Setenv(constants.PlacementReservationTTLEnvVar, "45")
	str.Equal(t, 45, seedCfg(t, nil).PlacementReservationTTLEnMinutes())
}

func TestPlacementReservationTTL_AbsurdValueIsATypoNotAnInstruction(t *testing.T) {
	// Past a day, "eventually reclaimed" means never.
	t.Setenv(constants.PlacementReservationTTLEnvVar, "100000")
	str.Equal(t, constants.DefaultPlacementReservationTTLEnMinutes,
		seedCfg(t, nil).PlacementReservationTTLEnMinutes())
}

func TestPlacementReconcileInterval(t *testing.T) {
	t.Setenv(constants.PlacementReconcileIntervalEnvVar, "5")
	str.Equal(t, 5, seedCfg(t, nil).PlacementReconcileIntervalSeconds())

	for _, raw := range []string{"", "0", "-1", "4000", "junk"} {
		t.Setenv(constants.PlacementReconcileIntervalEnvVar, raw)
		str.Equalf(t, constants.DefaultPlacementReconcileIntervalSeconds,
			seedCfg(t, nil).PlacementReconcileIntervalSeconds(), "raw=%q", raw)
	}
}

func TestPlacementNodeIDIsTrimmed(t *testing.T) {
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "  qa-A  ")
	str.Equal(t, "qa-A", seedCfg(t, nil).OrchestratorNodeID())

	t.Setenv(constants.OrchestratorNodeIDEnvVar, "   ")
	str.Empty(t, seedCfg(t, nil).OrchestratorNodeID())
}

func TestVMCPUOvercommitRatio(t *testing.T) {
	t.Setenv(constants.VMCPUOvercommitRatioEnvVar, "2")
	str.Equal(t, 2.0, seedCfg(t, nil).VMCPUOvercommitRatio())

	// NaN, Inf, zero and junk all fall back to "no oversubscription" rather than
	// to an infinite capacity that would admit everything.
	for _, raw := range []string{"", "0", "-1", "NaN", "+Inf", "1e400", "abc"} {
		t.Setenv(constants.VMCPUOvercommitRatioEnvVar, raw)
		str.Equalf(t, constants.DefaultVMCPUOvercommitRatio, seedCfg(t, nil).VMCPUOvercommitRatio(), "raw=%q", raw)
	}
}
