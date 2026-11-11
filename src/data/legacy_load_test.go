package data

// Loader units for LoadLegacyData (#424 additive helper, AD-5): side-effect
// freedom is the contract — dir snapshots around every mutation-shaped case,
// memoryDatabase pollution audited directly (same-package test).
//
// Log capture uses ONE TestMain-attached file (the 424-01 lesson:
// EnableLogFile handles are global; per-test attachments onto deleted temp
// dirs drop later writes across the package binary).

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var loaderLogFile string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "legacy-load-logs-")
	if err != nil {
		panic(err)
	}
	loaderLogFile = filepath.Join(dir, "legacy-load.log")
	basecontext.NewRootBaseContext().EnableLogFile(loaderLogFile)
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

const loaderFixtures = "../database/legacy_migrator/testdata"

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(loaderFixtures, name))
	require.NoError(t, err)
	return b
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func waitForLoaderLog(t *testing.T, substr string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		b, _ := os.ReadFile(loaderLogFile)
		if strings.Contains(string(b), substr) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("captured log never contained %q — got: %q", substr, string(b))
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func assertAllCollections(t *testing.T, d Data) {
	t.Helper()
	assert.Equal(t, "1.0.0", d.Schema.Version)
	require.NotNil(t, d.Configuration)
	assert.Equal(t, "config-1", d.Configuration.ID)
	require.NotNil(t, d.ReverseProxy)
	assert.Equal(t, "8443", d.ReverseProxy.Port)
	assert.Len(t, d.Users, 3)
	assert.Len(t, d.Claims, 4)
	assert.Len(t, d.Roles, 2)
	assert.Len(t, d.ApiKeys, 2)
	assert.Len(t, d.PackerTemplates, 1)
	assert.Len(t, d.ManifestsCatalog, 1)
	assert.Len(t, d.OrchestratorHosts, 1)
	assert.Len(t, d.HostsVMSnapshots, 1)
	assert.Len(t, d.ReverseProxyHosts, 1)
	assert.Len(t, d.CatalogManagers, 1)
	assert.Len(t, d.Jobs, 2)
	assert.Len(t, d.VMSnapshots, 1)
	assert.Len(t, d.EnrollmentTokens, 1)
	assert.Len(t, d.UserConfigs, 1)
	// embedded graph survived decode
	assert.Len(t, d.Users[1].Roles, 2)
	assert.Len(t, d.Users[1].Claims, 2)
	assert.Len(t, d.Roles[0].Claims, 2)
	assert.Len(t, d.Jobs[0].Steps, 2)
}

func TestLoadLegacyDataPlaintextGolden(t *testing.T) {
	// Strongest form of the spec's "no globals" row: earlier data-package
	// tests legitimately pollute memoryDatabase via NewJsonDatabase, so a
	// bare nil-before assertion is suite-order dependent. Instead: park the
	// global (restored on exit, tests are sequential), forcing ANY write by
	// the loader — including re-assignment — to fail this test.
	saved := memoryDatabase
	memoryDatabase = nil
	t.Cleanup(func() { memoryDatabase = saved })

	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, readFixture(t, "golden.json"), 0o600))
	before := dirNames(t, dir)

	d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.NoError(t, err)
	assertAllCollections(t, d)

	assert.Equal(t, before, dirNames(t, dir), "loader must not create/modify/delete anything")
	assert.Nil(t, memoryDatabase, "no singleton pollution")
}

func TestLoadLegacyDataPartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, readFixture(t, "partial.json"), 0o600))

	d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.NoError(t, err)
	require.Len(t, d.Users, 1)
	assert.Equal(t, "solo", d.Users[0].Username)
	assert.Nil(t, d.Configuration)
	assert.Nil(t, d.ReverseProxy)
	assert.Empty(t, d.Jobs)
	assert.Empty(t, d.Roles)
}

func TestLoadLegacyDataEncryptedGolden(t *testing.T) {
	golden := readFixture(t, "golden.json")
	keyDir := t.TempDir()
	pemPath := filepath.Join(keyDir, "k.pem")
	require.NoError(t, security.GenPrivateRsaKey(pemPath, 2048))
	pemBytes, err := os.ReadFile(pemPath)
	require.NoError(t, err)

	cipher, err := security.EncryptString(string(pemBytes), string(golden))
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, cipher, 0o600))
	before := dirNames(t, dir)
	sumBefore := sha256.Sum256(cipher)

	// The env carries the BASE64 of the PEM — exactly what Config decodes.
	t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, base64.StdEncoding.EncodeToString(pemBytes))

	d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.NoError(t, err)
	assertAllCollections(t, d)

	// Struct equality vs the plaintext decode of the same golden.
	var want Data
	require.NoError(t, json.Unmarshal(golden, &want))
	assert.Equal(t, want, d)

	assert.Equal(t, before, dirNames(t, dir))
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, sumBefore, sha256.Sum256(after), "ciphertext untouched")
}

func TestLoadLegacyDataEncryptedWithoutKey(t *testing.T) {
	golden := readFixture(t, "golden.json")
	keyDir := t.TempDir()
	pemPath := filepath.Join(keyDir, "k.pem")
	require.NoError(t, security.GenPrivateRsaKey(pemPath, 2048))
	pemBytes, err := os.ReadFile(pemPath)
	require.NoError(t, err)
	cipher, err := security.EncryptString(string(pemBytes), string(golden))
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, cipher, 0o600))
	t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, "")

	require.NotPanics(t, func() {
		d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
		require.Error(t, err)
		assert.Empty(t, d.Users)
		assert.Contains(t, err.Error(), "ENCRYPTION_PRIVATE_KEY")
	})
}

func TestLoadLegacyDataCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, readFixture(t, "corrupt.json"), 0o600))
	t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, "")

	d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.Error(t, err)
	assert.Equal(t, Data{}, d)
}

func TestLoadLegacyDataAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.ErrorIs(t, err, ErrNoLegacyDataFile)
	assert.Equal(t, 404, errors.GetSystemErrorCode(err))
	assert.Equal(t, Data{}, d)
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr), "absent file is never created")
}

func TestLoadLegacyDataZeroByte(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.ErrorIs(t, err, ErrLegacyDataFileEmpty)
	assert.Equal(t, 404, errors.GetSystemErrorCode(err))
	assert.Equal(t, Data{}, d)

	// The absent-path branch of the same sentinel: still zero files created.
	missing := filepath.Join(dir, "sub", "data.json")
	_, err = LoadLegacyData(basecontext.NewRootBaseContext(), missing)
	assert.ErrorIs(t, err, ErrNoLegacyDataFile)
}

func TestLoadLegacyDataResidualSaveFallback(t *testing.T) {
	golden := readFixture(t, "golden.json")
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, []byte("}{ not json ~~~"), 0o600))
	savePath := path + ".save"
	require.NoError(t, os.WriteFile(savePath, golden, 0o600))
	t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, "")

	d, err := LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.NoError(t, err)
	require.Len(t, d.Users, 3)

	waitForLoaderLog(t, "decoded residual") // warn: fallback taken, visibly

	// Read-only recovery: nothing consumed, nothing rewritten (OQ-1 freeze).
	_, statErr := os.Stat(savePath)
	require.NoError(t, statErr, ".save must survive the read")
	mainBytes, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "}{ not json ~~~", string(mainBytes), "main file bytes unchanged")
}
