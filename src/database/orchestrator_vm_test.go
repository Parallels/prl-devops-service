package database

// Orchestrator VM rows + RP views (PBI #421 task 14).

import (
	"errors"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestSQLOrchestratorVMs_UpsertAndReads(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	bare := basecontext.NewBaseContext()

	h1 := ohCreate(t, g, root, models.OrchestratorHost{Host: "v1.local", State: "healthy"})
	if err := g.UpdateOrchestratorHostTimestamp(root, h1.ID, "2026-03-01T00:00:00Z", ""); err != nil {
		t.Fatal(err)
	}
	h2 := ohCreate(t, g, root, models.OrchestratorHost{Host: "v2.local"}) // unhealthy
	if err := g.UpdateOrchestratorHostTimestamp(root, h2.ID, "2026-05-01T00:00:00Z", ""); err != nil {
		t.Fatal(err)
	}

	// Twins run authz-free (bare ctx) and Fold-match the host id (:704).
	if err := g.UpsertOrchestratorHostVM(bare, strings.ToUpper(h1.ID), models.VirtualMachine{ID: "vm-a1", Name: "alpha", State: "running"}); err != nil {
		t.Fatal(err)
	}
	if err := g.UpsertOrchestratorHostVM(bare, h1.ID, models.VirtualMachine{ID: "vm-a2", Name: "beta"}); err != nil {
		t.Fatal(err)
	}
	if err := g.UpsertOrchestratorHostVM(bare, h2.ID, models.VirtualMachine{ID: "vm-b1", Name: "gamma"}); err != nil {
		t.Fatal(err)
	}

	vms, err := g.GetOrchestratorHostVirtualMachines(root, h1.ID, "")
	if err != nil || len(vms) != 2 || vms[0].ID != "vm-a1" || vms[1].ID != "vm-a2" {
		t.Fatalf("host-scoped list: %v %+v", err, vms)
	}
	// NO health/enabled gate on the host-scoped read: h2 is unhealthy yet serves.
	if vms, err := g.GetOrchestratorHostVirtualMachines(root, h2.ID, ""); err != nil || len(vms) != 1 {
		t.Fatalf("unhealthy host gate: %v %+v", err, vms)
	}
	// Global list: AUTHZ (bare sees NO hosts -> empty list, no error)
	// + healthy-only contributor gate.
	if bareList, err := g.GetOrchestratorVirtualMachines(bare, ""); err != nil || len(bareList) != 0 {
		t.Fatalf("bare authz list: %v %+v", err, bareList)
	}
	all, err := g.GetOrchestratorVirtualMachines(root, "")
	if err != nil || len(all) != 2 {
		t.Fatalf("global list must skip unhealthy h2: %v %+v", err, all)
	}
	if all[0].ID != "vm-a1" { // UpdatedAt-desc host order (h1 03-01 > h2 excluded anyway)
		t.Fatalf("order: %+v", all)
	}
	fed, err := g.GetOrchestratorVirtualMachines(root, "name=alpha")
	if err != nil || len(fed) != 1 || fed[0].ID != "vm-a1" {
		t.Fatalf("filter: %v %+v", err, fed)
	}

	// Single VM: Fold on ID only (Name does NOT match, unlike Delete).
	vm, err := g.GetOrchestratorHostVirtualMachine(root, h1.ID, "VM-A1")
	if err != nil || vm.Name != "alpha" {
		t.Fatalf("fold single get: %v %+v", err, vm)
	}
	if _, err := g.GetOrchestratorHostVirtualMachine(root, h1.ID, "alpha"); !errors.Is(err, interfaces.ErrOrchestratorHostVirtualMachineNotFound) {
		t.Fatalf("name must NOT match single get: %v", err)
	}
	if _, err := g.GetOrchestratorHostVirtualMachine(bare, h1.ID, "vm-a1"); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("authz leak: %v", err)
	}
}

