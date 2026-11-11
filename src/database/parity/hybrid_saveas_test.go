//go:build !race

package parity

// SaveAs delegate probe (task 16 hybrid rep). Build-gated to !race:
// legacy SaveAs swaps j.filename WITHOUT the lock while RunBackup's
// goroutine (started inside NewJsonDatabase, backup.go:16-26) reads
// j.filename — a happens-before violation the race detector flags on ANY
// SaveAs after engine creation, even when the accesses are far apart in
// wall time. That is a PRE-EXISTING legacy defect (T5 backlog / #421
// upstream PR note), not something this port may fix (ER-5). Under -race
// the stub variant logs a skip; plain gate runs and the human run execute
// the full probe.

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
)

var saveSuffix = regexp.MustCompile(`\.\d+\.save `)

func probeSaveAsDelegate(t *testing.T, ctx basecontext.ApiContext, legacy *data.JsonDatabase, sql datainterfaces.DatabaseService) {
	// The delegate must write the legacy snapshot through, and surface
	// identical errors. Legacy QUIRK pinned through the delegate: SaveAs
	// REWRITES the target into the legacy DB's own directory
	// (main.go:203-206), keeping only the base name.
	snap := filepath.Join(t.TempDir(), "snap.json")
	if err := sql.SaveAs(ctx, snap); err != nil {
		t.Fatalf("delegate SaveAs: %v", err)
	}
	landed := filepath.Join(filepath.Dir(legacy.Filename()), filepath.Base(snap))
	if fi, err := os.Stat(landed); err != nil || fi.Size() == 0 {
		t.Fatalf("SaveAs did not reach the legacy writer at %s: %v", landed, err)
	}
	// Error identity through the same rewrite: Base("") = "." lands on the
	// legacy DIRECTORY — both handles must fail writing a directory alike.
	// The service wraps the error opaquely (no Unwrap chain) and the
	// message embeds a random .save temp suffix, so compare the messages
	// with that suffix canonicalised.
	lErr := legacy.SaveAs(ctx, "")
	sErr := sql.SaveAs(ctx, "")
	if (lErr == nil) != (sErr == nil) {
		t.Fatalf("SaveAs-dir-target drift: legacy=%v sql=%v", lErr, sErr)
	}
	if lErr != nil {
		norm := saveSuffix.ReplaceAllString(lErr.Error(), ".save ")
		if norm != saveSuffix.ReplaceAllString(sErr.Error(), ".save ") {
			t.Fatalf("SaveAs-dir-target mismatch:\nlegacy=%v\nsql  =%v", lErr, sErr)
		}
	}
}
