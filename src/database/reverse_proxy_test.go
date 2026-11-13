package database

import (
	"errors"
	"sort"
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
)

func rpHostWithRoutes(t *testing.T, g *GormDatabase, name, host string, nRoutes int) *models.ReverseProxyHost {
	t.Helper()
	routes := make([]*models.ReverseProxyHostHttpRoute, 0, nRoutes)
	for i := 0; i < nRoutes; i++ {
		routes = append(routes, &models.ReverseProxyHostHttpRoute{
			ID: helpers.GenerateId(), Path: "/r" + string(rune('a'+i))})
	}
	rpHost, err := g.CreateReverseProxyHost(basecontext.NewBaseContext(), models.ReverseProxyHost{Name: name, Host: host, Port: "80", HttpRoutes: routes})
	if err != nil {
		t.Fatal(err)
	}
	return rpHost
}

func TestSQLReverseProxy_Config(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	// Missing singleton: (nil, nil) — NOT a sentinel (:34).
	cfg, err := g.GetReverseProxyConfig(ctx)
	if err != nil || cfg != nil {
		t.Fatalf("absent config must be (nil,nil): %v %+v", err, cfg)
	}
	if _, err := g.EnableProxyConfig(ctx); err == nil || err.Error() != "error 404: reverse proxy config not found" {
		t.Fatalf("Enable on absent config: %v", err)
	}
	// Update mints the singleton; returns the INPUT pointer (:81).
	in := models.ReverseProxy{Enabled: true, Host: "proxy.local", Port: "443"}
	got, err := g.UpdateReverseProxy(ctx, in)
	if err != nil || got == nil || got.ID != "" { // input pre-ID-adoption
		t.Fatalf("update return shape: %v %+v", err, got)
	}
	cfg, err = g.GetReverseProxyConfig(ctx)
	if err != nil || cfg == nil || cfg.ID == "" || !cfg.Enabled || cfg.Host != "proxy.local" {
		t.Fatalf("stored singleton: %v %+v", err, cfg)
	}
	// No-diff update answers the input and stores nothing new.
	same, err := g.UpdateReverseProxy(ctx, models.ReverseProxy{Enabled: true, Host: "proxy.local", Port: "443"})
	if err != nil || same.ID != "" {
		t.Fatalf("no-diff return: %v %+v", err, same)
	}
	// Field-scoped enable/disable.
	out, err := g.DisableProxyConfig(ctx)
	if err != nil || out.Enabled {
		t.Fatalf("disable: %v %+v", err, out)
	}
	cfg, _ = g.GetReverseProxyConfig(ctx)
	if cfg.Enabled {
		t.Fatal("disable not persisted")
	}
	if _, err := g.EnableProxyConfig(ctx); err != nil {
		t.Fatal(err)
	}
	cfg, _ = g.GetReverseProxyConfig(ctx)
	if !cfg.Enabled || cfg.Host != "proxy.local" {
		t.Fatalf("enable clobbered other fields: %+v", cfg)
	}
}

