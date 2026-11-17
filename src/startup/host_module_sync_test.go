package startup

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	str "github.com/stretchr/testify/assert"
)

type fakePD struct {
	results     []bool // consumed by successive probe calls; last value sticks
	idx         int
	hasPD       bool
	refreshSeen int
}

func (f *fakePD) next() bool {
	v := f.results[f.idx]
	if f.idx < len(f.results)-1 {
		f.idx++
	}
	return v
}

func (f *fakePD) IsParallelsDesktopAvailable() bool  { return f.next() }
func (f *fakePD) RefreshParallelsDesktopState() bool { f.refreshSeen++; return f.next() }
func (f *fakePD) HasParallelsDesktop() bool          { return f.hasPD }

type fakeModules struct {
	enabled        map[string]bool
	persistInvoked int
}

func (f *fakeModules) IsModuleEnabled(module string) bool { return f.enabled[module] }
func (f *fakeModules) EnableModule(module string)         { f.enabled[module] = true }
func (f *fakeModules) DisableModule(module string)        { delete(f.enabled, module) }

func TestSyncHostModule_LateLicenseRecoveredAtBoot(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	pd := &fakePD{results: []bool{false, false, true}, hasPD: true}
	cfg := &fakeModules{enabled: map[string]bool{}}
	persisted := 0

	syncHostModuleWithParallelsDesktop(ctx, cfg, pd, 6, time.Millisecond, func() { persisted++ })

	str.True(t, cfg.enabled[constants.HOST_MODE], "host must be enabled once PD reports available")
	str.Equal(t, 1, persisted)
}

func TestSyncHostModule_NoRecheckWithoutPdBinaries(t *testing.T) {
	// PD binaries absent: the loop must not run at all (fast boot for PD-free
	// Macs), and an enabled host module is disabled as before.
	ctx := basecontext.NewBaseContext()
	pd := &fakePD{results: []bool{false}, hasPD: false}
	cfg := &fakeModules{enabled: map[string]bool{constants.HOST_MODE: true}}
	persisted := 0

	start := time.Now()
	syncHostModuleWithParallelsDesktop(ctx, cfg, pd, 6, 5*time.Second, func() { persisted++ })

	str.WithinDuration(t, start, time.Now(), time.Second)
	str.NotContains(t, cfg.enabled, constants.HOST_MODE)
	str.Zero(t, pd.refreshSeen)
	str.Zero(t, persisted)
}

func TestSyncHostModule_ExhaustedWindowKeepsLegacyDisable(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	pd := &fakePD{results: []bool{false}, hasPD: true}
	cfg := &fakeModules{enabled: map[string]bool{constants.HOST_MODE: true}}

	syncHostModuleWithParallelsDesktop(ctx, cfg, pd, 3, time.Millisecond, nil)

	str.NotContains(t, cfg.enabled, constants.HOST_MODE, "still-unavailable PD after the window disables host as before")
	str.Equal(t, 3, pd.refreshSeen)
}

func TestWatchHostModuleRecovery_EnablesAndRestartsOnceOnFlip(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	pd := &fakePD{results: []bool{false, true}, hasPD: true}
	cfg := &fakeModules{enabled: map[string]bool{}}
	persisted, restarts := 0, 0
	done := make(chan struct{})

	go func() {
		watchHostModuleRecovery(ctx, cfg, pd, 5*time.Millisecond, func() { persisted++ }, func() { restarts++ })
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not act on availability flip")
	}

	str.True(t, cfg.enabled[constants.HOST_MODE])
	str.Equal(t, 1, persisted)
	str.Equal(t, 1, restarts, "must exit-once, not keep restarting")
}

func TestWatchHostModuleRecovery_IdleWhileHostEnabled(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	pd := &fakePD{results: []bool{true}, hasPD: true}
	cfg := &fakeModules{enabled: map[string]bool{constants.HOST_MODE: true}}
	restarts := 0

	go watchHostModuleRecovery(ctx, cfg, pd, 5*time.Millisecond, nil, func() { restarts++ })
	time.Sleep(40 * time.Millisecond)

	str.Zero(t, pd.refreshSeen, "already-enabled host must not be probed")
	str.Zero(t, restarts)
}
