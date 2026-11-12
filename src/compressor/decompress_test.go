package compressor

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
)

func TestDetectFileType(t *testing.T) {
	testCases := []struct {
		name         string
		header       []byte
		expectedType string
		expectedErr  error
	}{
		{
			// A gzip stream is reported as tar.gz: detectFileType does not peek into the
			// decompressed payload to tell a bare gzip member apart from a tar.gz.
			name:         "gzip file",
			header:       []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
			expectedType: "tar.gz",
			expectedErr:  nil,
		},
		{
			name:         "truncated gzip file",
			header:       []byte{0x1F},
			expectedType: "unknown",
			expectedErr:  errors.New("file format not recognized as gzip or tar"),
		},
		{
			name: "tar file",
			header: func() []byte {
				tarHeader := make([]byte, 512)
				copy(tarHeader[257:], []byte("ustar\x00"))
				return tarHeader
			}(),
			expectedType: "tar",
			expectedErr:  nil,
		},
		{
			name: "tar file with some leading data",
			header: func() []byte {
				tarHeader := make([]byte, 512)
				copy(tarHeader[257:], []byte("ustar\x00"))
				return tarHeader
			}(),
			expectedType: "tar",
			expectedErr:  nil,
		},
		{
			name:         "truncated tar file",
			header:       []byte("ustar"),
			expectedType: "unknown",
			expectedErr:  errors.New("file format not recognized as gzip or tar"),
		},
		{
			name:         "unknown file",
			header:       []byte{0x01, 0x02, 0x03, 0x04},
			expectedType: "unknown",
			expectedErr:  errors.New("file format not recognized as gzip or tar"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileType, err := detectFileType(tc.header)

			if fileType != tc.expectedType {
				t.Errorf("Expected file type %q, but got %q", tc.expectedType, fileType)
			}

			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error %q, but got nil", tc.expectedErr)
				} else if err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error %q, but got %q", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("Expected nil error, but got: %q", err)
			}
		})
	}
}

// tarEntry describes a single entry of an archive built by the tests below.
type tarEntry struct {
	name     string
	typeflag byte
	linkname string
	content  string
}

