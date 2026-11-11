package parity

// DomainSuite "orchestrator-vms-snapshots" (PBI #421 task 14): the VM
// twins (authz-free raw-scan hosts, EXACT-== VM ids, idempotent-nil
// Remove/State/Uptime), the global VM list (AUTHZ hosts, healthy+enabled
// contributor gate, host UpdatedAt-desc then rowid), ReplaceOrchestratorHostVMs
// (whole-slice swap; SQL stamps the host_id carrier — blanked below), the
// RP views written through UpdateOrchestratorHost (legacy embedded objects
// vs SQL relation rows — payload uses fixed ids so views compare directly,
// carriers blanked), and both snapshot stores (EXACT hostId keys, per-VMId
// key upsert, global/host namespace isolation, empty-list carrier
// deviation pinned on the SQL side only).

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func seedOHVM(t testing.TB, db datainterfaces.DatabaseService) {
	seedOH(t, db)
	ctx := ohCtx()
	id := func(host string) string {
		list, err := db.GetOrchestratorHosts(ctx, "host="+host)
		if err != nil || len(list) != 1 {
			t.Fatalf("resolve %s: %v", host, err)
		}
		return list[0].ID
	}
	up := func(hostID string, vm models.VirtualMachine) {
		if err := db.UpsertOrchestratorHostVM(ctx, hostID, vm); err != nil {
			t.Fatalf("seed vm %s: %v", vm.ID, err)
		}
	}
	up(id("h1.local"), models.VirtualMachine{ID: "va-1", Name: "alpha", State: "running"})
	up(id("h1.local"), models.VirtualMachine{ID: "va-2", Name: "beta", State: "stopped"})
	up(id("h2.local"), models.VirtualMachine{ID: "vb-1", Name: "gamma", State: "running"}) // unhealthy host
	up(id("h3.local"), models.VirtualMachine{ID: "vc-1", Name: "delta", State: "running"}) // secret host
}

func ohvmIDOf(t testing.TB, db datainterfaces.DatabaseService, host string) string {
	list, err := db.GetOrchestratorHosts(ohCtx(), "host="+host)
	if err != nil {
		t.Fatalf("resolve %s: %v", host, err)
	}
	if len(list) != 1 {
		return "" // ghost host: the empty id matches nothing (HostNotFound)
	}
	return list[0].ID
}

func normaliseVM(in []models.VirtualMachine) []models.VirtualMachine {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.VirtualMachine(nil), in...)
	for i := range out {
		out[i].HostId = "" // carrier: SQL-stamped, legacy verbatim
	}
	return out
}

func normaliseSnap(in []models.VMSnapshot) []models.VMSnapshot {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.VMSnapshot(nil), in...)
	for i := range out {
		out[i].HostID, out[i].VMID = "", "" // json:"-" carriers, SQL-stamped
	}
	return out
}

func normaliseSnapRec(rec *models.HostsVMSnapshotsRecord) *models.HostsVMSnapshotsRecord {
	if rec == nil {
		return nil
	}
	out := &models.HostsVMSnapshotsRecord{HostId: "", VMSnapshots: map[string][]models.VMSnapshot{}}
	for k, v := range rec.VMSnapshots {
		out.VMSnapshots[k] = normaliseSnap(v)
	}
	return out
}

func TestParityDomain_OrchestratorVMsSnapshots(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "orchestrator-vms-snapshots",
		Seed: seedOHVM,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkOHVMLists,
			checkOHVMMutations,
			checkOHSnapshots,
		},
	})
}