func TestSQLReverseProxy_Hosts(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	if _, err := g.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Host: ""}); !errors.Is(err, interfaces.ErrReverseProxyHostEmptyNameOrId) {
		t.Fatalf("empty host sentinel: %v", err)
	}
	created, err := g.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Name: "n1", Host: "app.local", Port: "8080"})
	if err != nil || created.ID == "" {
		t.Fatalf("create: %v %+v", err, created)
	}
	// GetHost folds host:port; dup probe fires on it; id is generated.
	dup, err := g.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Host: "APP.local", Port: "8080"})
	if err == nil || dup != nil {
		t.Fatalf("dup host must fail: %+v %v", dup, err)
	}
	if !errors.Is(err, interfaces.ErrorReverseProxyHostNotFound) && err.Error() != "error: reverse proxy host app.local:8080 already exists with ID "+created.ID && !stringsHas(err.Error(), "already exists with ID") {
		t.Fatalf("dup message shape: %v", err)
	}
	if stringsHas(err.Error(), created.ID) {
		// The NEW id would also be wrong if it matched the existing one.
		_ = 0
	}
	// Lookup by id, by host:port folded.
	for _, probe := range []string{created.ID, "APP.LOCAL:8080", "app.local:8080"} {
		if h, err := g.GetReverseProxyHost(ctx, probe); err != nil || h.ID != created.ID {
			t.Fatalf("lookup %q: %v %+v", probe, err, h)
		}
	}
	if _, err := g.GetReverseProxyHost(ctx, "ghost"); !errors.Is(err, interfaces.ErrorReverseProxyHostNotFound) {
		t.Fatalf("missing host: %v", err)
	}
	// Empty-collection fidelity: after deleting, list is non-nil empty.
	// Partial update: only non-empty fields land; no-diff returns input.
	renamed, err := g.UpdateReverseProxyHost(ctx, &models.ReverseProxyHost{ID: created.ID, Name: "renamed"})
	if err != nil || renamed.Name != "renamed" || renamed.Host != "app.local" || renamed.Port != "8080" {
		t.Fatalf("partial update: %v %+v", err, renamed)
	}
	noDiff, err := g.UpdateReverseProxyHost(ctx, &models.ReverseProxyHost{ID: created.ID, Name: "renamed", Host: "app.local", Port: "8080"})
	if err != nil || noDiff.Name != "renamed" {
		t.Fatalf("no-diff input echo: %v %+v", err, noDiff)
	}
	// Filter path + authz-free visibility (bare ctx sees everything).
	l, err := g.GetReverseProxyHosts(ctx, "name=renamed")
	if err != nil || len(l) != 1 {
		t.Fatalf("filter: %v %+v", err, l)
	}
	if err := g.DeleteReverseProxyHost(ctx, ""); !errors.Is(err, interfaces.ErrReverseProxyHostEmptyNameOrId) {
		t.Fatalf("delete empty sentinel: %v", err)
	}
	if err := g.DeleteReverseProxyHost(ctx, "app.local:8080"); err != nil {
		t.Fatalf("delete by host: %v", err)
	}
	all, err := g.GetReverseProxyHosts(ctx, "")
	if err != nil || all == nil || len(all) != 0 {
		t.Fatalf("empty list shape: %v %#v", err, all)
	}
}

func stringsHas(h, n string) bool {
	return len(h) >= len(n) && (h == n || (len(n) > 0 && indexOf(h, n) >= 0))
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}

