package parallelsdesktop

import (
	"fmt"
	"sync"
	"testing"
	"time"

	basecontext "github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestListenToParallelsEvents(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)
	defer s.StopListeners()

	// Test initial start
	s.listenToParallelsEvents(ctx)
	assert.True(t, s.eventsProcessing)

	// Test calling again (should not start another listener)
	s.listenToParallelsEvents(ctx)
	assert.True(t, s.eventsProcessing)

	time.Sleep(100 * time.Millisecond)

	// Test concurrent calls
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.listenToParallelsEvents(ctx)
		}()
	}
	wg.Wait()
	assert.True(t, s.eventsProcessing)

	// Simulate events channel being full (hard to test directly, but ensure no panic)
	for i := 0; i < 1000; i++ {
		select {
		case eventsChannel <- models.ParallelsServiceEvent{}:
		default:
			// Channel full, skip
		}
	}
	time.Sleep(50 * time.Millisecond)
	assert.True(t, s.eventsProcessing)
}

func TestStopListeners(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)

	// Stop without starting
	s.StopListeners()
	assert.False(t, s.eventsProcessing)

	// Start and then stop
	s.listenToParallelsEvents(ctx)
	assert.True(t, s.eventsProcessing)

	s.StopListeners()
	assert.False(t, s.eventsProcessing)

	// Stop again (should be idempotent)
	s.StopListeners()
	assert.False(t, s.eventsProcessing)

	// Test with multiple stops concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.StopListeners()
		}()
	}
	wg.Wait()
	assert.False(t, s.eventsProcessing)
}

func TestProcessEvent(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)

	// Mock VMs
	s.cachedLocalVms = []models.ParallelsVM{
		{ID: "test-vm-1", State: "stopped"},
		{ID: "test-vm-2", State: "running"},
	}

	// Test vm_state_changed
	event1 := models.ParallelsServiceEvent{
		EventName:      "vm_state_changed",
		VMID:           "test-vm-1",
		AdditionalInfo: &models.AdditionalInfo{VmStateName: "running"},
	}
	s.processEvent(ctx, event1)
	assert.Equal(t, "running", s.cachedLocalVms[0].State)

	// Test vm_added
	event2 := models.ParallelsServiceEvent{
		EventName: "vm_added",
		VMID:      "test-vm-3",
	}
	initialLen := len(s.cachedLocalVms)
	s.processEvent(ctx, event2)
	time.Sleep(100 * time.Millisecond) // Allow refreshCache to complete
	assert.GreaterOrEqual(t, len(s.cachedLocalVms), initialLen)

	// Test vm_unregistered
	event3 := models.ParallelsServiceEvent{
		EventName: "vm_unregistered",
		VMID:      "test-vm-2",
	}
	s.processEvent(ctx, event3)
	time.Sleep(100 * time.Millisecond)

	// Test unknown event (should log and do nothing)
	event4 := models.ParallelsServiceEvent{
		EventName: "unknown",
		VMID:      "test-vm-4",
	}
	s.processEvent(ctx, event4) // Should not panic

	// Test nil event (edge case)
	s.processEvent(ctx, models.ParallelsServiceEvent{}) // Should handle gracefully

	// Test event with empty fields
	event5 := models.ParallelsServiceEvent{
		EventName: "",
		VMID:      "",
	}
	s.processEvent(ctx, event5)

	// Concurrent events
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := models.ParallelsServiceEvent{
				EventName:      "vm_state_changed",
				VMID:           fmt.Sprintf("test-vm-%d", i%2+1),
				AdditionalInfo: &models.AdditionalInfo{VmStateName: fmt.Sprintf("state-%d", i)},
			}
			s.processEvent(ctx, event)
		}(i)
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
}

func TestProcessVmStateChanged(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)

	s.cachedLocalVms = []models.ParallelsVM{
		{ID: "test-vm-1", State: "stopped"},
		{ID: "test-vm-2", State: "running"},
	}

	// Update existing VM
	event := models.ParallelsServiceEvent{
		VMID:           "test-vm-1",
		AdditionalInfo: &models.AdditionalInfo{VmStateName: "running"},
	}
	s.processVmStateChanged(ctx, event)
	assert.Equal(t, "running", s.cachedLocalVms[0].State)

	// No update for non-existing VM
	event2 := models.ParallelsServiceEvent{
		VMID:           "non-existing",
		AdditionalInfo: &models.AdditionalInfo{VmStateName: "paused"},
	}
	originalState := s.cachedLocalVms[0].State
	s.processVmStateChanged(ctx, event2)
	assert.Equal(t, originalState, s.cachedLocalVms[0].State) // No change

	// Nil AdditionalInfo
	event3 := models.ParallelsServiceEvent{
		VMID: "test-vm-2",
	}
	s.processVmStateChanged(ctx, event3)
	assert.Equal(t, "running", s.cachedLocalVms[1].State) // No change

	// Concurrent updates with non-empty states
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := models.ParallelsServiceEvent{
				VMID:           "test-vm-1",
				AdditionalInfo: &models.AdditionalInfo{VmStateName: fmt.Sprintf("state-%d", i+1)},
			}
			s.processVmStateChanged(ctx, event)
		}(i)
	}
	wg.Wait()
	// Last one wins, but since concurrent, just check it's not empty
	assert.NotEmpty(t, s.cachedLocalVms[0].State)
}

func TestGetFilteredUsers(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)

	// Normal case
	users, err := s.getFilteredUsers(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, users)

	// Test with concurrent calls
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.getFilteredUsers(ctx)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}

func TestProcessVmAdded(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)

	initialLen := len(s.cachedLocalVms)
	event := models.ParallelsServiceEvent{
		VMID: "new-vm",
	}
	s.processVmAdded(ctx, event)
	time.Sleep(100 * time.Millisecond) // Allow refreshCache
	assert.GreaterOrEqual(t, len(s.cachedLocalVms), initialLen)

	// Test with empty VMID
	event2 := models.ParallelsServiceEvent{
		VMID: "",
	}
	s.processVmAdded(ctx, event2)
	time.Sleep(100 * time.Millisecond)

	// Concurrent adds
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := models.ParallelsServiceEvent{
				VMID: fmt.Sprintf("vm-%d", i),
			}
			s.processVmAdded(ctx, event)
		}(i)
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
}

func TestProcessVmUnregistered(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)

	event := models.ParallelsServiceEvent{
		VMID: "removed-vm",
	}
	s.processVmUnregistered(ctx, event)
	time.Sleep(100 * time.Millisecond)

	// Test with empty VMID
	event2 := models.ParallelsServiceEvent{
		VMID: "",
	}
	s.processVmUnregistered(ctx, event2)
	time.Sleep(100 * time.Millisecond)

	// Concurrent unregisters
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := models.ParallelsServiceEvent{
				VMID: fmt.Sprintf("vm-%d", i),
			}
			s.processVmUnregistered(ctx, event)
		}(i)
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
}

func TestRefreshCache(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	s := New(ctx)

	// Normal refresh
	s.refreshCache(ctx)
	assert.NotNil(t, s.cachedLocalVms)

	// Refresh again
	s.refreshCache(ctx)
	assert.NotNil(t, s.cachedLocalVms)

	// Concurrent refresh
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.refreshCache(ctx)
		}()
	}
	wg.Wait()
	assert.NotNil(t, s.cachedLocalVms)
}
