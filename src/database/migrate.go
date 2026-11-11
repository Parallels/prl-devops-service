package database

// Production AutoMigrate registry (PBI #421 task 18; the closing boot
// wiring). Tasks 05-16 pinned this exact list in their test registries
// (roles_test.go header); it graduates here as the single source, and the
// test copies now delegate. #424's migration runner adds schema_migrations
// bookkeeping on top; model list additions ride the domain that owns them.

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"gorm.io/gorm"
)

// AllGormModels lists every #419 table model, dependency-free order (gorm
// sorts FKs internally). Slice history: the parity harness and package
// database units each carried a duplicate of this list "until the task-18
// boot wiring" — this is that wiring; both registries now call here.
func AllGormModels() []interface{} {
	return []interface{}{
		&models.User{}, &models.Role{}, &models.Claim{}, &models.RoleClaims{}, &models.ApiKey{},
		&models.DatabaseSchema{},
		&models.CatalogManager{}, &models.CatalogManifest{}, &models.Job{}, &models.PackerTemplate{},
		&models.UserConfig{}, &models.ReverseProxy{}, &models.ReverseProxyHost{},
		&models.OrchestratorEnrollmentToken{}, &models.Configuration{},
		&models.OrchestratorHost{}, &models.VMSnapshot{}, &models.VirtualMachine{},
	}
}

// RunAutoMigrate applies the registry to an open connection. DB_AUTO_MIGRATE
// gates the CALL at boot (provider); the function itself is unconditional so
// tests and tooling share it.
func RunAutoMigrate(ctx basecontext.ApiContext, gormDB *gorm.DB) error {
	if gormDB == nil {
		return ErrDatabaseUnreachable
	}
	if err := gormDB.AutoMigrate(AllGormModels()...); err != nil {
		return err
	}
	ctx.LogInfof("[Database] SQL AutoMigrate complete (%d models)", len(AllGormModels()))
	return nil
}