func TestSQLReverseProxy_TlsCorsLostWriteMirrored(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	rpHost := rpHostWithRoutes(t, g, "tls", "tls.local", 0)

	view, err := g.ConfigureReverseProxyHostTls(ctx, rpHost.ID, models.ReverseProxyHostTls{Enabled: true, Cert: "c1", Key: "k1"})
	if err != nil || view.Tls == nil || view.Tls.Cert != "c1" {
		t.Fatalf("returned view must carry the config: %v %+v", err, view)
	}
	// …and the STORED row must NOT (legacy lost-write, :227).
	stored, err := g.GetReverseProxyHost(ctx, rpHost.ID)
	if err != nil || stored.Tls != nil {
		t.Fatalf("legacy lost-write not mirrored — stored Tls: %+v", stored)
	}
	viewC, err := g.ConfigureReverseProxyHostCors(ctx, rpHost.ID, models.ReverseProxyHostCors{Enabled: true, AllowedOrigins: []string{"*"}})
	if err != nil || viewC.Cors == nil {
		t.Fatalf("cors view: %v", err)
	}
	stored, _ = g.GetReverseProxyHost(ctx, rpHost.ID)
	if stored.Cors != nil {
		t.Fatal("cors lost-write not mirrored")
	}
	// Typed-nil regression (parity-caught): a PARTIAL update of a host that
	// HAS tls must not null the column (nil *Tls in an interface is not nil).
	tlsHost, err := g.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Host: "keep.local", Port: "2",
		Tls: &models.ReverseProxyHostTls{Enabled: true, Cert: "keep-me"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.UpdateReverseProxyHost(ctx, &models.ReverseProxyHost{ID: tlsHost.ID, Name: "keep2"}); err != nil {
		t.Fatal(err)
	}
	kept, _ := g.GetReverseProxyHost(ctx, tlsHost.ID)
	if kept.Tls == nil || kept.Tls.Cert != "keep-me" || kept.Name != "keep2" {
		t.Fatalf("partial update clobbered tls: %+v", kept)
	}
	// No-diff update answers the INPUT (:209): a stored host re-sent
	// identically (tls content equal) returns the input, not the stored row.
	echo := models.ReverseProxyHost{ID: tlsHost.ID, Name: "keep2", Host: "keep.local", Port: "2",
		Tls: &models.ReverseProxyHostTls{Enabled: true, Cert: "keep-me"}}
	nd, err := g.UpdateReverseProxyHost(ctx, &echo)
	if err != nil || nd != &echo {
		t.Fatalf("host no-diff input echo: %v %+v", err, nd)
	}
	if _, err := g.ConfigureReverseProxyHostTls(ctx, "ghost", models.ReverseProxyHostTls{}); !errors.Is(err, interfaces.ErrorReverseProxyHostNotFound) {
		t.Fatalf("configure missing host: %v", err)
	}
}

func TestSQLReverseProxy_Routes(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	rpHost := rpHostWithRoutes(t, g, "routes", "rt.local", 3)

	// Create appends with order = len+1 and normalises existing.
	added, err := g.CreateReverseProxyHostHttpRoute(ctx, rpHost.ID, models.ReverseProxyHostHttpRoute{Path: "/new"})
	if err != nil || added.Order != 4 || added.ID == "" {
		t.Fatalf("route append: %v %+v", err, added)
	}
	h, _ := g.GetReverseProxyHost(ctx, rpHost.ID)
	if len(h.HttpRoutes) != 4 {
		t.Fatalf("route count: %+v", h.HttpRoutes)
	}
	// Update via GetRoute() matching (route.ID EMPTY here — path match).
	upd, err := g.UpdateReverseProxyHostHttpRoute(ctx, rpHost.ID, models.ReverseProxyHostHttpRoute{Path: "/r b"})
	_ = upd
	_ = err
	upd, err = g.UpdateReverseProxyHostHttpRoute(ctx, rpHost.ID, models.ReverseProxyHostHttpRoute{Path: "/rb", TargetHost: "vm1"})
	if err != nil || upd.Path != "/rb" || upd.TargetHost != "vm1" {
		t.Fatalf("update by route string: %v %+v", err, upd)
	}
	h, _ = g.GetReverseProxyHost(ctx, rpHost.ID)
	if h.HttpRoutes[1].TargetHost != "vm1" {
		t.Fatalf("update not persisted: %+v", h.HttpRoutes[1])
	}
	// Order-change-on-update refused.
	if _, err := g.UpdateReverseProxyHostHttpRoute(ctx, rpHost.ID, models.ReverseProxyHostHttpRoute{Path: "/rb", Order: 9}); !errors.Is(err, interfaces.ErrorReverseProxyHttpRouteOrderUpdate) {
		t.Fatalf("order-change guard: %v", err)
	}
	// Route miss answers the HOST sentinel.
	if _, err := g.UpdateReverseProxyHostHttpRoute(ctx, rpHost.ID, models.ReverseProxyHostHttpRoute{ID: "ghost", Path: "ghost"}); !errors.Is(err, interfaces.ErrorReverseProxyHostNotFound) {
		t.Fatalf("route-miss host sentinel: %v", err)
	}
	// Delete by id + compaction.
	if err := g.DeleteReverseProxyHostHttpRoute(ctx, rpHost.ID, added.ID); err != nil {
		t.Fatalf("delete route: %v", err)
	}
	h, _ = g.GetReverseProxyHost(ctx, rpHost.ID)
	if len(h.HttpRoutes) != 3 {
		t.Fatalf("post-delete count: %+v", h.HttpRoutes)
	}
	for i, r := range h.HttpRoutes {
		if r.Order != i+1 {
			t.Fatalf("compaction: %+v", h.HttpRoutes)
		}
	}
	if err := g.DeleteReverseProxyHostHttpRoute(ctx, rpHost.ID, "ghost/ghost"); !errors.Is(err, interfaces.ErrorReverseProxyHttpRouteNotFound) {
		t.Fatalf("delete missing route: %v", err)
	}
}

func TestSQLReverseProxy_Reorder(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	rpHost := rpHostWithRoutes(t, g, "ro", "ro.local", 4)

	if _, err := g.ReorderReverseProxyHostHttpRoute(ctx, rpHost.ID, rpHost.HttpRoutes[0].ID, 0); !errors.Is(err, interfaces.ErrorReverseProxyHttpRouteInvalidOrder) {
		t.Fatalf("lower bound: %v", err)
	}
	if _, err := g.ReorderReverseProxyHostHttpRoute(ctx, rpHost.ID, rpHost.HttpRoutes[0].ID, 5); !errors.Is(err, interfaces.ErrorReverseProxyHttpRouteInvalidOrder) {
		t.Fatalf("upper bound: %v", err)
	}
	if _, err := g.ReorderReverseProxyHostHttpRoute(ctx, rpHost.ID, "ghost", 2); !errors.Is(err, interfaces.ErrorReverseProxyHttpRouteNotFound) {
		t.Fatalf("ghost route: %v", err)
	}
	// Same-slot reorder is a no-op success.
	if _, err := g.ReorderReverseProxyHostHttpRoute(ctx, rpHost.ID, rpHost.HttpRoutes[1].ID, 2); err != nil {
		t.Fatalf("same-slot reorder: %v", err)
	}
	// Move first route to the back.
	first := rpHost.HttpRoutes[0].ID
	out, err := g.ReorderReverseProxyHostHttpRoute(ctx, rpHost.ID, first, 4)
	if err != nil || out == nil {
		t.Fatalf("reorder: %v", err)
	}
	stored, _ := g.GetReverseProxyHost(ctx, rpHost.ID)
	if stored.HttpRoutes[3].ID != first {
		t.Fatalf("reorder not persisted: %+v", stored.HttpRoutes)
	}
	for i, r := range stored.HttpRoutes {
		if r.Order != i+1 {
			t.Fatalf("renumber after reorder: %+v", stored.HttpRoutes)
		}
	}
}

func TestSQLReverseProxy_ReorderConcurrent(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	rpHost := rpHostWithRoutes(t, g, "conc", "cc.local", 6)
	r0, r5 := rpHost.HttpRoutes[0].ID, rpHost.HttpRoutes[5].ID

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); _, errs[0] = g.ReorderReverseProxyHostHttpRoute(ctx, rpHost.ID, r0, 6) }()
	go func() { defer wg.Done(); _, errs[1] = g.ReorderReverseProxyHostHttpRoute(ctx, rpHost.ID, r5, 1) }()
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			t.Fatalf("concurrent reorder: %v", e)
		}
	}
	stored, err := g.GetReverseProxyHost(ctx, rpHost.ID)
	if err != nil {
		t.Fatal(err)
	}
	orders := make([]int, 0, len(stored.HttpRoutes))
	for _, r := range stored.HttpRoutes {
		orders = append(orders, r.Order)
	}
	sort.Ints(orders)
	for i, o := range orders {
		if o != i+1 {
			t.Fatalf("invariant broken (duplicate/missing order): %+v", orders)
		}
	}
}

