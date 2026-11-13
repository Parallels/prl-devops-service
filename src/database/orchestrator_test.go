package database

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

func ohCreate(t *testing.T, g *GormDatabase, ctx basecontext.ApiContext, host models.OrchestratorHost) *models.OrchestratorHost {
	t.Helper()
	h, err := g.CreateOrchestratorHost(ctx, host)
	if err != nil {
		t.Fatalf("create %s: %v", host.Host, err)
	}
	return h
}

func TestSQLOrchestratorHosts_ListPipeline(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	bare := basecontext.NewBaseContext()

	// Ordering: UpdatedAt desc (D-4), ties keep rowid.
	mk := func(name, updated string) *models.OrchestratorHost {
		h := ohCreate(t, g, root, models.OrchestratorHost{Host: name})
		if err := g.UpdateOrchestratorHostTimestamp(root, h.ID, updated, ""); err != nil {
			t.Fatal(err)
		}
		return h
	}
	mk("h1.local", "2026-01-01T00:00:00Z")
	mk("h2.local", "2026-03-01T00:00:00Z")
	mk("h3.local", "2026-02-01T00:00:00Z")
	list, err := g.GetOrchestratorHosts(root, "")
	if err != nil || len(list) != 3 {
		t.Fatalf("list: %v %+v", err, len(list))
	}
	if list[0].Host != "h2.local" || list[1].Host != "h3.local" || list[2].Host != "h1.local" {
		t.Fatalf("UpdatedAt-desc order: %s %s %s", list[0].Host, list[1].Host, list[2].Host)
	}
	// Relations never preloaded (#419 D-7).
	for _, h := range list {
		if h.VirtualMachines != nil || h.ReverseProxy != nil || h.ReverseProxyHosts != nil {
			t.Fatalf("relation preload leaked: %+v", h)
		}
	}
	// Authz: required-roles host invisible to bare ctx (matrix row 1).
	secret := ohCreate(t, g, root, models.OrchestratorHost{Host: "secret.local", RequiredRoles: []string{"admin"}})
	if _, err := g.GetOrchestratorHost(bare, secret.ID); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("authz leak to bare ctx: %v", err)
	}
	if _, err := g.GetOrchestratorHost(root, secret.ID); err != nil {
		t.Fatalf("root lookup: %v", err)
	}
	// Active filter: Enabled only (and it runs BEFORE the property filter).
	dis := ohCreate(t, g, root, models.OrchestratorHost{Host: "off.local"})
	if _, err := g.UpdateOrchestratorHost(root, &models.OrchestratorHost{ID: dis.ID, Host: "off.local", Enabled: false, OsVersion: "x"}); err != nil {
		t.Fatal(err)
	}
	act, err := g.GetActiveOrchestratorHosts(root, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range act {
		if h.Host == "off.local" {
			t.Fatal("disabled host in active list")
		}
	}
	if len(act) != 4 { // h1..h3 + secret (root sees it)
		t.Fatalf("active count: %d", len(act))
	}
	// GetHost-folded probe (url build: schema://host:port/prefix).
	h1, err := g.GetOrchestratorHost(root, "H2.LOCAL")
	if err != nil {
		t.Fatalf("folded host probe: %v", err)
	}
	_ = h1
}

