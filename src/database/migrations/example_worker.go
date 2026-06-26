package migrations

import (
	"github.com/Parallels/prl-devops-service/errors"
	"gorm.io/gorm"
)

// ExampleSeedWorker demonstrates how to create a seed worker
type ExampleSeedWorker struct {
	db *gorm.DB
}

// NewExampleSeedWorker creates a new example seed worker
func NewExampleSeedWorker(db *gorm.DB) *ExampleSeedWorker {
	return &ExampleSeedWorker{
		db: db,
	}
}

// GetName returns the name of this seed
func (e *ExampleSeedWorker) GetName() string {
	return "example-seed"
}

// GetDescription returns the description of this seed
func (e *ExampleSeedWorker) GetDescription() string {
	return "Example seed that creates a sample table"
}

// GetVersion returns the version number
func (e *ExampleSeedWorker) GetVersion() int {
	return 1
}

// Up applies the seed
func (e *ExampleSeedWorker) Up() *errors.Diagnostics {
	diag := errors.NewDiagnostics("example_seed_up")
	defer diag.Complete()

	// Example: Create a sample table
	sql := `
		CREATE TABLE IF NOT EXISTS example_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`

	if err := e.db.Exec(sql).Error; err != nil {
		diag.AddError("CREATE_TABLE_FAILED", "Failed to create example table", "example_seed", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	// Example: Insert some sample data
	sampleData := []map[string]interface{}{
		{"name": "Sample 1", "description": "First sample record"},
		{"name": "Sample 2", "description": "Second sample record"},
	}

	for _, data := range sampleData {
		if err := e.db.Table("example_table").Create(data).Error; err != nil {
			diag.AddError("INSERT_DATA_FAILED", "Failed to insert sample data", "example_seed", map[string]interface{}{
				"error": err.Error(),
				"data":  data,
			})
			return diag
		}
	}

	diag.AddPathEntry("example_seed_applied", "example_seed", map[string]interface{}{
		"table_created":    "example_table",
		"records_inserted": len(sampleData),
	})

	return diag
}

// Down rolls back the seed
func (e *ExampleSeedWorker) Down() *errors.Diagnostics {
	diag := errors.NewDiagnostics("example_seed_down")
	defer diag.Complete()

	// Drop the table we created
	if err := e.db.Exec("DROP TABLE IF EXISTS example_table").Error; err != nil {
		diag.AddError("DROP_TABLE_FAILED", "Failed to drop example table", "example_seed", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	diag.AddPathEntry("example_seed_rolled_back", "example_seed", map[string]interface{}{
		"table_dropped": "example_table",
	})

	return diag
}

// Usage example:
// func main() {
//     db := // your database connection
//     seedService := NewSeedService(db)
//
//     // Register the example seed
//     exampleWorker := NewExampleSeedWorker(db)
//     seedService.Register(exampleWorker)
//
//     // Run all seeds
//     ctx := appctx.NewContext(nil)
//     diag := seedService.RunAll(ctx)
//
//     if diag.HasErrors() {
//         fmt.Printf("Seeds failed: %s\n", diag.GetSummary())
//     } else {
//         fmt.Println("All seeds applied successfully")
//     }
// }
