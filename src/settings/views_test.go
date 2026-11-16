package settings

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

func findEffective(entries []apimodels.SettingEffectiveEntry, key string) *apimodels.SettingEffectiveEntry {
	for i := range entries {
		if entries[i].Key == key {
			return &entries[i]
		}
	}
	return nil
}

func TestHostNodeViewShadowingAndSources(t *testing.T) {
	svc := newTestService(t, false, true)
	if _, err := svc.SetLocalValue("MAX_VMS_PER_HOST", "3"); err != nil {
		t.Fatal(err)
	}
	bundle := map[string]string{"MAX_VMS_PER_HOST": "7"}
	if _, err := svc.ApplyBundle(9, BundleHash(bundle), bundle); err != nil {
		t.Fatal(err)
	}

	entries := svc.HostNodeView()
	entry := findEffective(entries, "MAX_VMS_PER_HOST")
	if entry == nil {
		t.Fatal("entry missing")
	}
	// Managed wins; local marked as shadowed — visible but inert.
	if entry.Effective != "7" || entry.Source != SourceManaged {
		t.Fatalf("expected managed 7, got %s from %s", entry.Effective, entry.Source)
	}
	if entry.ShadowedBy != SourceManaged {
		t.Fatalf("local edit hidden behind managed must be marked, got %q", entry.ShadowedBy)
	}
	if entry.LocalValue == nil || *entry.LocalValue != "3" {
		t.Fatal("local layer must still be reported")
	}
}

func TestHostNodeViewNoShadowWhenManagedAbsent(t *testing.T) {
	svc := newTestService(t, false, true)
	if _, err := svc.SetLocalValue("MAX_VMS_PER_HOST", "3"); err != nil {
		t.Fatal(err)
	}
	entry := findEffective(svc.HostNodeView(), "MAX_VMS_PER_HOST")
	if entry == nil {
		t.Fatal("entry missing")
	}
	if entry.Effective != "3" || entry.Source != SourceLocal {
		t.Fatalf("expected local 3, got %s/%s", entry.Effective, entry.Source)
	}
	if entry.ShadowedBy != "" {
		t.Fatalf("unshadowed local edit must not carry a marker, got %q", entry.ShadowedBy)
	}
}

func TestEffectiveSourceProbing(t *testing.T) {
	svc := newTestService(t, false, true)

	// 1. Nothing stored, nothing in env: catalog default.
	entry := findEffective(svc.HostNodeView(), "MAX_VMS_PER_HOST")
	if entry.Effective != entry.DefaultValue || entry.Source != SourceDefault {
		t.Fatalf("expected catalog default, got %q/%s", entry.Effective, entry.Source)
	}

	// 2. Environment layer surfaces with the right source.
	t.Setenv("MAX_VMS_PER_HOST", "2")
	entry = findEffective(svc.HostNodeView(), "MAX_VMS_PER_HOST")
	if entry.Effective != "2" || entry.Source != SourceEnv {
		t.Fatalf("expected env layer, got %q/%s", entry.Effective, entry.Source)
	}

	// 3. DB layer beats env and reports db source.
	if _, err := svc.SetLocalValue("MAX_VMS_PER_HOST", "4"); err != nil {
		t.Fatal(err)
	}
	entry = findEffective(svc.HostNodeView(), "MAX_VMS_PER_HOST")
	if entry.Effective != "4" || entry.Source != SourceLocal {
		t.Fatalf("expected db layer over env, got %q/%s", entry.Effective, entry.Source)
	}
}

func TestOrchestratorNodeViewSources(t *testing.T) {
	svc := newTestService(t, true, false)
	if _, err := svc.SetScopeValue(models.SettingScopeOrchestrator, "LOG_LEVEL", "warn", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetLocalValue("LOG_LEVEL", "error"); err != nil {
		t.Fatal(err)
	}

	entry := findEffective(svc.OrchestratorNodeView(), "LOG_LEVEL")
	if entry.Effective != "warn" || entry.Source != SourceOrchestrator {
		t.Fatalf("expected orchestrator_db warn, got %s/%s", entry.Effective, entry.Source)
	}
	if entry.ShadowedBy != SourceOrchestrator {
		t.Fatalf("local edit shadowed by orchestrator scope must be marked, got %q", entry.ShadowedBy)
	}

	// Host-scope rows never appear in the orchestrator's own effective view.
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "ENABLE_CANARY", "true", "tester"); err != nil {
		t.Fatal(err)
	}
	entry = findEffective(svc.OrchestratorNodeView(), "ENABLE_CANARY")
	if entry.Effective == "true" && entry.Source != SourceDefault {
		t.Fatal("hosts-scope value must not answer for the orchestrator process")
	}
}

func TestHostEffectiveForRemoteView(t *testing.T) {
	svc := newTestService(t, true, false)
	if _, err := svc.SetScopeValue(models.SettingScopeHosts, "MAX_VMS_PER_HOST", "5", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetScopeValue(models.HostSettingScope("h1"), "MAX_VMS_PER_HOST", "1", "tester"); err != nil {
		t.Fatal(err)
	}
	// The orchestrator's own env must never leak into a REMOTE host view.
	t.Setenv("ENABLE_CANARY", "true")

	entries := svc.HostEffectiveFor("h1")
	entry := findEffective(entries, "MAX_VMS_PER_HOST")
	if entry.Effective != "1" || entry.Source != SourceHost {
		t.Fatalf("expected per-host 1, got %s/%s", entry.Effective, entry.Source)
	}
	if entry.ShadowedBy != SourceHost {
		t.Fatalf("per-host over hosts-global should be marked, got %q", entry.ShadowedBy)
	}
	canary := findEffective(entries, "ENABLE_CANARY")
	if canary.Source != SourceDefault {
		t.Fatalf("remote view must not probe this process's env, got %s", canary.Source)
	}
	// Orchestrator-only keys are absent from the host view entirely.
	if findEffective(entries, "ORCHESTRATOR_PULL_FREQUENCY_SECONDS") != nil {
		t.Fatal("orchestrator-only key must not appear in host view")
	}
}

func TestNodeInfoManagedMarker(t *testing.T) {
	svc := newTestService(t, false, true)
	if svc.NodeInfo().Managed {
		t.Fatal("fresh node is unmanaged")
	}
	bundle := map[string]string{"MAX_VMS_PER_HOST": "7"}
	if _, err := svc.ApplyBundle(9, BundleHash(bundle), bundle); err != nil {
		t.Fatal(err)
	}
	info := svc.NodeInfo()
	if !info.Managed || info.ManagedHash != BundleHash(bundle) {
		t.Fatalf("apply must mark managed + hash: %+v", info)
	}
}