func TestSQLOrchestratorHosts_CreateDelete(t *testing.T) {
	g := newSQLDB(t)
	ctx := tplRootCtx()

	if _, err := g.CreateOrchestratorHost(ctx, models.OrchestratorHost{}); !errors.Is(err, interfaces.ErrOrchestratorHostEmptyName) {
		t.Fatalf("empty host sentinel: %v", err)
	}
	h := ohCreate(t, g, ctx, models.OrchestratorHost{
		Host: "vm-host.local", Port: "8888",
		VirtualMachines: []models.VirtualMachine{{ID: "vm-1", Name: "one"}, {ID: "vm-2", Name: "two"}},
	})
	if h.ID == "" || !h.Enabled || h.CreatedAt == "" || h.UpdatedAt == "" {
		t.Fatalf("create stamps: %+v", h)
	}
	if len(h.VirtualMachines) != 2 { // returned view keeps the payload
		t.Fatalf("returned VMs: %+v", h.VirtualMachines)
	}
	var n int64
	g.gormDB.Table("virtual_machines").Where("host_id = ?", h.ID).Count(&n)
	if n != 2 {
		t.Fatalf("VM rows not landed with HostId carrier: %d", n)
	}
	// Dup probe: 400 message carries the FRESH id, not the existing one.
	dup, err := g.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "VM-HOST.LOCAL", Port: "8888"})
	if err == nil || dup != nil {
		t.Fatalf("dup must fail: %+v %v", dup, err)
	}
	if !strings.Contains(err.Error(), "already exists with ID") {
		t.Fatalf("dup message: %v", err)
	}
	if strings.Contains(err.Error(), h.ID) {
		t.Fatalf("dup message leaked the EXISTING id (legacy names the fresh one): %v", err)
	}
	// Delete matches RAW host (no port, no url fold).
	if err := g.DeleteOrchestratorHost(ctx, ""); !errors.Is(err, interfaces.ErrOrchestratorHostEmptyIdOrHost) {
		t.Fatalf("empty id sentinel: %v", err)
	}
	if err := g.DeleteOrchestratorHost(ctx, "VM-HOST.LOCAL"); err != nil {
		t.Fatalf("delete by raw host: %v", err)
	}
	g.gormDB.Table("virtual_machines").Where("host_id = ?", h.ID).Count(&n)
	if n != 0 {
		t.Fatalf("delete did not cascade VM rows: %d", n)
	}
	if err := g.DeleteOrchestratorHost(ctx, "ghost"); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("delete missing: %v", err)
	}
}

func TestSQLOrchestratorHosts_UpdateSemantics(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	h := ohCreate(t, g, ctx, models.OrchestratorHost{Host: "u.local", Port: "1"})
	if _, err := g.UpdateOrchestratorHostWebsocketStatus(ctx, h.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := g.UpdateOrchestratorHostDetails(ctx, &models.OrchestratorHost{ID: h.ID, Host: "u.local", Port: "1",
		VirtualMachines: []models.VirtualMachine{{ID: "seed-vm"}}}); err != nil {
		t.Fatal(err)
	}
	created := func() *models.OrchestratorHost {
		row, err := g.orchHostByID(h.ID)
		if err != nil {
			t.Fatal(err)
		}
		return row
	}
	before := created()

	// Update: writes OsVersion, must NOT touch HasWebsocketEvents / VM rows
	// / CreatedAt / ID.
	in := models.OrchestratorHost{ID: h.ID, Host: "u.local", Port: "1", OsVersion: "23H2",
		HasWebsocketEvents: true, Enabled: true, CreatedAt: "FORGED"}
	// Diff must fire (OsVersion differs) — input has NO VM payload; the
	// legacy set never assigns VirtualMachines, so payload-less input must
	// not empty the relation either.
	after, err := g.UpdateOrchestratorHost(ctx, &in)
	if err != nil {
		t.Fatal(err)
	}
	if after.OsVersion != "23H2" || !after.HasWebsocketEvents || after.CreatedAt != before.CreatedAt || after.ID != h.ID {
		t.Fatalf("Update field set violated: %+v", after)
	}
	var n int64
	g.gormDB.Table("virtual_machines").Where("host_id = ?", h.ID).Count(&n)
	if n != 1 {
		t.Fatalf("Update clobbered VM relation rows: %d", n)
	}
	// No-diff echoes the INPUT pointer. NOTE: Diff() DOES compare
	// HasWebsocketEvents (:245) while the update set NEVER writes it — a
	// caller echoing a stale ws flag re-fires the update on BOTH engines;
	// the input below mirrors the stored flag exactly, like legacy callers.
	echo, err := g.UpdateOrchestratorHost(ctx, &in)
	if err != nil || echo != &in {
		t.Fatalf("no-diff echo: %v %p %p", err, echo, &in)
	}
	// Not-found.
	if _, err := g.UpdateOrchestratorHost(ctx, &models.OrchestratorHost{ID: "ghost"}); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("update ghost: %v", err)
	}
	if _, err := g.UpdateOrchestratorHost(ctx, &models.OrchestratorHost{}); !errors.Is(err, interfaces.ErrOrchestratorHostEmptyIdOrHost) {
		t.Fatalf("update empty id: %v", err)
	}

	// Details: dup pre-check BEFORE existence, names the EXISTING id.
	other := ohCreate(t, g, ctx, models.OrchestratorHost{Host: "taken.local"})
	_, err = g.UpdateOrchestratorHostDetails(ctx, &models.OrchestratorHost{ID: "ghost-id", Host: "taken.local"})
	if err == nil || !strings.Contains(err.Error(), other.ID) {
		t.Fatalf("details dup must name existing id (ghost-id accepted?): %v", err)
	}
	// Details set: OsVersion is NOT in it (stays), VM payload REPLACES rows.
	d, err := g.UpdateOrchestratorHostDetails(ctx, &models.OrchestratorHost{ID: h.ID, Host: "u.local", Port: "1",
		OsVersion: "IGNORED", VirtualMachines: []models.VirtualMachine{{ID: "new-vm"}}})
	if err != nil || d.OsVersion != "23H2" {
		t.Fatalf("Details wrote OsVersion (not in its set): %v %+v", err, d)
	}
	g.gormDB.Table("virtual_machines").Where("host_id = ?", h.ID).Count(&n)
	if n != 1 {
		t.Fatalf("Details VM replace count: %d", n)
	}
	var ids []string
	g.gormDB.Table("virtual_machines").Where("host_id = ?", h.ID).Pluck("id", &ids)
	if len(ids) != 1 || ids[0] != "new-vm" {
		t.Fatalf("Details VM replace ids: %v", ids)
	}
}