func TestSQLOrchestratorVMs_TwinsIdempotent(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	bare := basecontext.NewBaseContext()
	h := ohCreate(t, g, root, models.OrchestratorHost{Host: "t.local", State: "healthy"})

	up := func(vm models.VirtualMachine) {
		t.Helper()
		if err := g.UpsertOrchestratorHostVM(bare, h.ID, vm); err != nil {
			t.Fatal(err)
		}
	}
	up(models.VirtualMachine{ID: "vm-x", Name: "x1", State: "running"})
	up(models.VirtualMachine{ID: "vm-y", Name: "y"})
	// In-place replace: same id, EXACT match (uppercase id must NOT match).
	up(models.VirtualMachine{ID: "vm-x", Name: "x2", State: "stopped", Uptime: "5m"})
	vms, _ := g.GetOrchestratorHostVirtualMachines(root, h.ID, "")
	if len(vms) != 2 || vms[0].ID != "vm-x" || vms[0].Name != "x2" || vms[0].Uptime != "5m" || vms[1].ID != "vm-y" {
		t.Fatalf("in-place replace: %+v", vms) // rowid kept: vm-x still first
	}
	if err := g.UpsertOrchestratorHostVM(bare, h.ID, models.VirtualMachine{ID: "VM-X", Name: "ghost"}); err != nil {
		t.Fatal(err)
	}
	vms, _ = g.GetOrchestratorHostVirtualMachines(root, h.ID, "")
	if len(vms) != 3 { // EXACT ==: VM-X is a NEW row, not a replace
		t.Fatalf("exact-match upsert: %+v", vms)
	}
	if err := g.RemoveOrchestratorHostVM(bare, h.ID, "VM-X"); err != nil {
		t.Fatal(err)
	}

	// Field-scoped twins; absent VM is nil (idempotent), absent host errors.
	if err := g.UpdateOrchestratorHostVMState(bare, h.ID, "vm-x", "paused"); err != nil {
		t.Fatal(err)
	}
	if err := g.UpdateOrchestratorHostVMUptime(bare, h.ID, "vm-x", "42h"); err != nil {
		t.Fatal(err)
	}
	vms, _ = g.GetOrchestratorHostVirtualMachines(root, h.ID, "")
	if vms[0].State != "paused" || vms[0].Uptime != "42h" || vms[0].Name != "x2" {
		t.Fatalf("field scope: %+v", vms[0])
	}
	for _, idemp := range []error{
		g.UpdateOrchestratorHostVMState(bare, h.ID, "ghost-vm", "x"),
		g.UpdateOrchestratorHostVMUptime(bare, h.ID, "ghost-vm", "x"),
		g.RemoveOrchestratorHostVM(bare, h.ID, "ghost-vm"),
	} {
		if idemp != nil {
			t.Fatalf("idempotent twin: %v", idemp)
		}
	}
	if err := g.UpdateOrchestratorHostVMState(bare, "no-such-host", "vm-x", "x"); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("twin host-miss: %v", err)
	}
	// Upsert with empty host id has NO empty-id sentinel: falls through to HostNotFound.
	if err := g.UpsertOrchestratorHostVM(bare, "", models.VirtualMachine{ID: "z"}); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("empty-host upsert: %v", err)
	}

	// DeleteOrchestratorVirtualMachine: host Fold ID||RAW Host; VM Fold ID||NAME.
	if err := g.DeleteOrchestratorVirtualMachine(bare, "T.LOCAL", "X2"); err != nil {
		t.Fatalf("delete by name (host by raw Host fold): %v", err)
	}
	if err := g.DeleteOrchestratorVirtualMachine(bare, h.ID, "vm-x"); !errors.Is(err, interfaces.ErrOrchestratorHostVirtualMachineNotFound) {
		t.Fatalf("double delete: %v", err)
	}
	if err := g.DeleteOrchestratorVirtualMachine(bare, "ghost.local", "vm-y"); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("delete host-miss: %v", err)
	}
	if err := g.DeleteOrchestratorVirtualMachine(bare, "", "vm-y"); !errors.Is(err, interfaces.ErrOrchestratorHostEmptyIdOrHost) {
		t.Fatalf("delete empty-id: %v", err)
	}
	if err := g.DeleteOrchestratorVirtualMachine(bare, h.ID, "vm-y"); err != nil {
		t.Fatal(err)
	}
	vms, _ = g.GetOrchestratorHostVirtualMachines(root, h.ID, "")
	if len(vms) != 0 {
		t.Fatalf("post-delete residue: %+v", vms)
	}
}

