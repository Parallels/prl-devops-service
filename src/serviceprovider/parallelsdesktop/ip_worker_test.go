package parallelsdesktop

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

// --- fixtures ---------------------------------------------------------------

func ipTestVM(id, name, osType, state, ip string) models.ParallelsVM {
	return models.ParallelsVM{
		ID:                id,
		Name:              name,
		OS:                osType,
		State:             state,
		User:              "testuser",
		InternalIpAddress: ip,
	}
}

// ipTestService returns a test service with one cached VM and deterministic
// IP-probe hooks: waits are virtual (they advance the fake clock) and attempts
// are provided by attemptFn.
func ipTestService(t *testing.T, vm models.ParallelsVM, attemptFn func(ctx basecontext.ApiContext, vmID string) (string, bool)) (*ParallelsService, *time.Time, *[]time.Duration) {
	t.Helper()
	service := createTestParallelsService()
	service.cachedLocalVms = []models.ParallelsVM{vm}

	now := time.Unix(1_700_000_000, 0)
	var mu sync.Mutex
	var waits []time.Duration

	service.ipProbeNow = func() time.Time { return now }
	service.ipProbeWait = func(ctx context.Context, d time.Duration) bool {
		mu.Lock()
		waits = append(waits, d)
		mu.Unlock()
		select {
		case <-ctx.Done():
			return false
		default:
		}
		now = now.Add(d) // virtual sleep
		return true
	}
	service.ipProbeAttempt = attemptFn
	return service, &now, &waits
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting: %s", msg)
}

// --- eligibility -------------------------------------------------------------

func TestShouldProbeIp(t *testing.T) {
	cases := []struct {
		name string
		vm   models.ParallelsVM
		want bool
	}{
		{"running-no-ip-any-os", ipTestVM("a", "linux-box", "linux", "running", ""), true},
		{"running-no-ip-unknown-tools", ipTestVM("b", "any", "", "running", ""), true},
		{"running-no-ip-tools-not-installed", func() models.ParallelsVM {
			vm := ipTestVM("c", "any", "linux", "running", "")
			vm.GuestTools.State = "not installed"
			return vm
		}(), false},
		{"running-no-ip-tools-installed", func() models.ParallelsVM {
			vm := ipTestVM("d", "any", "linux", "running", "")
			vm.GuestTools.State = "guest tools installed"
			return vm
		}(), true},
		{"stopped-no-ip", ipTestVM("e", "any", "linux", "stopped", ""), false},
		{"running-with-ip-linux", ipTestVM("f", "any", "linux", "running", "10.0.0.1"), false},
		{"macos-with-ip-unresolved", ipTestVM("g", "runner", "macosx", "running", "10.0.0.2"), true},
		{"macos-by-name-with-ip-unresolved", ipTestVM("h", "mac-mini-01", "unknown", "running", "10.0.0.3"), true},
		{"macos-with-ip-resolved", func() models.ParallelsVM {
			vm := ipTestVM("i", "mac-mini-01", "macosx", "running", "10.0.0.4")
			vm.IpStatus = IpStatusResolved
			return vm
		}(), false},
		{"macos-with-failed-status-rearms", func() models.ParallelsVM {
			vm := ipTestVM("j", "mac-mini-01", "macosx", "running", "10.0.0.5")
			vm.IpStatus = IpStatusFailed
			return vm
		}(), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldProbeIp(tc.vm); got != tc.want {
				t.Errorf("shouldProbeIp(%+v) = %v, want %v", tc.vm, got, tc.want)
			}
		})
	}
}

func TestVmToolsProbablyInstalled(t *testing.T) {
	cases := map[string]struct {
		state string
		want  bool
	}{
		"empty/unknown":      {"", true},
		"installed":          {"guest tools installed", true},
		"installed-variant":  {"installed", true},
		"not-installed":      {"not installed", false},
		"not-installed-2":    {"guest tools are not installed", false},
		"notinstalled":       {"notinstalled", false},
		"not-installed-dash": {"not-installed", false},
		"updating":           {"updating", true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			vm := models.ParallelsVM{}
			vm.GuestTools.State = tc.state
			if got := vmToolsProbablyInstalled(vm); got != tc.want {
				t.Errorf("vmToolsProbablyInstalled(%q) = %v, want %v", tc.state, got, tc.want)
			}
		})
	}
}

