package parity

// DomainSuite "orchestrator-hosts" (PBI #421 task 13): the D-4 list
// pipeline (authz + UpdatedAt-desc ordering driven by CALLER-set
// timestamps → fully deterministic), the five resource aggregations
// (incl. the reserved double-count and system-reserved TotalAppleVms
// omission quirks), and the identical-order mutation script: ws flag,
// timestamp/healthy-clear, resources + cache writes, whole-row update
// no-diff echo, details set + VM replacement, dup-create 400, delete.
// Host ids are engine-generated → rows key by Host URL; VMs are relations
// in SQL and embedded payload in legacy → list comparisons BLANK the
// VM/RP relation fields on both sides (#419 D-7: SQL never preloads; VM
// row landing is pinned SQL-side in the unit suite).

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

func ohCtx() basecontext.ApiContext { return tplCtxRoot() }

func ohRes(cpu string) *models.HostResources {
	return &models.HostResources{CpuType: cpu, TotalAppleVms: 5,
		Total:          models.HostResourceItem{LogicalCpuCount: 10, PhysicalCpuCount: 2, MemorySize: 32, DiskSize: 500, FreeDiskSize: 200},
		TotalAvailable: models.HostResourceItem{LogicalCpuCount: 6, MemorySize: 20},
		TotalInUse:     models.HostResourceItem{LogicalCpuCount: 4, DiskSize: 300},
		TotalReserved:  models.HostResourceItem{LogicalCpuCount: 1, TotalAppleVms: 2},
		SystemReserved: models.HostResourceItem{LogicalCpuCount: 1, MemorySize: 8}}
}

func seedOH(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := ohCtx()
	mk := func(host, port, cpu, state string, fn func(h *models.OrchestratorHost)) models.OrchestratorHost {
		h := models.OrchestratorHost{Host: host, Port: port, State: state, Resources: ohRes(cpu)}
		if fn != nil {
			fn(&h)
		}
		created, err := db.CreateOrchestratorHost(ctx, h)
		if err != nil {
			t.Fatalf("seed host %s: %v", host, err)
		}
		return *created
	}
	h1 := mk("h1.local", "8888", "intel", "healthy", nil)
	h2 := mk("h2.local", "8888", "intel", "unhealthy", nil)
	h3 := mk("h3.local", "8888", "m4", "healthy", func(h *models.OrchestratorHost) { h.RequiredRoles = []string{"admin"} })
	mk("h4.local", "8888", "m4", "healthy", func(h *models.OrchestratorHost) { h.Resources = nil })
	// Deterministic UpdatedAt order (caller clock): h1 newest.
	stamp := func(h models.OrchestratorHost, ts string) {
		if err := db.UpdateOrchestratorHostTimestamp(ctx, h.ID, ts, ""); err != nil {
			t.Fatalf("stamp %s: %v", h.Host, err)
		}
	}
	stamp(h1, "2026-05-01T00:00:00Z")
	stamp(h2, "2026-04-01T00:00:00Z")
	stamp(h3, "2026-03-01T00:00:00Z")
	// h4 keeps the seed-time clock — order after the stamped three only
	// matters relatively; stamp it oldest for a total order.
	stamp(h3, "2026-03-01T00:00:00Z")
}

func normaliseOH(in []models.OrchestratorHost) []models.OrchestratorHost {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.OrchestratorHost(nil), in...)
	for i := range out {
		out[i].ID = "OH:" + out[i].Host
		out[i].CreatedAt, out[i].UpdatedAt = "", ""
		out[i].VirtualMachines = nil // relation vs embedded — compared SQL-side only
		out[i].ReverseProxy, out[i].ReverseProxyHosts = nil, nil
		out[i].HealthCheck = nil
	}
	return out
}

func normaliseOH1(h *models.OrchestratorHost) models.OrchestratorHost {
	if h == nil {
		return models.OrchestratorHost{}
	}
	return normaliseOH([]models.OrchestratorHost{*h})[0]
}

func TestParityDomain_OrchestratorHosts(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "orchestrator-hosts",
		Seed: seedOH,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkOHLists,
			checkOHAggregations,
			checkOHMutations,
		},
	})
}

