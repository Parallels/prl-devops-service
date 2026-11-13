package interfaces

// ReverseProxyService mirrors reverse_proxy method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type ReverseProxyService interface {
	GetReverseProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error)
	EnableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error)
	DisableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error)
	UpdateReverseProxy(ctx basecontext.ApiContext, rp models.ReverseProxy) (*models.ReverseProxy, error)
	GetReverseProxyHosts(ctx basecontext.ApiContext, filter string) ([]models.ReverseProxyHost, error)
	GetReverseProxyHost(ctx basecontext.ApiContext, idOrName string) (*models.ReverseProxyHost, error)
	CreateReverseProxyHost(ctx basecontext.ApiContext, rpHost models.ReverseProxyHost) (*models.ReverseProxyHost, error)
	DeleteReverseProxyHost(ctx basecontext.ApiContext, idOrName string) error
	UpdateReverseProxyHost(ctx basecontext.ApiContext, rpHost *models.ReverseProxyHost) (*models.ReverseProxyHost, error)
	ConfigureReverseProxyHostTls(ctx basecontext.ApiContext, rpHostId string, tlsConfig models.ReverseProxyHostTls) (*models.ReverseProxyHost, error)
	ConfigureReverseProxyHostCors(ctx basecontext.ApiContext, rpHostId string, corsConfig models.ReverseProxyHostCors) (*models.ReverseProxyHost, error)
	CreateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error)
	DeleteReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string) error
	UpdateReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostHttpRoute) (*models.ReverseProxyHostHttpRoute, error)
	ReorderReverseProxyHostHttpRoute(ctx basecontext.ApiContext, rpHostId, routeId string, newOrder int) (*models.ReverseProxyHost, error)
	UpdateReverseProxyHostTcpRoute(ctx basecontext.ApiContext, rpHostId string, route models.ReverseProxyHostTcpRoute) (*models.ReverseProxyHostTcpRoute, error)
}
