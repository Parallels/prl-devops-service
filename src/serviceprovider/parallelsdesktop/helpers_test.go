package parallelsdesktop

import (
	"errors"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

func TestFindVmInCache(t *testing.T) {
	service := createTestParallelsService()
	service.cachedLocalVms = createMockVMs()

	t.Run("ByName", func(t *testing.T) {
		vm := service.findVmInCache("TestVM1")
		if vm == nil || vm.ID != "vm-test-123" {
			t.Fatalf("expected vm-test-123 by name, got %v", vm)
		}
	})
	t.Run("ById", func(t *testing.T) {
		vm := service.findVmInCache("vm-test-456")
		if vm == nil || vm.Name != "TestVM2" {
			t.Fatalf("expected TestVM2 by id, got %v", vm)
		}
	})
	t.Run("CaseInsensitive", func(t *testing.T) {
		if vm := service.findVmInCache("testvm1"); vm == nil {
			t.Fatal("expected case-insensitive name match")
		}
	})
	t.Run("Miss", func(t *testing.T) {
		if vm := service.findVmInCache("ghost-vm"); vm != nil {
			t.Fatalf("expected nil for missing vm, got %v", vm)
		}
	})
}

func TestVmNotFoundBackoffLifecycle(t *testing.T) {
	// createTestParallelsService deliberately leaves notFoundVm nil, which
	// also exercises nil-map safety for services built without New().
	service := createTestParallelsService()

	if service.vmNotFoundBackoffActive("MissingVM") {
		t.Fatal("unknown vm must not be in backoff")
	}

	service.markVmNotFound("MissingVM")
	if !service.vmNotFoundBackoffActive("missingvm") {
		t.Fatal("backoff must be active after mark (case-insensitive key)")
	}

	service.clearVmNotFoundBackoff("MISSINGVM")
	if service.vmNotFoundBackoffActive("MissingVM") {
		t.Fatal("backoff must be inactive after clear")
	}

	service.markVmNotFound("ExpiredVM")
	service.syncMu.Lock()
	service.notFoundVm["expiredvm"] = time.Now().Add(-2 * CacheFindNotFoundTTL)
	service.syncMu.Unlock()
	if service.vmNotFoundBackoffActive("ExpiredVM") {
		t.Fatal("expired entry must not keep the backoff active")
	}
	service.syncMu.Lock()
	_, stillThere := service.notFoundVm["expiredvm"]
	service.syncMu.Unlock()
	if stillThere {
		t.Fatal("expired entry must be evicted on sight")
	}
}

func TestFindVmInCacheAndSystemReturnsCachedVmImmediately(t *testing.T) {
	service := createTestParallelsService()
	service.cachedLocalVms = createMockVMs()
	service.findProbeWait = func(time.Duration) {
		t.Error("cache hit must not sleep")
	}
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		t.Error("cache hit must not trigger a system scan")
		return nil, nil
	}

	vm, err := service.findVmInCacheAndSystem(createTestContext(), "TestVM1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil || vm.ID != "vm-test-123" {
		t.Fatalf("unexpected vm: %v", vm)
	}
}

func TestFindVmInCacheAndSystemBacksOffAfterConfirmedMissing(t *testing.T) {
	service := createTestParallelsService()
	scanCalls := 0
	service.findProbeWait = func(time.Duration) {}
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		scanCalls++
		return []models.ParallelsVM{}, nil // scan succeeds, VM genuinely absent
	}
	ctx := createTestContext()

	// First lookup pays the full retry + scan cycle once.
	if _, err := service.findVmInCacheAndSystem(ctx, "gone-vm"); !errors.Is(err, ErrVirtualMachineNotFound) {
		t.Fatalf("expected ErrVirtualMachineNotFound, got %v", err)
	}
	if scanCalls != 1 {
		t.Fatalf("first lookup must scan exactly once, got %d", scanCalls)
	}

	// While the miss is fresh, later lookups (reverse proxy reloads, polling
	// clients) must fail fast: no sleeps, no extra scans.
	service.findProbeWait = func(time.Duration) {
		t.Error("backoff window must not run retry sleeps")
	}
	start := time.Now()
	_, err := service.findVmInCacheAndSystem(ctx, "GONE-VM")
	elapsed := time.Since(start)
	if !errors.Is(err, ErrVirtualMachineNotFound) {
		t.Fatalf("expected ErrVirtualMachineNotFound during backoff, got %v", err)
	}
	if scanCalls != 1 {
		t.Fatalf("backoff window must not re-scan the system, scanCalls=%d", scanCalls)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("backoff window must fail fast, took %v", elapsed)
	}
}

func TestFindVmNotFoundMarkedOnlyOnConfirmedAbsence(t *testing.T) {
	service := createTestParallelsService()
	service.findProbeWait = func(time.Duration) {}
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		return nil, errors.New("system scan failed")
	}

	if _, err := service.findVmInCacheAndSystem(createTestContext(), "flaky-vm"); !errors.Is(err, ErrVirtualMachineNotFound) {
		t.Fatalf("expected ErrVirtualMachineNotFound, got %v", err)
	}
	if service.vmNotFoundBackoffActive("flaky-vm") {
		t.Fatal("a failed system scan is not evidence of absence and must not start the backoff")
	}
}

