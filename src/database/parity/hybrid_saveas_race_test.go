//go:build race

package parity

// Under -race the SaveAs probe is skipped — legacy's RunBackup goroutine
// reads j.filename unguarded, so any test-side SaveAs trips the detector
// without temporal overlap. See hybrid_saveas_test.go (the !race variant)
// for the probe itself and the T5 backlog rationale.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
)

func probeSaveAsDelegate(t *testing.T, ctx basecontext.ApiContext, legacy *data.JsonDatabase, sql datainterfaces.DatabaseService) {
	t.Log("SaveAs delegate probe skipped under -race (legacy backup-goroutine race, backlog T5)")
}
