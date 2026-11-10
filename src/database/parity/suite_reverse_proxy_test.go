package parity

// DomainSuite "reverse-proxy" (PBI #421 task 12): config singleton + hosts
// + embedded serialised-JSON routes. Route ids and host ids are
// engine-generated → hosts key by Name, routes by Path (unique per host in
// this fixture); the dup-create 400 message embeds a generated id, so it is
// normalised at " already exists with ID ". The concurrent reorder stress
// lives in the SQL unit test ONLY: the task file's premise that legacy
// serialises reorders behind dataMutex.Lock is STALE — reverse_proxy.go
// takes no locks at all (checked :1-425), so a legacy-side two-goroutine
// reorder is genuinely racy and cannot be asserted. Sequential reorder
// parity (identical order on both engines) is asserted here.

import (
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func rpCtx() basecontext.ApiContext { return basecontext.NewBaseContext() } // domain has NO authz

func seedRP(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := rpCtx()
	if _, err := db.UpdateReverseProxy(ctx, models.ReverseProxy{Enabled: true, Host: "proxy.local", Port: "443"}); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	mk := func(name, host, port string, routes []*models.ReverseProxyHostHttpRoute, tls *models.ReverseProxyHostTls, tcp *models.ReverseProxyHostTcpRoute) {
		if _, err := db.CreateReverseProxyHost(ctx, models.ReverseProxyHost{
			Name: name, Host: host, Port: port, HttpRoutes: routes, Tls: tls, TcpRoute: tcp,
		}); err != nil {
			t.Fatalf("seed host %s: %v", name, err)
		}
	}
	ordered := func(paths ...string) []*models.ReverseProxyHostHttpRoute {
		var out []*models.ReverseProxyHostHttpRoute
		for i, p := range paths {
			out = append(out, &models.ReverseProxyHostHttpRoute{Order: i + 1, Path: p})
		}
		return out
	}
	mk("web", "web.local", "80", ordered("/a", "/b"), &models.ReverseProxyHostTls{Enabled: true, Cert: "c"}, nil)
	mk("api", "api.local", "8080", ordered("/x", "/y", "/z"), nil, nil)
	mk("tcp", "tcp.local", "", nil, nil, &models.ReverseProxyHostTcpRoute{TargetPort: "22"})
	mk("bare", "bare.local", "80", nil, nil, nil)
}

func normaliseHosts(in []models.ReverseProxyHost) []models.ReverseProxyHost {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.ReverseProxyHost(nil), in...)
	for i := range out {
		out[i].ID = strings.ToUpper(out[i].Name) // key, no collision with real ids
		for _, r := range out[i].HttpRoutes {
			r.ID = "route:" + r.Path
		}
		if out[i].TcpRoute != nil {
			out[i].TcpRoute.ID = "tcp:" + out[i].Name
		}
	}
	return out
}

func normaliseHost(h *models.ReverseProxyHost) models.ReverseProxyHost {
	if h == nil {
		return models.ReverseProxyHost{}
	}
	return normaliseHosts([]models.ReverseProxyHost{*h})[0]
}

func normIDMsg(err error) string { // dup-create embeds a per-engine id
	if err == nil {
		return "<nil>"
	}
	s := err.Error()
	if i := strings.Index(s, " already exists with ID "); i >= 0 {
		return s[:i+len(" already exists with ID ")] + "<gen>"
	}
	return s
}

func TestParityDomain_ReverseProxy(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "reverse-proxy",
		Seed: seedRP,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkRPConfig,
			checkRPHostLists,
			checkRPMutations,
		},
	})
}