func TestSQLOrchestratorHosts_FieldScopedWriters(t *testing.T) {
	g := newSQLDB(t)
	ctx := tplRootCtx()
	h := ohCreate(t, g, ctx, models.OrchestratorHost{Host: "f.local", State: "unhealthy"})
	stored := func() *models.OrchestratorHost {
		row, _ := g.orchHostByID(h.ID)
		return row
	}
	// Timestamp: caller clock + healthy clears LastUnhealthy fields.
	if _, err := g.UpdateOrchestratorHostDetails(ctx, &models.OrchestratorHost{ID: h.ID, Host: "f.local",
		State: "unhealthy", LastUnhealthy: "2026-01-01", LastUnhealthyErrorMessage: "boom"}); err != nil {
		t.Fatal(err)
	}
	if err := g.UpdateOrchestratorHostTimestamp(ctx, h.ID, "2026-05-05T00:00:00Z", "healthy"); err != nil {
		t.Fatal(err)
	}
	row := stored()
	if row.UpdatedAt != "2026-05-05T00:00:00Z" || row.State != "healthy" || row.LastUnhealthy != "" || row.LastUnhealthyErrorMessage != "" {
		t.Fatalf("timestamp/healthy-clear: %+v", row)
	}
	// Empty state leaves State untouched.
	if err := g.UpdateOrchestratorHostTimestamp(ctx, h.ID, "2026-05-06T00:00:00Z", ""); err != nil {
		t.Fatal(err)
	}
	if stored().State != "healthy" {
		t.Fatal("empty state clobbered State")
	}
	if err := g.UpdateOrchestratorHostTimestamp(ctx, "", "x", ""); !errors.Is(err, interfaces.ErrOrchestratorHostEmptyIdOrHost) {
		t.Fatalf("timestamp sentinel: %v", err)
	}
	if err := g.UpdateOrchestratorHostTimestamp(ctx, "ghost", "x", ""); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("timestamp ghost: %v", err)
	}
	// Websocket: changed-flag semantics.
	changed, err := g.UpdateOrchestratorHostWebsocketStatus(ctx, h.ID, true)
	if err != nil || !changed {
		t.Fatalf("first set must report changed: %v %v", changed, err)
	}
	changed, err = g.UpdateOrchestratorHostWebsocketStatus(ctx, h.ID, true)
	if err != nil || changed {
		t.Fatalf("same value must report unchanged: %v %v", changed, err)
	}
	if _, err := g.UpdateOrchestratorHostWebsocketStatus(ctx, "", true); !errors.Is(err, interfaces.ErrOrchestratorHostEmptyIdOrHost) {
		t.Fatalf("ws sentinel: %v", err)
	}
	if _, err := g.UpdateOrchestratorHostWebsocketStatus(ctx, "ghost", true); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("ws ghost: %v", err)
	}
	// Resources: persisted, UpdatedAt NOT bumped (:817 contract).
	updatedAtBefore := stored().UpdatedAt
	res := &models.HostResources{CpuType: "m4", Total: models.HostResourceItem{LogicalCpuCount: 8}}
	if err := g.UpdateOrchestratorHostResources(ctx, h.ID, res); err != nil {
		t.Fatal(err)
	}
	after := stored()
	if after.Resources == nil || after.Resources.CpuType != "m4" || after.UpdatedAt != updatedAtBefore {
		t.Fatalf("resources write bumped UpdatedAt or lost payload: %+v", after.Resources)
	}
	got, err := g.GetOrchestratorHostResources(ctx, h.ID)
	if err != nil || got.CpuType != "m4" {
		t.Fatalf("GetHostResources: %v %+v", err, got)
	}
	// Nil resources -> host exists but reads NotFound via :542 branch.
	if err := g.UpdateOrchestratorHostResources(ctx, h.ID, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := g.GetOrchestratorHostResources(ctx, h.ID); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("nil resources read: %v", err)
	}
	if err := g.UpdateOrchestratorHostResources(ctx, "ghost", res); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("resources ghost: %v", err)
	}
	// CacheConfig round trip.
	cc := &apimodels.CatalogCacheConfig{Enabled: true, Folder: "/var/cache", MaxSize: 42}
	if err := g.UpdateOrchestratorHostCacheConfig(ctx, h.ID, cc); err != nil {
		t.Fatal(err)
	}
	if got := stored().CacheConfig; got == nil || got.Folder != "/var/cache" || got.MaxSize != 42 {
		t.Fatalf("cache config round trip: %+v", got)
	}
	if err := g.UpdateOrchestratorHostCacheConfig(ctx, "", cc); !errors.Is(err, interfaces.ErrOrchestratorHostEmptyIdOrHost) {
		t.Fatalf("cache sentinel: %v", err)
	}
}

