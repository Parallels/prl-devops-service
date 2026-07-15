package data

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupConcurrencyTestDB creates a database instance for concurrency testing
func setupConcurrencyTestDB(t *testing.T) (*JsonDatabase, string, basecontext.ApiContext) {
	tmpDir, err := os.MkdirTemp("", "prl-devops-concurrency-test-*")
	require.NoError(t, err)

	dbFile := filepath.Join(tmpDir, "concurrency_test_db.json")

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	_ = config.New(ctx)
	memoryDatabase = nil

	db := NewJsonDatabase(ctx, dbFile)
	require.NotNil(t, db)

	return db, tmpDir, ctx
}

// TestConcurrentReadsWrites tests concurrent read and write operations
func TestConcurrentReadsWrites(t *testing.T) {
	db, tmpDir, ctx := setupConcurrencyTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	const (
		numGoroutines = 50
		numOperations = 100
	)

	var (
		wg          sync.WaitGroup
		createCount int32
		readCount   int32
		updateCount int32
		errorCount  int32
	)

	// Start multiple goroutines performing mixed operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				operation := rand.Intn(100)

				// 40% reads, 30% creates, 30% updates
				switch {
				case operation < 40:
					// Read operation
					_, err := db.GetOrchestratorHosts(ctx, "")
					if err != nil && err != ErrDatabaseNotConnected {
						atomic.AddInt32(&errorCount, 1)
						t.Logf("Read error: %v", err)
					} else {
						atomic.AddInt32(&readCount, 1)
					}

				case operation < 70:
					// Create operation
					host := models.OrchestratorHost{
						Host:        fmt.Sprintf("concurrent-host-%d-%d.example.com", id, j),
						Description: fmt.Sprintf("Created by goroutine %d", id),
					}
					_, err := db.CreateOrchestratorHost(ctx, host)
					if err != nil && err != ErrDatabaseNotConnected {
						atomic.AddInt32(&errorCount, 1)
						t.Logf("Create error: %v", err)
					} else {
						atomic.AddInt32(&createCount, 1)
					}

				default:
					// Update operation
					hosts, err := db.GetOrchestratorHosts(ctx, "")
					if err == nil && len(hosts) > 0 {
						// Update a random host
						hostToUpdate := &hosts[rand.Intn(len(hosts))]
						hostToUpdate.Description = fmt.Sprintf("Updated by goroutine %d at %d", id, time.Now().Unix())
						_, err = db.UpdateOrchestratorHost(ctx, hostToUpdate)
						if err != nil && err != ErrDatabaseNotConnected {
							atomic.AddInt32(&errorCount, 1)
							t.Logf("Update error: %v", err)
						} else {
							atomic.AddInt32(&updateCount, 1)
						}
					}
				}

				// Small random sleep to simulate real-world timing
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify no race conditions occurred
	t.Logf("Operations completed: Creates=%d, Reads=%d, Updates=%d, Errors=%d",
		atomic.LoadInt32(&createCount),
		atomic.LoadInt32(&readCount),
		atomic.LoadInt32(&updateCount),
		atomic.LoadInt32(&errorCount))

	// Verify database is still consistent
	hosts, err := db.GetOrchestratorHosts(ctx, "")
	assert.NoError(t, err)
	t.Logf("Total hosts in database: %d", len(hosts))

	// Error count should be low (some errors expected due to concurrent operations)
	assert.Less(t, int(atomic.LoadInt32(&errorCount)), numGoroutines*numOperations/10,
		"Too many errors occurred during concurrent operations")
}