// --- preservation merge ------------------------------------------------------

func TestPreserveResolvedIp(t *testing.T) {
	cached := ipTestVM("a", "mac", "macosx", "running", "10.10.10.5")
	cached.IpStatus = IpStatusResolved

	t.Run("live-silence-keeps-cached-ip", func(t *testing.T) {
		live := ipTestVM("a", "mac", "macosx", "running", "")
		merged := preserveResolvedIp(cached, live)
		if merged.InternalIpAddress != "10.10.10.5" {
			t.Errorf("expected cached IP to survive, got %q", merged.InternalIpAddress)
		}
		if merged.IpStatus != IpStatusResolved {
			t.Errorf("expected resolved status to survive, got %q", merged.IpStatus)
		}
	})

	t.Run("live-ip-wins", func(t *testing.T) {
		live := ipTestVM("a", "mac", "macosx", "running", "10.10.10.9")
		live.IpStatus = IpStatusResolved
		merged := preserveResolvedIp(cached, live)
		if merged.InternalIpAddress != "10.10.10.9" {
			t.Errorf("expected live IP to win, got %q", merged.InternalIpAddress)
		}
	})

	t.Run("live-stopped-drops-ip", func(t *testing.T) {
		live := ipTestVM("a", "mac", "macosx", "stopped", "")
		merged := preserveResolvedIp(cached, live)
		if merged.InternalIpAddress != "" {
			t.Errorf("expected no IP for stopped VM, got %q", merged.InternalIpAddress)
		}
	})

	t.Run("failed-status-carried-through", func(t *testing.T) {
		failed := cached
		failed.IpStatus = IpStatusFailed
		failed.IpStatusDetail = "no IP address obtained"
		failed.InternalIpAddress = ""
		live := ipTestVM("a", "mac", "macosx", "running", "")
		merged := preserveResolvedIp(failed, live)
		if merged.IpStatus != IpStatusFailed || merged.IpStatusDetail == "" {
			t.Errorf("expected failed status to carry through, got %q/%q", merged.IpStatus, merged.IpStatusDetail)
		}
	})

	t.Run("live-status-wins", func(t *testing.T) {
		live := ipTestVM("a", "mac", "macosx", "running", "")
		live.IpStatus = IpStatusResolving
		merged := preserveResolvedIp(cached, live)
		if merged.IpStatus != IpStatusResolving {
			t.Errorf("expected live status to win, got %q", merged.IpStatus)
		}
	})
}

// --- worker lifecycle ---------------------------------------------------------

func TestIpPokeWorkerResolvesAndStops(t *testing.T) {
	vm := ipTestVM("vm-mac-1", "mac-runner", "macosx", "running", "")
	var calls int32
	service, _, waits := ipTestService(t, vm, func(_ basecontext.ApiContext, id string) (string, bool) {
		if atomic.AddInt32(&calls, 1) >= 3 {
			return "10.211.55.99", true
		}
		return "", false
	})

	service.startIpPokeWorker(service.ctx, vm.ID)

	waitFor(t, 2*time.Second, func() bool {
		cached, ok := service.snapshotVm(vm.ID)
		return ok && cached.IpStatus == IpStatusResolved
	}, "worker to resolve the IP")

	cached, _ := service.snapshotVm(vm.ID)
	if cached.InternalIpAddress != "10.211.55.99" {
		t.Errorf("expected resolved IP in cache, got %q", cached.InternalIpAddress)
	}
	if cached.IpStatusDetail != "" {
		t.Errorf("expected empty detail after resolve, got %q", cached.IpStatusDetail)
	}
	waitFor(t, 2*time.Second, func() bool {
		service.syncMu.Lock()
		defer service.syncMu.Unlock()
		return len(service.ipPokeWorkers) == 0
	}, "worker to unregister")

	// Backoff schedule before the resolve on attempt 3: 10s then 20s.
	time.Sleep(20 * time.Millisecond)
	if len(*waits) != 2 || (*waits)[0] != 10*time.Second || (*waits)[1] != 20*time.Second {
		t.Errorf("expected backoff waits [10s 20s], got %v", *waits)
	}
}

