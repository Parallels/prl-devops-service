package startup

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
)

const (
	// pdBootRecheckAttempts / pdBootRecheckDelay bound the boot-time window
	// during which a macOS node re-probes Parallels Desktop when the first
	// probe failed but the PD binaries are on disk (VM service starting
	// later than this service via launchd, license activation racing the
	// service start, transient prlsrvctl failures).
	pdBootRecheckAttempts = 6
	pdBootRecheckDelay    = 5 * time.Second
	// pdAvailabilityWatchInterval is how often a running macOS node re-checks
	// whether Parallels Desktop became available after boot.
	pdAvailabilityWatchInterval = 60 * time.Second
)

// parallelsDesktopState is the slice of the service provider the host-module
// sync needs; kept small so the sync logic is testable without Parallels.
type parallelsDesktopState interface {
	IsParallelsDesktopAvailable() bool
	RefreshParallelsDesktopState() bool
	HasParallelsDesktop() bool
}

// moduleConfig is the slice of the config service the sync mutates.
type moduleConfig interface {
	IsModuleEnabled(module string) bool
	EnableModule(module string)
	DisableModule(module string)
}

// syncHostModuleWithParallelsDesktop keeps HOST_MODE in sync with actual
// Parallels Desktop availability on macOS at service start.
//
// It checks both directions so that a freshly deployed node that just had PD
// installed (or where PD was not ready at the previous boot) gets the module
// enabled automatically. The single availability snapshot taken by
// InitServices is not trusted when it says "unavailable" while the PD
// binaries are present: that snapshot is false not only when PD is genuinely
// missing, but also whenever prlsrvctl answered late or the license was not
// yet active at probe time — all transient at boot. In that case the probe is
// re-run a bounded number of times before the module decision is made, so a
// healthy macOS node does not silently lose host-mode routes (/cache,
// /virtualmachines, host settings) for the whole process lifetime.
func syncHostModuleWithParallelsDesktop(
	ctx basecontext.ApiContext,
	cfg moduleConfig,
	provider parallelsDesktopState,
	attempts int,
	delay time.Duration,
	persist func(),
) {
	pdAvailable := provider.IsParallelsDesktopAvailable()
	for attempt := 0; !pdAvailable && attempt < attempts && provider.HasParallelsDesktop(); attempt++ {
		time.Sleep(delay)
		pdAvailable = provider.RefreshParallelsDesktopState()
	}

	if cfg.IsModuleEnabled(constants.HOST_MODE) && !pdAvailable {
		ctx.LogWarnf("Parallels Desktop is not available, disabling host module")
		cfg.DisableModule(constants.HOST_MODE)
	} else if !cfg.IsModuleEnabled(constants.HOST_MODE) && pdAvailable {
		ctx.LogInfof("Parallels Desktop is available; auto-enabling host module")
		cfg.EnableModule(constants.HOST_MODE)
		if persist != nil {
			persist()
		}
	}
}

// watchHostModuleRecovery is the runtime counterpart of the boot sync: PD can
// become available long after startup (license activated later, PD installed
// while the service runs). Route registration only happens at process start,
// so when the flip to "available" is observed the module is enabled,
// persisted, and the service exits cleanly — the launchd/systemd unit is
// KeepAlive/Restart-based and brings the service back with host-mode routes
// registered.
//
// The watcher only ever acts on the enable direction: isLicensed is sticky
// (RefreshInfo never clears it), so a transient probe failure can never tear
// the host module back down or trigger a restart loop. A genuine loss of PD
// is handled by the boot sync on the next service restart.
func watchHostModuleRecovery(
	ctx basecontext.ApiContext,
	cfg moduleConfig,
	provider parallelsDesktopState,
	interval time.Duration,
	persist func(),
	restart func(),
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if cfg.IsModuleEnabled(constants.HOST_MODE) {
			continue
		}
		if !provider.RefreshParallelsDesktopState() {
			continue
		}
		ctx.LogInfof("Parallels Desktop became available after boot; enabling host module and restarting to register host routes")
		cfg.EnableModule(constants.HOST_MODE)
		if persist != nil {
			persist()
		}
		if restart != nil {
			restart()
		}
		return
	}
}