func TestSQLOrchestratorVMs_Replace(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	bare := basecontext.NewBaseContext()
	h := ohCreate(t, g, root, models.OrchestratorHost{Host: "r.local", State: "healthy"})

	set := func(vms ...models.VirtualMachine) {
		t.Helper()
		if err := g.ReplaceOrchestratorHostVMs(bare, h.ID, vms); err != nil {
			t.Fatal(err)
		}
	}
	// Foreign HostId in payload is overridden by the carrier (#419 D-3).
	set(models.VirtualMachine{ID: "r1", Name: "one", HostId: "elsewhere"},
		models.VirtualMachine{ID: "r2", Name: "two"},
		models.VirtualMachine{ID: "r3", Name: "three"})
	vms, err := g.GetOrchestratorHostVirtualMachines(root, h.ID, "")
	if err != nil || len(vms) != 3 || vms[0].ID != "r1" || vms[2].ID != "r3" {
		t.Fatalf("replace 3: %v %+v", err, vms)
	}
	for _, v := range vms {
		if v.HostId != h.ID {
			t.Fatalf("carrier stamp: %+v", v)
		}
	}
	// Overlapping one: r2 keeps identity but gets the NEW row (slice swap).
	set(models.VirtualMachine{ID: "r2", Name: "two-new"})
	vms, _ = g.GetOrchestratorHostVirtualMachines(root, h.ID, "")
	if len(vms) != 1 || vms[0].Name != "two-new" {
		t.Fatalf("replace 1: %+v", vms)
	}
	set() // empty payload = wipe
	if vms, _ = g.GetOrchestratorHostVirtualMachines(root, h.ID, ""); len(vms) != 0 {
		t.Fatalf("replace 0: %+v", vms)
	}
	if err := g.ReplaceOrchestratorHostVMs(bare, "ghost", nil); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("replace host-miss: %v", err)
	}
}

func TestSQLOrchestratorUpdate_RPWriteRegression(t *testing.T) {
	// Mirrors legacy :253-254: Update REPLACES the RP relation rows from the
	// payload (nil clears), Diff sees them, and reads attach them.
	g := newSQLDB(t)
	root := tplRootCtx()
	h := ohCreate(t, g, root, models.OrchestratorHost{Host: "rp.local"})

	row, err := g.GetOrchestratorHost(root, h.ID)
	if err != nil || row.ReverseProxy != nil || row.ReverseProxyHosts != nil {
		t.Fatalf("fresh host RP attach: %v %+v", err, row)
	}
	// Update carrying config + two hosts.
	in, _ := g.GetOrchestratorHost(root, h.ID)
	in.Description = "d2"
	in.ReverseProxy = &models.ReverseProxy{ID: "cfg-1", Port: "443", Enabled: true}
	in.ReverseProxyHosts = []*models.ReverseProxyHost{{ID: "rp-1", Host: "a.rp"}, {ID: "rp-2", Host: "b.rp"}}
	if _, err := g.UpdateOrchestratorHost(root, in); err != nil {
		t.Fatal(err)
	}
	cfg, err := g.GetOrchestratorReverseProxyConfig(root, h.ID)
	if err != nil || cfg == nil || cfg.ID != "cfg-1" || cfg.Port != "443" || cfg.HostID != h.ID {
		t.Fatalf("config view: %v %+v", err, cfg)
	}
	rph, err := g.GetOrchestratorReverseProxyHosts(root, h.ID, "")
	if err != nil || len(rph) != 2 || rph[0].ID != "rp-1" || rph[1].ID != "rp-2" {
		t.Fatalf("hosts view: %v %+v", err, rph)
	}
	one, err := g.GetOrchestratorReverseProxyHost(root, h.ID, "rp-2")
	if err != nil || one.Host != "b.rp" {
		t.Fatalf("single view: %v %+v", err, one)
	}
	if _, err := g.GetOrchestratorReverseProxyHost(root, h.ID, "rp-ghost"); !errors.Is(err, interfaces.ErrOrchestratorReverseProxyHostNotFound) {
		t.Fatalf("single view miss: %v", err)
	}
	// Single read now carries the relations (legacy shared-object wire).
	row, _ = g.GetOrchestratorHost(root, h.ID)
	if row.ReverseProxy == nil || len(row.ReverseProxyHosts) != 2 {
		t.Fatalf("read attach: %+v", row)
	}
	// Same payload back in: no diff on RP, Description differs -> rewritten
	// verbatim, views unchanged.
	in2, _ := g.GetOrchestratorHost(root, h.ID)
	in2.Description = "d3"
	if _, err := g.UpdateOrchestratorHost(root, in2); err != nil {
		t.Fatal(err)
	}
	if rph, _ := g.GetOrchestratorReverseProxyHosts(root, h.ID, ""); len(rph) != 2 {
		t.Fatalf("rewrite drift: %+v", rph)
	}
	// LIST rows never attach -> nil payload CLEARS the rows (:253 assigns nil).
	l, _ := g.GetOrchestratorHosts(root, "")
	in3 := l[0]
	if in3.ReverseProxy != nil || in3.ReverseProxyHosts != nil {
		t.Fatal("list leaked attach")
	}
	in3.Description = "d4"
	if _, err := g.UpdateOrchestratorHost(root, &in3); err != nil {
		t.Fatal(err)
	}
	if cfg, err := g.GetOrchestratorReverseProxyConfig(root, h.ID); err != nil || cfg != nil {
		t.Fatalf("nil-clear config: %v %+v", err, cfg)
	}
	if rph, err := g.GetOrchestratorReverseProxyHosts(root, h.ID, ""); err != nil || len(rph) != 0 {
		t.Fatalf("nil-clear hosts: %v %+v", err, rph)
	}
	// No-diff echo with attached row is untouched.
	in5, _ := g.GetOrchestratorHost(root, h.ID)
	echo, err := g.UpdateOrchestratorHost(root, in5)
	if err != nil || echo != in5 {
		t.Fatalf("echo pointer: %v %p %p", err, echo, in5)
	}
}