func checkRPConfig(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := rpCtx()
	lc, lerr := legacy.GetReverseProxyConfig(ctx)
	sc, serr := sql.GetReverseProxyConfig(ctx)
	assertSameError(t, "GetReverseProxyConfig", lerr, serr)
	if (lc == nil) != (sc == nil) {
		t.Fatalf("config nilness diverges: %+v vs %+v", lc, sc)
	}
	l, s := *lc, *sc
	l.ID, s.ID = "", "" // singleton id is engine-minted
	AssertSameList(t, "config", []models.ReverseProxy{l}, []models.ReverseProxy{s})
	// Field-scoped flips, same order on both.
	for _, fn := range []func(db datainterfaces.DatabaseService) (*models.ReverseProxy, error){
		func(db datainterfaces.DatabaseService) (*models.ReverseProxy, error) {
			return db.DisableProxyConfig(ctx)
		},
		func(db datainterfaces.DatabaseService) (*models.ReverseProxy, error) {
			return db.EnableProxyConfig(ctx)
		},
	} {
		ld, lerr := fn(legacy)
		sd, serr := fn(sql)
		assertSameError(t, "flip", lerr, serr)
		ld.ID, sd.ID = "", ""
		AssertSameList(t, "flip view", []models.ReverseProxy{*ld}, []models.ReverseProxy{*sd})
	}
}

func checkRPHostLists(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := rpCtx()
	for _, f := range []string{"", "name=web", "host=api.local", "port=80", "id=nope", "name=WEB", "host=tcp.local"} {
		lt, lerr := legacy.GetReverseProxyHosts(ctx, f)
		st, serr := sql.GetReverseProxyHosts(ctx, f)
		assertSameError(t, "GetReverseProxyHosts["+f+"]", lerr, serr)
		AssertSameList(t, "GetReverseProxyHosts["+f+"]", normaliseHosts(lt), normaliseHosts(st))
	}
	for _, probe := range []string{"WEB", "api.local:8080", "API.LOCAL:8080", "nope"} {
		lh, lerr := legacy.GetReverseProxyHost(ctx, probe)
		sh, serr := sql.GetReverseProxyHost(ctx, probe)
		assertSameError(t, "GetReverseProxyHost["+probe+"]", lerr, serr)
		AssertSameList(t, "GetReverseProxyHost["+probe+"]", []models.ReverseProxyHost{normaliseHost(lh)}, []models.ReverseProxyHost{normaliseHost(sh)})
	}
}

