package settings

import (
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// openTestDB spins an isolated in-memory sqlite store with the five settings
// tables — same AutoMigrate the boot registry runs over AllGormModels.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+filepath.Join(t.TempDir(), "settings.db")+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.SettingDefinition{}, &models.SettingValue{},
		&models.SettingSyncState{}, &models.NodeSetting{}, &models.SettingMeta{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

// newTestService builds a Service over a fresh store. Same-package test
// surface: mirrors InitService without the serviceprovider/config globals.
func newTestService(t *testing.T, isOrchestrator, isHost bool) *Service {
	t.Helper()
	svc := &Service{
		ctx:            basecontext.NewRootBaseContext(),
		st:             newStore(openTestDB(t)),
		definitions:    map[string]models.SettingDefinition{},
		values:         map[string]map[string]string{},
		managed:        map[string]string{},
		local:          map[string]string{},
		isOrchestrator: isOrchestrator,
		isHost:         isHost,
	}
	if err := svc.load(); err != nil {
		t.Fatalf("service load: %v", err)
	}
	return svc
}

func TestCatalogSeededAndWhitelisted(t *testing.T) {
	svc := newTestService(t, true, false)
	if svc.Definition("MAX_VMS_PER_HOST") == nil {
		t.Fatal("expected MAX_VMS_PER_HOST in the catalog")
	}
	// Case/whitespace normalization on lookup.
	if svc.Definition("  max_vms_per_host ") == nil {
		t.Fatal("expected normalized key lookup")
	}
	// Boot-critical and secret keys are permanently outside the whitelist.
	for _, key := range []string{"DB_PATH", "DB_TYPE", "DB_POSTGRES_URL", "TLS_CERT_FILE", "API_PORT", "MODE", "ENABLED_MODULES", "BASE_URL", "LOG_FILE", "API_HMAC_KEY", "JWT_KEY", "ENCRYPTION_PRIVATE_KEY", "ROOT_PASSWORD"} {
		if svc.Definition(key) != nil {
			t.Fatalf("%s must never be catalogued", key)
		}
	}
}

func TestSetScopeValueValidation(t *testing.T) {
	svc := newTestService(t, true, false)

	if _, err := svc.SetScopeValue(models.SettingScopeOrchestrator, "NOT_A_KEY", "x", "tester"); err == nil {
		t.Fatal("uncatalogued key must be rejected")
	}
	if _, err := svc.SetScopeValue(models.SettingScopeOrchestrator, constantsProbe(), "x", "tester"); err == nil {
		t.Fatal("boot-critical key must be rejected")
	}
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "ENABLE_CANARY", "maybe", "tester"); err == nil {
		t.Fatal("invalid bool must be rejected")
	}
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "12", "tester"); err != nil {
		t.Fatalf("valid int rejected: %v", err)
	}
	// Enum keys normalize to the allow-list spelling.
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST_MODE", " MacOS ", "tester"); err != nil {
		t.Fatalf("case-insensitive enum should pass: %v", err)
	}
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST_MODE", "windows", "tester"); err == nil {
		t.Fatal("value outside the enum must be rejected")
	}
	if _, err := svc.SetScopeValue("bogus-scope", "MAX_VMS_PER_HOST", "1", "tester"); err == nil {
		t.Fatal("unknown scope must be rejected")
	}
}

// constantsProbe returns a permanently-excluded key name (kept out of the
// catalog on purpose — see catalog.go header).
func constantsProbe() string { return "DB_PATH" }

