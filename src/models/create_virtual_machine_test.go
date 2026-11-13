package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateVirtualMachineRequestTargetTags(t *testing.T) {
	// A clean list survives validation untouched.
	request := CreateVirtualMachineRequest{Name: "vm", Architecture: "arm64", TargetTags: []string{"pool=build", "os=sonoma"}}
	require.NoError(t, request.Validate())
	require.Equal(t, []string{"pool=build", "os=sonoma"}, request.TargetTags)

	// Entries are trimmed and deduplicated case-insensitively, first spelling
	// and order kept.
	request = CreateVirtualMachineRequest{Name: "vm", Architecture: "arm64", TargetTags: []string{" Pool=Build ", "pool=build"}}
	require.NoError(t, request.Validate())
	require.Equal(t, []string{"Pool=Build"}, request.TargetTags)

	// An entry that is empty after trimming is a client error.
	request = CreateVirtualMachineRequest{Name: "vm", Architecture: "arm64", TargetTags: []string{"pool=build", ""}}
	err := request.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "target_tags cannot contain empty values")

	// An empty list normalizes to nil: no constraint, never match-nothing.
	request = CreateVirtualMachineRequest{Name: "vm", Architecture: "arm64", TargetTags: []string{}}
	require.NoError(t, request.Validate())
	require.Nil(t, request.TargetTags)

	// Absent stays absent — request shape is byte-compatible for callers that
	// never use the field.
	request = CreateVirtualMachineRequest{Name: "vm", Architecture: "arm64"}
	require.NoError(t, request.Validate())
	require.Nil(t, request.TargetTags)
}

func TestNormalizeTagConstraints(t *testing.T) {
	normalized, err := NormalizeTagConstraints(nil)
	require.NoError(t, err)
	require.Nil(t, normalized)

	normalized, err = NormalizeTagConstraints([]string{"a", "A ", " b", "a"})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, normalized)

	_, err = NormalizeTagConstraints([]string{"pool=build", " "})
	require.Error(t, err)
	require.Contains(t, err.Error(), "target_tags cannot contain empty values")

	normalized, err = NormalizeTagConstraints([]string{})
	require.NoError(t, err)
	require.Nil(t, normalized)
}