// TestConcurrentSaveOperations tests concurrent save operations
func TestConcurrentSaveOperations(t *testing.T) {
	db, tmpDir, ctx := setupConcurrencyTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	const (
		numGoroutines = 20
		numSaves      = 10
	)

	var (
		wg         sync.WaitGroup
		saveCount  int32
		errorCount int32
	)

	// Create some initial data
	for i := 0; i < 10; i++ {
		host := models.OrchestratorHost{
			Host: fmt.Sprintf("save-test-host-%d.example.com", i),
		}
		_, err := db.CreateOrchestratorHost(ctx, host)
		require.NoError(t, err)
	}

	// Start multiple goroutines performing saves
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numSaves; j++ {
				// Modify data
				db.dataMutex.Lock()
				db.data.Users = append(db.data.Users, models.User{
					ID:    fmt.Sprintf("concurrent-user-%d-%d", id, j),
					Email: fmt.Sprintf("user%d-%d@test.com", id, j),
				})
				db.dataMutex.Unlock()

				// Save
				err := db.SaveNow(ctx)
				if err != nil {
					atomic.AddInt32(&errorCount, 1)
					t.Logf("Save error from goroutine %d: %v", id, err)
				} else {
					atomic.AddInt32(&saveCount, 1)
				}

				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}(i)
	}

	// Wait for all saves to complete
	wg.Wait()

	t.Logf("Save operations completed: Successful=%d, Errors=%d",
		atomic.LoadInt32(&saveCount),
		atomic.LoadInt32(&errorCount))

	// Final save to ensure all data is persisted
	err := db.SaveNow(ctx)
	assert.NoError(t, err)

	// Verify file exists and is not corrupted
	fileInfo, err := os.Stat(db.Filename())
	assert.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0))

	// Verify data integrity by reloading
	memoryDatabase = nil
	db2 := NewJsonDatabase(ctx, db.Filename())
	defer func() {
		if db2.cancel != nil {
			close(db2.cancel)
		}
	}()

	db2.dataMutex.RLock()
	userCount := len(db2.data.Users)
	hostCount := len(db2.data.OrchestratorHosts)
	db2.dataMutex.RUnlock()

	t.Logf("After reload: Users=%d, Hosts=%d", userCount, hostCount)
	assert.Greater(t, userCount, 0, "Users should have been saved")
	assert.Equal(t, 10, hostCount, "All hosts should be preserved")
}

