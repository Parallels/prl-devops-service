package parallelsdesktop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/processlauncher"
)

// Helper functions
func createTestContext() basecontext.ApiContext {
	return basecontext.NewRootBaseContext()
}

func createTestParallelsService() *ParallelsService {
	ctx := createTestContext()
	service := &ParallelsService{
		ctx:              ctx,
		eventsProcessing: false,
		cachedLocalVms:   []models.ParallelsVM{},
		executable:       "/usr/local/bin/prlctl",
		ProcessLauncher:  &processlauncher.MockProcessLauncher{},
	}
	return service
}

func createMockVMs() []models.ParallelsVM {
	return []models.ParallelsVM{
		{
			ID:          "vm-test-123",
			Name:        "TestVM1",
			Description: "Test Virtual Machine 1",
			State:       "running",
			User:        "testuser",
		},
		{
			ID:          "vm-test-456",
			Name:        "TestVM2",
			Description: "Test Virtual Machine 2",
			State:       "stopped",
			User:        "testuser",
		},
	}
}

// Tests
func TestListenToParallelsEvents(t *testing.T) {
	t.Run("AlreadyProcessing", func(t *testing.T) {
		service := createTestParallelsService()
		service.eventsProcessing = true
		service.listenToParallelsEvents(service.ctx)
		if !service.eventsProcessing {
			t.Error("Expected eventsProcessing to remain true")
		}
	})

	t.Run("WithMockProcessLauncher", func(t *testing.T) {
		service := createTestParallelsService()
		mockOutput := "{}"
		r, w, _ := os.Pipe()
		go func() {
			w.WriteString(mockOutput + "\n")
			time.Sleep(100 * time.Millisecond)
			w.Close()
		}()
		mockProcessLauncher := &processlauncher.MockProcessLauncher{
			LaunchFunc: func(cmd *exec.Cmd) (*os.File, error) {
				return r, nil
			},
		}
		service.ProcessLauncher = mockProcessLauncher
		if service.eventsProcessing {
			t.Error("Expected eventsProcessing to be false initially")
		}
	})
}

func TestProcessEventsChannel(t *testing.T) {
	t.Run("ProcessesEventsSuccessfully", func(t *testing.T) {
		service := createTestParallelsService()
		ctx, cancel := context.WithCancel(context.Background())
		service.listenerCtx = ctx
		service.cancelFunc = cancel
		go service.processEventsChannel(service.ctx)
		testEvent := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-test-123",
			EventName: "vm_state_changed",
			AdditionalInfo: &models.AdditionalInfo{
				VmStateName: "running",
			},
		}
		time.Sleep(50 * time.Millisecond)
		select {
		case eventsChannel <- testEvent:
		case <-time.After(100 * time.Millisecond):
			t.Error("Failed to send event to channel")
		}
		time.Sleep(50 * time.Millisecond)
		cancel()
		time.Sleep(50 * time.Millisecond)
	})
}

func TestStopListeners(t *testing.T) {
	t.Run("StopsWhenProcessing", func(t *testing.T) {
		service := createTestParallelsService()
		ctx, cancel := context.WithCancel(context.Background())
		service.listenerCtx = ctx
		service.cancelFunc = cancel
		service.eventsProcessing = true
		service.StopListeners()
		if service.eventsProcessing {
			t.Error("Expected eventsProcessing to be false after stopping")
		}
	})

	t.Run("NoOpWhenNotProcessing", func(t *testing.T) {
		service := createTestParallelsService()
		service.eventsProcessing = false
		service.StopListeners()
		if service.eventsProcessing {
			t.Error("Expected eventsProcessing to remain false")
		}
	})
}

func TestProcessEvent(t *testing.T) {
	service := createTestParallelsService()

	t.Run("VmStateChanged", func(t *testing.T) {
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-test-123",
			EventName: "vm_state_changed",
			AdditionalInfo: &models.AdditionalInfo{
				VmStateName: "running",
			},
		}
		service.cachedLocalVms = createMockVMs()
		service.processEvent(service.ctx, event)
		service.RLock()
		found := false
		for _, vm := range service.cachedLocalVms {
			if vm.ID == "vm-test-123" {
				found = true
				if vm.State != "running" {
					t.Errorf("Expected VM state to be 'running', got '%s'", vm.State)
				}
			}
		}
		service.RUnlock()
		if !found {
			t.Error("Expected to find VM in cache")
		}
	})

	t.Run("UnsupportedEvent", func(t *testing.T) {
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-test-123",
			EventName: "unsupported_event",
		}
		service.processEvent(service.ctx, event)
	})
}

