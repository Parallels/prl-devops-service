package parity

// #425 task 03 — infra wave: reverse proxy (legacy fixtures from
// src/data/reverse_proxy_test.go, transcribed 1:1). The domain has NO authz
// layer (suite_reverse_proxy_test.go:23), so every scenario runs under
// authCtxPlain(). Includes the route-reorder invariant (reverse_proxy:342
// family) exercised through DatabaseService methods ONLY — engine-agnostic
// by construction — and the order-compacting delete.

import (
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParityReverseProxyInfraWave(t *testing.T) {
	testutil.RunOnEngines(t, reverseProxyScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
}

func rpRoute(id, path string) models.ReverseProxyHostHttpRoute {
	return models.ReverseProxyHostHttpRoute{
		ID: id, Path: path, TargetHost: "backend.local", TargetPort: "80",
	}
}

func reverseProxyScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "rp/config_enable_disable", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			_, err := db.UpdateReverseProxy(ctx, models.ReverseProxy{})
			require.NoError(t, err)
			enabled, err := db.EnableProxyConfig(ctx)
			require.NoError(t, err)
			assert.True(t, enabled.Enabled)
			disabled, err := db.DisableProxyConfig(ctx)
			require.NoError(t, err)
			assert.False(t, disabled.Enabled)
		}},
		{Name: "rp/config_update_readback", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			updated, err := db.UpdateReverseProxy(ctx, models.ReverseProxy{
				Enabled: true, Host: "example.com", Port: "443",
			})
			require.NoError(t, err)
			assert.True(t, updated.Enabled)
			assert.Equal(t, "example.com", updated.Host)
			loaded, err := db.GetReverseProxyConfig(ctx)
			require.NoError(t, err)
			assert.Equal(t, "443", loaded.Port)
		}},
		{Name: "rp/host_create_update_tls_cors", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			created, err := db.CreateReverseProxyHost(ctx, models.ReverseProxyHost{
				ID: "infra-rp-host1", Host: "app.example.com", Port: "8080",
			})
			require.NoError(t, err)
			assert.Equal(t, "app.example.com", created.Host)
			updatedHost := *created
			updatedHost.Port = "8443"
			updated, err := db.UpdateReverseProxyHost(ctx, &updatedHost)
			require.NoError(t, err)
			assert.Equal(t, "8443", updated.Port)
			tlsed, err := db.ConfigureReverseProxyHostTls(ctx, created.ID, models.ReverseProxyHostTls{Enabled: true})
			require.NoError(t, err)
			require.NotNil(t, tlsed.Tls)
			assert.True(t, tlsed.Tls.Enabled)
			corsed, err := db.ConfigureReverseProxyHostCors(ctx, created.ID, models.ReverseProxyHostCors{Enabled: true})
			require.NoError(t, err)
			require.NotNil(t, corsed.Cors)
			assert.True(t, corsed.Cors.Enabled)
		}},
		{Name: "rp/tcp_route_on_clean_host", Run: func(t *testing.T, _ ctxT, db dbT) {
			// updateTcpRoute only allowed while the host carries NO http routes
			// (data/reverse_proxy.go guard order: WithHttpRouteHost BEFORE NotFound).
			// PINNED: Create* regenerates entity IDs on BOTH engines — all
			// subsequent calls use the returned IDs.
			ctx := authCtxPlain()
			createdHost, err := db.CreateReverseProxyHost(ctx, models.ReverseProxyHost{
				Host: "tcp.example.com", Port: "9000",
				TcpRoute: &models.ReverseProxyHostTcpRoute{TargetHost: "t.local", TargetPort: "9001"},
			})
			require.NoError(t, err)
			require.NotNil(t, createdHost.TcpRoute)
			tcpID := createdHost.TcpRoute.ID
			require.NotEmpty(t, tcpID, "tcp route id assigned at create")
			updated, err := db.UpdateReverseProxyHostTcpRoute(ctx, createdHost.ID, models.ReverseProxyHostTcpRoute{
				ID: tcpID, TargetHost: "t2.local", TargetPort: "9002",
			})
			require.NoError(t, err)
			assert.Equal(t, "9002", updated.TargetPort)
			loaded, err := db.GetReverseProxyHost(ctx, createdHost.ID)
			require.NoError(t, err)
			require.NotNil(t, loaded.TcpRoute)
			assert.Equal(t, "t2.local", loaded.TcpRoute.TargetHost)
		}},
		{Name: "rp/routes_reorder_and_compacting_delete", Run: func(t *testing.T, _ ctxT, db dbT) {
			// Mirror of reverse_proxy_test.go:185-248 (first route bumped to
			// last, then the middle row deleted ⇒ orders compact to {1,2},
			// no gaps). Route IDs are engine-assigned (PINNED) — captured.
			ctx := authCtxPlain()
			host, err := db.CreateReverseProxyHost(ctx, models.ReverseProxyHost{
				Host: "routes.example.com", Port: "80",
			})
			require.NoError(t, err)
			rid := make([]string, 0, 3)
			for _, r := range []models.ReverseProxyHostHttpRoute{rpRoute("r1", "/one"), rpRoute("r2", "/two"), rpRoute("r3", "/three")} {
				created, err := db.CreateReverseProxyHostHttpRoute(ctx, host.ID, r)
				require.NoError(t, err)
				rid = append(rid, created.ID)
			}
			reordered, err := db.ReorderReverseProxyHostHttpRoute(ctx, host.ID, rid[0], 3)
			require.NoError(t, err)
			assert.Equal(t, map[string]int{rid[0]: 3, rid[1]: 1, rid[2]: 2}, routeOrders(reordered))
			require.NoError(t, db.DeleteReverseProxyHostHttpRoute(ctx, host.ID, rid[2]))
			loaded, err := db.GetReverseProxyHost(ctx, host.ID)
			require.NoError(t, err)
			require.Len(t, loaded.HttpRoutes, 2)
			assert.Equal(t, map[string]int{rid[0]: 2, rid[1]: 1}, routeOrders(loaded), "orders compact, no gap")
		}},
		{Name: "rp/reorder_concurrent_invariant", Run: func(t *testing.T, _ ctxT, db dbT) {
			// #421 TestReorderInvariant_Concurrent, DatabaseService-methods-only:
			// concurrent reorders must still land on a VALID permutation — no
			// duplicate/zero orders, no gaps.
			ctx := authCtxPlain()
			host, err := db.CreateReverseProxyHost(ctx, models.ReverseProxyHost{
				Host: "conc.example.com", Port: "80",
			})
			require.NoError(t, err)
			rid := make([]string, 0, 3) // PINNED: engine-assigned ids
			for _, r := range []models.ReverseProxyHostHttpRoute{rpRoute("c1", "/a"), rpRoute("c2", "/b"), rpRoute("c3", "/c")} {
				created, err := db.CreateReverseProxyHostHttpRoute(ctx, host.ID, r)
				require.NoError(t, err)
				rid = append(rid, created.ID)
			}
			var wg sync.WaitGroup
			if _, isLegacy := db.(*data.JsonDatabase); isLegacy {
				// FINDING (T15 candidate): legacy ReorderReverseProxyHost-
				// HttpRoute mutates the SHARED route slice outside dataMutex
				// (normalizeReverseProxyHttpRoutesOrder, reverse_proxy.go:25
				// via :357) — concurrent invocation is a data race BY
				// CONSTRUCTION under -race. #421's concurrent invariant was
				// SQL-only, so the JSON path was never hammered. Running the
				// invariant SEQUENTIALLY on legacy; the race is a production
				// finding (fix scope = JSON engine, #426 territory), not a
				// test this suite may enable.
				for i := range rid {
					_, err := db.ReorderReverseProxyHostHttpRoute(ctx, host.ID, rid[i], i+1)
					require.NoError(t, err)
				}
			} else {
				for i := range rid {
					wg.Add(1)
					go func(routeID string, pos int) {
						defer wg.Done()
						_, err := db.ReorderReverseProxyHostHttpRoute(ctx, host.ID, routeID, pos)
						assert.NoError(t, err) // invariant must not rely on swallowed errors
					}(rid[i], i+1)
				}
			}
			wg.Wait()
			loaded, err := db.GetReverseProxyHost(ctx, host.ID)
			require.NoError(t, err)
			require.Len(t, loaded.HttpRoutes, 3)
			orders := routeOrders(loaded)
			assert.ElementsMatch(t, []int{1, 2, 3}, []int{orders[rid[0]], orders[rid[1]], orders[rid[2]]},
				"final orders always a valid permutation")
		}},
		{Name: "rp/error_edges", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			host, err := db.CreateReverseProxyHost(ctx, models.ReverseProxyHost{
				Host: "err.example.com", Port: "80",
			})
			require.NoError(t, err)
			route, err := db.CreateReverseProxyHostHttpRoute(ctx, host.ID, rpRoute("e1", "/x"))
			require.NoError(t, err)
			_, err = db.ReorderReverseProxyHostHttpRoute(ctx, host.ID, "no-such-route", 1)
			require.Error(t, err)
			assert.Equal(t, data.ErrorReverseProxyHttpRouteNotFound, err)
			_, err = db.ReorderReverseProxyHostHttpRoute(ctx, "no-such-host", route.ID, 1)
			require.Error(t, err)
			assert.Equal(t, data.ErrorReverseProxyHostNotFound, err)
			// order changes go through Reorder only (D-4):
			_, err = db.UpdateReverseProxyHostHttpRoute(ctx, host.ID, models.ReverseProxyHostHttpRoute{
				ID: route.ID, Path: "/x", TargetHost: "backend.local", TargetPort: "80", Order: 5,
			})
			require.Error(t, err)
			assert.Equal(t, data.ErrorReverseProxyHttpRouteOrderUpdate, err)
		}},
	}
}

func routeOrders(h *models.ReverseProxyHost) map[string]int {
	out := map[string]int{}
	for _, r := range h.HttpRoutes {
		out[r.ID] = r.Order
	}
	return out
}
