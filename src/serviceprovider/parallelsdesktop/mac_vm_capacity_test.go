package parallelsdesktop

import (
	"sync"
	"testing"
)

func TestConcurrentMacVMStartsClaimOnlyRemainingSlot(t *testing.T) {
	s := &ParallelsService{}
	var wg sync.WaitGroup
	accepted := make(chan string, 2)
	for _, id := range []string{"second", "third"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			s.macAdmissionMu.Lock()
			defer s.macAdmissionMu.Unlock()
			if s.claimMacVMSlot(id, []string{"first"}) == nil {
				accepted <- id
			}
		}(id)
	}
	wg.Wait()
	close(accepted)
	if len(accepted) != 1 {
		t.Fatalf("expected exactly one accepted start, got %d", len(accepted))
	}
	for id := range accepted {
		if got := s.pendingMacVMs([]string{"first", id}); got != 0 {
			t.Fatalf("double counted active reservation: %d", got)
		}
	}
}

func TestPausedMacVMCanResumeAtCapacity(t *testing.T) {
	s := &ParallelsService{}
	if err := s.claimMacVMSlot("paused", []string{"running", "paused"}); err != nil {
		t.Fatal(err)
	}
	if err := s.claimMacVMSlot("third", []string{"running", "paused"}); err == nil {
		t.Fatal("third VM admitted")
	}
	if err := s.claimMacVMSlot("paused", []string{"running", "paused"}); err == nil {
		t.Fatal("duplicate start admitted")
	}
}
