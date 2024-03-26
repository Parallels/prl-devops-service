package migrations

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersion(t *testing.T) {
	tests := []struct {
		version       string
		targetVersion string
		expected      VersionComparisonResult
		expectedErr   error
	}{
		{
			version:       "0.2.3",
			targetVersion: "1.2.3",
			expected:      VersionLowerThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.3",
			targetVersion: "1.2.3",
			expected:      VersionEqualToTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.3",
			targetVersion: "1.2.4",
			expected:      VersionLowerThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.4",
			targetVersion: "1.2.3",
			expected:      VersionHigherThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.3",
			targetVersion: "1.3.0",
			expected:      VersionLowerThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.3.0",
			targetVersion: "1.2.3",
			expected:      VersionHigherThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.3",
			targetVersion: "2.0.0",
			expected:      VersionLowerThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "2.0.0",
			targetVersion: "1.2.3",
			expected:      VersionHigherThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.3",
			targetVersion: "1.2",
			expected:      VersionHigherThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2",
			targetVersion: "1.2.3",
			expected:      VersionLowerThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.3",
			targetVersion: "1",
			expected:      VersionHigherThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1",
			targetVersion: "1.2.3",
			expected:      VersionLowerThanTarget,
			expectedErr:   nil,
		},
		{
			version:       "1.2.3",
			targetVersion: "1.2.3.4",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "1.2.3.4",
			targetVersion: "1.2.3",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "1.2.3",
			targetVersion: "1.2.3.0",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "1.2.3.0",
			targetVersion: "1.2.3",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "a.b.c",
			targetVersion: "c.d.e",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "0.b.c",
			targetVersion: "a.d.e",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "0.b.c",
			targetVersion: "0.d.e",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "0.0.c",
			targetVersion: "0.d.e",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "0.0.c",
			targetVersion: "0.0.e",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
		{
			version:       "0.0.0",
			targetVersion: "0.0.e",
			expected:      VersionLowerThanTarget,
			expectedErr:   errors.New("invalid version format"),
		},
	}

	for _, test := range tests {
		result, err := compareVersions(test.version, test.targetVersion)
		assert.Equalf(t, test.expectedErr, err, "expected %v, got %v for version %v and targetVersion %v", test.expectedErr, err, test.version, test.targetVersion)
		assert.Equalf(t, test.expected, result, "expected %v, got %v for version %v and targetVersion %v", test.expected, result, test.version, test.targetVersion)
	}
}
