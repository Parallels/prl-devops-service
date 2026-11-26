package config

// Placement domain getters (VM placement reservations). Tolerant-int posture
// copied from DbGhostJobTimeoutMinutes (database.go:46): a missing, zero,
// negative or overflowing value falls back to the default instead of disabling
// enforcement, because a mis-set knob must never silently remove a safeguard.

import (
	"math"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
)

// PlacementReservationEnabled reports whether the orchestrator claims capacity
// before dispatching a create. This is the documented rollback lever: setting
// PLACEMENT_RESERVATION_ENABLED=false reproduces pre-feature placement.
func (c *Config) PlacementReservationEnabled() bool {
	raw := strings.TrimSpace(strings.ToLower(c.GetKey(constants.PlacementReservationEnabledEnvVar)))
	if raw == "" {
		return constants.DefaultPlacementReservationEnabled
	}

	return helpers.StringToBool(raw)
}

// PlacementReservationTTLEnMinutes returns the reservation lease in minutes.
// Out-of-range is treated as a typo, not an instruction: a negative or absurd
// value falls back to the default rather than disabling the backstop or holding
// capacity forever, with the same posture as the reconcile interval below.
func (c *Config) PlacementReservationTTLEnMinutes() int {
	ttl := c.GetIntKey(constants.PlacementReservationTTLEnvVar)
	if ttl <= 0 {
		return constants.DefaultPlacementReservationTTLEnMinutes
	}

	if ttl > constants.MaxPlacementReservationTTLEnMinutes || int64(ttl) > int64((1<<63-1)/60000000000) {
		if c.ctx != nil {
			c.ctx.LogWarnf("Invalid %s value %d; using %d minutes instead",
				constants.PlacementReservationTTLEnvVar, ttl, constants.DefaultPlacementReservationTTLEnMinutes)
		}

		return constants.DefaultPlacementReservationTTLEnMinutes
	}

	return ttl
}

// PlacementReconcileIntervalSeconds returns the sweeper cadence in seconds.
func (c *Config) PlacementReconcileIntervalSeconds() int {
	interval := c.GetIntKey(constants.PlacementReconcileIntervalEnvVar)
	if interval <= 0 || interval > 3600 {
		return constants.DefaultPlacementReconcileIntervalSeconds
	}

	return interval
}

// OrchestratorNodeID returns the pinned replica identity ("" = derive it).
func (c *Config) OrchestratorNodeID() string {
	return strings.TrimSpace(c.GetKey(constants.OrchestratorNodeIDEnvVar))
}

// VMCPUAdmissionEnabled reports whether this host enforces CPU/memory capacity
// on create. Off unless explicitly enabled.
func (c *Config) VMCPUAdmissionEnabled() bool {
	raw := strings.TrimSpace(strings.ToLower(c.GetKey(constants.VMCPUAdmissionEnabledEnvVar)))
	if raw == "" {
		return false
	}

	return helpers.StringToBool(raw)
}

// VMCPUOvercommitRatio multiplies host capacity for the admission gate. Values
// <= 0 or unparseable fall back to 1.0 (no oversubscription) rather than to
// unlimited, and NaN/Inf are rejected outright: they would make the gate
// compare against an infinite capacity and admit everything.
func (c *Config) VMCPUOvercommitRatio() float64 {
	raw := strings.TrimSpace(c.GetKey(constants.VMCPUOvercommitRatioEnvVar))
	if raw == "" {
		return constants.DefaultVMCPUOvercommitRatio
	}

	ratio, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(ratio) || math.IsInf(ratio, 0) || ratio <= 0 {
		if c.ctx != nil {
			c.ctx.LogWarnf("Invalid %s value %q; using %g",
				constants.VMCPUOvercommitRatioEnvVar, raw, constants.DefaultVMCPUOvercommitRatio)
		}

		return constants.DefaultVMCPUOvercommitRatio
	}

	return ratio
}