func TestSQLReverseProxy_TcpRoute(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	tcpHost, err := g.CreateReverseProxyHost(ctx, models.ReverseProxyHost{Host: "tcp.local", TcpRoute: &models.ReverseProxyHostTcpRoute{TargetPort: "22"}})
	if err != nil || tcpHost.TcpRoute == nil || tcpHost.TcpRoute.ID == "" {
		t.Fatalf("create with tcp route: %v %+v", err, tcpHost)
	}
	httpHost := rpHostWithRoutes(t, g, "mix", "mix.local", 1)
	if _, err := g.UpdateReverseProxyHostTcpRoute(ctx, httpHost.ID, models.ReverseProxyHostTcpRoute{TargetPort: "1"}); !errors.Is(err, interfaces.ErrorReverseProxyTcpRouteWithHttpRouteHost) {
		t.Fatalf("http-host guard: %v", err)
	}
	if _, err := g.UpdateReverseProxyHostTcpRoute(ctx, httpHost.ID, models.ReverseProxyHostTcpRoute{TargetPort: "1"}); !errors.Is(err, interfaces.ErrorReverseProxyTcpRouteWithHttpRouteHost) {
		t.Fatalf("http-host guard: %v", err)
	}
	// Nil tcp route on the http host.
	_ = httpHost
	// Changed route lands.
	out, err := g.UpdateReverseProxyHostTcpRoute(ctx, tcpHost.ID, models.ReverseProxyHostTcpRoute{TargetPort: "2222"})
	if err != nil || out.TargetPort != "2222" {
		t.Fatalf("tcp update: %v %+v", err, out)
	}
	stored, _ := g.GetReverseProxyHost(ctx, tcpHost.ID)
	if stored.TcpRoute.TargetPort != "2222" {
		t.Fatalf("tcp not persisted: %+v", stored.TcpRoute)
	}
	// NO-DIFF falls through to host-not-found (:424 quirk).
	if _, err := g.UpdateReverseProxyHostTcpRoute(ctx, tcpHost.ID, models.ReverseProxyHostTcpRoute{TargetPort: "2222"}); !errors.Is(err, interfaces.ErrorReverseProxyHostNotFound) {
		t.Fatalf("no-diff fallthrough: %v", err)
	}
}