func TestIpPokeWorkerGivesUpAfterTimeout(t *testing.T) {
	vm := ipTestVM("vm-mac-2", "mac-runner", "macosx", "running", "")
	service, now, waits := ipTestService(t, vm, func(_ basecontext.ApiContext, _ string) (string, bool) {
		return "", false // never resolves
	})

	service.startIpPokeWorker(service.ctx, vm.ID)

	waitFor(t, 2*time.Second, func() bool {
		cached, ok := service.snapshotVm(vm.ID)
		return ok && cached.IpStatus == IpStatusFailed
	}, "worker to give up with failed status")

	cached, _ := service.snapshotVm(vm.ID)
	if !strings.Contains(cached.IpStatusDetail, "no IP address obtained after") {
		t.Errorf("unexpected failure detail: %q", cached.IpStatusDetail)
	}
	if cached.InternalIpAddress != "" {
		t.Errorf("failed VM should have no IP, got %q", cached.InternalIpAddress)
	}

	// The virtual clock must have been bounded by the 280s lifetime and the
	// wait schedule must be 10s-based, +10s per round, capped at 60s.
	if elapsed := (*now).Sub(time.Unix(1_700_000_000, 0)); elapsed > ipProbeTimeout {
		t.Errorf("virtual clock overshot the probe timeout: %v", elapsed)
	}
	var prev time.Duration
	for i, w := range *waits {
		want := ipProbeBaseDelay + time.Duration(i)*ipProbeBackoffStep
		if want > ipProbeMaxDelay {
			want = ipProbeMaxDelay
		}
		if w > want {
			t.Errorf("wait[%d]=%v exceeds schedule cap %v", i, w, want)
		}
		if w < prev-60*time.Second {
			t.Errorf("wait[%d]=%v out of order after %v", i, w, prev)
		}
		prev = want
	}
}

func TestIpPokeWorkerDeduplicates(t *testing.T) {
	vm := ipTestVM("vm-mac-3", "mac-runner", "macosx", "running", "")
	entered := make(chan struct{})
	release := make(chan struct{})
	var calls int32
	service, _, _ := ipTestService(t, vm, func(_ basecontext.ApiContext, _ string) (string, bool) {
		if atomic.AddInt32(&calls, 1) == 1 {
			close(entered)
			<-release // hold the worker inside its first attempt
		}
		return "10.211.55.1", true
	})

	service.startIpPokeWorker(service.ctx, vm.ID)
	<-entered
	service.startIpPokeWorker(service.ctx, vm.ID) // must be a no-op

	service.syncMu.Lock()
	n := len(service.ipPokeWorkers)
	service.syncMu.Unlock()
	if n != 1 {
		t.Fatalf("expected exactly 1 worker after double start, got %d", n)
	}

	close(release)
	waitFor(t, 2*time.Second, func() bool {
		cached, ok := service.snapshotVm(vm.ID)
		return ok && cached.IpStatus == IpStatusResolved
	}, "worker to resolve after release")
}

func TestIpPokeWorkerStopIsClean(t *testing.T) {
	vm := ipTestVM("vm-mac-4", "mac-runner", "macosx", "running", "")
	entered := make(chan struct{})
	release := make(chan struct{})
	var calls int32
	service, _, _ := ipTestService(t, vm, func(_ basecontext.ApiContext, _ string) (string, bool) {
		if atomic.AddInt32(&calls, 1) == 1 {
			close(entered)
			<-release // hold the worker inside its first attempt
		}
		return "", false
	})

	service.startIpPokeWorker(service.ctx, vm.ID)
	<-entered
	service.stopIpPokeWorker(vm.ID)
	close(release) // let the held attempt return; worker must exit via cancelled ctx

	waitFor(t, 2*time.Second, func() bool {
		service.syncMu.Lock()
		defer service.syncMu.Unlock()
		return len(service.ipPokeWorkers) == 0
	}, "worker to unregister after stop")

	cached, _ := service.snapshotVm(vm.ID)
	if cached.IpStatus == IpStatusFailed {
		t.Errorf("a stopped worker must not mark the VM failed: %q", cached.IpStatus)
	}
	if cached.IpStatus == IpStatusResolved {
		t.Errorf("unexpected resolved status: %q", cached.IpStatus)
	}
}

