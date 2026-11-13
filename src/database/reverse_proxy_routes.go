package database

// Embedded-route methods of the reverse-proxy domain (PBI #421 task 12
// split, pre-authorised in the task file: main file crossed the 200-line
// projection). Read-modify-write of the http_routes / tcp_route
// serialised-JSON columns; the four tx sites of D-9 live here — reorder
// rewrites EVERY route order inside one Transaction so concurrent
// reorders serialise (concurrency stress pinned in tests).

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

func rpNormalizeRouteOrder(routes []*models.ReverseProxyHostHttpRoute) {
	for i, route := range routes {
		route.Order = i + 1
	}
}

func rpRoutesJSON(routes []*models.ReverseProxyHostHttpRoute) (map[string]interface{}, error) {
	b, err := json.Marshal(routes)
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return map[string]interface{}{"http_routes": string(b)}, nil
}

func (g *GormDatabase) CreateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rpHost, err := g.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}
	route.ID = helpers.GenerateId()
	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		var stored []models.ReverseProxyHost
		if err := tx.Table("reverse_proxy_hosts").Where("id = ?", rpHost.ID).Find(&stored).Error; err != nil {
			return mapGormError(err, nil, nil)
		}
		if len(stored) == 0 {
			return nil // legacy: no host row found ⇒ route returned unstored, no error (:275)
		}
		rpNormalizeRouteOrder(stored[0].HttpRoutes)
		route.Order = len(stored[0].HttpRoutes) + 1
		stored[0].HttpRoutes = append(stored[0].HttpRoutes, &route)
		set, err := rpRoutesJSON(stored[0].HttpRoutes)
		if err != nil {
			return err
		}
		return mapGormError(tx.Table("reverse_proxy_hosts").Where("id = ?", rpHost.ID).Updates(set).Error, nil, nil)
	})
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (g *GormDatabase) DeleteReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	rpHost, err := g.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return err
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		routes := rpHost.HttpRoutes
		routeDeleted := false
		for i, r := range routes {
			if strings.EqualFold(r.ID, routeId) || strings.EqualFold(r.GetRoute(), routeId) {
				routes = append(append([]*models.ReverseProxyHostHttpRoute(nil), routes[:i]...), routes[i+1:]...)
				routeDeleted = true
				break
			}
		}
		if !routeDeleted {
			return interfaces.ErrorReverseProxyHttpRouteNotFound
		}
		rpNormalizeRouteOrder(routes)
		set, err := rpRoutesJSON(routes)
		if err != nil {
			return err
		}
		return mapGormError(tx.Table("reverse_proxy_hosts").Where("id = ?", rpHost.ID).Updates(set).Error, nil, nil)
	})
}

func (g *GormDatabase) UpdateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rpHost, err := g.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}
	var out *models.ReverseProxyHostHttpRoute
	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		for i, r := range rpHost.HttpRoutes {
			// Matching on GetRoute() equality, NOT the incoming id (:324).
			if strings.EqualFold(r.ID, route.ID) || strings.EqualFold(r.GetRoute(), route.GetRoute()) {
				if route.Order != 0 && route.Order != r.Order {
					return interfaces.ErrorReverseProxyHttpRouteOrderUpdate
				}
				route.Order = r.Order
				if r.Diff(route) {
					rpHost.HttpRoutes[i] = &route
					rpNormalizeRouteOrder(rpHost.HttpRoutes)
					set, err := rpRoutesJSON(rpHost.HttpRoutes)
					if err != nil {
						return err
					}
					if err := mapGormError(tx.Table("reverse_proxy_hosts").Where("id = ?", rpHost.ID).Updates(set).Error, nil, nil); err != nil {
						return err
					}
					out = &route
					return nil
				}
				out = r
				return nil
			}
		}
		return interfaces.ErrorReverseProxyHostNotFound // route miss answers the HOST sentinel (:340)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Reorder is the invariant call site: the whole permutation + renumber is
// ONE update of the serialised column inside one tx, so concurrent
// reorders serialise instead of interleaving half-written orders. (The
// Postgres clause.Locking{Strength:"UPDATE"} hardening belongs to task 18
// alongside the pg driver wiring; SQLite serialises writers anyway.)
func (g *GormDatabase) ReorderReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string, newOrder int) (*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rpHost, err := g.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}
	if newOrder < 1 || newOrder > len(rpHost.HttpRoutes) {
		return nil, interfaces.ErrorReverseProxyHttpRouteInvalidOrder
	}
	rpNormalizeRouteOrder(rpHost.HttpRoutes)
	currentIndex := -1
	for i, r := range rpHost.HttpRoutes {
		if strings.EqualFold(r.ID, routeId) {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		return nil, interfaces.ErrorReverseProxyHttpRouteNotFound
	}
	targetIndex := newOrder - 1
	if currentIndex == targetIndex {
		return rpHost, nil
	}
	route := rpHost.HttpRoutes[currentIndex]
	rpHost.HttpRoutes = append(rpHost.HttpRoutes[:currentIndex], rpHost.HttpRoutes[currentIndex+1:]...)
	if targetIndex >= len(rpHost.HttpRoutes) {
		rpHost.HttpRoutes = append(rpHost.HttpRoutes, route)
	} else {
		rpHost.HttpRoutes = append(rpHost.HttpRoutes[:targetIndex], append([]*models.ReverseProxyHostHttpRoute{route}, rpHost.HttpRoutes[targetIndex:]...)...)
	}
	rpNormalizeRouteOrder(rpHost.HttpRoutes)
	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		set, err := rpRoutesJSON(rpHost.HttpRoutes)
		if err != nil {
			return err
		}
		return mapGormError(tx.Table("reverse_proxy_hosts").Where("id = ?", rpHost.ID).Updates(set).Error, nil, nil)
	})
	if err != nil {
		return nil, err
	}
	return g.rpHostByID(rpHost.ID)
}

func (g *GormDatabase) UpdateReverseProxyHostTcpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostTcpRoute) (*models.ReverseProxyHostTcpRoute, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rpHost, err := g.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}
	if len(rpHost.HttpRoutes) > 0 {
		return nil, interfaces.ErrorReverseProxyTcpRouteWithHttpRouteHost
	}
	if rpHost.TcpRoute == nil {
		return nil, interfaces.ErrorReverseProxyTcpRouteNotFound
	}
	if rpHost.TcpRoute.Diff(route) {
		b, err := json.Marshal(route)
		if err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		if err := mapGormError(g.gormDB.Table("reverse_proxy_hosts").Where("id = ?", rpHost.ID).
			Update("tcp_route", string(b)).Error, nil, nil); err != nil {
			return nil, err
		}
		return &route, nil
	}
	// No-diff falls off the end of legacy's if-block into the host-not-found
	// sentinel (:396-424 — mirrored, however odd it reads).
	return nil, interfaces.ErrorReverseProxyHostNotFound
}
