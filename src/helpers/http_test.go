package helpers

import "testing"

func TestCleanUrlSuffixAndPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no prefix or suffix",
			input:    "test",
			expected: "test",
		},
		{
			name:     "prefix only",
			input:    "/test",
			expected: "test",
		},
		{
			name:     "suffix only",
			input:    "test/",
			expected: "test",
		},
		{
			name:     "prefix and suffix",
			input:    "/test/",
			expected: "test",
		},
		{
			name:     "prefix and suffix",
			input:    "/test/",
			expected: "test",
		},
		{
			name:     "prefix and suffix",
			input:    "/test/",
			expected: "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := CleanUrlSuffixAndPrefix(tc.input)
			if actual != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, actual)
			}
		})
	}
}

func TestJoinUrl(t *testing.T) {
	testCases := []struct {
		name        string
		input       []string
		expected    string
		expectError bool
	}{
		{
			name:        "single segment",
			input:       []string{"test"},
			expected:    "test",
			expectError: false,
		},
		{
			name:        "multiple segments",
			input:       []string{"test", "path", "to", "resource"},
			expected:    "test/path/to/resource",
			expectError: false,
		},
		{
			name:        "empty segments",
			input:       []string{"test", "", "path", "", "to", "resource"},
			expected:    "test/path/to/resource",
			expectError: false,
		},
		{
			name:        "segments with prefix and suffix",
			input:       []string{"/test/", "/path/", "/to/", "/resource/"},
			expected:    "/test/path/to/resource/",
			expectError: false,
		},
		{
			name:        "segments with prefix and suffix and protocol",
			input:       []string{"http://", "localhost", "/test/", "/path/", "/to/", "/resource/"},
			expected:    "http://localhost/test/path/to/resource/",
			expectError: false,
		},
		{
			name:        "segments with prefix and suffix protocol and port",
			input:       []string{"http://", "localhost", ":80", "/test/", "/path/", "/to/", "/resource"},
			expected:    "http://localhost:80/test/path/to/resource",
			expectError: false,
		},
		{
			name:        "segments with prefix and suffix protocol and port and spaces",
			input:       []string{"http://", "localhost", ":80", "/test//", "", "/path/", "//to/", "/resource/"},
			expected:    "http://localhost:80/test/path/to/resource/",
			expectError: false,
		},
		{
			name:        "segments with wrong and suffix should generate error",
			input:       []string{"://", "httxxp://", "localhost", "", "/test/", "/path/", "/to/", "/resource/"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := JoinUrl(tc.input)
			if err != nil {
				tc.expectError = true
			} else {
				if actual.String() != tc.expected {
					t.Errorf("Expected %s, but got %s", tc.expected, actual.String())
				}
			}
		})
	}

	// Test empty segments
	_, err := JoinUrl([]string{})
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}
