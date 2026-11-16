package settings

// Orchestrator-side sync engine: pushes full managed bundles to hosts over
// the host's own API (POST /v1/settings/apply). Pushes are idempotent — the
// hash comparison on both ends makes repeated delivery free.
//
// Import direction is strict: settings must never import orchestrator (the
// orchestrator imports settings and wires triggers). Everything the engine
// needs comes from serviceprovider (database + http client), which both
// sides share.

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	dmodels "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	apiclient "github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
)

// hostApplyPath matches the versioned host-side endpoint registered by the
// host settings controller (same "/v1/..."-relative convention the proven
// host callers such as get_host_virtual_machines use).
const hostApplyPath = "/v1/settings/apply"

type syncEngine struct {
	ctx basecontext.ApiContext
	svc *Service

	mu       sync.Mutex
	inflight map[string]bool
	stop     chan struct{}
}

var (
	engineMu sync.Mutex
	engine   *syncEngine
)

// getEngineLocked returns the live engine, lazily creating a bare one
// (without the retry loop) if a trigger fires before StartSyncEngine ran.
// Nil only when the settings domain is inert on this node.
func getEngineLocked(ctx basecontext.ApiContext) *syncEngine {
	engineMu.Lock()
	defer engineMu.Unlock()
	if engine == nil {
		if Get() == nil || !config.Get().IsOrchestrator() {
			return nil
		}
		engine = &syncEngine{ctx: ctx, svc: Get(), inflight: map[string]bool{}, stop: make(chan struct{})}
	}
	return engine
}

// StartSyncEngine wires the engine on orchestrator nodes and starts a slow
// periodic retry pass for hosts stuck pending/error (offline hosts converge
// via the healthy-transition trigger; the retry covers misses without pings).
func StartSyncEngine(ctx basecontext.ApiContext) {
	svc := Get()
	if svc == nil || !config.Get().IsOrchestrator() {
		return
	}
	engineMu.Lock()
	defer engineMu.Unlock()
	if engine != nil {
		return
	}
	engine = &syncEngine{ctx: ctx, svc: svc, inflight: map[string]bool{}, stop: make(chan struct{})}
	interval := time.Duration(config.Get().OrchestratorPullFrequency()) * time.Second
	if interval < 10*time.Second {
		interval = 10 * time.Second
	}
	go engine.retryLoop(interval)
	ctx.LogInfof("settings sync engine started (retry interval %v)", interval)
}

// retryLoop converges hosts whose state is pending/error and no live trigger
// has fired for them (host offline when the value changed).
func (e *syncEngine) retryLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-e.stop:
			return
		case <-ticker.C:
			for _, state := range e.svc.SyncStates() {
				if state.Status == dmodels.SettingSyncStatusPending ||
					state.Status == dmodels.SettingSyncStatusError ||
					state.Status == dmodels.SettingSyncStatusNever {
					if !e.svc.HostIsCurrent(state.HostID) {
						go func(id string) { _ = e.syncHost(id) }(state.HostID)
					}
				}
			}
		}
	}
}

// StopSyncEngine halts the periodic retry pass (tests / shutdown).
func StopSyncEngine() {
	engineMu.Lock()
	defer engineMu.Unlock()
	if engine != nil {
		close(engine.stop)
		engine = nil
	}
}

// OnHostOnline is the healthy-transition trigger (orchestrator calls it from
// its health handler). It never blocks the caller: the push runs in the
// background, deduplicated per host, and is a no-op when the host already
// holds the desired bundle.
func OnHostOnline(ctx basecontext.ApiContext, hostID string) {
	e := getEngineLocked(ctx)
	if e == nil {
		return
	}
	if e.svc.HostIsCurrent(hostID) {
		return
	}
	go func(id string) { _ = e.syncHost(id) }(hostID)
}

// SyncHostNow pushes one host synchronously (manual API trigger).
func SyncHostNow(ctx basecontext.ApiContext, hostID string) apimodels.SettingsSyncResult {
	e := getEngineLocked(ctx)
	if e == nil {
		return apimodels.SettingsSyncResult{HostID: hostID, Status: "error", Error: "settings service unavailable"}
	}
	return e.syncHost(hostID)
}