func buildTar(t *testing.T, entries []tarEntry) *bytes.Buffer {
	t.Helper()

	buffer := &bytes.Buffer{}
	writer := tar.NewWriter(buffer)

	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.typeflag,
			Mode:     0o644,
		}
		if entry.typeflag == tar.TypeDir {
			header.Mode = 0o755
		}
		if entry.typeflag != tar.TypeDir && entry.typeflag != tar.TypeSymlink {
			header.Size = int64(len(entry.content))
		}
		if entry.typeflag == tar.TypeSymlink {
			header.Linkname = entry.linkname
		}

		if err := writer.WriteHeader(header); err != nil {
			t.Fatalf("failed to write header for %q: %v", entry.name, err)
		}
		if entry.content != "" {
			if _, err := writer.Write([]byte(entry.content)); err != nil {
				t.Fatalf("failed to write content for %q: %v", entry.name, err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	return buffer
}

// An entry name that walks above the destination has to fail the extraction, not just
// be logged.
func TestProcessTarFileRejectsTraversal(t *testing.T) {
	base := t.TempDir()
	destination := filepath.Join(base, "out")
	victim := filepath.Join(base, "outside.txt")

	archive := buildTar(t, []tarEntry{
		{name: "../outside.txt", typeflag: tar.TypeReg, content: "pwned"},
	})

	if err := processTarFile(basecontext.NewBaseContext(), tar.NewReader(archive), destination, "", ""); err == nil {
		t.Fatal("an entry walking above the destination was accepted")
	}

	if _, err := os.Stat(victim); err == nil {
		t.Fatalf("wrote %q outside the destination", victim)
	}
}

// An archive that plants a symlink pointing somewhere else on the host must not be able
// to use it to move the following entry's write out of the destination.
func TestProcessTarFileDoesNotMaterialiseLinks(t *testing.T) {
	base := t.TempDir()
	destination := filepath.Join(base, "out")
	outside := filepath.Join(base, "outside")
	victim := filepath.Join(outside, "pwned.txt")

	if err := os.MkdirAll(outside, 0o750); err != nil {
		t.Fatalf("failed to create the outside directory: %v", err)
	}

	archive := buildTar(t, []tarEntry{
		{name: "escape", typeflag: tar.TypeSymlink, linkname: outside},
		{name: "escape/pwned.txt", typeflag: tar.TypeReg, content: "pwned"},
	})

	if err := processTarFile(basecontext.NewBaseContext(), tar.NewReader(archive), destination, "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(victim); err == nil {
		t.Fatalf("wrote %q outside the destination", victim)
	}

	info, err := os.Lstat(filepath.Join(destination, "escape"))
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("created a symlink inside the destination")
	}
}

// The ordinary case still has to work, including a relative destination and directories
// that are only implied by a file name rather than listed as entries.
func TestProcessTarFileExtractsArchive(t *testing.T) {
	t.Chdir(t.TempDir())

	archive := buildTar(t, []tarEntry{
		{name: "vm", typeflag: tar.TypeDir},
		{name: "vm/hd.img", typeflag: tar.TypeReg, content: "disk contents"},
		{name: "vm/notes/readme.txt", typeflag: tar.TypeReg, content: "hello"},
	})

	if err := processTarFile(basecontext.NewBaseContext(), tar.NewReader(archive), "out", "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for file, expected := range map[string]string{
		filepath.Join("out", "vm", "hd.img"):              "disk contents",
		filepath.Join("out", "vm", "notes", "readme.txt"): "hello",
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		if string(data) != expected {
			t.Errorf("%s contains %q, expected %q", file, string(data), expected)
		}
	}
}

// macOS packs pull from catalog storage as /VmInfo.pvi style names, because the pack
// writers trimmed the pack root off each walked path and left a leading separator.
// Such a name is made relative to the destination, the way GNU tar strips a leading
// "/"; it is not a reason to fail the pull.
func TestProcessTarFileExtractsLeadingSlashEntries(t *testing.T) {
	base := t.TempDir()
	destination := filepath.Join(base, "out")

	archive := buildTar(t, []tarEntry{
		{name: "/VmInfo.pvi", typeflag: tar.TypeReg, content: "vm info"},
		{name: "/Sub/dir/file.txt", typeflag: tar.TypeReg, content: "nested"},
	})

	if err := processTarFile(basecontext.NewBaseContext(), tar.NewReader(archive), destination, "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for file, expected := range map[string]string{
		filepath.Join(destination, "VmInfo.pvi"):             "vm info",
		filepath.Join(destination, "Sub", "dir", "file.txt"): "nested",
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		if string(data) != expected {
			t.Errorf("%s contains %q, expected %q", file, string(data), expected)
		}
	}
}

// The leading separator is stripped, but a name that then walks above the destination
// is still refused.
func TestProcessTarFileRejectsTraversalAfterLeadingSlash(t *testing.T) {
	base := t.TempDir()
	destination := filepath.Join(base, "out")
	victim := filepath.Join(base, "outside.txt")

	archive := buildTar(t, []tarEntry{
		{name: "/../../outside.txt", typeflag: tar.TypeReg, content: "pwned"},
	})

	if err := processTarFile(basecontext.NewBaseContext(), tar.NewReader(archive), destination, "", ""); err == nil {
		t.Fatal("an entry walking above the destination was accepted")
	}

	if _, err := os.Stat(victim); err == nil {
		t.Fatalf("wrote %q outside the destination", victim)
	}
}

// A pack built by Compress must carry relative entry names and round-trip back through
// the decompressor: writers and this reader are the same codebase, and an archive that
// only the old lenient extractor accepted is a latent pull failure.
func TestCompressRoundTripsWithRelativeNames(t *testing.T) {
	base := t.TempDir()

	source := filepath.Join(base, "vm")
	if err := os.MkdirAll(filepath.Join(source, "Sub"), 0o750); err != nil {
		t.Fatalf("failed to create the source tree: %v", err)
	}
	for file, content := range map[string]string{
		"VmInfo.pvi":   "vm info",
		"Sub/file.txt": "nested",
	} {
		if err := os.WriteFile(filepath.Join(source, filepath.FromSlash(file)), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", file, err)
		}
	}

	tarFilePath, err := Compress(basecontext.NewBaseContext(), source, "vm.tar", base, "")
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		t.Fatalf("failed to open %s: %v", tarFilePath, err)
	}
	defer tarFile.Close()

	names := map[string]bool{}
	reader := tar.NewReader(tarFile)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read the archive: %v", err)
		}
		if strings.HasPrefix(header.Name, "/") {
			t.Errorf("entry %q keeps a leading separator", header.Name)
		}
		names[header.Name] = true
	}
	for _, want := range []string{"VmInfo.pvi", "Sub/file.txt"} {
		if !names[want] {
			t.Errorf("the archive does not contain the entry %q (entries: %v)", want, names)
		}
	}

	if _, err := tarFile.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to rewind the archive: %v", err)
	}
	destination := filepath.Join(base, "out")
	if err := processTarFile(basecontext.NewBaseContext(), tar.NewReader(tarFile), destination, "", ""); err != nil {
		t.Fatalf("decompressing a freshly compressed pack failed: %v", err)
	}
	for file, expected := range map[string]string{
		filepath.Join(destination, "VmInfo.pvi"):      "vm info",
		filepath.Join(destination, "Sub", "file.txt"): "nested",
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		if string(data) != expected {
			t.Errorf("%s contains %q, expected %q", file, string(data), expected)
		}
	}
}
