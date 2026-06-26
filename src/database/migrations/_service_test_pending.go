package migrations

import (
	"context"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/cjlapao/common-go-logger"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	logging.InitializeWithConfig(logging.LogConfig{
		Level: "error",
	})
}

// setupTestDB creates a test database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func TestNewSeedService(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)

	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.workers)
	assert.NotNil(t, service.applied)
	assert.Equal(t, 0, len(service.workers))
	assert.Equal(t, 0, len(service.applied))
}

func TestSeedService_Register(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)

	// Test registering a worker
	worker1 := NewMockSeedWorker("test-seed-1", "Test seed 1")
	service.Register(worker1)

	assert.Equal(t, 1, len(service.GetRegisteredSeeds()))
	assert.Contains(t, service.GetRegisteredSeeds(), "test-seed-1")

	// Test registering another worker
	worker2 := NewMockSeedWorker("test-seed-2", "Test seed 2")
	service.Register(worker2)

	assert.Equal(t, 2, len(service.GetRegisteredSeeds()))
	assert.Contains(t, service.GetRegisteredSeeds(), "test-seed-1")
	assert.Contains(t, service.GetRegisteredSeeds(), "test-seed-2")

	// Test registering duplicate worker (should not add)
	service.Register(worker1)
	assert.Equal(t, 2, len(service.GetRegisteredSeeds()))
}

func TestSeedService_RunAll_Success(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register successful workers
	worker1 := NewMockSeedWorker("test-seed-1", "Test seed 1")
	worker2 := NewMockSeedWorker("test-seed-2", "Test seed 2")

	service.Register(worker1)
	service.Register(worker2)

	// Run all seeds
	diag := service.RunAll(ctx)

	// Should succeed
	assert.False(t, diag.HasErrors())
	assert.True(t, worker1.UpCalled)
	assert.True(t, worker2.UpCalled)
	assert.False(t, worker1.DownCalled)
	assert.False(t, worker2.DownCalled)

	// Check applied seeds
	applied := service.GetAppliedSeeds()
	assert.Equal(t, 2, len(applied))
	assert.Contains(t, applied, "test-seed-1")
	assert.Contains(t, applied, "test-seed-2")

	// Check pending seeds (should be empty)
	pending := service.GetPendingSeeds()
	assert.Equal(t, 0, len(pending))
}

func TestSeedService_RunAll_WithAlreadyApplied(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register workers
	worker1 := NewMockSeedWorker("test-seed-1", "Test seed 1")
	worker2 := NewMockSeedWorker("test-seed-2", "Test seed 2")

	service.Register(worker1)
	service.Register(worker2)

	// Run first time
	diag := service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Reset mock state
	worker1.UpCalled = false
	worker2.UpCalled = false

	// Run second time - should skip already applied
	diag = service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Should not call Up again
	assert.False(t, worker1.UpCalled)
	assert.False(t, worker2.UpCalled)
}

func TestSeedService_RunAll_Failure(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register workers - first succeeds, second fails
	worker1 := NewMockSeedWorker("test-seed-1", "Test seed 1")
	worker2 := NewMockSeedWorker("test-seed-2", "Test seed 2")
	worker2.FailOnUp = true

	service.Register(worker1)
	service.Register(worker2)

	// Run all seeds
	diag := service.RunAll(ctx)

	// Should fail
	assert.True(t, diag.HasErrors())
	assert.True(t, worker1.UpCalled)
	assert.True(t, worker2.UpCalled)
	assert.True(t, worker2.DownCalled) // Should attempt rollback

	// Check applied seeds (only first should be applied)
	applied := service.GetAppliedSeeds()
	assert.Equal(t, 1, len(applied))
	assert.Contains(t, applied, "test-seed-1")
	assert.NotContains(t, applied, "test-seed-2")
}