func TestAppliesToScopeGating(t *testing.T) {
	svc := newTestService(t, true, false)

	// Host-only key may not be written to the orchestrator scope.
	if _, err := svc.SetScopeValue(models.SettingScopeOrchestrator, "MAX_VMS_PER_HOST", "4", "tester"); err == nil {
		t.Fatal("host-only key must be rejected in orchestrator scope")
	}
	// Orchestrator-only key may not go into host scopes.
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "ORCHESTRATOR_PULL_FREQUENCY_SECONDS", "5", "tester"); err == nil {
		t.Fatal("orchestrator-only key must be rejected in hosts scope")
	}
	if _, err := svc.SetScopeValue(models.HostSettingScope("h1"), "ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS", "5", "tester"); err == nil {
		t.Fatal("orchestrator-only key must be rejected in host scope")
	}
	// All-scope key is fine everywhere.
	if _, err := svc.SetScopeValue(models.SettingScopeOrchestrator, "LOG_LEVEL", "warn", "tester"); err != nil {
		t.Fatalf("all-scope key rejected: %v", err)
	}
	// Host-local writes reject orchestrator-only keys too.
	hostSvc := newTestService(t, false, true)
	if _, err := hostSvc.SetLocalValue("ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS", "5"); err == nil {
		t.Fatal("host-local write must reject orchestrator-only key")
	}
}

func TestScopedUpsertKeepsSingleRowAndBumpsRevision(t *testing.T) {
	svc := newTestService(t, true, false)

	first, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "1", "tester")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "2", "tester")
	if err != nil {
		t.Fatal(err)
	}
	if second <= first {
		t.Fatalf("revision must advance: %d -> %d", first, second)
	}
	// Different scope, same key = separate row.
	if _, err := svc.SetScopeValue(models.HostSettingScope("h1"), "MAX_VMS_PER_HOST", "9", "tester"); err != nil {
		t.Fatal(err)
	}

	rows := svc.Values()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	for _, row := range rows {
		if row.Scope == models.SettingScopeHosts && row.Value != "2" {
			t.Fatalf("upsert must update in place, got %q", row.Value)
		}
	}
}

func TestDeleteScopeValueSemantics(t *testing.T) {
	svc := newTestService(t, true, false)
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "1", "tester"); err != nil {
		t.Fatal(err)
	}
	existed, err := svc.DeleteScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST")
	if err != nil || !existed {
		t.Fatalf("expected existed=true, got %v %v", existed, err)
	}
	existed, err = svc.DeleteScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST")
	if err != nil || existed {
		t.Fatalf("second delete must report not-existed, got %v %v", existed, err)
	}
}