func checkRPMutations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := rpCtx()
	// run applies the same closure to both engines and compares outcomes.
	run := func(what string, call func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error)) {
		lh, lerr := call(legacy)
		sh, serr := call(sql)
		assertSameError(t, what, lerr, serr)
		AssertSameList(t, what, []models.ReverseProxyHost{normaliseHost(lh)}, []models.ReverseProxyHost{normaliseHost(sh)})
	}
	web := func(db datainterfaces.DatabaseService) string {
		h, err := db.GetReverseProxyHost(ctx, "web.local:80")
		if err != nil {
			t.Fatalf("web lookup: %v", err)
		}
		return h.ID
	}
	api := func(db datainterfaces.DatabaseService) string {
		h, err := db.GetReverseProxyHost(ctx, "api.local:8080")
		if err != nil {
			t.Fatalf("api lookup: %v", err)
		}
		return h.ID
	}

	// Partial host update (name only).
	run("UpdateReverseProxyHost", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		id := web(db)
		return db.UpdateReverseProxyHost(ctx, &models.ReverseProxyHost{ID: id, Name: "web2"})
	})
	// TLS lost-write: view carries, store does not.
	run("ConfigureTLS", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		return db.ConfigureReverseProxyHostTls(ctx, web(db), models.ReverseProxyHostTls{Enabled: true, Cert: "new"})
	})
	run("ReloadAfterTLS", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		return db.GetReverseProxyHost(ctx, web(db))
	})
	// New host with route.
	run("CreateHost", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		return db.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Name: "extra", Host: "extra.local", Port: "90",
			HttpRoutes: []*models.ReverseProxyHostHttpRoute{{Path: "/e", Order: 1}}})
	})
	_, le := legacy.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Name: "extra", Host: "extra.local", Port: "90"})
	_, se := sql.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Name: "extra", Host: "extra.local", Port: "90"})
	if normIDMsg(le) != normIDMsg(se) {
		t.Errorf("dup-create message: %q vs %q", normIDMsg(le), normIDMsg(se))
	}
	// Route append + path-match update.
	run("CreateRoute", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		_, err := db.CreateReverseProxyHostHttpRoute(ctx, api(db), models.ReverseProxyHostHttpRoute{Path: "/w"})
		if err != nil {
			return nil, err
		}
		return db.GetReverseProxyHost(ctx, api(db))
	})
	run("UpdateRoute", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		_, err := db.UpdateReverseProxyHostHttpRoute(ctx, api(db), models.ReverseProxyHostHttpRoute{Path: "/y", TargetHost: "vm1"})
		if err != nil {
			return nil, err
		}
		return db.GetReverseProxyHost(ctx, api(db))
	})
	_, le = legacy.UpdateReverseProxyHostHttpRoute(ctx, api(legacy), models.ReverseProxyHostHttpRoute{Path: "/y", Order: 9})
	_, se = sql.UpdateReverseProxyHostHttpRoute(ctx, api(sql), models.ReverseProxyHostHttpRoute{Path: "/y", Order: 9})
	assertSameError(t, "UpdateRouteOrderGuard", le, se)
	// Sequential reorder (concurrent stress is SQL-only: legacy is lock-free).
	routeID := func(db datainterfaces.DatabaseService, hostProbe, path string) string {
		h, err := db.GetReverseProxyHost(ctx, hostProbe)
		if err != nil {
			t.Fatalf("route lookup %s: %v", hostProbe, err)
		}
		for _, r := range h.HttpRoutes {
			if r.Path == path {
				return r.ID
			}
		}
		t.Fatalf("route %s on %s missing", path, hostProbe)
		return ""
	}
	run("Reorder", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		return db.ReorderReverseProxyHostHttpRoute(ctx, api(db), routeID(db, api(db), "/z"), 1)
	})
	_, le = legacy.ReorderReverseProxyHostHttpRoute(ctx, api(legacy), routeID(legacy, api(legacy), "/z"), 99)
	_, se = sql.ReorderReverseProxyHostHttpRoute(ctx, api(sql), routeID(sql, api(sql), "/z"), 99)
	assertSameError(t, "ReorderBounds", le, se)
	run("DeleteRoute", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		if err := db.DeleteReverseProxyHostHttpRoute(ctx, api(db), routeID(db, api(db), "/x")); err != nil {
			return nil, err
		}
		return db.GetReverseProxyHost(ctx, api(db))
	})
	// TCP: change lands; no-diff answers host-not-found.
	tcpProbe := func(db datainterfaces.DatabaseService) string {
		h, err := db.GetReverseProxyHost(ctx, "tcp.local")
		if err != nil {
			t.Fatalf("tcp host lookup: %v", err)
		}
		return h.ID
	}
	run("UpdateTcp", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		id := tcpProbe(db)
		if _, err := db.UpdateReverseProxyHostTcpRoute(ctx, id, models.ReverseProxyHostTcpRoute{TargetPort: "2222"}); err != nil {
			return nil, err
		}
		return db.GetReverseProxyHost(ctx, id)
	})
	_, le = legacy.UpdateReverseProxyHostTcpRoute(ctx, tcpProbe(legacy), models.ReverseProxyHostTcpRoute{TargetPort: "2222"})
	_, se = sql.UpdateReverseProxyHostTcpRoute(ctx, tcpProbe(sql), models.ReverseProxyHostTcpRoute{TargetPort: "2222"})
	assertSameError(t, "UpdateTcpNoDiff", le, se)
	// Delete host (by id, folded match) + final list.
	run("DeleteHost", func(db datainterfaces.DatabaseService) (*models.ReverseProxyHost, error) {
		if err := db.DeleteReverseProxyHost(ctx, tcpProbe(db)); err != nil {
			return nil, err
		}
		return &models.ReverseProxyHost{}, nil
	})
	lt, lerr := legacy.GetReverseProxyHosts(ctx, "")
	st, serr := sql.GetReverseProxyHosts(ctx, "")
	assertSameError(t, "final hosts", lerr, serr)
	AssertSameList(t, "final hosts", normaliseHosts(lt), normaliseHosts(st))
}
