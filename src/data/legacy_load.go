package data

// #424 additive helper — approved in PBI #424 tasks (AD-5's practical form:
// reuse the legacy DECODE rules without touching JsonDatabase). main.go is
// frozen; the sentinels below live here, same package, same idiom.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/cjlapao/common-go/helper"
)

var (
	ErrNoLegacyDataFile    = errors.NewWithCode("legacy data file not found", 404)
	ErrLegacyDataFileEmpty = errors.NewWithCode("legacy data file is empty", 404)
)

// LoadLegacyData reads a legacy data file into the Data struct WITHOUT side
// effects: no file creation, no backup (contrast loadFromFile's j.Backup at
// main.go:535), no save, no singleton, no goroutines, no RecoverOngoingJobs
// (main.go:578). Read-only residual recovery (*.save → *.save_bak → *.panic,
// Load's order at main.go:135-157) decodes the newest glob match as a
// FALLBACK ONLY — never SaveNow, never removeAllBackupSavedFiles (contrast
// recoverFromResidualSaveFiles, main.go:476-484): the legacy store is frozen
// (OQ-1). Decoding mirrors loadFromFile (main.go:550-570) EXACTLY — keep the
// two in sync if the legacy format ever changes; migration is a one-time
// consumer, legacy stays canonical.
//
// #424 additive helper — approved in PBI #424 tasks. Do not call from legacy
// paths.
func LoadLegacyData(ctx basecontext.ApiContext, filename string) (Data, error) {
	var empty Data
	info, err := os.Stat(filename)
	if err != nil || info.Size() == 0 {
		if os.IsNotExist(err) {
			return empty, ErrNoLegacyDataFile
		}
		if err != nil {
			return empty, errors.NewFromError(err)
		}
		return empty, ErrLegacyDataFileEmpty
	}
	raw, err := helper.ReadFromFile(filename) // same reader as loadFromFile:539
	if err != nil {
		return empty, errors.NewFromError(err)
	}
	data, err := decodeLegacyBytes(ctx, raw, filename) // plaintext → encrypted (decrypt), main.go:550-570
	if err == nil {
		return data, nil
	}
	for _, glob := range []string{"*.save", "*.save_bak", "*.panic"} { // read-only recovery
		matches, _ := filepath.Glob(filename + glob)
		if len(matches) == 0 {
			continue
		}
		r2, rErr := helper.ReadFromFile(matches[len(matches)-1]) // newest, legacy semantics
		if rErr != nil || len(r2) == 0 {
			continue
		}
		if d2, dErr := decodeLegacyBytes(ctx, r2, matches[len(matches)-1]); dErr == nil {
			ctx.LogWarnf("[LegacyMigration] main file unreadable; decoded residual %s (read-only)", matches[len(matches)-1])
			return d2, nil
		}
	}
	return empty, err
}

// decodeLegacyBytes is the exact main.go:550-570 sequence, extracted so
// loadFromFile could call it later — but this task does NOT refactor
// loadFromFile (frozen; no behaviour risk).
func decodeLegacyBytes(ctx basecontext.ApiContext, content []byte, origin string) (Data, error) {
	var data Data
	err := json.Unmarshal(content, &data)
	if err == nil {
		return data, nil
	}
	cfg := config.Get()
	if cfg.EncryptionPrivateKey() == "" {
		return data, errors.NewWithCodeAndDescription(
			fmt.Sprintf("failed to parse legacy data file %s: %v (encrypted file? set ENCRYPTION_PRIVATE_KEY)", origin, err), 500,
			"the legacy data store may be encrypted; the migration needs the encryption key")
	}
	plain, err := security.DecryptString(cfg.EncryptionPrivateKey(), content)
	if err != nil {
		return data, errors.NewFromError(err)
	}
	if err := json.Unmarshal([]byte(plain), &data); err != nil {
		return data, errors.NewFromError(err)
	}
	return data, nil
}
