package helpers

import (
	"testing"
)

func TestVersionLessThan(t *testing.T) {
	testCases := []struct {
		name     string
		v1       *Version
		v2       *Version
		expected bool
	}{
		{
			name:     "v1 < v2",
			v1:       NewVersion("1.0.0"),
			v2:       NewVersion("1.1.0"),
			expected: true,
		},
		{
			name:     "v1 > v2",
			v1:       NewVersion("2.0.0"),
			v2:       NewVersion("1.1.0"),
			expected: false,
		},
		{
			name:     "v1 = v2",
			v1:       NewVersion("1.1.0"),
			v2:       NewVersion("1.1.0"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.v1.LessThan(tc.v2)
			if actual != tc.expected {
				t.Errorf("Expected %t, but got %t", tc.expected, actual)
			}
		})
	}
}
