package legacy_migrator

// Mapping knowledge (plan D-7): one CollectionSpec per legacy collection,
// ordered FK-safe. Scalar collections are #419-tagged models whose legacy
// struct IS the target model (data/models carries the gorm tags since
// #419-wave-b) — so the "copy" is identity: no JSON round-trip, string
// dates verbatim (D-6). Table/PK strings come from the tagged models
// themselves (GORM resolves both), never hand-typed.

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CollectionSpec struct {
	Name        string
	LegacyCount func(d *data.Data) int
	Migrate     func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunkSize int) (int64, error)
	// Validate feeds task 05's expected-vs-actual counters; scalar rows are
	// a plain model Count.
	Validate func(db *gorm.DB, d *data.Data) (expected, actual int64, err error)
}

// collectionPlan() lives in mapping_relational.go since task 04 (the full
// ordered plan needs the relational specs in scope); scalarPlan() stays here
// as the scalar half, reused by the relational plan and its tests.

func upsertAll[T any](ctx basecontext.ApiContext, tx *gorm.DB, rows []T, pk string, chunk int) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	res := tx.Omit(clause.Associations). // edges are ONLY the edgeSpecs' business: no implicit M2M full-saves from parent Creates (task 04)
						Clauses(clause.OnConflict{Columns: []clause.Column{{Name: pk}}, UpdateAll: true}).
						CreateInBatches(&rows, chunk) // AC-2: re-run = same PK ⇒ update, never duplicate
	// Caveat: RowsAffected ≠ len(rows) on MySQL-family update-in-place;
	// SQLite/Postgres semantics used here report per-statement affected rows.
	return res.RowsAffected, res.Error
}

func scalarSpec[T any](name string, pk string,
	rows func(d *data.Data) []T,
	count func(db *gorm.DB) (int64, error),
) CollectionSpec {
	return CollectionSpec{
		Name:        name,
		LegacyCount: func(d *data.Data) int { return len(rows(d)) },
		Migrate: func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunk int) (int64, error) {
			return upsertAll(ctx, tx, rows(d), pk, chunk)
		},
		Validate: func(db *gorm.DB, d *data.Data) (int64, int64, error) {
			a, err := count(db)
			return int64(len(rows(d))), a, err
		},
	}
}

// pointerSpec renders the one-row pointer collections (ReverseProxy):
// nil ⇒ 0 rows; else a one-element slice through the same engine.
// (Configuration moved to configurationSpec — a PK upsert cannot hold its
// singleton invariant once the app has seeded a placeholder row.)
func pointerSpec[T any](name string, pk string,
	row func(d *data.Data) *T,
	count func(db *gorm.DB) (int64, error),
) CollectionSpec {
	rows := func(d *data.Data) []T {
		r := row(d)
		if r == nil {
			return nil
		}
		return []T{*r}
	}
	return CollectionSpec{
		Name:        name,
		LegacyCount: func(d *data.Data) int { return len(rows(d)) },
		Migrate: func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunk int) (int64, error) {
			return upsertAll(ctx, tx, rows(d), pk, chunk)
		},
		Validate: func(db *gorm.DB, d *data.Data) (int64, int64, error) {
			a, err := count(db)
			return int64(len(rows(d))), a, err
		},
	}
}

func countOf[M any](db *gorm.DB) (int64, error) {
	var n int64
	err := db.Model(new(M)).Count(&n).Error
	return n, err
}

// configurationSpec handles the configurations singleton beyond PK upsert.
// SQL mode seed-on-reads the installation id on first read
// (database/configuration.go), so a legacy store arriving or changing AFTER
// that seeding re-migrates (fingerprint differs ⇒ marker gate opens) and a
// PK-keyed upsert sees two different ids: it INSERTs the legacy row beside
// the seeded placeholder — exactly the "singleton table holds 2 rows"
// corruption takeConfiguration warns about. The legacy id IS the
// installation identity, so the frozen legacy row REPLACES the table
// wholesale (idempotent: re-runs keep one row). nil legacy ⇒ no-op — there
// is nothing to reconcile against, the seeded row stays.
func configurationSpec() CollectionSpec {
	legacyRow := func(d *data.Data) *models.Configuration { return d.Configuration }
	legacyCount := func(d *data.Data) int {
		if legacyRow(d) == nil {
			return 0
		}
		return 1
	}
	return CollectionSpec{
		Name:        "Configuration",
		LegacyCount: legacyCount,
		Migrate: func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunk int) (int64, error) {
			r := legacyRow(d)
			if r == nil {
				return 0, nil
			}
			// GORM refuses a WHERE-less DELETE; the whole (singleton) table
			// is the target by design. Also repairs an already-duplicated
			// table when the marker gate re-opens on a changed source.
			if err := tx.Where("1 = 1").Delete(&models.Configuration{}).Error; err != nil {
				return 0, err
			}
			if err := tx.Omit(clause.Associations).Create(r).Error; err != nil {
				return 0, err
			}
			return 1, nil
		},
		Validate: func(db *gorm.DB, d *data.Data) (int64, int64, error) {
			a, err := countOf[models.Configuration](db)
			return int64(legacyCount(d)), a, err
		},
	}
}

func scalarPlan() []CollectionSpec {
	return []CollectionSpec{
		configurationSpec(), // singleton-aware replace, see configurationSpec doc
		// Data-declaration-order stretch: ApiKeys, PackerTemplates …
		scalarSpec("ApiKeys", "id",
			func(d *data.Data) []models.ApiKey { return d.ApiKeys },
			func(db *gorm.DB) (int64, error) { return countOf[models.ApiKey](db) }),
		scalarSpec("PackerTemplates", "id",
			func(d *data.Data) []models.PackerTemplate { return d.PackerTemplates },
			func(db *gorm.DB) (int64, error) { return countOf[models.PackerTemplate](db) }),
		// … hosts before their referrers (FK-safe), then the rest.
		scalarSpec("OrchestratorHosts", "id",
			func(d *data.Data) []models.OrchestratorHost { return d.OrchestratorHosts },
			func(db *gorm.DB) (int64, error) { return countOf[models.OrchestratorHost](db) }),
		pointerSpec("ReverseProxy", "id",
			func(d *data.Data) *models.ReverseProxy { return d.ReverseProxy },
			func(db *gorm.DB) (int64, error) { return countOf[models.ReverseProxy](db) }),
		scalarSpec("ReverseProxyHosts", "id",
			func(d *data.Data) []models.ReverseProxyHost { return d.ReverseProxyHosts },
			func(db *gorm.DB) (int64, error) { return countOf[models.ReverseProxyHost](db) }),
		scalarSpec("CatalogManagers", "id",
			func(d *data.Data) []models.CatalogManager { return d.CatalogManagers },
			func(db *gorm.DB) (int64, error) { return countOf[models.CatalogManager](db) }),
		scalarSpec("Jobs", "id",
			func(d *data.Data) []models.Job { return d.Jobs },
			func(db *gorm.DB) (int64, error) { return countOf[models.Job](db) }),
		scalarSpec("EnrollmentTokens", "id",
			func(d *data.Data) []models.OrchestratorEnrollmentToken { return d.EnrollmentTokens },
			func(db *gorm.DB) (int64, error) { return countOf[models.OrchestratorEnrollmentToken](db) }),
		scalarSpec("UserConfigs", "id",
			func(d *data.Data) []models.UserConfig { return d.UserConfigs },
			func(db *gorm.DB) (int64, error) { return countOf[models.UserConfig](db) }),
	}
}
