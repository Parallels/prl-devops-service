package mappers

import (
	config_models "github.com/Parallels/prl-devops-service/config"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoReverseProxyHostTcpRouteToApi(m data_models.ReverseProxyHostTcpRoute) models.ReverseProxyHostTcpRoute {
	return models.ReverseProxyHostTcpRoute{
		ID:         m.ID,
		TargetPort: m.TargetPort,
		TargetHost: m.TargetHost,
		TargetVmId: m.TargetVmId,
	}
}

func DtoReverseProxyHostTcpRoutesToApi(m []data_models.ReverseProxyHostTcpRoute) []models.ReverseProxyHostTcpRoute {
	var routes []models.ReverseProxyHostTcpRoute
	for _, route := range m {
		routes = append(routes, DtoReverseProxyHostTcpRouteToApi(route))
	}

	return routes
}

func ApiReverseProxyHostTcpRouteToDto(m models.ReverseProxyHostTcpRoute) data_models.ReverseProxyHostTcpRoute {
	return data_models.ReverseProxyHostTcpRoute{
		ID:         m.ID,
		TargetPort: m.TargetPort,
		TargetHost: m.TargetHost,
		TargetVmId: m.TargetVmId,
	}
}

func ApiReverseProxyHostTcpRoutesToDto(m []models.ReverseProxyHostTcpRoute) []data_models.ReverseProxyHostTcpRoute {
	var routes []data_models.ReverseProxyHostTcpRoute
	for _, route := range m {
		routes = append(routes, ApiReverseProxyHostTcpRouteToDto(route))
	}

	return routes
}

func ApiReverseProxyHostTcpRouteCreateRequestToDto(m models.ReverseProxyHostTcpRouteCreateRequest) data_models.ReverseProxyHostTcpRoute {
	return data_models.ReverseProxyHostTcpRoute{
		TargetPort: m.TargetPort,
		TargetHost: m.TargetHost,
		TargetVmId: m.TargetVmId,
	}
}

func ConfigReverseProxyHostTcpRouteToDto(m config_models.ReverseProxyConfigServerTcpRoute) data_models.ReverseProxyHostTcpRoute {
	return data_models.ReverseProxyHostTcpRoute{
		ID:         helpers.GenerateId(),
		TargetPort: m.TargetPort,
		TargetHost: m.TargetHost,
	}
}

func ConfigReverseProxyHostTcpRoutesToDto(m []config_models.ReverseProxyConfigServerTcpRoute) []data_models.ReverseProxyHostTcpRoute {
	var routes []data_models.ReverseProxyHostTcpRoute
	for _, route := range m {
		routes = append(routes, ConfigReverseProxyHostTcpRouteToDto(route))
	}

	return routes
}