func TestSeedService_RunAll_FailureWithRollbackFailure(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register worker that fails on both up and down
	worker := NewMockSeedWorker("test-seed-1", "Test seed 1")
	worker.FailOnUp = true
	worker.FailOnDown = true

	service.Register(worker)

	// Run all seeds
	diag := service.RunAll(ctx)

	// Should fail with both up and rollback errors
	assert.True(t, diag.HasErrors())
	assert.True(t, worker.UpCalled)
	assert.True(t, worker.DownCalled)

	// Should have both SEED_FAILED and ROLLBACK_FAILED errors
	errorCount := 0
	for _, item := range diag.Items {
		if item.Type == diagnostics.DiagnosticTypeError {
			errorCount++
		}
	}
	assert.GreaterOrEqual(t, errorCount, 2)
}

func TestSeedService_Rollback_Success(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register and apply a worker
	worker := NewMockSeedWorker("test-seed-1", "Test seed 1")
	service.Register(worker)

	// Apply the seed first
	diag := service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Reset mock state
	worker.DownCalled = false

	// Rollback the seed
	diag = service.Rollback(ctx, "test-seed-1")

	// Should succeed
	assert.False(t, diag.HasErrors())
	assert.True(t, worker.DownCalled)

	// Check applied seeds (should be empty)
	applied := service.GetAppliedSeeds()
	assert.Equal(t, 0, len(applied))

	// Check pending seeds (should contain the rolled back seed)
	pending := service.GetPendingSeeds()
	assert.Equal(t, 1, len(pending))
	assert.Contains(t, pending, "test-seed-1")
}

func TestSeedService_Rollback_NotApplied(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register a worker but don't apply it
	worker := NewMockSeedWorker("test-seed-1", "Test seed 1")
	service.Register(worker)

	// Try to rollback
	diag := service.Rollback(ctx, "test-seed-1")

	// Should fail
	assert.True(t, diag.HasErrors())
	assert.False(t, worker.DownCalled)
}

func TestSeedService_Rollback_NotFound(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Try to rollback non-existent seed
	diag := service.Rollback(ctx, "non-existent-seed")

	// Should fail
	assert.True(t, diag.HasErrors())
}

func TestSeedService_Rollback_Failure(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register and apply a worker
	worker := NewMockSeedWorker("test-seed-1", "Test seed 1")
	service.Register(worker)

	// Apply the seed first
	diag := service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Set worker to fail on down
	worker.FailOnDown = true

	// Try to rollback
	diag = service.Rollback(ctx, "test-seed-1")

	// Should fail
	assert.True(t, diag.HasErrors())
	assert.True(t, worker.DownCalled)

	// Seed should still be applied (rollback failed)
	applied := service.GetAppliedSeeds()
	assert.Equal(t, 1, len(applied))
	assert.Contains(t, applied, "test-seed-1")
}

func TestSeedService_GetAppliedSeeds(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Initially should be empty
	applied := service.GetAppliedSeeds()
	assert.Equal(t, 0, len(applied))

	// Register and apply a worker
	worker := NewMockSeedWorker("test-seed-1", "Test seed 1")
	service.Register(worker)

	diag := service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Should now have one applied seed
	applied = service.GetAppliedSeeds()
	assert.Equal(t, 1, len(applied))
	assert.Contains(t, applied, "test-seed-1")
}

func TestSeedService_GetPendingSeeds(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register workers
	worker1 := NewMockSeedWorker("test-seed-1", "Test seed 1")
	worker2 := NewMockSeedWorker("test-seed-2", "Test seed 2")

	service.Register(worker1)
	service.Register(worker2)

	// Initially all should be pending
	pending := service.GetPendingSeeds()
	assert.Equal(t, 2, len(pending))
	assert.Contains(t, pending, "test-seed-1")
	assert.Contains(t, pending, "test-seed-2")

	// Apply one seed
	diag := service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Should now have one pending
	pending = service.GetPendingSeeds()
	assert.Equal(t, 0, len(pending)) // Both were applied
}

