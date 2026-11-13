package legacy_migrator

// Detection (AC-1, plan D-3): the legacy-store presence answer computed
// from os.Stat ALONE. Callers (task 06) log absent/empty at debug level —
// "skip silently" in AC-1 terms is a nil error with Present=false, never a
// swallowed failure.

import (
	"os"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
)

// Detection is the AC-1 answer. Reason is human-readable ("absent", "empty", "present").
type Detection struct {
	Present   bool
	SizeBytes int64
	ModTime   time.Time
	Reason    string
}

// DetectLegacyStore stats the legacy file. Pure: os.Stat only.
// BANNED here (review gate): JsonDatabase.IsDataFileEmpty — it CREATES the
// file (src/data/main.go:645-655) — and anything that constructs JsonDatabase.
// `path` comes from the startup hook via JsonDatabase.Filename() — never
// re-derived here (single resolution point; AD-3/#420).
func DetectLegacyStore(ctx basecontext.ApiContext, path string) (Detection, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Detection{Reason: "absent"}, nil // AC-1: skip silently, no error
		}
		// Permission errors etc. are NOT "absent": a real error beats a
		// silent skip — the startup policy (D-9) decides fail-vs-continue.
		return Detection{}, errors.NewFromError(err)
	}
	if info.Size() == 0 {
		return Detection{Reason: "empty"}, nil // fresh install (brief: empty file = skip)
	}
	return Detection{Present: true, SizeBytes: info.Size(), ModTime: info.ModTime(), Reason: "present"}, nil
}