func TestStartIpPokeWorkerSkipsIneligible(t *testing.T) {
	vm := ipTestVM("vm-mac-5", "mac-runner", "macosx", "running", "")
	vm.GuestTools.State = "not installed"
	service, _, _ := ipTestService(t, vm, nil)

	service.startIpPokeWorker(service.ctx, vm.ID)

	service.syncMu.Lock()
	n := len(service.ipPokeWorkers)
	service.syncMu.Unlock()
	if n != 0 {
		t.Errorf("expected no worker for tools-missing VM, got %d", n)
	}
}

func TestNotifyMissingIpsArmsOnlyEligible(t *testing.T) {
	eligible := ipTestVM("vm-a", "mac-runner", "macosx", "running", "")
	stopped := ipTestVM("vm-b", "mac-stop", "macosx", "stopped", "")
	linuxWithIp := ipTestVM("vm-c", "linux", "linux", "running", "10.0.0.1")

	service := createTestParallelsService()
	service.cachedLocalVms = []models.ParallelsVM{eligible, stopped, linuxWithIp}
	service.ipProbeWait = func(ctx context.Context, _ time.Duration) bool {
		<-ctx.Done()
		return false
	}
	service.ipProbeAttempt = func(_ basecontext.ApiContext, _ string) (string, bool) { return "", false }

	service.NotifyMissingIps([]models.ParallelsVM{eligible, stopped, linuxWithIp})

	waitFor(t, 2*time.Second, func() bool {
		service.syncMu.Lock()
		defer service.syncMu.Unlock()
		_, ok := service.ipPokeWorkers[eligible.ID]
		return ok && len(service.ipPokeWorkers) == 1
	}, "only the eligible VM to get a worker")

	service.stopAllIpPokeWorkers()
	waitFor(t, 2*time.Second, func() bool {
		service.syncMu.Lock()
		defer service.syncMu.Unlock()
		return len(service.ipPokeWorkers) == 0
	}, "all workers to stop")
}

func TestResetIpStatusForStop(t *testing.T) {
	vm := ipTestVM("vm-mac-6", "mac-runner", "macosx", "running", "10.211.55.7")
	vm.IpStatus = IpStatusFailed
	vm.IpStatusDetail = "boom"
	service := createTestParallelsService()
	service.cachedLocalVms = []models.ParallelsVM{vm}

	service.resetIpStatusForStop(service.ctx, vm.ID)

	cached, ok := service.snapshotVm(vm.ID)
	if !ok {
		t.Fatal("VM missing from cache")
	}
	if cached.IpStatus != "" || cached.IpStatusDetail != "" {
		t.Errorf("expected cleared status, got %q/%q", cached.IpStatus, cached.IpStatusDetail)
	}
	if cached.InternalIpAddress != "10.211.55.7" {
		t.Errorf("expected last known IP to be kept, got %q", cached.InternalIpAddress)
	}
}

func TestMutateCachedVmNoopReturnsNil(t *testing.T) {
	vm := ipTestVM("vm-mac-7", "mac-runner", "macosx", "running", "")
	service := createTestParallelsService()
	service.cachedLocalVms = []models.ParallelsVM{vm}

	changed := service.mutateCachedVm(vm.ID, func(*models.ParallelsVM) {})
	if changed != nil {
		t.Error("no-op mutation should report no change")
	}
	changed = service.mutateCachedVm("does-not-exist", func(e *models.ParallelsVM) { e.IpStatus = IpStatusResolving })
	if changed != nil {
		t.Error("unknown VM should report no change")
	}
	changed = service.mutateCachedVm(vm.ID, func(e *models.ParallelsVM) { e.IpStatus = IpStatusResolving })
	if changed == nil || changed.IpStatus != IpStatusResolving {
		t.Error("expected changed VM copy for real mutation")
	}
}