// SyncAllNow fans a synchronous full push out to every enabled host (manual
// API trigger). Returns one result row per host.
func SyncAllNow(ctx basecontext.ApiContext, target string) []apimodels.SettingsSyncResult {
	e := getEngineLocked(ctx)
	if e == nil {
		return []apimodels.SettingsSyncResult{{Status: "error", Error: "settings service unavailable"}}
	}
	db, err := serviceprovider.GetDatabaseService(e.ctx)
	if err != nil || db == nil {
		return []apimodels.SettingsSyncResult{{Status: "error", Error: "database unavailable"}}
	}
	hosts, err := db.GetOrchestratorHosts(e.ctx, "")
	if err != nil {
		return []apimodels.SettingsSyncResult{{Status: "error", Error: err.Error()}}
	}
	results := make([]apimodels.SettingsSyncResult, 0, len(hosts))
	for i := range hosts {
		host := hosts[i]
		if target != "" && target != "all" && target != host.ID {
			continue
		}
		results = append(results, e.syncHost(host.ID))
	}
	return results
}

// pendingGate serializes pushes per host; returns false when one is running.
func (e *syncEngine) begin(hostID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.inflight[hostID] {
		return false
	}
	e.inflight[hostID] = true
	return true
}

func (e *syncEngine) end(hostID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.inflight, hostID)
}

// syncHost performs the actual push: skip when current, deliver the full
// bundle otherwise, record the outcome.
func (e *syncEngine) syncHost(hostID string) apimodels.SettingsSyncResult {
	result := apimodels.SettingsSyncResult{HostID: hostID}
	db, err := serviceprovider.GetDatabaseService(e.ctx)
	if err != nil || db == nil {
		result.Status = "error"
		result.Error = "database unavailable"
		return result
	}
	host, err := db.GetOrchestratorHost(e.ctx, hostID)
	if err != nil || host == nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("host %s not found", hostID)
		return result
	}
	result.HostName = host.Host
	if !host.Enabled {
		result.Status = dmodels.SettingSyncStatusPending
		result.Skipped = true
		return result
	}

	bundle, hash, revision := e.svc.DesiredBundleForHost(hostID)
	state := e.svc.SyncStateFor(hostID)
	if state.Status == dmodels.SettingSyncStatusSynced && state.ContentHash == hash {
		result.Status = dmodels.SettingSyncStatusSynced
		result.Skipped = true
		result.Unchanged = true
		result.Revision = revision
		return result
	}

	if !e.begin(hostID) {
		result.Status = dmodels.SettingSyncStatusPending
		result.Skipped = true
		return result
	}
	defer e.end(hostID)

	payload := apimodels.SettingsSyncRequest{Revision: revision, Hash: hash, Values: bundle}
	response, err := e.push(host, payload)
	if err != nil {
		e.svc.RecordSyncFailure(hostID, err.Error())
		result.Status = dmodels.SettingSyncStatusError
		result.Error = err.Error()
		return result
	}
	e.svc.RecordSyncSuccess(hostID, revision, hash)
	result.Status = dmodels.SettingSyncStatusSynced
	result.Unchanged = response.Unchanged
	result.Revision = revision
	return result
}

// push delivers the bundle to one host, authenticating with the per-host API
// key minted at enrollment (same posture as every orchestrator host call).
func (e *syncEngine) push(host *dmodels.OrchestratorHost, payload apimodels.SettingsSyncRequest) (apimodels.SettingsSyncResponse, error) {
	var response apimodels.SettingsSyncResponse
	base := host.GetHost()
	if base == "" {
		return response, fmt.Errorf("host %s has no resolvable endpoint", host.ID)
	}
	url, err := helpers.JoinUrl([]string{base, hostApplyPath})
	if err != nil {
		return response, err
	}

	client := apiclient.NewHttpClient(e.ctx)
	client.WithHeader("X-SOURCE", "ORCHESTRATOR_REQUEST")
	client.WithHeader("X-SOURCE-ID", "ORCHESTRATOR_REQUEST")
	if host.Authentication != nil {
		if host.Authentication.ApiKey != "" {
			client.AuthorizeWithApiKey(host.Authentication.ApiKey)
		} else if host.Authentication.Username != "" {
			client.AuthorizeWithUsernameAndPassword(host.Authentication.Username, host.Authentication.Password)
		}
	}
	client.WithTimeout(30 * time.Second)

	rsp, err := client.Post(url.String(), payload, &response)
	if err != nil {
		return response, err
	}
	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusAccepted {
		return response, fmt.Errorf("host %s rejected settings bundle with status %d", host.ID, rsp.StatusCode)
	}
	return response, nil
}
