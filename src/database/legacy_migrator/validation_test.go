package legacy_migrator

// AC-6 units (task 05): the warn-only contract — every failure shape
// (mismatch, count-query error, formula drift) must surface as an issue +
// warn line and must NEVER surface as an error.

import (
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestValidateCountsHappyGolden(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	ctx := newRecordingCtx()

	rep, err := RunMigration(ctx, db, Options{SourcePath: path})
	require.NoError(t, err)
	assert.Empty(t, rep.Validation, "golden must validate clean")

	joined := strings.Join(ctx.all(), "\n")
	assert.Contains(t, joined, "legacy migration validation finished: all collections matched")
	oks := ctx.count(func(l string) bool { return strings.Contains(l, "validation ok: ") })
	assert.Equal(t, len(collectionPlan()), oks, "one info line per collection (AC-4 support)")
}

func TestValidateCountsSeededMismatch(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	ctx := basecontext.NewRootBaseContext()

	rep, err := RunMigration(ctx, db, Options{SourcePath: path})
	require.NoError(t, err)
	require.Empty(t, rep.Validation)

	// post-migration tampering: one extra row the legacy store never had
	require.NoError(t, db.Create(&models.ApiKey{ID: "ak-intruder", Name: "ghost"}).Error)

	rec := newRecordingCtx()
	issues := ValidateCounts(rec, db, mustLoad(t, path), collectionPlan())
	require.Len(t, issues, 1)
	assert.Equal(t, "ApiKeys", issues[0].Collection)
	assert.Equal(t, int64(2), issues[0].Expected)
	assert.Equal(t, int64(3), issues[0].Actual)
	assert.Equal(t, 1, rec.count(func(l string) bool {
		return strings.Contains(l, "validation MISMATCH ApiKeys: legacy=2 new=3")
	}))
}

func TestValidateCountsQueryError(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	ctx := basecontext.NewRootBaseContext()
	_, err := RunMigration(ctx, db, Options{SourcePath: path})
	require.NoError(t, err)

	require.NoError(t, db.Migrator().DropTable(&models.CatalogManager{})) // simulate broken count path

	rec := newRecordingCtx()
	var issues []ValidationIssue
	require.NotPanics(t, func() {
		issues = ValidateCounts(rec, db, mustLoad(t, path), collectionPlan())
	})
	var found bool
	for _, i := range issues {
		if i.Collection == "CatalogManagers" {
			found = true
			assert.Equal(t, int64(-1), i.Actual, "count-query failure recorded as Actual=-1")
		}
	}
	assert.True(t, found, "failed count surfaced as an issue")
	assert.Equal(t, 1, rec.count(func(l string) bool {
		return strings.Contains(l, "CatalogManagers count query failed")
	}))
}

func TestValidateCountsFormulaDrift(t *testing.T) {
	db := migDB(t)
	drifted := CollectionSpec{
		Name:        "Drifty",
		LegacyCount: func(d *data.Data) int { return 7 },
		Migrate:     nil, // never called by ValidateCounts
		Validate: func(g *gorm.DB, d *data.Data) (int64, int64, error) {
			return 3, 3, nil // formula says 3 while the accessor says 7
		},
	}
	rec := newRecordingCtx()
	issues := ValidateCounts(rec, db, &data.Data{}, []CollectionSpec{drifted})

	require.Len(t, issues, 1) // the 7-vs-3 mismatch also lands
	assert.Equal(t, "Drifty", issues[0].Collection)
	assert.Equal(t, int64(7), issues[0].Expected)
	assert.Equal(t, int64(3), issues[0].Actual)
	joined := strings.Join(rec.all(), "\n")
	assert.Contains(t, joined, "Drifty accessor=7 formula=3 — mapping doc drift?")
	assert.Contains(t, joined, "validation MISMATCH Drifty: legacy=7 new=3")
}

func TestValidateCountsDryRun(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	ctx := newRecordingCtx()

	_, err := RunMigration(ctx, db, Options{SourcePath: path, DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, 0, ctx.count(func(l string) bool { return strings.Contains(l, "validation") }),
		"dry-run commits nothing ⇒ validates nothing (AC-7)")
}

func mustLoad(t *testing.T, path string) *data.Data {
	t.Helper()
	d, err := data.LoadLegacyData(basecontext.NewRootBaseContext(), path)
	require.NoError(t, err)
	return &d
}