func TestSQLOrchestratorHosts_Aggregations(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	mk := func(cpu, state string, enabled bool, res *models.HostResources) {
		h, err := g.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: cpu + "-" + state + ".local", State: state, Resources: res})
		if err != nil {
			t.Fatal(err)
		}
		if !enabled {
			if _, err := g.UpdateOrchestratorHost(ctx, &models.OrchestratorHost{ID: h.ID, Host: h.Host, State: state, Enabled: false, Resources: res}); err != nil {
				t.Fatal(err)
			}
		}
	}
	res := func(cpu string) *models.HostResources {
		return &models.HostResources{CpuType: cpu, TotalAppleVms: 5,
			Total:          models.HostResourceItem{LogicalCpuCount: 10, TotalAppleVms: 3},
			TotalAvailable: models.HostResourceItem{LogicalCpuCount: 6, MemorySize: 2},
			TotalInUse:     models.HostResourceItem{LogicalCpuCount: 4},
			TotalReserved:  models.HostResourceItem{LogicalCpuCount: 1, TotalAppleVms: 2},
			SystemReserved: models.HostResourceItem{LogicalCpuCount: 1}}
	}
	mk("intel", "healthy", true, res("intel"))
	mk("intel", "healthy", true, res("intel"))
	mk("m4", "healthy", true, res("m4"))
	mk("intel", "unhealthy", true, res("intel")) // excluded: state
	mk("m4", "healthy", false, res("m4"))        // excluded: disabled
	mk("", "healthy", true, nil)                 // excluded: nil resources

	avail := g.GetOrchestratorAvailableResources(ctx)
	if len(avail) != 2 {
		t.Fatalf("available groups: %+v", avail)
	}
	if avail["intel"].LogicalCpuCount != 12 || avail["intel"].MemorySize != 4 {
		t.Fatalf("available sum: %+v", avail["intel"])
	}
	tot := g.GetOrchestratorTotalResources(ctx)
	if tot["intel"].LogicalCpuCount != 20 || tot["intel"].TotalAppleVms != 6 {
		t.Fatalf("total sum: %+v", tot["intel"])
	}
	inuse := g.GetOrchestratorInUseResources(ctx)
	if inuse["m4"].LogicalCpuCount != 4 {
		t.Fatalf("in-use sum: %+v", inuse["m4"])
	}
	// Reserved: TotalAppleVms = host-level 5 + reserved 2 = 7 PER HOST.
	rsv := g.GetOrchestratorReservedResources(ctx)
	if rsv["intel"].TotalAppleVms != 14 || rsv["intel"].LogicalCpuCount != 2 {
		t.Fatalf("reserved double-count quirk: %+v", rsv["intel"])
	}
	// System reserved: NO TotalAppleVms at all.
	sys := g.GetOrchestratorSystemReservedResources(ctx)
	if sys["intel"].TotalAppleVms != 0 || sys["intel"].LogicalCpuCount != 2 {
		t.Fatalf("system-reserved TotalAppleVms leak: %+v", sys["intel"])
	}
	// Guard-free-ness: aggregations answer EMPTY, never error, cold.
	cold := NewGormDatabase(ctx, nil)
	if m := cold.GetOrchestratorAvailableResources(ctx); m == nil || len(m) != 0 {
		t.Fatalf("cold aggregation must answer empty non-nil map: %#v", m)
	}
	if _, err := cold.GetOrchestratorHostResources(ctx, "x"); !errors.Is(err, interfaces.ErrOrchestratorHostNotFound) {
		t.Fatalf("cold GetHostResources (no legacy guard): %v", err)
	}
}

