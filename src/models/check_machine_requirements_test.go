package models

import (
	"encoding/json"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/require"
)

func TestCheckMachineRequirementsRequestValidation(t *testing.T) {
	var request CheckMachineRequirementsRequest
	require.Error(t, request.Validate())
	require.Contains(t, request.Validate().Error(), "either catalog_id or requirements must be provided")

	request = CheckMachineRequirementsRequest{CatalogId: "ubuntu", Requirements: &MachineRequirements{Cpu: 2}}
	require.Error(t, request.Validate())
	require.Contains(t, request.Validate().Error(), "cannot both be provided")

	request = CheckMachineRequirementsRequest{CatalogId: "ubuntu", Architecture: "arm64", Connection: "http://catalog", CatalogManagerId: "mgr"}
	require.Error(t, request.Validate())
	require.Contains(t, request.Validate().Error(), "connection and catalog_manager_id cannot both be provided")

	// catalog_id requires an architecture, and the version defaults to latest.
	request = CheckMachineRequirementsRequest{CatalogId: "ubuntu"}
	require.Error(t, request.Validate())
	require.Contains(t, request.Validate().Error(), "architecture is required")

	request = CheckMachineRequirementsRequest{CatalogId: "ubuntu", Architecture: "arm64"}
	require.NoError(t, request.Validate())
	require.Equal(t, constants.LATEST_TAG, request.Version)

	// Requirements get the same defaults the VM creation path uses.
	request = CheckMachineRequirementsRequest{Requirements: &MachineRequirements{}}
	require.NoError(t, request.Validate())
	require.Equal(t, 2, request.Requirements.Cpu)
	require.Equal(t, 2048, request.Requirements.Memory)
}

func TestCheckMachineRequirementsRequestTargetTags(t *testing.T) {
	// A clean list survives validation untouched.
	request := CheckMachineRequirementsRequest{Requirements: &MachineRequirements{}, TargetTags: []string{"pool=build", "os=sonoma"}}
	require.NoError(t, request.Validate())
	require.Equal(t, []string{"pool=build", "os=sonoma"}, request.TargetTags)

	// Entries are trimmed and deduplicated case-insensitively, first spelling
	// and order kept.
	request = CheckMachineRequirementsRequest{Requirements: &MachineRequirements{}, TargetTags: []string{" Pool=Build ", "pool=build"}}
	require.NoError(t, request.Validate())
	require.Equal(t, []string{"Pool=Build"}, request.TargetTags)

	// An entry that is empty after trimming is a client error.
	request = CheckMachineRequirementsRequest{Requirements: &MachineRequirements{}, TargetTags: []string{"pool=build", "  "}}
	err := request.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "target_tags cannot contain empty values")

	// An empty list normalizes to nil: no constraint, never match-nothing.
	request = CheckMachineRequirementsRequest{Requirements: &MachineRequirements{}, TargetTags: []string{}}
	require.NoError(t, request.Validate())
	require.Nil(t, request.TargetTags)
}

func TestMachineRequirementsGetSpecsType(t *testing.T) {
	require.Equal(t, "pvm", (&MachineRequirements{}).GetSpecsType())
	require.Equal(t, "macvm", (&MachineRequirements{Os: "macos"}).GetSpecsType())
	require.Equal(t, "macvm", (&MachineRequirements{Os: "macOS"}).GetSpecsType())
	require.Equal(t, "macvm", (&MachineRequirements{Type: "macvm"}).GetSpecsType())
	require.Equal(t, "pvm", (&MachineRequirements{Os: "linux"}).GetSpecsType())
}

func TestCheckMachineRequirementsResponseSerialization(t *testing.T) {
	slimData, err := json.Marshal(CheckMachineRequirementsResponse{Passed: true})
	require.NoError(t, err)
	require.NotContains(t, string(slimData), `"hosts"`)
	require.NotContains(t, string(slimData), `"message"`)

	fullData, err := json.Marshal(CheckMachineRequirementsResponse{
		Passed:  false,
		Message: "No host meets the requirements",
		Hosts: []MachineRequirementsHostResult{{
			Host:   "host-a",
			Passed: false,
			Checks: []RequirementCheckResult{{Name: "cpu", Passed: false, Required: NewRequirementCheckValue(4), Available: NewRequirementCheckValue(2), Unit: "logical_cpus", Message: "CPU: required 4 logical CPUs, available 2"}},
		}},
	})
	require.NoError(t, err)
	require.Contains(t, string(fullData), `"hosts"`)
	require.Contains(t, string(fullData), `"required":4`)
	require.Contains(t, string(fullData), `"available":2`)
}
