package legacy_migrator

// Mapping knowledge (plan D-7): one CollectionSpec per legacy collection,
// ordered FK-safe. Scalar collections are #419-tagged models whose legacy
// struct IS the target model (data/models carries the gorm tags since
// #419-wave-b) — so the "copy" is identity: no JSON round-trip, string
// dates verbatim (D-6). Table/PK strings come from the tagged models
// themselves (GORM resolves both), never hand-typed.

import (
	"fmt"
	"reflect"
	"strings"

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

func upsertAll[T any](ctx basecontext.ApiContext, name string, tx *gorm.DB, rows []T, pk string, chunk int) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	rows, dropped := dedupeByPK(rows, pk)
	if len(dropped) > 0 {
		// Worth a warning rather than silence: the legacy store held two records
		// that the relational one cannot, and which copy survived was previously
		// decided by whichever engine happened to be running.
		ctx.LogWarnf("[Migration] %s: %d row(s) share a primary key with an earlier row, first kept (e.g. %s)",
			name, len(dropped), strings.Join(firstKeys(dropped, 5), ", "))
	}

	res := tx.Omit(clause.Associations). // edges are ONLY the edgeSpecs' business: no implicit M2M full-saves from parent Creates (task 04)
						Clauses(clause.OnConflict{Columns: []clause.Column{{Name: pk}}, UpdateAll: true}).
						CreateInBatches(&rows, chunk) // AC-2: re-run = same PK ⇒ update, never duplicate
	// Caveat: RowsAffected ≠ len(rows) on MySQL-family update-in-place;
	// SQLite/Postgres semantics used here report per-statement affected rows.
	return res.RowsAffected, res.Error
}

// dedupeByPK collapses rows that would land on the same primary key within one
// batch, keeping the first — the rule the typed rows() functions already follow
// ("canonical first; first-seen order = deterministic file order").
//
// This is the backstop for the collections whose rows() has no dedupe to speak
// of: vmRows flattens every host's embedded VM slice, and a VM stored under two
// hosts arrives twice. SQLite's ON CONFLICT DO UPDATE lets the later copy
// overwrite the earlier one and nobody notices; PostgreSQL rejects the entire
// statement — "ON CONFLICT DO UPDATE command cannot affect row a second time"
// (SQLSTATE 21000) — which aborts the migration and rolls the whole install
// back. Both engines were looking at the same data; only one of them admitted it.
//
// Rows whose key cannot be read (not a struct, no primary key tagged) are
// passed through untouched: guessing at a key would drop data, and an
// un-collapsible batch is the engine's problem to report, not this one's.
func dedupeByPK[T any](rows []T, pk string) ([]T, []string) {
	keyOf, ok := primaryKeyOf[T](pk)
	if !ok {
		return rows, nil
	}

	seen := make(map[string]struct{}, len(rows))
	out := make([]T, 0, len(rows))
	var dropped []string

	for _, row := range rows {
		key := keyOf(row)
		if _, dup := seen[key]; dup {
			dropped = append(dropped, key)
			continue
		}
		seen[key] = struct{}{}
		out = append(out, row)
	}

	return out, dropped
}

// primaryKeyOf resolves the field backing the pk column once, per instantiation
// of the generic, and returns an accessor for it.
func primaryKeyOf[T any](pk string) (func(T) string, bool) {
	t := reflect.TypeOf(new(T)).Elem()
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	idx, ok := pkFieldIndex(t, pk)
	if !ok {
		return nil, false
	}

	return func(row T) string {
		v := reflect.ValueOf(row)
		for _, step := range idx {
			v = v.Field(step)
		}
		if !v.IsValid() {
			return ""
		}
		return fmt.Sprintf("%v", v.Interface())
	}, true
}

func pkFieldIndex(t reflect.Type, pk string) ([]int, bool) {
	// The gorm tag is authoritative — these models declare `gorm:"primaryKey"`.
	for i := 0; i < t.NumField(); i++ {
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primaryKey") {
			return []int{i}, true
		}
	}

	// Otherwise match the column name against the field. Underscores are dropped
	// on both sides: the column is `host_id` and the field HostID, and an
	// acronym-aware snake_case is more machinery than this comparison needs.
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if strings.EqualFold(collapse(f.Name), collapse(pk)) {
			return []int{i}, true
		}
	}

	return nil, false
}

func collapse(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "_", "")
}

func firstKeys(keys []string, limit int) []string {
	if len(keys) <= limit {
		return keys
	}
	return keys[:limit]
}

func scalarSpec[T any](name string, pk string,
	rows func(d *data.Data) []T,
	count func(db *gorm.DB) (int64, error),
) CollectionSpec {
	return CollectionSpec{
		Name:        name,
		LegacyCount: func(d *data.Data) int { return len(rows(d)) },
		Migrate: func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunk int) (int64, error) {
			return upsertAll(ctx, name, tx, rows(d), pk, chunk)
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
			return upsertAll(ctx, name, tx, rows(d), pk, chunk)
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