func TestProcessVmStateChanged(t *testing.T) {
	t.Run("UpdatesExistingVM", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-test-123",
			EventName: "vm_state_changed",
			AdditionalInfo: &models.AdditionalInfo{
				VmStateName: "suspended",
			},
		}
		service.processVmStateChanged(service.ctx, event)
		service.RLock()
		found := false
		for _, vm := range service.cachedLocalVms {
			if vm.ID == "vm-test-123" {
				found = true
				if vm.State != "suspended" {
					t.Errorf("Expected state 'suspended', got '%s'", vm.State)
				}
			}
		}
		service.RUnlock()
		if !found {
			t.Error("VM not found in cache")
		}
	})

	t.Run("NoAdditionalInfo", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-test-123",
			EventName: "vm_state_changed",
		}
		originalState := service.cachedLocalVms[0].State
		service.processVmStateChanged(service.ctx, event)
		service.RLock()
		if service.cachedLocalVms[0].State != originalState {
			t.Error("State should not have changed without AdditionalInfo")
		}
		service.RUnlock()
	})

	t.Run("ConcurrentStateUpdates", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()
		var wg sync.WaitGroup
		numGoroutines := 10
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(iteration int) {
				defer wg.Done()
				event := models.ParallelsServiceEvent{
					Timestamp: fmt.Sprintf("2024-01-01 12:00:%02d", iteration),
					VMID:      "vm-test-123",
					EventName: "vm_state_changed",
					AdditionalInfo: &models.AdditionalInfo{
						VmStateName: fmt.Sprintf("state-%d", iteration),
					},
				}
				service.processVmStateChanged(service.ctx, event)
			}(i)
		}
		wg.Wait()
		service.RLock()
		if len(service.cachedLocalVms) == 0 {
			t.Error("Cache should not be empty")
		}
		service.RUnlock()
	})
}

func TestProcessVmAdded(t *testing.T) {
	t.Run("RefreshesCache", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-new-789",
			EventName: "vm_added",
		}
		service.processVmAdded(service.ctx, event)
	})
}

func TestProcessVmUnregistered(t *testing.T) {
	t.Run("RefreshesCache", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-test-123",
			EventName: "vm_unregistered",
		}
		service.processVmUnregistered(service.ctx, event)
	})
}

func TestRefreshCache(t *testing.T) {
	t.Run("RefreshesCache", func(t *testing.T) {
		service := createTestParallelsService()
		initialCache := createMockVMs()
		service.cachedLocalVms = initialCache

		// Call refreshCache - behavior depends on environment
		service.refreshCache(service.ctx)

		service.RLock()
		cacheLen := len(service.cachedLocalVms)
		service.RUnlock()

		// In dev environment with prlctl: cache will have real VMs (cacheLen >= 0)
		// In CI/CD without prlctl: cache will be cleared (cacheLen == 0)
		// Both are valid - just verify cache is not nil and function doesn't panic
		if service.cachedLocalVms == nil {
			t.Error("Expected cache to be initialized (not nil)")
		}

		// Log the result for visibility
		t.Logf("Cache refresh completed with %d VMs (environment-dependent)", cacheLen)
	})

	t.Run("ConcurrentRefresh", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()
		var wg sync.WaitGroup
		numGoroutines := 5
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				service.refreshCache(service.ctx)
			}()
		}
		wg.Wait()
		service.RLock()
		cacheNotNil := service.cachedLocalVms != nil
		service.RUnlock()
		if !cacheNotNil {
			t.Error("Expected cache to be non-nil after concurrent refreshes")
		}
	})
}

func TestIsEventSupported(t *testing.T) {
	tests := []struct {
		name     string
		event    models.ParallelsServiceEvent
		expected bool
	}{
		{"SupportedVmStateChanged", models.ParallelsServiceEvent{EventName: "vm_state_changed"}, true},
		{"SupportedVmAdded", models.ParallelsServiceEvent{EventName: "vm_added"}, true},
		{"SupportedVmUnregistered", models.ParallelsServiceEvent{EventName: "vm_unregistered"}, true},
		{"UnsupportedEvent", models.ParallelsServiceEvent{EventName: "vm_deleted"}, false},
		{"EmptyEventName", models.ParallelsServiceEvent{EventName: ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEventSupported(tt.event)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for event %s", tt.expected, result, tt.event.EventName)
			}
		})
	}
}

