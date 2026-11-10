package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/reverse_proxy delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetReverseProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	return g.legacy.GetReverseProxyConfig(ctx)
}

func (g *GormDatabase) EnableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	return g.legacy.EnableProxyConfig(ctx)
}

func (g *GormDatabase) DisableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	return g.legacy.DisableProxyConfig(ctx)
}

func (g *GormDatabase) UpdateReverseProxy(ctx basecontext.ApiContext, rp models.ReverseProxy) (*models.ReverseProxy, error) {
	return g.legacy.UpdateReverseProxy(ctx, rp)
}

func (g *GormDatabase) GetReverseProxyHosts(ctx basecontext.ApiContext, filter string) ([]models.ReverseProxyHost, error) {
	return g.legacy.GetReverseProxyHosts(ctx, filter)
}

func (g *GormDatabase) GetReverseProxyHost(ctx basecontext.ApiContext, idOrName string) (*models.ReverseProxyHost, error) {
	return g.legacy.GetReverseProxyHost(ctx, idOrName)
}

func (g *GormDatabase) CreateReverseProxyHost(ctx basecontext.ApiContext, rpHost models.ReverseProxyHost) (*models.ReverseProxyHost, error) {
	return g.legacy.CreateReverseProxyHost(ctx, rpHost)
}

func (g *GormDatabase) DeleteReverseProxyHost(ctx basecontext.ApiContext, idOrName string) error {
	return g.legacy.DeleteReverseProxyHost(ctx, idOrName)
}

func (g *GormDatabase) UpdateReverseProxyHost(ctx basecontext.ApiContext, rpHost *models.ReverseProxyHost) (*models.ReverseProxyHost, error) {
	return g.legacy.UpdateReverseProxyHost(ctx, rpHost)
}

func (g *GormDatabase) ConfigureReverseProxyHostTls(ctx basecontext.ApiContext, rpHostId string, tlsConfig models.ReverseProxyHostTls) (*models.ReverseProxyHost, error) {
	return g.legacy.ConfigureReverseProxyHostTls(ctx, rpHostId, tlsConfig)
}

func (g *GormDatabase) ConfigureReverseProxyHostCors(ctx basecontext.ApiContext, rpHostId string, corsConfig models.ReverseProxyHostCors) (*models.ReverseProxyHost, error) {
	return g.legacy.ConfigureReverseProxyHostCors(ctx, rpHostId, corsConfig)
}

func (g *GormDatabase) CreateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error) {
	return g.legacy.CreateReverseProxyHostHttpRoute(ctx, rpHostId, route)
}

func (g *GormDatabase) DeleteReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string) error {
	return g.legacy.DeleteReverseProxyHostHttpRoute(ctx, rpHostId, routeId)
}

func (g *GormDatabase) UpdateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error) {
	return g.legacy.UpdateReverseProxyHostHttpRoute(ctx, rpHostId, route)
}

func (g *GormDatabase) ReorderReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string, newOrder int) (*models.ReverseProxyHost, error) {
	return g.legacy.ReorderReverseProxyHostHttpRoute(ctx, rpHostId, routeId, newOrder)
}

func (g *GormDatabase) UpdateReverseProxyHostTcpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostTcpRoute) (*models.ReverseProxyHostTcpRoute, error) {
	return g.legacy.UpdateReverseProxyHostTcpRoute(ctx, rpHostId, route)
}