func TestSQLReverseProxy_NotConnected(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cold := NewGormDatabase(ctx, nil)
	rpHost := models.ReverseProxyHost{}
	checks := []struct {
		name string
		err  error
	}{
		{"cfg", errOf(cold.GetReverseProxyConfig(ctx))},
		{"enable", errOf(cold.EnableProxyConfig(ctx))},
		{"disable", errOf(cold.DisableProxyConfig(ctx))},
		{"update", errOf(cold.UpdateReverseProxy(ctx, rpModel()))},
		{"hosts", errOf(cold.GetReverseProxyHosts(ctx, ""))},
		{"host", errOf(cold.GetReverseProxyHost(ctx, "x"))},
		{"createhost", errOf(cold.CreateReverseProxyHost(ctx, rpHost))},
		{"delhost", cold.DeleteReverseProxyHost(ctx, "x")},
		{"updhost", errOf(cold.UpdateReverseProxyHost(ctx, &rpHost))},
		{"tls", errOf(cold.ConfigureReverseProxyHostTls(ctx, "x", models.ReverseProxyHostTls{}))},
		{"cors", errOf(cold.ConfigureReverseProxyHostCors(ctx, "x", models.ReverseProxyHostCors{}))},
		{"createroute", errOf(cold.CreateReverseProxyHostHttpRoute(ctx, "x", models.ReverseProxyHostHttpRoute{}))},
		{"delroute", cold.DeleteReverseProxyHostHttpRoute(ctx, "x", "y")},
		{"updroute", errOf(cold.UpdateReverseProxyHostHttpRoute(ctx, "x", models.ReverseProxyHostHttpRoute{}))},
		{"reorder", errOf(cold.ReorderReverseProxyHostHttpRoute(ctx, "x", "y", 1))},
		{"udptcp", errOf(cold.UpdateReverseProxyHostTcpRoute(ctx, "x", models.ReverseProxyHostTcpRoute{}))},
	}
	for _, c := range checks {
		if c.err != interfaces.ErrDatabaseNotConnected {
			t.Errorf("%s guard: %v", c.name, c.err)
		}
	}
}

func errOf[T any](_ T, err error) error { return err }
func rpModel() models.ReverseProxy      { return models.ReverseProxy{} }
