package migrations

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/migrations/workers"
	logging "github.com/cjlapao/common-go-logger"
	"gorm.io/gorm"
)

// InitializeAndRun initializes the migration service and runs all default migrations.
//
// This function automatically seeds a fresh database with:
//   - 90+ system claims (VM operations, user management, roles, etc.)
//   - 3 default roles (USER, ADMIN, SUPER_USER)
//   - Role-claim associations based on constants.RoleClaimsMap
//   - Default admin user (username: "admin", password: "admin" or ROOT_PASSWORD env var)
//
// The migration system is idempotent - it's safe to call this multiple times.
// Already-applied migrations are tracked in the _migrations table and will be skipped.
//
// Usage in application startup:
//
//	func main() {
//	    // Initialize database connection
//	    db, err := gorm.Open(sqlite.Open("prl-devops.db"), &gorm.Config{})
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Run auto-migrations for schema
//	    db.AutoMigrate(&models.User{}, &models.Role{}, &models.Claim{}, ...)
//
//	    // Seed database with default data
//	    if err := migrations.InitializeAndRun(db); err != nil {
//	        log.Fatal("Migration failed:", err)
//	    }
//
//	    // Continue with application startup...
//	}
//
// Environment Variables:
//   - ROOT_PASSWORD: Custom password for admin user (default: "admin")
//
// Security Note:
// Always change the default admin password after first login!
func InitializeAndRun(db *gorm.DB) error {
	logger := logging.Get()
	logger.Info("Initializing database migrations...")

	// Initialize migration service
	migrationService := Initialize(db)

	// Register all default workers in order
	migrationService.Register(workers.NewDefaultClaimsWorker(db))    // Order: 10
	migrationService.Register(workers.NewDefaultRolesWorker(db))     // Order: 20
	migrationService.Register(workers.NewRoleClaimsWorker(db))       // Order: 30
	migrationService.Register(workers.NewDefaultAdminUserWorker(db)) // Order: 40

	// Run all migrations
	ctx := basecontext.NewRootBaseContext()
	diag := migrationService.RunAll(*ctx)

	if diag.HasErrors() {
		logger.Error("Migration failed: %v", diag.GetSummary())
		errors := diag.GetErrors()
		if len(errors) > 0 {
			return fmt.Errorf("migration failed: %s", errors[0].Message)
		}
		return fmt.Errorf("migration failed: %s", diag.GetSummary())
	}

	logger.Info("Database migrations completed successfully")
	return nil
}