func TestSQLOrchestratorVMs_ColdGuards(t *testing.T) {
	cold := NewGormDatabase(basecontext.NewBaseContext(), nil)
	ctx := tplRootCtx()
	want := interfaces.ErrDatabaseNotConnected
	var got []error
	if _, e := cold.GetHostVMSnapshots(ctx, "h"); true {
		got = append(got, e)
	}
	got = append(got, cold.SetHostVMSnapshots(ctx, "h", models.VMSnapshots{}))
	_, e1 := cold.GetOrchestratorVirtualMachines(ctx, "")
	_, e2 := cold.GetOrchestratorHostVirtualMachines(ctx, "h", "")
	_, e3 := cold.GetOrchestratorHostVirtualMachine(ctx, "h", "v")
	_, e4 := cold.GetOrchestratorReverseProxyHosts(ctx, "h", "")
	_, e5 := cold.GetOrchestratorReverseProxyHost(ctx, "h", "v")
	_, e6 := cold.GetOrchestratorReverseProxyConfig(ctx, "h")
	_, e7 := cold.GetOrchestratorHostVirtualMachines(ctx, "h", "")
	got = append(got, e1, e2, e3, e4, e5, e6, e7,
		cold.DeleteOrchestratorVirtualMachine(ctx, "h", "v"),
		cold.UpsertOrchestratorHostVM(ctx, "h", models.VirtualMachine{}),
		cold.RemoveOrchestratorHostVM(ctx, "h", "v"),
		cold.UpdateOrchestratorHostVMState(ctx, "h", "v", "s"),
		cold.UpdateOrchestratorHostVMUptime(ctx, "h", "v", "u"),
		cold.ReplaceOrchestratorHostVMs(ctx, "h", nil),
		cold.SetListVMSnapshotsByVMId("v", models.VMSnapshots{}))
	if _, err := cold.GetListVMSnapshotsByVMId("v"); !errors.Is(err, want) {
		got = append(got, err)
	}
	for i, err := range got {
		if !errors.Is(err, want) {
			t.Fatalf("cold guard %d: %v", i, err)
		}
	}
}
