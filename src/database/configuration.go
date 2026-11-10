package database

// SQL implementation of the configuration singleton — the installation's
// identity row (PBI #421 task 08). NOTE on the filename: the task spec said
// `config.go`, but that file already exists (task-04 connection-config
// contract, approved) — the domain lives here instead, 1:1 with its delegate
// removal. Ledger deviation.
//
// Legacy data/config.go is pure singleton semantics: seed-on-read creation
// (GetConfiguration/GetId/SeedId each lazily mint the one row with a
// generated installation id) and a single-write SetId guarded by
// "ID already exists". No method here touches query.ParseFilter, order_layer,
// index_layer, authorize_layer or OwnRecords — none exists in the legacy
// source either. No transaction sites (D-9): every method is one read or one
// statement-level targeted write; the read-then-write races are closed by
// INSERT ... WHERE NOT EXISTS and an id=='' guarded update, then re-read —
// legacy serialised these on dataMutex, single-conn sqlite serialises the
// statements. (The spec's "pinned primary key" cannot be literal here: the
// primary key IS the generated installation id.)

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/security"
	log "github.com/cjlapao/common-go-logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// generateInstallationId calls the SAME primitives as legacy generateId
// (config.go:75-82): crypto-random 32 chars, uuid fallback. Call-wise port —
// the primitives live in security/uuid; no generator body is duplicated.
func generateInstallationId() string {
	id, err := security.GenerateCryptoRandomString(32)
	if err != nil {
		id = uuid.New().String()
	}
	return id
}

// takeConfiguration reads the singleton row (first by rowid — legacy
// physically could not hold two; extra rows are corruption, defended per
// spec: take first, log). Not found ⇒ gorm.ErrRecordNotFound passes through
// for the seed path to handle.
func (g *GormDatabase) takeConfiguration() (*models.Configuration, error) {
	var cfg models.Configuration
	err := g.gormDB.Table("configurations").Order("rowid").Limit(1).Take(&cfg).Error
	if err != nil {
		return nil, err
	}
	var n int64
	if err := g.gormDB.Table("configurations").Count(&n).Error; err == nil && n > 1 {
		log.Get().Log("configurations singleton table holds %d rows (corruption); using first by rowid", log.Warning, n)
	}
	return &cfg, nil
}

// seedConfigurationRow implements legacy's seed-on-read: create the row only
// if none exists (single statement — race-closed), then re-read.
func (g *GormDatabase) seedConfigurationRow() (*models.Configuration, error) {
	if err := g.gormDB.Exec(
		"INSERT INTO configurations (id) SELECT ? WHERE NOT EXISTS (SELECT 1 FROM configurations)",
		generateInstallationId()).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	cfg, err := g.takeConfiguration()
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return cfg, nil
}

func (g *GormDatabase) GetConfiguration(ctx basecontext.ApiContext) (*models.Configuration, error) {
	if !g.IsConnected() {
		return nil, datainterfaces.ErrDatabaseNotConnected
	}
	cfg, err := g.takeConfiguration()
	if err == gorm.ErrRecordNotFound {
		// Legacy seed-on-read (config.go:17-20) — mirrored, not "fixed" into an error.
		return g.seedConfigurationRow()
	}
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return cfg, nil
}

func (g *GormDatabase) GetId(ctx basecontext.ApiContext) (string, error) {
	if !g.IsConnected() {
		return "", datainterfaces.ErrDatabaseNotConnected
	}
	cfg, err := g.GetConfiguration(ctx) // seed-on-read included
	if err != nil {
		return "", err
	}
	return cfg.ID, nil
}

// SeedId mirrors config.go:35-52: seed-on-read, then fill an empty ID. The
// fill is an id==” guarded single-statement update — racing callers either
// no-op on the insert or match zero rows on the update, and every caller
// returns the RE-READ value, so all callers agree (spec's race-closure, no
// tx needed).
func (g *GormDatabase) SeedId(ctx basecontext.ApiContext) (string, error) {
	if !g.IsConnected() {
		return "", datainterfaces.ErrDatabaseNotConnected
	}
	cfg, err := g.seedConfigurationRow()
	if err != nil {
		return "", err
	}
	if cfg.ID == "" {
		if err := g.gormDB.Table("configurations").
			Where("id = ?", "").Update("id", generateInstallationId()).Error; err != nil {
			return "", mapGormError(err, nil, nil)
		}
		cfg, err = g.takeConfiguration()
		if err != nil {
			return "", mapGormError(err, nil, nil)
		}
	}
	return cfg.ID, nil
}

// SetId mirrors config.go:57-73 exactly: the id is only writable while the
// stored value is empty; a second write answers the legacy plain error
// ("ID already exists", same message — not a sentinel). Empty incoming id on
// a fresh installation creates the empty singleton (legacy creates
// &Configuration{} then assigns "").
func (g *GormDatabase) SetId(ctx basecontext.ApiContext, id string) (string, error) {
	if !g.IsConnected() {
		return "", datainterfaces.ErrDatabaseNotConnected
	}
	cfg, err := g.takeConfiguration()
	if err == gorm.ErrRecordNotFound {
		// Fresh installation: legacy creates the empty singleton then assigns the
		// id — one statement here inserts the id directly; racing callers no-op
		// on the NOT EXISTS and fall through to the existing-value path below.
		if err := g.gormDB.Exec(
			"INSERT INTO configurations (id) SELECT ? WHERE NOT EXISTS (SELECT 1 FROM configurations)",
			id).Error; err != nil {
			return "", mapGormError(err, nil, nil)
		}
		cfg, err = g.takeConfiguration()
		if err != nil {
			return "", mapGormError(err, nil, nil)
		}
		if cfg.ID == id { // our insert won the race — assignment complete
			return id, nil
		}
	} else if err != nil {
		return "", mapGormError(err, nil, nil)
	}
	if cfg.ID != "" {
		return "", errors.New("ID already exists") // legacy plain error, message-pinned
	}
	if id != "" {
		if err := g.gormDB.Table("configurations").
			Where("id = ?", "").Update("id", id).Error; err != nil {
			return "", mapGormError(err, nil, nil)
		}
		return id, nil
	}
	return cfg.ID, nil
}