func BenchmarkProcessVmStateChanged(b *testing.B) {
	service := createTestParallelsService()
	service.cachedLocalVms = createMockVMs()
	event := models.ParallelsServiceEvent{
		Timestamp: "2024-01-01 12:00:00",
		VMID:      "vm-test-123",
		EventName: "vm_state_changed",
		AdditionalInfo: &models.AdditionalInfo{
			VmStateName: "running",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.processVmStateChanged(service.ctx, event)
	}
}

func TestRaceConditions(t *testing.T) {
	t.Run("ConcurrentCacheAccess", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()
		var wg sync.WaitGroup
		numReaders := 10
		numWriters := 5
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				service.RLock()
				_ = len(service.cachedLocalVms)
				service.RUnlock()
			}()
		}
		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(iteration int) {
				defer wg.Done()
				event := models.ParallelsServiceEvent{
					Timestamp: fmt.Sprintf("2024-01-01 12:00:%02d", iteration),
					VMID:      "vm-test-123",
					EventName: "vm_state_changed",
					AdditionalInfo: &models.AdditionalInfo{
						VmStateName: "running",
					},
				}
				service.processVmStateChanged(service.ctx, event)
			}(i)
		}
		wg.Wait()
	})
}

func TestProcessEventAdditional(t *testing.T) {
	service := createTestParallelsService()

	t.Run("VmAdded", func(t *testing.T) {
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-new-789",
			EventName: "vm_added",
		}
		service.cachedLocalVms = createMockVMs()
		service.processEvent(service.ctx, event)
	})

	t.Run("VmUnregistered", func(t *testing.T) {
		event := models.ParallelsServiceEvent{
			Timestamp: "2024-01-01 12:00:00",
			VMID:      "vm-test-123",
			EventName: "vm_unregistered",
		}
		service.cachedLocalVms = createMockVMs()
		service.processEvent(service.ctx, event)
	})
}

func TestProcessVmStateChangedEdgeCases(t *testing.T) {
	t.Run("MultipleVMsWithSameEvent", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()

		events := []models.ParallelsServiceEvent{
			{
				Timestamp: "2024-01-01 12:00:00",
				VMID:      "vm-test-123",
				EventName: "vm_state_changed",
				AdditionalInfo: &models.AdditionalInfo{
					VmStateName: "running",
				},
			},
			{
				Timestamp: "2024-01-01 12:00:01",
				VMID:      "vm-test-456",
				EventName: "vm_state_changed",
				AdditionalInfo: &models.AdditionalInfo{
					VmStateName: "stopped",
				},
			},
		}

		for _, event := range events {
			service.processVmStateChanged(service.ctx, event)
		}

		service.RLock()
		vm1Found := false
		vm2Found := false
		for _, vm := range service.cachedLocalVms {
			if vm.ID == "vm-test-123" && vm.State == "running" {
				vm1Found = true
			}
			if vm.ID == "vm-test-456" && vm.State == "stopped" {
				vm2Found = true
			}
		}
		service.RUnlock()

		if !vm1Found || !vm2Found {
			t.Error("Expected both VMs to be updated")
		}
	})
}

func TestGetFilteredUsers(t *testing.T) {
	t.Run("ExecutesWithoutPanic", func(t *testing.T) {
		service := createTestParallelsService()
		_, err := service.getFilteredUsers(service.ctx)
		_ = err
	})
}

func TestConcurrentEventProcessing(t *testing.T) {
	t.Run("MixedEvents", func(t *testing.T) {
		service := createTestParallelsService()
		service.cachedLocalVms = createMockVMs()

		var wg sync.WaitGroup

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				event := models.ParallelsServiceEvent{
					Timestamp: fmt.Sprintf("2024-01-01 12:00:%02d", idx),
					VMID:      "vm-test-123",
					EventName: "vm_state_changed",
					AdditionalInfo: &models.AdditionalInfo{
						VmStateName: "running",
					},
				}
				service.processEvent(service.ctx, event)
			}(i)
		}

		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				event := models.ParallelsServiceEvent{
					Timestamp: fmt.Sprintf("2024-01-01 12:00:%02d", idx+10),
					VMID:      fmt.Sprintf("vm-new-%d", idx),
					EventName: "vm_added",
				}
				service.processEvent(service.ctx, event)
			}(i)
		}

		wg.Wait()
	})
}
