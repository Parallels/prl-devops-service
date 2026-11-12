package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

var (
	ErrReverseProxyHostEmptyNameOrId           = interfaces.ErrReverseProxyHostEmptyNameOrId
	ErrorReverseProxyHostNotFound              = interfaces.ErrorReverseProxyHostNotFound
	ErrorReverseProxyHttpRouteNotFound         = interfaces.ErrorReverseProxyHttpRouteNotFound
	ErrorReverseProxyHttpRouteInvalidOrder     = interfaces.ErrorReverseProxyHttpRouteInvalidOrder
	ErrorReverseProxyHttpRouteOrderUpdate      = interfaces.ErrorReverseProxyHttpRouteOrderUpdate
	ErrorReverseProxyTcpRouteNotFound          = interfaces.ErrorReverseProxyTcpRouteNotFound
	ErrorReverseProxyTcpRouteWithHttpRouteHost = interfaces.ErrorReverseProxyTcpRouteWithHttpRouteHost
)

func normalizeReverseProxyHttpRoutesOrder(routes []*models.ReverseProxyHostHttpRoute) {
	for i, route := range routes {
		route.Order = i + 1
	}
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) GetReverseProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	return j.data.ReverseProxy, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) EnableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if j.data.ReverseProxy == nil {
		return nil, errors.NewWithCode("reverse proxy config not found", 404)
	}

	j.data.ReverseProxy.Enabled = true
	_ = j.SaveNow(ctx)

	return j.data.ReverseProxy, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) DisableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if j.data.ReverseProxy == nil {
		return nil, errors.NewWithCode("reverse proxy config not found", 404)
	}

	j.data.ReverseProxy.Enabled = false
	_ = j.SaveNow(ctx)

	return j.data.ReverseProxy, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) UpdateReverseProxy(ctx basecontext.ApiContext, rp models.ReverseProxy) (*models.ReverseProxy, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if j.data.ReverseProxy.Diff(rp) {
		rpCopy := rp
		if j.data.ReverseProxy != nil {
			rpCopy.ID = j.data.ReverseProxy.ID
		} else {
			rpCopy.ID = helpers.GenerateId()
		}
		j.data.ReverseProxy = &rpCopy
		_ = j.SaveNow(ctx)
		return &rp, nil
	}

	return &rp, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) GetReverseProxyHosts(ctx basecontext.ApiContext, filter string) ([]models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.ReverseProxyHosts, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) GetReverseProxyHost(ctx basecontext.ApiContext, idOrName string) (*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	rpHosts, err := j.GetReverseProxyHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, rpHost := range rpHosts {
		hostname := rpHost.GetHost()
		if strings.EqualFold(rpHost.ID, idOrName) || strings.EqualFold(hostname, idOrName) {
			return &rpHost, nil
		}
	}

	return nil, ErrorReverseProxyHostNotFound
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) CreateReverseProxyHost(ctx basecontext.ApiContext, rpHost models.ReverseProxyHost) (*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if rpHost.Host == "" {
		return nil, ErrReverseProxyHostEmptyNameOrId
	}

	rpHost.ID = helpers.GenerateId()

	if u, _ := j.GetReverseProxyHost(ctx, rpHost.GetHost()); u != nil {
		return nil, errors.NewWithCodef(400, "reverse proxy host %s already exists with ID %s", rpHost.GetHost(), rpHost.ID)
	}

	if rpHost.TcpRoute != nil {
		rpHost.TcpRoute.ID = helpers.GenerateId()
	}
	if rpHost.HttpRoutes != nil {
		for _, route := range rpHost.HttpRoutes {
			route.ID = helpers.GenerateId()
		}
	}

	j.data.ReverseProxyHosts = append(j.data.ReverseProxyHosts, rpHost)
	_ = j.SaveNow(ctx)

	return &rpHost, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) DeleteReverseProxyHost(ctx basecontext.ApiContext, idOrName string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if idOrName == "" {
		return ErrReverseProxyHostEmptyNameOrId
	}

	for i, rpHost := range j.data.ReverseProxyHosts {
		if strings.EqualFold(rpHost.ID, idOrName) || strings.EqualFold(rpHost.GetHost(), idOrName) {
			j.data.ReverseProxyHosts = append(j.data.ReverseProxyHosts[:i], j.data.ReverseProxyHosts[i+1:]...)
			_ = j.SaveNow(ctx)
			return nil
		}
	}

	return ErrorReverseProxyHostNotFound
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) UpdateReverseProxyHost(ctx basecontext.ApiContext, rpHost *models.ReverseProxyHost) (*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if rpHost.ID == "" {
		return nil, ErrReverseProxyHostEmptyNameOrId
	}

	for i, h := range j.data.ReverseProxyHosts {
		if strings.EqualFold(h.ID, rpHost.ID) ||
			strings.EqualFold(h.GetHost(), rpHost.ID) ||
			strings.EqualFold(h.GetHost(), rpHost.GetHost()) {
			if h.Diff(*rpHost) {
				if rpHost.Name != "" {
					j.data.ReverseProxyHosts[i].Name = rpHost.Name
				}
				if rpHost.Host != "" {
					j.data.ReverseProxyHosts[i].Host = rpHost.Host
				}
				if rpHost.Port != "" {
					j.data.ReverseProxyHosts[i].Port = rpHost.Port
				}
				if rpHost.Tls != nil {
					j.data.ReverseProxyHosts[i].Tls = rpHost.Tls
				}
				if rpHost.Cors != nil {
					j.data.ReverseProxyHosts[i].Cors = rpHost.Cors
				}

				_ = j.SaveNow(ctx)
				return &j.data.ReverseProxyHosts[i], nil
			}

			return rpHost, nil
		}
	}

	return nil, ErrorReverseProxyHostNotFound
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) ConfigureReverseProxyHostTls(ctx basecontext.ApiContext, rpHostId string, tlsConfig models.ReverseProxyHostTls) (*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	rpHost, err := j.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}

	if rpHost.Tls.Diff(tlsConfig) {
		rpHost.Tls = &tlsConfig
		_ = j.SaveNow(ctx)
		return rpHost, nil
	}

	return rpHost, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) ConfigureReverseProxyHostCors(ctx basecontext.ApiContext, rpHostId string, corsConfig models.ReverseProxyHostCors) (*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	rpHost, err := j.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}

	if rpHost.Cors.Diff(corsConfig) {
		rpHost.Cors = &corsConfig
		_ = j.SaveNow(ctx)
		return rpHost, nil
	}

	return rpHost, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) CreateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	rpHost, err := j.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}

	route.ID = helpers.GenerateId()
	for i, h := range j.data.ReverseProxyHosts {
		if h.ID == rpHost.ID {
			normalizeReverseProxyHttpRoutesOrder(j.data.ReverseProxyHosts[i].HttpRoutes)
			route.Order = len(j.data.ReverseProxyHosts[i].HttpRoutes) + 1
			j.data.ReverseProxyHosts[i].HttpRoutes = append(j.data.ReverseProxyHosts[i].HttpRoutes, &route)
			j.SaveNow(ctx)
			return &route, nil
		}
	}

	return &route, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) DeleteReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	rpHost, err := j.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return err
	}

	routeDeleted := false
	for i, r := range rpHost.HttpRoutes {
		if strings.EqualFold(r.ID, routeId) || strings.EqualFold(r.GetRoute(), routeId) {
			rpHost.HttpRoutes = append(rpHost.HttpRoutes[:i], rpHost.HttpRoutes[i+1:]...)
			routeDeleted = true
			break
		}
	}
	if !routeDeleted {
		return ErrorReverseProxyHttpRouteNotFound
	}

	normalizeReverseProxyHttpRoutesOrder(rpHost.HttpRoutes)

	for i, h := range j.data.ReverseProxyHosts {
		if h.ID == rpHost.ID {
			j.data.ReverseProxyHosts[i].HttpRoutes = rpHost.HttpRoutes
			_ = j.SaveNow(ctx)
			return nil
		}
	}

	return ErrorReverseProxyHttpRouteNotFound
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) UpdateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	rpHost, err := j.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}

	for i, r := range rpHost.HttpRoutes {
		if strings.EqualFold(r.ID, route.ID) || strings.EqualFold(r.GetRoute(), route.GetRoute()) {
			if route.Order != 0 && route.Order != r.Order {
				return nil, ErrorReverseProxyHttpRouteOrderUpdate
			}
			route.Order = r.Order
			if r.Diff(route) {
				rpHost.HttpRoutes[i] = &route
				normalizeReverseProxyHttpRoutesOrder(rpHost.HttpRoutes)
				_ = j.SaveNow(ctx)
				return rpHost.HttpRoutes[i], nil
			}

			return r, nil
		}
	}

	return nil, ErrorReverseProxyHostNotFound
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) ReorderReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string, newOrder int) (*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	rpHost, err := j.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}

	if newOrder < 1 || newOrder > len(rpHost.HttpRoutes) {
		return nil, ErrorReverseProxyHttpRouteInvalidOrder
	}

	normalizeReverseProxyHttpRoutesOrder(rpHost.HttpRoutes)

	currentIndex := -1
	for i, r := range rpHost.HttpRoutes {
		if strings.EqualFold(r.ID, routeId) {
			currentIndex = i
			break
		}
	}

	if currentIndex < 0 {
		return nil, ErrorReverseProxyHttpRouteNotFound
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
	normalizeReverseProxyHttpRoutesOrder(rpHost.HttpRoutes)

	for i, h := range j.data.ReverseProxyHosts {
		if h.ID == rpHost.ID {
			j.data.ReverseProxyHosts[i].HttpRoutes = rpHost.HttpRoutes
			_ = j.SaveNow(ctx)
			return &j.data.ReverseProxyHosts[i], nil
		}
	}

	return nil, ErrorReverseProxyHostNotFound
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.
func (j *JsonDatabase) UpdateReverseProxyHostTcpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostTcpRoute) (*models.ReverseProxyHostTcpRoute, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	rpHost, err := j.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}

	if len(rpHost.HttpRoutes) > 0 {
		return nil, ErrorReverseProxyTcpRouteWithHttpRouteHost
	}

	if rpHost.TcpRoute == nil {
		return nil, ErrorReverseProxyTcpRouteNotFound
	}

	if rpHost.TcpRoute.Diff(route) {
		for i, h := range j.data.ReverseProxyHosts {
			if h.ID == rpHost.ID {
				j.data.ReverseProxyHosts[i].TcpRoute = &route
				_ = j.SaveNow(ctx)
				return &route, nil
			}
		}
	}

	return nil, ErrorReverseProxyHostNotFound
}
