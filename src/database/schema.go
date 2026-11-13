package database

// SQL implementation of the schema-version meta (PBI #421 task 08) on the
// single-row schema_metadata table (#419 OQ-3 default naming via
// DatabaseSchema.TableName; AD-4: this table stores ONLY the legacy schema
// string — #424's data-migration completion marker is a separate table and
// is never touched from here). Legacy data/schema.go is a plain string
// getter/setter: GetSchemaVersion on an empty table returns the zero value
// "" with NO error (first-boot path the seeds and #424 rely on),
// UpdateSchemaVersion stores whatever string it is given — validation is
// deliberately NOT invented. No filter/authz/order wiring exists in the
// legacy source; no transaction sites (D-9): single-row read or single
// statement write, serialised by the single connection.

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"gorm.io/gorm"
)

func (g *GormDatabase) GetSchemaVersion(ctx basecontext.ApiContext) (string, error) {
	if !g.IsConnected() {
		return "", interfaces.ErrDatabaseNotConnected
	}
	var row models.DatabaseSchema
	err := g.gormDB.Table("schema_metadata").Order("rowid").Limit(1).Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil // legacy zero-value Data.Schema.Version — empty, not error
	}
	if err != nil {
		return "", mapGormError(err, nil, nil)
	}
	return row.Version, nil
}

// UpdateSchemaVersion upserts the single meta row without a transaction:
// INSERT ... WHERE NOT EXISTS when the table is empty, otherwise a targeted
// update keyed on the existing row's primary key. Racing inserts collapse on
// the NOT EXISTS clause; the update-by-old-PK can only match the one row.
func (g *GormDatabase) UpdateSchemaVersion(ctx basecontext.ApiContext, version string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	var n int64
	if err := g.gormDB.Table("schema_metadata").Count(&n).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	if n == 0 {
		if err := g.gormDB.Exec(
			"INSERT INTO schema_metadata (version) SELECT ? WHERE NOT EXISTS (SELECT 1 FROM schema_metadata)",
			version).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		return nil
	}
	var row models.DatabaseSchema
	if err := g.gormDB.Table("schema_metadata").Order("rowid").Limit(1).Take(&row).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	if row.Version == version {
		return nil
	}
	if err := g.gormDB.Table("schema_metadata").
		Where("version = ?", row.Version).Update("version", version).Error; err != nil {
		return mapGormError(err, nil, nil)
	}
	return nil
}