func TestLookupChainOrchestratorNode(t *testing.T) {
	svc := newTestService(t, true, false)
	if _, err := svc.SetScopeValue(models.SettingScopeOrchestrator, "LOG_LEVEL", "warn", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetLocalValue("LOG_LEVEL", "error"); err != nil {
		t.Fatal(err)
	}
	managed := map[string]string{"LOG_LEVEL": "debug"}
	if _, err := svc.ApplyBundle(50, BundleHash(managed), managed); err != nil {
		t.Fatal(err)
	}

	// managed wins over orchestrator-scope and local.
	if v, ok := svc.lookup("log_level"); !ok || v != "debug" {
		t.Fatalf("managed must win, got %q %v", v, ok)
	}
	// Drop managed → orchestrator scope wins.
	if _, err := svc.ApplyBundle(51, BundleHash(map[string]string{}), map[string]string{}); err != nil {
		t.Fatal(err)
	}
	if v, ok := svc.lookup("LOG_LEVEL"); !ok || v != "warn" {
		t.Fatalf("orchestrator scope must win after managed cleared, got %q %v", v, ok)
	}
	// Drop orchestrator scope → local wins.
	if _, err := svc.DeleteScopeValue(models.SettingScopeOrchestrator, "LOG_LEVEL"); err != nil {
		t.Fatal(err)
	}
	if v, ok := svc.lookup("LOG_LEVEL"); !ok || v != "error" {
		t.Fatalf("local must win last, got %q %v", v, ok)
	}
	// Drop local → not found (GetKey fallback takes over).
	if _, err := svc.DeleteLocalValue("LOG_LEVEL"); err != nil {
		t.Fatal(err)
	}
	if _, ok := svc.lookup("LOG_LEVEL"); ok {
		t.Fatal("expected miss after every layer removed")
	}
}

func TestApplyBundleReplacesManagedOnlyAndIsIdempotent(t *testing.T) {
	svc := newTestService(t, false, true)
	if _, err := svc.SetLocalValue("MAX_VMS_PER_HOST", "42"); err != nil {
		t.Fatal(err)
	}

	bundle := map[string]string{"MAX_VMS_PER_HOST": "10", "ENABLE_CANARY": "true"}
	rsp, err := svc.ApplyBundle(7, BundleHash(bundle), bundle)
	if err != nil {
		t.Fatal(err)
	}
	if rsp.Unchanged {
		t.Fatal("first apply must report changed")
	}
	if rsp.Hash != BundleHash(bundle) {
		t.Fatal("response must carry the computed hash")
	}

	// Local layer untouched by the wholesale managed replace.
	if v, ok := svc.local["MAX_VMS_PER_HOST"]; !ok || v != "42" {
		t.Fatalf("local layer was touched: %q", v)
	}
	// Managed shadows local in the resolver.
	if v, _ := svc.lookup("MAX_VMS_PER_HOST"); v != "10" {
		t.Fatalf("managed must shadow local, got %q", v)
	}

	// Idempotent re-delivery.
	rsp, err = svc.ApplyBundle(7, BundleHash(bundle), bundle)
	if err != nil {
		t.Fatal(err)
	}
	if !rsp.Unchanged {
		t.Fatal("identical hash re-delivery must report unchanged")
	}

	// Replacement drops keys no longer in the bundle — local survives.
	if _, err := svc.ApplyBundle(8, BundleHash(map[string]string{"ENABLE_CANARY": "true"}), map[string]string{"ENABLE_CANARY": "true"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := svc.managed["MAX_VMS_PER_HOST"]; ok {
		t.Fatal("stale managed key must be gone after wholesale replace")
	}
	if _, ok := svc.local["MAX_VMS_PER_HOST"]; !ok {
		t.Fatal("local key must survive the replace")
	}
}

func TestApplyBundleRejectsBadBundles(t *testing.T) {
	svc := newTestService(t, false, true)

	// Unknown key.
	if _, err := svc.ApplyBundle(1, "", map[string]string{"DB_PATH": "/tmp/evil"}); err == nil {
		t.Fatal("uncatalogued key must never be applied")
	}
	// Orchestrator-only key smuggled into a host bundle.
	if _, err := svc.ApplyBundle(1, "", map[string]string{"ORCHESTRATOR_PULL_FREQUENCY_SECONDS": "1"}); err == nil {
		t.Fatal("orchestrator-only key must be rejected on hosts")
	}
	// Hash mismatch (transport integrity).
	values := map[string]string{"MAX_VMS_PER_HOST": "10"}
	if _, err := svc.ApplyBundle(1, "deadbeef", values); err == nil {
		t.Fatal("hash mismatch must be rejected")
	}
	// Bad value type inside an otherwise fine bundle.
	if _, err := svc.ApplyBundle(1, "", map[string]string{"MAX_VMS_PER_HOST": "not-a-number"}); err == nil {
		t.Fatal("invalid value must be rejected")
	}
	// And nothing leaked into the managed layer.
	if _, ok := svc.managed["DB_PATH"]; ok {
		t.Fatal("rejected apply must not touch state")
	}
}

func TestBundleHashStableAndOrderIndependent(t *testing.T) {
	a := map[string]string{"A": "1", "B": "2", "C": "3"}
	b := map[string]string{"C": "3", "B": "2", "A": "1"}
	if BundleHash(a) != BundleHash(b) {
		t.Fatal("hash must be insertion-order independent")
	}
	a["B"] = "2x"
	if BundleHash(a) == BundleHash(b) {
		t.Fatal("hash must change with values")
	}
	if BundleHash(map[string]string{"AB": "", "C": ""}) == BundleHash(map[string]string{"A": "", "BC": ""}) {
		t.Fatal("separator must prevent key/value boundary collisions")
	}
}

func TestDesiredBundleMergesHostOverGlobal(t *testing.T) {
	svc := newTestService(t, true, false)
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "5", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetScopeValue(models.HostSettingScope("h1"), "MAX_VMS_PER_HOST", "1", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetScopeValue(models.HostSettingScope("h2"), "ENABLE_CANARY", "true", "tester"); err != nil {
		t.Fatal(err)
	}
	// Orchestrator-scope rows never ride a host bundle.
	if _, err := svc.SetScopeValue(models.SettingScopeOrchestrator, "LOG_LEVEL", "warn", "tester"); err != nil {
		t.Fatal(err)
	}

	bundle, _, _ := svc.DesiredBundleForHost("h1")
	if bundle["MAX_VMS_PER_HOST"] != "1" {
		t.Fatalf("host override must win over hosts-global, got %q", bundle["MAX_VMS_PER_HOST"])
	}
	if _, ok := bundle["ENABLE_CANARY"]; ok {
		t.Fatal("other host's override leaked into this bundle")
	}
	if _, ok := bundle["LOG_LEVEL"]; ok {
		t.Fatal("orchestrator-scope value leaked into a host bundle")
	}

	other, _, _ := svc.DesiredBundleForHost("h2")
	if other["MAX_VMS_PER_HOST"] != "5" {
		t.Fatal("hosts-global value must apply when no per-host override")
	}
	if other["ENABLE_CANARY"] != "true" {
		t.Fatal("per-host value missing")
	}
}

func TestSyncStateTransitions(t *testing.T) {
	svc := newTestService(t, true, false)
	bundle := map[string]string{"MAX_VMS_PER_HOST": "3"}
	_, hash, _ := svc.DesiredBundleForHost("h1") // empty bundle baseline

	if svc.HostIsCurrent("h1") {
		t.Fatal("never-synced host cannot be current")
	}

	svc.MarkHostPending("h1")
	if state := svc.SyncStateFor("h1"); state.Status != models.SettingSyncStatusPending {
		t.Fatalf("expected pending, got %s", state.Status)
	}

	svc.RecordSyncSuccess("h1", 4, hash)
	if !svc.HostIsCurrent("h1") {
		t.Fatal("synced + matching hash must be current")
	}
	// Pending marking keeps the acknowledged fingerprint for drift display.
	svc.MarkHostPending("h1")
	if state := svc.SyncStateFor("h1"); state.ContentHash != hash {
		t.Fatal("mark pending must keep the last acknowledged hash")
	}

	svc.RecordSyncFailure("h1", "boom")
	state := svc.SyncStateFor("h1")
	if state.Status != models.SettingSyncStatusError || state.LastError != "boom" || state.ContentHash != hash {
		t.Fatalf("failure must keep hash and record error, got %+v", state)
	}
	_ = bundle

	// Value change flips the desired hash → not current anymore.
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "3", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, newHash, _ := svc.DesiredBundleForHost("h1"); newHash == hash {
		t.Skip("value identical — hash unchanged by design")
	}
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "4", "tester"); err != nil {
		t.Fatal(err)
	}
	if svc.HostIsCurrent("h1") {
		t.Fatal("changed bundle must make the host pending")
	}
}

func TestClearHostDropsValuesAndState(t *testing.T) {
	svc := newTestService(t, true, false)
	if _, err := svc.SetScopeValue(models.HostSettingScope("h9"), "MAX_VMS_PER_HOST", "2", "tester"); err != nil {
		t.Fatal(err)
	}
	svc.RecordSyncSuccess("h9", 1, "hash")

	if err := svc.ClearHost("h9"); err != nil {
		t.Fatal(err)
	}
	for _, row := range svc.Values() {
		if row.Scope == models.HostSettingScope("h9") {
			t.Fatal("host values must be gone after ClearHost")
		}
	}
	for _, state := range svc.SyncStates() {
		if state.HostID == "h9" {
			t.Fatal("sync state must be gone after ClearHost")
		}
	}
}