func checkOHVMLists(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := ohCtx()
	bare := basecontext.NewBaseContext()

	// Global: healthy+enabled contributors only, authz-gated, host
	// UpdatedAt-desc then rowid: [va-1, va-2] (h1) then [vc-1] (h3).
	for _, probe := range []struct {
		ctx  basecontext.ApiContext
		name string
		f    string
	}{
		{ctx, "root", ""},
		{ctx, "root", "state=running"},
		{ctx, "root", "name=beta"},
		{ctx, "root", "id=nope"},
		{bare, "bardenyall", ""},
	} {
		l, lerr := legacy.GetOrchestratorVirtualMachines(probe.ctx, probe.f)
		s, serr := sql.GetOrchestratorVirtualMachines(probe.ctx, probe.f)
		assertSameError(t, "GetOrchestratorVirtualMachines["+probe.name+"/"+probe.f+"]", lerr, serr)
		AssertSameList(t, "GetOrchestratorVirtualMachines["+probe.name+"/"+probe.f+"]", normaliseVM(l), normaliseVM(s))
	}
	// Host-scoped reads carry NO health gate: unhealthy h2 still serves.
	for _, host := range []string{"h1.local", "h2.local", "h3.local", "ghost.local"} {
		lh, sh := ohvmIDOf(t, legacy, host), ohvmIDOf(t, sql, host)
		l, lerr := legacy.GetOrchestratorHostVirtualMachines(ctx, lh, "")
		s, serr := sql.GetOrchestratorHostVirtualMachines(ctx, sh, "")
		assertSameError(t, "HostVMs["+host+"]", lerr, serr)
		AssertSameList(t, "HostVMs["+host+"]", normaliseVM(l), normaliseVM(s))
	}
	// Single VM: Fold id hits, Name does NOT match.
	for _, probe := range []string{"VA-1", "alpha", "ghost"} {
		lh, sh := ohvmIDOf(t, legacy, "h1.local"), ohvmIDOf(t, sql, "h1.local")
		l, lerr := legacy.GetOrchestratorHostVirtualMachine(ctx, lh, probe)
		s, serr := sql.GetOrchestratorHostVirtualMachine(ctx, sh, probe)
		assertSameError(t, "HostVM["+probe+"]", lerr, serr)
		var ln, sn []models.VirtualMachine
		if l != nil {
			ln = normaliseVM([]models.VirtualMachine{*l})
		}
		if s != nil {
			sn = normaliseVM([]models.VirtualMachine{*s})
		}
		AssertSameList(t, "HostVM["+probe+"]", ln, sn)
	}
}