func TestFindVmRetryDelaysGrowExponentiallyUpToCap(t *testing.T) {
	service := createTestParallelsService()
	var delays []time.Duration
	service.findProbeWait = func(d time.Duration) { delays = append(delays, d) }
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		return []models.ParallelsVM{}, nil
	}

	if _, err := service.findVmInCacheAndSystem(createTestContext(), "missing-vm"); !errors.Is(err, ErrVirtualMachineNotFound) {
		t.Fatalf("expected ErrVirtualMachineNotFound, got %v", err)
	}
	if len(delays) != CacheFindNumberOfRetries-1 {
		t.Fatalf("expected %d backoff sleeps, got %d: %v", CacheFindNumberOfRetries-1, len(delays), delays)
	}
	expected := CacheFindRetryBaseDelay
	for i, d := range delays {
		if d != expected {
			t.Fatalf("delay %d: expected %v, got %v", i, expected, d)
		}
		if d > CacheFindRetryMaxDelay {
			t.Fatalf("delay %d exceeds cap: %v", i, d)
		}
		expected *= 2
		if expected > CacheFindRetryMaxDelay {
			expected = CacheFindRetryMaxDelay
		}
	}
	if !service.vmNotFoundBackoffActive("missing-vm") {
		t.Fatal("successful scan with absent vm must start the not-found backoff")
	}
}

func TestGetVmFastSkipsRetryWaits(t *testing.T) {
	// A reverse proxy start runs route resolution inside a 10s operation
	// timeout: the fast path must reach the scan without any retry sleeps.
	service := createTestParallelsService()
	scanCalls := 0
	service.findProbeWait = func(time.Duration) {
		t.Error("GetVmFast must not run retry sleeps")
	}
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		scanCalls++
		return nil, errors.New("scan failed")
	}

	if vm, err := service.GetVmFast(createTestContext(), "ghost-vm"); err == nil || vm != nil {
		t.Fatalf("expected nil vm with error, got %v / %v", vm, err)
	}
	if scanCalls != 1 {
		t.Fatalf("GetVmFast must run exactly one scan, got %d", scanCalls)
	}
	if service.vmNotFoundBackoffActive("ghost-vm") {
		t.Fatal("a failed scan must not mark the vm missing")
	}

	// While a confirmed miss is fresh, the fast path short-circuits too.
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		scanCalls++
		return []models.ParallelsVM{}, nil
	}
	if _, err := service.GetVmFast(createTestContext(), "gone-vm"); !errors.Is(err, ErrVirtualMachineNotFound) {
		t.Fatalf("expected ErrVirtualMachineNotFound, got %v", err)
	}
	scansAfterConfirm := scanCalls
	if _, err := service.GetVmFast(createTestContext(), "gone-vm"); !errors.Is(err, ErrVirtualMachineNotFound) {
		t.Fatalf("expected ErrVirtualMachineNotFound on repeat, got %v", err)
	}
	if scanCalls != scansAfterConfirm {
		t.Fatalf("backoff window must not re-scan, scanCalls went %d -> %d", scansAfterConfirm, scanCalls)
	}
}

func TestGetVmFastResolvesVmFoundByScan(t *testing.T) {
	service := createTestParallelsService()
	service.findProbeWait = func(time.Duration) {
		t.Error("GetVmFast must not run retry sleeps")
	}
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		return createMockVMs(), nil
	}

	vm, err := service.GetVmFast(createTestContext(), "TestVM2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vm == nil || vm.ID != "vm-test-456" {
		t.Fatalf("unexpected vm: %v", vm)
	}
	if len(service.cachedLocalVms) != len(createMockVMs()) {
		t.Fatalf("expected the scan to refresh the cache, got %d entries", len(service.cachedLocalVms))
	}
}

func TestFindVmInCacheAndSystemProbeHitClearsStaleBackoff(t *testing.T) {
	service := createTestParallelsService()
	// A stale (expired) entry: it must not short-circuit, and the eventual
	// cache hit must evict it.
	service.markVmNotFound("late-vm")
	service.syncMu.Lock()
	service.notFoundVm["late-vm"] = time.Now().Add(-2 * CacheFindNotFoundTTL)
	service.syncMu.Unlock()

	service.findProbeWait = func(time.Duration) {
		svc := models.ParallelsVM{ID: "late-vm", Name: "LateVM", State: "running"}
		service.Lock()
		service.cachedLocalVms = append(service.cachedLocalVms, svc)
		service.Unlock()
	}
	service.findSystemVms = func(basecontext.ApiContext, string) ([]models.ParallelsVM, error) {
		t.Error("system scan must not run once a retry probe finds the vm")
		return nil, nil
	}

	vm, err := service.findVmInCacheAndSystem(createTestContext(), "late-vm")
	if err != nil || vm == nil || vm.Name != "LateVM" {
		t.Fatalf("expected LateVM from refreshed cache, got %v / %v", vm, err)
	}
	service.syncMu.Lock()
	_, stillThere := service.notFoundVm["late-vm"]
	service.syncMu.Unlock()
	if stillThere {
		t.Fatal("cache hit must clear the not-found backoff entry")
	}
}
