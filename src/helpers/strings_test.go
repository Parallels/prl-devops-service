package helpers

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// Archive entry names are attacker controlled, so no name may resolve outside the
// directory it is extracted into. The sibling case is the one a plain string prefix
// test lets through: destination <base>/dest accepts <base>/dest-evil/x.txt.
func TestSanitizeArchivePath(t *testing.T) {
	base := t.TempDir()

	for _, dir := range []string{"dest", "dest-evil"} {
		if err := os.MkdirAll(filepath.Join(base, dir), 0o750); err != nil {
			t.Fatalf("failed to create %s: %v", dir, err)
		}
	}

	tests := []struct {
		name        string
		destination string
		entry       string
		expectError bool
		goos        string // run only on this GOOS
		notGoos     string // skip on this GOOS
	}{
		{name: "plain file", destination: "dest", entry: "file.txt"},
		{name: "nested file", destination: "dest", entry: "a/b/file.txt"},
		{name: "entry for the destination itself", destination: "dest", entry: "."},
		{name: "dot slash prefixed", destination: "dest", entry: "./file.txt"},
		{name: "leading separator is made relative", destination: "dest", entry: "/VmInfo.pvi"},
		{name: "absolute path is made relative", destination: "dest", entry: "/etc/passwd"},
		{name: "double leading separator is made relative", destination: "dest", entry: "//etc/passwd"},
		{name: "walks above the destination", destination: "dest", entry: "../../etc/passwd", expectError: true},
		{name: "walks above after descending", destination: "dest", entry: "a/../../outside.txt", expectError: true},
		{name: "walks above after the leading separator", destination: "dest", entry: "/../../etc/passwd", expectError: true},
		{name: "sibling sharing the destination prefix", destination: "dest", entry: "../dest-evil/x.txt", expectError: true},
		{name: "empty filepath after the strip", destination: "dest", entry: "/", expectError: true},
		{name: "windows volume path", destination: "dest", entry: `C:\Windows\win.ini`, goos: "windows"},
		{name: "windows volume only", destination: "dest", entry: "C:", expectError: true, goos: "windows"},
		{name: "no destination given", destination: "", entry: "file.txt", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.goos != "" && runtime.GOOS != tt.goos {
				t.Skipf("requires GOOS %s", tt.goos)
			}
			if tt.notGoos != "" && runtime.GOOS == tt.notGoos {
				t.Skipf("not applicable on GOOS %s", tt.notGoos)
			}

			destination := tt.destination
			if destination != "" {
				destination = filepath.Join(base, destination)
			}

			got, err := SanitizeArchivePath(destination, tt.entry)
			if tt.expectError {
				if err == nil {
					t.Fatalf("entry %q was accepted, resolved to %q", tt.entry, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for entry %q: %v", tt.entry, err)
			}

			root, err := filepath.Abs(destination)
			if err != nil {
				t.Fatalf("failed to resolve destination: %v", err)
			}
			rel, err := filepath.Rel(root, got)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				t.Fatalf("entry %q resolved outside the destination: %q is not inside %q", tt.entry, got, root)
			}
		})
	}
}

// Tar entry names are relative by the format's own definition. Trimming the walked
// root off a path instead leaves a leading separator on every entry, which older
// extractors only tolerated by accident, so writers must go through this helper.
func TestArchiveEntryName(t *testing.T) {
	tests := []struct {
		name        string
		root        string
		filePath    string
		want        string
		expectError bool
	}{
		{name: "top level file", root: "/pack", filePath: "/pack/VmInfo.pvi", want: "VmInfo.pvi"},
		{name: "nested file", root: "/pack", filePath: "/pack/sub/dir/file.txt", want: "sub/dir/file.txt"},
		{name: "the root itself", root: "/pack", filePath: "/pack", expectError: true},
		{name: "outside the root", root: "/pack", filePath: "/pack/../outside.txt", expectError: true},
		{name: "sibling sharing the root prefix", root: "/pack", filePath: "/pack-evil/x.txt", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ArchiveEntryName(tt.root, tt.filePath)
			if tt.expectError {
				if err == nil {
					t.Fatalf("path %q inside root %q was accepted, produced %q", tt.filePath, tt.root, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.filePath, err)
			}
			if got != tt.want {
				t.Errorf("ArchiveEntryName(%q, %q) = %q, want %q", tt.root, tt.filePath, got, tt.want)
			}
		})
	}
}
