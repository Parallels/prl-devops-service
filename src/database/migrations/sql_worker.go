package migrations

import (
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"gorm.io/gorm"
)

// SQLMigrationWorker represents a migration worker that executes SQL files
type SQLMigrationWorker struct {
	Name        string
	Description string
	Version     int
	UpFile      string
	DownFile    string
	db          *gorm.DB
}

// NewSQLMigrationWorker creates a new SQL migration worker
func NewSQLMigrationWorker(db *gorm.DB, name, description string, version int, upFile, downFile string) *SQLMigrationWorker {
	return &SQLMigrationWorker{
		Name:        name,
		Description: description,
		Version:     version,
		UpFile:      upFile,
		DownFile:    downFile,
		db:          db,
	}
}

func (w *SQLMigrationWorker) GetName() string {
	return w.Name
}

func (w *SQLMigrationWorker) GetDescription() string {
	return w.Description
}

func (w *SQLMigrationWorker) GetVersion() int {
	return w.Version
}

func (w *SQLMigrationWorker) GetOrder() int {
	return w.Version // For SQL migrations, version can double as order
}

func (w *SQLMigrationWorker) Up(ctx basecontext.BaseContext) *errors.Diagnostics {
	diag := errors.NewDiagnostics(fmt.Sprintf("%s_up", w.Name))

	// Read SQL file
	content, err := os.ReadFile(w.UpFile)
	if err != nil {
		diag.AddError("read_file_error", fmt.Sprintf("failed to read up migration file: %v", err), w.Name, nil)
		return diag
	}

	// Execute SQL
	// Split by semicolon? Or just execute as one block?
	// standard mechanism usually splits by statement if driver doesn't support multiple stats.
	// But let's assume we can execute the whole thing or split by ;
	// GORM Exec can execute raw SQL.
	sql := string(content)

	err = w.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sql).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		diag.AddError("execution_error", fmt.Sprintf("failed to execute up migration: %v", err), w.Name, nil)
		return diag
	}

	return diag
}

func (w *SQLMigrationWorker) Down(ctx basecontext.BaseContext) *errors.Diagnostics {
	diag := errors.NewDiagnostics(fmt.Sprintf("%s_down", w.Name))

	if w.DownFile == "" {
		return diag // No down migration
	}

	// Read SQL file
	content, err := os.ReadFile(w.DownFile)
	if err != nil {
		diag.AddError("read_file_error", fmt.Sprintf("failed to read down migration file: %v", err), w.Name, nil)
		return diag
	}

	sql := string(content)

	err = w.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sql).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		diag.AddError("execution_error", fmt.Sprintf("failed to execute down migration: %v", err), w.Name, nil)
		return diag
	}

	return diag
}