func checkOHVMMutations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := ohCtx()
	lh1, sh1 := ohvmIDOf(t, legacy, "h1.local"), ohvmIDOf(t, sql, "h1.local")
	ghost := "no-such-host"
	runPair := func(name string, lf, sf func() error) {
		lerr, serr := lf(), sf()
		assertSameError(t, name, lerr, serr)
	}
	blank := func(c *models.ReverseProxy) *models.ReverseProxy {
		if c == nil {
			return nil
		}
		cp := *c
		cp.HostID = "" // SQL stamps the carrier, legacy stores the payload verbatim
		return &cp
	}

	// --- RP views written through UpdateOrchestratorHost (fidelity fix) ----
	in, err := legacy.GetOrchestratorHost(ctx, lh1)
	if err != nil {
		t.Fatal(err)
	}
	in.Description = "rp-set"
	in.ReverseProxy = &models.ReverseProxy{ID: "cfg-1", Port: "443", Enabled: true}
	in.ReverseProxyHosts = []*models.ReverseProxyHost{{ID: "rp-1", Host: "a.rp", Port: "8080"}, {ID: "rp-2", Host: "b.rp"}}
	if _, err := legacy.UpdateOrchestratorHost(ctx, in); err != nil {
		t.Fatal(err)
	}
	in2, err := sql.GetOrchestratorHost(ctx, sh1)
	if err != nil {
		t.Fatal(err)
	}
	in2.Description = "rp-set"
	in2.ReverseProxy = &models.ReverseProxy{ID: "cfg-1", Port: "443", Enabled: true}
	in2.ReverseProxyHosts = []*models.ReverseProxyHost{{ID: "rp-1", Host: "a.rp", Port: "8080"}, {ID: "rp-2", Host: "b.rp"}}
	if _, err := sql.UpdateOrchestratorHost(ctx, in2); err != nil {
		t.Fatal(err)
	}
	lc, lcerr := legacy.GetOrchestratorReverseProxyConfig(ctx, lh1)
	sc, scerr := sql.GetOrchestratorReverseProxyConfig(ctx, sh1)
	assertSameError(t, "RPConfig", lcerr, scerr)
	if !reflect.DeepEqual(blank(lc), blank(sc)) {
		t.Errorf("RP config: legacy=%#v sql=%#v", lc, sc)
	}
	lh, lherr := legacy.GetOrchestratorReverseProxyHosts(ctx, lh1, "")
	sh, sherr := sql.GetOrchestratorReverseProxyHosts(ctx, sh1, "")
	assertSameError(t, "RPHosts", lherr, sherr)
	AssertSameList(t, "RPHosts", normaliseRPHosts(lh), normaliseRPHosts(sh))
	for _, probe := range []string{"rp-2", "A.RP:8080", "rp-ghost"} {
		l1, l1err := legacy.GetOrchestratorReverseProxyHost(ctx, lh1, probe)
		s1, s1err := sql.GetOrchestratorReverseProxyHost(ctx, sh1, probe)
		assertSameError(t, "RPHost["+probe+"]", l1err, s1err)
		AssertSameList(t, "RPHost["+probe+"]", normaliseRPHostsPtr(l1), normaliseRPHostsPtr(s1))
	}

	// --- twins + delete + replace (legacy THEN sql, errors compared) ------
	runPair("State",
		func() error { return legacy.UpdateOrchestratorHostVMState(ctx, lh1, "va-1", "paused") },
		func() error { return sql.UpdateOrchestratorHostVMState(ctx, sh1, "va-1", "paused") })
	runPair("Uptime",
		func() error { return legacy.UpdateOrchestratorHostVMUptime(ctx, lh1, "va-1", "42h") },
		func() error { return sql.UpdateOrchestratorHostVMUptime(ctx, sh1, "va-1", "42h") })
	runPair("State-vm-miss",
		func() error { return legacy.UpdateOrchestratorHostVMState(ctx, lh1, "ghost-vm", "x") },
		func() error { return sql.UpdateOrchestratorHostVMState(ctx, sh1, "ghost-vm", "x") })
	runPair("State-host-miss",
		func() error { return legacy.UpdateOrchestratorHostVMState(ctx, ghost, "va-1", "x") },
		func() error { return sql.UpdateOrchestratorHostVMState(ctx, ghost, "va-1", "x") })
	runPair("Uptime-vm-miss",
		func() error { return legacy.UpdateOrchestratorHostVMUptime(ctx, lh1, "ghost-vm", "x") },
		func() error { return sql.UpdateOrchestratorHostVMUptime(ctx, sh1, "ghost-vm", "x") })
	runPair("Remove-vm-miss",
		func() error { return legacy.RemoveOrchestratorHostVM(ctx, lh1, "ghost-vm") },
		func() error { return sql.RemoveOrchestratorHostVM(ctx, sh1, "ghost-vm") })
	runPair("Remove-host-miss",
		func() error { return legacy.RemoveOrchestratorHostVM(ctx, ghost, "va-1") },
		func() error { return sql.RemoveOrchestratorHostVM(ctx, ghost, "va-1") })
	// EXACT == on ids: "VA-1" is an APPEND, then a precise remove.
	runPair("Upsert-case-new",
		func() error {
			return legacy.UpsertOrchestratorHostVM(ctx, lh1, models.VirtualMachine{ID: "VA-1", Name: "case-new"})
		},
		func() error {
			return sql.UpsertOrchestratorHostVM(ctx, sh1, models.VirtualMachine{ID: "VA-1", Name: "case-new"})
		})
	runPair("Upsert-va-1-replace",
		func() error {
			return legacy.UpsertOrchestratorHostVM(ctx, lh1, models.VirtualMachine{ID: "va-1", Name: "alpha2", State: "paused", Uptime: "42h"})
		},
		func() error {
			return sql.UpsertOrchestratorHostVM(ctx, sh1, models.VirtualMachine{ID: "va-1", Name: "alpha2", State: "paused", Uptime: "42h"})
		})
	runPair("Remove-case",
		func() error { return legacy.RemoveOrchestratorHostVM(ctx, lh1, "VA-1") },
		func() error { return sql.RemoveOrchestratorHostVM(ctx, sh1, "VA-1") })
	runPair("Delete-by-fold-host-and-name",
		func() error { return legacy.DeleteOrchestratorVirtualMachine(ctx, "H1.LOCAL", "BETA") },
		func() error { return sql.DeleteOrchestratorVirtualMachine(ctx, "H1.LOCAL", "BETA") })
	runPair("Delete-double",
		func() error { return legacy.DeleteOrchestratorVirtualMachine(ctx, lh1, "va-2") },
		func() error { return sql.DeleteOrchestratorVirtualMachine(ctx, sh1, "va-2") })
	runPair("Delete-host-miss",
		func() error { return legacy.DeleteOrchestratorVirtualMachine(ctx, "ghost.local", "x") },
		func() error { return sql.DeleteOrchestratorVirtualMachine(ctx, "ghost.local", "x") })
	runPair("Delete-empty-id",
		func() error { return legacy.DeleteOrchestratorVirtualMachine(ctx, "", "x") },
		func() error { return sql.DeleteOrchestratorVirtualMachine(ctx, "", "x") })
	runPair("Replace-2",
		func() error {
			return legacy.ReplaceOrchestratorHostVMs(ctx, lh1, []models.VirtualMachine{{ID: "ra", Name: "one", HostId: "elsewhere"}, {ID: "rb", Name: "two"}})
		},
		func() error {
			return sql.ReplaceOrchestratorHostVMs(ctx, sh1, []models.VirtualMachine{{ID: "ra", Name: "one", HostId: "elsewhere"}, {ID: "rb", Name: "two"}})
		})

	// Host-scoped + global lists at the Replace-2 state.
	lv, lverr := legacy.GetOrchestratorHostVirtualMachines(ctx, lh1, "")
	sv, sverr := sql.GetOrchestratorHostVirtualMachines(ctx, sh1, "")
	assertSameError(t, "HostVMs/replace2", lverr, sverr)
	AssertSameList(t, "HostVMs/replace2", normaliseVM(lv), normaliseVM(sv))
	l, lerr := legacy.GetOrchestratorVirtualMachines(ctx, "")
	s, serr := sql.GetOrchestratorVirtualMachines(ctx, "")
	assertSameError(t, "VMs/post-replace", lerr, serr)
	AssertSameList(t, "VMs/post-replace", normaliseVM(l), normaliseVM(s))

	runPair("Replace-0",
		func() error { return legacy.ReplaceOrchestratorHostVMs(ctx, lh1, nil) },
		func() error { return sql.ReplaceOrchestratorHostVMs(ctx, sh1, nil) })
	runPair("Replace-host-miss",
		func() error { return legacy.ReplaceOrchestratorHostVMs(ctx, ghost, nil) },
		func() error { return sql.ReplaceOrchestratorHostVMs(ctx, ghost, nil) })

	// Clear the RP payload with an explicit nil payload on BOTH sides:
	// legacy assigns the nil embedded fields (:253-254), SQL deletes the
	// rows. Legacy LIST rows carry the embedded slice (shared pointer), SQL
	// LIST rows never attach — nil the payload to keep the script symmetric.
	lall, _ := legacy.GetOrchestratorHosts(ctx, "host=h1.local")
	sall, _ := sql.GetOrchestratorHosts(ctx, "host=h1.local")
	lall[0].Description = "rp-clear"
	lall[0].ReverseProxy, lall[0].ReverseProxyHosts = nil, nil
	sall[0].Description = "rp-clear"
	if _, err := legacy.UpdateOrchestratorHost(ctx, &lall[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := sql.UpdateOrchestratorHost(ctx, &sall[0]); err != nil {
		t.Fatal(err)
	}
	lc, lcerr = legacy.GetOrchestratorReverseProxyConfig(ctx, lh1)
	sc, scerr = sql.GetOrchestratorReverseProxyConfig(ctx, sh1)
	assertSameError(t, "RPConfig/cleared", lcerr, scerr)
	if !reflect.DeepEqual(blank(lc), blank(sc)) {
		t.Errorf("RP config after clear: legacy=%#v sql=%#v", lc, sc)
	}
	lh, lherr = legacy.GetOrchestratorReverseProxyHosts(ctx, lh1, "")
	sh, sherr = sql.GetOrchestratorReverseProxyHosts(ctx, sh1, "")
	assertSameError(t, "RPHosts/cleared", lherr, sherr)
	AssertSameList(t, "RPHosts/cleared", normaliseRPHosts(lh), normaliseRPHosts(sh))
}

func normaliseRPHosts(in []*models.ReverseProxyHost) []*models.ReverseProxyHost {
	out := make([]*models.ReverseProxyHost, 0, len(in))
	for _, h := range in {
		if h == nil {
			out = append(out, nil)
			continue
		}
		cp := *h
		cp.HostID = ""
		out = append(out, &cp)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normaliseRPHostsPtr(in *models.ReverseProxyHost) []*models.ReverseProxyHost {
	if in == nil {
		return nil
	}
	return normaliseRPHosts([]*models.ReverseProxyHost{in})
}

func checkOHSnapshots(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := ohCtx()
	lh1, sh1 := ohvmIDOf(t, legacy, "h1.local"), ohvmIDOf(t, sql, "h1.local")

	setH := func(db datainterfaces.DatabaseService, hostID, vmID string, snaps ...models.VMSnapshot) error {
		return db.SetHostVMSnapshots(ctx, hostID, models.VMSnapshots{VMId: vmID, VMSnapshot: snaps})
	}
	if err := setH(legacy, lh1, "vm-a", models.VMSnapshot{ID: "s1", Name: "one"}, models.VMSnapshot{ID: "s2", Name: "two"}); err != nil {
		t.Fatal(err)
	}
	if err := setH(sql, sh1, "vm-a", models.VMSnapshot{ID: "s1", Name: "one"}, models.VMSnapshot{ID: "s2", Name: "two"}); err != nil {
		t.Fatal(err)
	}
	if err := setH(legacy, lh1, "vm-b", models.VMSnapshot{ID: "s3", Name: "three"}); err != nil {
		t.Fatal(err)
	}
	if err := setH(sql, sh1, "vm-b", models.VMSnapshot{ID: "s3", Name: "three"}); err != nil {
		t.Fatal(err)
	}
	lr, lerr := legacy.GetHostVMSnapshots(ctx, lh1)
	sr, serr := sql.GetHostVMSnapshots(ctx, sh1)
	assertSameError(t, "HostVMSnapshots", lerr, serr)
	if !reflect.DeepEqual(normaliseSnapRec(lr), normaliseSnapRec(sr)) {
		t.Errorf("host snapshots: legacy=%#v sql=%#v", lr, sr)
	}
	// Key-level upsert: vm-a replaced, vm-b survives.
	if err := setH(legacy, lh1, "vm-a", models.VMSnapshot{ID: "s1b", Name: "one-prime"}); err != nil {
		t.Fatal(err)
	}
	if err := setH(sql, sh1, "vm-a", models.VMSnapshot{ID: "s1b", Name: "one-prime"}); err != nil {
		t.Fatal(err)
	}
	lr, lerr = legacy.GetHostVMSnapshots(ctx, lh1)
	sr, serr = sql.GetHostVMSnapshots(ctx, sh1)
	assertSameError(t, "HostVMSnapshots/upsert", lerr, serr)
	if !reflect.DeepEqual(normaliseSnapRec(lr), normaliseSnapRec(sr)) {
		t.Errorf("host snapshots upsert: legacy=%#v sql=%#v", lr, sr)
	}
	// Miss + EXACT-hostId probes (Fold must not apply: upper-cased own id
	// is a miss on BOTH engines; probes stay engine-scoped).
	for _, probe := range []struct{ lprobe, sprobe, name string }{
		{"ghost-host", "ghost-host", "ghost"},
		{strings.ToUpper(lh1), strings.ToUpper(sh1), "case-miss"},
	} {
		_, lerr := legacy.GetHostVMSnapshots(ctx, probe.lprobe)
		_, serr := sql.GetHostVMSnapshots(ctx, probe.sprobe)
		assertSameError(t, "HostVMSnapshots/miss["+probe.name+"]", lerr, serr)
	}

	// Global store — engine-agnostic keys, full compare.
	setL := func(db datainterfaces.DatabaseService, vmID string, snaps ...models.VMSnapshot) error {
		return db.SetListVMSnapshotsByVMId(vmID, models.VMSnapshots{VMId: vmID, VMSnapshot: snaps})
	}
	if err := setL(legacy, "gvm", models.VMSnapshot{ID: "g1", Name: "first"}); err != nil {
		t.Fatal(err)
	}
	if err := setL(sql, "gvm", models.VMSnapshot{ID: "g1", Name: "first"}); err != nil {
		t.Fatal(err)
	}
	if err := setL(legacy, "gvm", models.VMSnapshot{ID: "g2", Name: "second"}, models.VMSnapshot{ID: "g3", Name: "third"}); err != nil {
		t.Fatal(err)
	}
	if err := setL(sql, "gvm", models.VMSnapshot{ID: "g2", Name: "second"}, models.VMSnapshot{ID: "g3", Name: "third"}); err != nil {
		t.Fatal(err)
	}
	// A host record for the SAME vm id must not leak into the global view.
	if err := setH(legacy, lh1, "gvm", models.VMSnapshot{ID: "hr1"}); err != nil {
		t.Fatal(err)
	}
	if err := setH(sql, sh1, "gvm", models.VMSnapshot{ID: "hr1"}); err != nil {
		t.Fatal(err)
	}
	lt, lerr := legacy.GetListVMSnapshotsByVMId("gvm")
	st, serr := sql.GetListVMSnapshotsByVMId("gvm")
	assertSameError(t, "ListVMSnapshots", lerr, serr)
	AssertSameList(t, "ListVMSnapshots", normaliseSnap(lt), normaliseSnap(st))
	lt, lerr = legacy.GetListVMSnapshotsByVMId("none")
	st, serr = sql.GetListVMSnapshotsByVMId("none")
	assertSameError(t, "ListVMSnapshots/miss", lerr, serr)
	AssertSameList(t, "ListVMSnapshots/miss", normaliseSnap(lt), normaliseSnap(st))
}
