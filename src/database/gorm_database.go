package database

// GormDatabase — the SQL-backed implementor of datainterfaces.DatabaseService
// (PBI #421 task 04, plan D-1 hybrid-delegate architecture).
//
// Day one: every domain method one-lines to the injected legacy handle (see
// delegate_*.go); tasks 05-16 replace delegates with SQL implementations;
// task 18 deletes the legacy handle. The type NEVER changes shape in between
// (R-1 discipline: tasks add/remove method bodies, never fields or the
// interface set).

import (
	"sync"
	"sync/atomic"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	sqlitedrv "github.com/glebarez/sqlite"
	postgresdrv "gorm.io/driver/postgres"
	"time"

	"gorm.io/gorm"
)

type GormDatabase struct {
	gormDB       *gorm.DB
	legacy       *data.JsonDatabase // hybrid delegate; deleted in task 18
	onJobTimeout func(models.Job)
	timeoutMu    sync.RWMutex
	connected    atomic.Bool
	closeOnce    sync.Once
	warnedBackup atomic.Bool
	dataSource   string
	// Task 15 additions (private, additive — R-1 governs the interface set
	// and delegate shape; the task spec mandates the injected clock):
	nowFn       func() time.Time // stale-window test hook (nil = real clock)
	staleCancel chan bool        // watcher cancel, legacy `chan bool` shape
}

// NewGormDatabase wraps an open GORM connection. legacy is the hybrid
// delegate for every not-yet-ported domain op; it may be nil only in tests
// that never touch delegated methods.
func NewGormDatabase(ctx basecontext.ApiContext, gormDB *gorm.DB, legacy *data.JsonDatabase) *GormDatabase {
	g := &GormDatabase{gormDB: gormDB, legacy: legacy}
	if gormDB != nil && gormDB.Dialector != nil {
		switch d := gormDB.Dialector.(type) {
		case *sqlitedrv.Dialector:
			g.dataSource = d.DSN // ":memory:" verbatim (D-7)
		case *postgresdrv.Dialector:
			// Never surface the DSN through Filename(): it carries credentials.
			// The opaque token keeps the escape-hatch contract, secrets stay
			// in config (#420 redaction invariants).
			g.dataSource = "postgres"
		}
	}
	g.connected.Store(gormDB != nil)
	return g
}

// DB exposes the raw handle (WAL audit #420/#421, #424 write path).
func (g *GormDatabase) DB() *gorm.DB { return g.gormDB }

func (g *GormDatabase) Connect(ctx basecontext.ApiContext) error {
	if g.gormDB == nil {
		g.connected.Store(false)
		return errors.NewWithCode("database operation failed: no gorm connection", 500)
	}
	sqlDB, err := g.gormDB.DB()
	if err != nil {
		g.connected.Store(false)
		return errors.NewWithCode("database operation failed", 500)
	}
	if err := sqlDB.Ping(); err != nil {
		g.connected.Store(false)
		return errors.NewWithCode("database operation failed", 500)
	}
	g.connected.Store(true)
	return nil
}

// Disconnect closes the underlying pool exactly once (idempotent).
// Task 15/18 note: stop the job-timeout watcher here once it is wired.
func (g *GormDatabase) Disconnect(ctx basecontext.ApiContext) error {
	var err error
	g.closeOnce.Do(func() {
		if g.gormDB != nil {
			if sqlDB, dbErr := g.gormDB.DB(); dbErr == nil {
				err = sqlDB.Close()
			}
		}
		g.connected.Store(false)
	})
	return err
}

func (g *GormDatabase) IsConnected() bool { return g.connected.Load() }

// dbFilename resolves the D-7 Filename() escape-hatch value from the active
// dialector: sqlite paths verbatim (":memory:" included), postgres DSNs are
// passed through REDACTED-free but only for logging parity, never secrets —
// callers treat it as opaque.
func (g *GormDatabase) dbFilename() string {
	return g.dataSource
}