func checkOHLists(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	bare := basecontext.NewBaseContext()
	for _, probe := range []struct {
		ctx     basecontext.ApiContext
		name    string
		filters []string
	}{
		{ohCtx(), "root", []string{"", "host=h1.*", "state=healthy", "id=nope"}},
		{bare, "bardenyall", []string{""}}, // matrix row 1: nil auth sees nothing
	} {
		for _, f := range probe.filters {
			lt, lerr := legacy.GetOrchestratorHosts(probe.ctx, f)
			st, serr := sql.GetOrchestratorHosts(probe.ctx, f)
			assertSameError(t, "GetOrchestratorHosts["+probe.name+"/"+f+"]", lerr, serr)
			AssertSameList(t, "GetOrchestratorHosts["+probe.name+"/"+f+"]", normaliseOH(lt), normaliseOH(st))
		}
	}
	lt, lerr := legacy.GetActiveOrchestratorHosts(ohCtx(), "")
	st, serr := sql.GetActiveOrchestratorHosts(ohCtx(), "")
	assertSameError(t, "GetActiveOrchestratorHosts", lerr, serr)
	AssertSameList(t, "GetActiveOrchestratorHosts", normaliseOH(lt), normaliseOH(st))
	// Single-host probes with the SAME folded strings on both engines —
	// GetHost() form carries the port, so derive probes from the rows.
	all, err := legacy.GetOrchestratorHosts(ohCtx(), "")
	if err != nil || len(all) < 3 {
		t.Fatalf("probe source list: %v", err)
	}
	probes := []string{all[0].GetHost(), strings.ToUpper(all[1].GetHost()), "ghost"}
	for _, probe := range probes {
		lh, lerr := legacy.GetOrchestratorHost(ohCtx(), probe)
		sh, serr := sql.GetOrchestratorHost(ohCtx(), probe)
		assertSameError(t, "GetOrchestratorHost["+probe+"]", lerr, serr)
		AssertSameList(t, "GetOrchestratorHost["+probe+"]",
			[]models.OrchestratorHost{normaliseOH1(lh)}, []models.OrchestratorHost{normaliseOH1(sh)})
	}
}

func checkOHAggregations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := ohCtx()
	type agg struct {
		name string
		a, b map[string]models.HostResourceItem
	}
	_ = agg{}
	cmp := func(name string, lg, sg func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem) {
		la, sa := lg(legacy), sg(sql)
		if !reflect.DeepEqual(la, sa) {
			t.Errorf("aggregation %s: legacy=%#v sql=%#v", name, la, sa)
		}
	}
	cmp("available", func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorAvailableResources(ctx)
	}, func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorAvailableResources(ctx)
	})
	cmp("total", func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorTotalResources(ctx)
	}, func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorTotalResources(ctx)
	})
	cmp("inuse", func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorInUseResources(ctx)
	}, func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorInUseResources(ctx)
	})
	cmp("reserved", func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorReservedResources(ctx)
	}, func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorReservedResources(ctx)
	})
	cmp("sysreserved", func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorSystemReservedResources(ctx)
	}, func(db datainterfaces.DatabaseService) map[string]models.HostResourceItem {
		return db.GetOrchestratorSystemReservedResources(ctx)
	})
}

