package legacy_migrator

// Detector units (task 01): AC-1 answers + the D-3 purity promise — the
// detector never touches the filesystem beyond os.Stat. Legacy
// IsDataFileEmpty is banned; these tests pin the replacement's behaviour.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestDetectLegacyStore(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()

	t.Run("absent: silent skip, no error", func(t *testing.T) {
		dir := t.TempDir()
		before := dirEntries(t, dir)
		det, err := DetectLegacyStore(ctx, filepath.Join(dir, "data.json"))
		require.NoError(t, err) // AC-1: not an error condition
		assert.False(t, det.Present)
		assert.Equal(t, "absent", det.Reason)
		assert.Equal(t, before, dirEntries(t, dir), "detector must not create anything")
	})

	t.Run("empty file: fresh install skip", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "data.json")
		require.NoError(t, os.WriteFile(path, nil, 0o600))
		before := dirEntries(t, dir)
		det, err := DetectLegacyStore(ctx, path)
		require.NoError(t, err)
		assert.False(t, det.Present)
		assert.Equal(t, "empty", det.Reason)
		assert.Equal(t, before, dirEntries(t, dir))
	})

	t.Run("present: size and mtime reported", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "data.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"users":[]}`), 0o600))
		before := dirEntries(t, dir)
		det, err := DetectLegacyStore(ctx, path)
		require.NoError(t, err)
		assert.True(t, det.Present)
		assert.Equal(t, "present", det.Reason)
		assert.Greater(t, det.SizeBytes, int64(0))
		assert.False(t, det.ModTime.IsZero())
		assert.Equal(t, before, dirEntries(t, dir))
	})

	t.Run("unreadable path is an error, not a silent skip", func(t *testing.T) {
		// A stat error that is NOT IsNotExist (here: a non-directory parent)
		// must surface — startup policy, not the detector, decides.
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file"), nil, 0o600))
		_, err := DetectLegacyStore(ctx, filepath.Join(dir, "file", "data.json"))
		assert.Error(t, err)
	})
}