func TestSeedService_GetRegisteredSeeds(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)

	// Initially should be empty
	registered := service.GetRegisteredSeeds()
	assert.Equal(t, 0, len(registered))

	// Register workers
	worker1 := NewMockSeedWorker("test-seed-1", "Test seed 1")
	worker2 := NewMockSeedWorker("test-seed-2", "Test seed 2")

	service.Register(worker1)
	service.Register(worker2)

	// Should now have two registered
	registered = service.GetRegisteredSeeds()
	assert.Equal(t, 2, len(registered))
	assert.Contains(t, registered, "test-seed-1")
	assert.Contains(t, registered, "test-seed-2")
}

func TestMigrationRecord_TableName(t *testing.T) {
	record := MigrationRecord{}
	assert.Equal(t, "_migrations", record.TableName())
}

func TestMockSeedWorker(t *testing.T) {
	worker := NewMockSeedWorker("test-worker", "Test description")
	ctx := appctx.NewContext(context.Background())

	assert.Equal(t, "test-worker", worker.GetName())
	assert.Equal(t, "Test description", worker.GetDescription())
	assert.Equal(t, 1, worker.GetVersion())

	// Test successful up
	diag := worker.Up(ctx)
	assert.False(t, diag.HasErrors())
	assert.True(t, worker.UpCalled)

	// Test successful down
	diag = worker.Down(ctx)
	assert.False(t, diag.HasErrors())
	assert.True(t, worker.DownCalled)

	// Test failing up
	worker.FailOnUp = true
	worker.UpCalled = false
	diag = worker.Up(ctx)
	assert.True(t, diag.HasErrors())
	assert.True(t, worker.UpCalled)

	// Test failing down
	worker.FailOnDown = true
	worker.DownCalled = false
	diag = worker.Down(ctx)
	assert.True(t, diag.HasErrors())
	assert.True(t, worker.DownCalled)
}

// Additional tests for better coverage

func TestSeedService_RunAll_EmptyWorkers(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Run with no workers registered
	diag := service.RunAll(ctx)

	// Should succeed with no errors
	assert.False(t, diag.HasErrors())

	// Check applied seeds (should be empty)
	applied := service.GetAppliedSeeds()
	assert.Equal(t, 0, len(applied))
}

func TestSeedService_RunAll_AllAlreadyApplied(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)
	ctx := appctx.NewContext(context.TODO())

	// Register workers
	worker1 := NewMockSeedWorker("test-seed-1", "Test seed 1")
	worker2 := NewMockSeedWorker("test-seed-2", "Test seed 2")

	service.Register(worker1)
	service.Register(worker2)

	// Apply all seeds
	diag := service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Reset mock state
	worker1.UpCalled = false
	worker2.UpCalled = false

	// Run again - all should be skipped
	diag = service.RunAll(ctx)
	assert.False(t, diag.HasErrors())

	// Should not call Up again
	assert.False(t, worker1.UpCalled)
	assert.False(t, worker2.UpCalled)
}

func TestSeedService_Register_DuplicateName(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)

	// Register worker with same name but different instance
	worker1 := NewMockSeedWorker("test-seed", "Test seed 1")
	worker2 := NewMockSeedWorker("test-seed", "Test seed 2") // Same name, different description

	service.Register(worker1)
	service.Register(worker2)

	// Should only have one registered (duplicate ignored)
	registered := service.GetRegisteredSeeds()
	assert.Equal(t, 1, len(registered))
	assert.Contains(t, registered, "test-seed")
}

func TestSeedService_ConcurrentAccess(t *testing.T) {
	db := setupTestDB()
	service := NewMigrationService(db)

	// Test concurrent registration
	done := make(chan bool, 2)

	go func() {
		worker := NewMockSeedWorker("concurrent-1", "Concurrent 1")
		service.Register(worker)
		done <- true
	}()

	go func() {
		worker := NewMockSeedWorker("concurrent-2", "Concurrent 2")
		service.Register(worker)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should have both workers registered
	registered := service.GetRegisteredSeeds()
	assert.Equal(t, 2, len(registered))
	assert.Contains(t, registered, "concurrent-1")
	assert.Contains(t, registered, "concurrent-2")
}
