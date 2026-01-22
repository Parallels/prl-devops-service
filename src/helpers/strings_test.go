package helpers

import (
	"testing"
)

func TestGetSizeByteFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		// Real-world df -h output cases (from macOS)
		{"926Gi", "926Gi", 926 * 1024 * 1024 * 1024, false},
		{"11Gi", "11Gi", 11 * 1024 * 1024 * 1024, false},
		{"4.3G", "4.3G", 4.3 * 1024 * 1024 * 1024, false},
		{"453k", "453k", 453 * 1024, false},
		{"1.8Ti", "1.8Ti", 1.8 * 1024 * 1024 * 1024 * 1024, false},

		// Each unit type (case-insensitive)
		{"1TB", "1TB", 1024 * 1024 * 1024 * 1024, false},
		{"1gb", "1gb", 1024 * 1024 * 1024, false},
		{"512MB", "512mb", 512 * 1024 * 1024, false},
		{"256KB", "256kb", 256 * 1024, false},
		{"1024B", "1024b", 1024, false},

		// Single-letter suffixes
		{"1T", "1T", 1024 * 1024 * 1024 * 1024, false},
		{"16G", "16G", 16 * 1024 * 1024 * 1024, false},
		{"512M", "512M", 512 * 1024 * 1024, false},
		{"256K", "256k", 256 * 1024, false},

		// Decimal values
		{"0.5gb", "0.5gb", 0.5 * 1024 * 1024 * 1024, false},
		{"2.25kb", "2.25kb", 2.25 * 1024, false},

		// Whitespace handling
		{" 1gb ", " 1gb ", 1024 * 1024 * 1024, false},
		{"1 gb", "1 gb", 1024 * 1024 * 1024, false},

		// Error cases
		{"Invalid unit", "1xyz", -1, true},
		{"No unit", "1024", -1, true},
		{"Invalid number", "abcgb", -1, true},
		{"Empty string", "", -1, true},
		{"Only unit", "gb", -1, true},

		// Parallels Desktop configuration format
		{"PD memory 8192Mb", "8192Mb", 8192 * 1024 * 1024, false},
		{"PD hdd0 131072Mb", "131072Mb", 131072 * 1024 * 1024, false},
		{"PD video 0Mb", "0Mb", 0, false},

		// Edge cases
		{"Zero", "0gb", 0, false},
		{"Negative", "-1gb", -1 * 1024 * 1024 * 1024, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetSizeByteFromString(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("For input '%s', expected %.2f, but got %.2f", tt.input, tt.expected, result)
				}
			}
		})
	}
}