// Heartbeat-vs-VM interleaving (task-13 concurrency proof): field-scoped
// timestamp writes and VM-row writes must never destroy each other.
func TestSQLOrchestratorHosts_HeartbeatDoesNotClobberVMs(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	h := ohCreate(t, g, ctx, models.OrchestratorHost{Host: "race.local",
		VirtualMachines: []models.VirtualMachine{{ID: "va"}, {ID: "vb"}}})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 25; i++ {
			if err := g.UpdateOrchestratorHostTimestamp(ctx, h.ID, "2026-06-06T00:00:00Z", "healthy"); err != nil {
				t.Errorf("heartbeat: %v", err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 25; i++ {
			if _, err := g.UpdateOrchestratorHost(ctx, &models.OrchestratorHost{ID: h.ID, Host: "race.local", Description: "tick"}); err != nil {
				t.Errorf("update: %v", err)
				return
			}
		}
	}()
	wg.Wait()
	var n int64
	g.gormDB.Table("virtual_machines").Where("host_id = ?", h.ID).Count(&n)
	if n != 2 {
		t.Fatalf("VM rows clobbered by heartbeat/update: %d", n)
	}
	row, _ := g.orchHostByID(h.ID)
	if row.UpdatedAt != "2026-06-06T00:00:00Z" && row.Description != "tick" {
		t.Fatalf("neither writer landed: %+v", row)
	}
}

func TestSQLOrchestratorHosts_NotConnected(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cold := NewGormDatabase(ctx, nil)
	checks := []struct {
		name string
		err  error
	}{
		{"list", errOf(cold.GetOrchestratorHosts(ctx, ""))},
		{"active", errOf(cold.GetActiveOrchestratorHosts(ctx, ""))},
		{"get", errOf(cold.GetOrchestratorHost(ctx, "x"))},
		{"create", errOf(cold.CreateOrchestratorHost(ctx, models.OrchestratorHost{}))},
		{"delete", cold.DeleteOrchestratorHost(ctx, "x")},
		{"update", errOf(cold.UpdateOrchestratorHost(ctx, &models.OrchestratorHost{ID: "x"}))},
		{"details", errOf(cold.UpdateOrchestratorHostDetails(ctx, &models.OrchestratorHost{ID: "x"}))},
		{"timestamp", cold.UpdateOrchestratorHostTimestamp(ctx, "x", "y", "")},
		{"ws", errOf(cold.UpdateOrchestratorHostWebsocketStatus(ctx, "x", true))},
		{"res-write", cold.UpdateOrchestratorHostResources(ctx, "x", nil)},
		{"cache-write", cold.UpdateOrchestratorHostCacheConfig(ctx, "x", nil)},
	}
	for _, c := range checks {
		if c.err != interfaces.ErrDatabaseNotConnected {
			t.Errorf("%s guard: %v", c.name, c.err)
		}
	}
}

func stripForDiff(o *models.OrchestratorHost) models.OrchestratorHost {
	c := *o
	c.ID, c.CreatedAt, c.UpdatedAt = "", "", ""
	c.HealthCheck = nil
	c.VirtualMachines = nil
	c.ReverseProxy, c.ReverseProxyHosts = nil, nil
	return c
}