// TestDeadlockPrevention tests that the database doesn't deadlock under stress
func TestDeadlockPrevention(t *testing.T) {
	db, tmpDir, ctx := setupConcurrencyTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	const (
		numGoroutines = 30
		duration      = 2 * time.Second
	)

	var (
		wg              sync.WaitGroup
		operationCount  int32
		timeoutOccurred atomic.Bool
	)

	// Create some initial data
	for i := 0; i < 5; i++ {
		host := models.OrchestratorHost{
			Host: fmt.Sprintf("deadlock-test-host-%d.example.com", i),
		}
		_, err := db.CreateOrchestratorHost(ctx, host)
		require.NoError(t, err)
	}

	// Channel to signal completion
	done := make(chan struct{})
	timeout := time.After(duration + 5*time.Second) // Extra time for cleanup

	// Start goroutines with random operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			localCtx := basecontext.NewRootBaseContext()
			localCtx.DisableLog()

			start := time.Now()
			for time.Since(start) < duration {
				select {
				case <-done:
					return
				default:
					// Perform random operation
					switch rand.Intn(5) {
					case 0:
						// Read all hosts
						_, _ = db.GetOrchestratorHosts(localCtx, "")
					case 1:
						// Get specific host
						hosts, err := db.GetOrchestratorHosts(localCtx, "")
						if err == nil && len(hosts) > 0 {
							_, _ = db.GetOrchestratorHost(localCtx, hosts[0].ID)
						}
					case 2:
						// Update a host
						hosts, err := db.GetOrchestratorHosts(localCtx, "")
						if err == nil && len(hosts) > 0 {
							host := &hosts[rand.Intn(len(hosts))]
							host.Description = fmt.Sprintf("Updated at %v", time.Now().UnixNano())
							_, _ = db.UpdateOrchestratorHost(localCtx, host)
						}
					case 3:
						// Get resources (read-heavy operation)
						_ = db.GetOrchestratorAvailableResources(localCtx)
					case 4:
						// Save data operation
						db.dataMutex.Lock()
						db.data.Users = append(db.data.Users, models.User{
							ID:    fmt.Sprintf("deadlock-user-%d-%d", id, time.Now().UnixNano()),
							Email: "test@example.com",
						})
						db.dataMutex.Unlock()
					}

					atomic.AddInt32(&operationCount, 1)
					time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for goroutines with timeout detection
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Logf("All operations completed successfully. Total operations: %d", atomic.LoadInt32(&operationCount))
	case <-timeout:
		timeoutOccurred.Store(true)
		t.Error("Test timed out - possible deadlock detected!")
		close(done)
	}

	// Verify database is still functional
	if !timeoutOccurred.Load() {
		hosts, err := db.GetOrchestratorHosts(ctx, "")
		assert.NoError(t, err)
		t.Logf("Final host count: %d", len(hosts))
	}
}

// TestRaceConditions tests for race conditions in data structures
func TestRaceConditions(t *testing.T) {
	db, tmpDir, ctx := setupConcurrencyTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	const (
		numGoroutines = 20
		numIterations = 50
	)

	var wg sync.WaitGroup

	// Test 1: Concurrent modifications to OrchestratorHosts
	t.Run("OrchestratorHostsConcurrency", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < numIterations; j++ {
					// Create
					host := models.OrchestratorHost{
						Host: fmt.Sprintf("race-host-%d-%d.example.com", id, j),
						Resources: &models.HostResources{
							CpuType: "arm64",
							Total: models.HostResourceItem{
								PhysicalCpuCount: int64(rand.Intn(16) + 1),
								LogicalCpuCount:  int64(rand.Intn(32) + 1),
							},
						},
					}
					_, _ = db.CreateOrchestratorHost(ctx, host)

					// Read
					_, _ = db.GetOrchestratorHosts(ctx, "")

					// Update if exists
					hosts, err := db.GetOrchestratorHosts(ctx, "")
					if err == nil && len(hosts) > 0 {
						h := &hosts[rand.Intn(len(hosts))]
						h.Description = fmt.Sprintf("Race test %d", id)
						_, _ = db.UpdateOrchestratorHost(ctx, h)
					}
				}
			}(i)
		}

		wg.Wait()
	})

	// Test 2: Concurrent resource aggregations
	t.Run("ResourceAggregationConcurrency", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < numIterations; j++ {
					_ = db.GetOrchestratorAvailableResources(ctx)
					_ = db.GetOrchestratorTotalResources(ctx)
					_ = db.GetOrchestratorInUseResources(ctx)
					_ = db.GetOrchestratorReservedResources(ctx)
					_ = db.GetOrchestratorSystemReservedResources(ctx)
				}
			}()
		}

		wg.Wait()
	})

	// Test 3: Concurrent save and read operations
	t.Run("SaveReadConcurrency", func(t *testing.T) {
		for i := 0; i < numGoroutines/2; i++ {
			// Readers
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < numIterations; j++ {
					db.dataMutex.RLock()
					_ = len(db.data.OrchestratorHosts)
					_ = len(db.data.Users)
					db.dataMutex.RUnlock()
					time.Sleep(time.Millisecond)
				}
			}()

			// Writers
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < numIterations; j++ {
					db.dataMutex.Lock()
					db.data.Users = append(db.data.Users, models.User{
						ID:    fmt.Sprintf("race-user-%d-%d", id, j),
						Email: fmt.Sprintf("race%d-%d@test.com", id, j),
					})
					db.dataMutex.Unlock()
					time.Sleep(time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
	})

	// Verify final state is consistent
	hosts, err := db.GetOrchestratorHosts(ctx, "")
	assert.NoError(t, err)
	t.Logf("Final state: %d hosts", len(hosts))

	db.dataMutex.RLock()
	userCount := len(db.data.Users)
	db.dataMutex.RUnlock()
	t.Logf("Final state: %d users", userCount)
	assert.Greater(t, userCount, 0)
}