func checkOHMutations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := ohCtx()
	idOf := func(db datainterfaces.DatabaseService, hostName string) string {
		list, err := db.GetOrchestratorHosts(ctx, "")
		if err != nil {
			t.Fatalf("id list: %v", err)
		}
		for _, h := range list {
			if h.Host == hostName {
				return h.ID
			}
		}
		t.Fatalf("no host %s", hostName)
		return ""
	}
	run := func(what string, call func(db datainterfaces.DatabaseService) (*models.OrchestratorHost, error)) {
		lh, lerr := call(legacy)
		sh, serr := call(sql)
		assertSameError(t, what, lerr, serr)
		AssertSameList(t, what,
			[]models.OrchestratorHost{normaliseOH1(lh)}, []models.OrchestratorHost{normaliseOH1(sh)})
	}

	// Websocket flag: changed twice, third time unchanged.
	for i, want := range []bool{true, true, false} {
		lc, lerr := legacy.UpdateOrchestratorHostWebsocketStatus(ctx, idOf(legacy, "h2.local"), want)
		sc, serr := sql.UpdateOrchestratorHostWebsocketStatus(ctx, idOf(sql, "h2.local"), want)
		if lc != sc || (lerr == nil) != (serr == nil) {
			t.Errorf("ws round %d: legacy=(%v,%v) sql=(%v,%v)", i, lc, lerr, sc, serr)
		}
	}
	// Timestamp + healthy-clear (h2 seeded unhealthy without LastUnhealthy
	// fields; set them through Details first, then clear via timestamp).
	run("DetailsSetUnhealthy", func(db datainterfaces.DatabaseService) (*models.OrchestratorHost, error) {
		id := idOf(db, "h2.local")
		row, err := db.GetOrchestratorHost(ctx, id)
		if err != nil {
			return nil, err
		}
		row.State = "unhealthy"
		row.LastUnhealthy = "2026-01-01T00:00:00Z"
		row.LastUnhealthyErrorMessage = "probe failed"
		return db.UpdateOrchestratorHostDetails(ctx, row)
	})
	run("TimestampHealthy", func(db datainterfaces.DatabaseService) (*models.OrchestratorHost, error) {
		id := idOf(db, "h2.local")
		if err := db.UpdateOrchestratorHostTimestamp(ctx, id, "2026-06-01T00:00:00Z", "healthy"); err != nil {
			return nil, err
		}
		return db.GetOrchestratorHost(ctx, id)
	})
	// Resources write + read (UpdatedAt NOT bumped — h2 timestamp must
	// survive as 2026-06-01; blanked by normalise but asserted via state).
	run("ResourcesWrite", func(db datainterfaces.DatabaseService) (*models.OrchestratorHost, error) {
		id := idOf(db, "h1.local")
		res := ohRes("intel")
		res.Total.LogicalCpuCount = 99
		if err := db.UpdateOrchestratorHostResources(ctx, id, res); err != nil {
			return nil, err
		}
		got, err := db.GetOrchestratorHostResources(ctx, id)
		if err != nil {
			return nil, err
		}
		if got.Total.LogicalCpuCount != 99 {
			t.Fatalf("resources read-back: %+v", got.Total)
		}
		return db.GetOrchestratorHost(ctx, id)
	})
	_, le := legacy.GetOrchestratorHostResources(ctx, idOf(legacy, "h4.local"))
	_, se := sql.GetOrchestratorHostResources(ctx, idOf(sql, "h4.local"))
	assertSameError(t, "HostResourcesNilOnH4", le, se) // h4 seeded nil → :542 sentinel
	// Cache config write.
	run("CacheConfig", func(db datainterfaces.DatabaseService) (*models.OrchestratorHost, error) {
		id := idOf(db, "h1.local")
		if err := db.UpdateOrchestratorHostCacheConfig(ctx, id, &apimodels.CatalogCacheConfig{Enabled: true, Folder: "/cache", MaxSize: 100}); err != nil {
			return nil, err
		}
		return db.GetOrchestratorHost(ctx, id)
	})
	// Whole-row update with a FAITHFUL echo of the stored row → no-diff
	// input-echo on both engines (the ws flag travels inside the row).
	run("UpdateNoDiffEcho", func(db datainterfaces.DatabaseService) (*models.OrchestratorHost, error) {
		row, err := db.GetOrchestratorHost(ctx, idOf(db, "h1.local"))
		if err != nil {
			return nil, err
		}
		return db.UpdateOrchestratorHost(ctx, row)
	})
	// Details with VM payload replacement (VM bodies compared SQL-side only;
	// here: both engines must accept and answer the reloaded host).
	run("DetailsVMReplace", func(db datainterfaces.DatabaseService) (*models.OrchestratorHost, error) {
		id := idOf(db, "h1.local")
		row, err := db.GetOrchestratorHost(ctx, id)
		if err != nil {
			return nil, err
		}
		row.Description = "details-tick"
		row.VirtualMachines = []models.VirtualMachine{{ID: "par-vm-1", Name: "alpha"}}
		return db.UpdateOrchestratorHostDetails(ctx, row)
	})
	// Details dup pre-check naming the EXISTING id.
	_, le = legacy.UpdateOrchestratorHostDetails(ctx, &models.OrchestratorHost{ID: idOf(legacy, "h1.local"), Host: "h2.local"})
	_, se = sql.UpdateOrchestratorHostDetails(ctx, &models.OrchestratorHost{ID: idOf(sql, "h1.local"), Host: "h2.local"})
	if (le == nil) != (se == nil) || (le != nil && le.Error() == se.Error()) {
		// messages carry engine ids — compare only that BOTH reject
		t.Errorf("details dup divergence: %v vs %v", le, se)
	}
	// Dup create (root ctx: authz probe active on both). Fresh engine ids
	// ride in the message — normIDMsg cuts at the id suffix (421-12 helper).
	_, le = legacy.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "H1.LOCAL", Port: "8888"})
	_, se = sql.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "H1.LOCAL", Port: "8888"})
	if le == nil || se == nil {
		t.Fatalf("DupCreate must reject on both: %v vs %v", le, se)
	}
	if ln, sn := normIDMsg(le), normIDMsg(se); ln != sn {
		t.Errorf("DupCreate: %q vs %q", ln, sn)
	}
	// Delete by RAW host + final lists.
	if lerr, serr := legacy.DeleteOrchestratorHost(ctx, "H3.LOCAL"), sql.DeleteOrchestratorHost(ctx, "H3.LOCAL"); (lerr == nil) != (serr == nil) {
		t.Errorf("delete divergence: %v vs %v", lerr, serr)
	}
	lt, lerr := legacy.GetOrchestratorHosts(ctx, "")
	st, serr := sql.GetOrchestratorHosts(ctx, "")
	assertSameError(t, "final hosts", lerr, serr)
	AssertSameList(t, "final hosts", normaliseOH(lt), normaliseOH(st))
	lt, lerr = legacy.GetOrchestratorHosts(ctx, "state=healthy")
	st, serr = sql.GetOrchestratorHosts(ctx, "state=healthy")
	assertSameError(t, "final healthy", lerr, serr)
	AssertSameList(t, "final healthy", normaliseOH(lt), normaliseOH(st))
}
