package mappers

import (
	config_models "github.com/Parallels/prl-devops-service/config"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoReverseProxyHostHttpRouteToApi(m data_models.ReverseProxyHostHttpRoute) models.ReverseProxyHostHttpRoute {
	return models.ReverseProxyHostHttpRoute{
		ID:              m.ID,
		Path:            m.Path,
		TargetVmId:      m.TargetVmId,
		TargetPort:      m.TargetPort,
		TargetHost:      m.TargetHost,
		Schema:          m.Schema,
		Pattern:         m.Pattern,
		RequestHeaders:  m.RequestHeaders,
		ResponseHeaders: m.ResponseHeaders,
	}
}

func DtoReverseProxyHostHttpRoutesFromApi(m []data_models.ReverseProxyHostHttpRoute) []models.ReverseProxyHostHttpRoute {
	var routes []models.ReverseProxyHostHttpRoute
	for _, route := range m {
		routes = append(routes, DtoReverseProxyHostHttpRouteToApi(route))
	}

	return routes
}

func ApiReverseProxyHostHttpRouteToDto(m models.ReverseProxyHostHttpRoute) data_models.ReverseProxyHostHttpRoute {
	return data_models.ReverseProxyHostHttpRoute{
		ID:              m.ID,
		Path:            m.Path,
		TargetVmId:      m.TargetVmId,
		TargetPort:      m.TargetPort,
		TargetHost:      m.TargetHost,
		Schema:          m.Schema,
		Pattern:         m.Pattern,
		RequestHeaders:  m.RequestHeaders,
		ResponseHeaders: m.ResponseHeaders,
	}
}

func ApiReverseProxyHostHttpRoutesToDto(m []models.ReverseProxyHostHttpRoute) []data_models.ReverseProxyHostHttpRoute {
	var routes []data_models.ReverseProxyHostHttpRoute
	for _, route := range m {
		routes = append(routes, ApiReverseProxyHostHttpRouteToDto(route))
	}

	return routes
}

func ApiReverseProxyHostCreateHttpRouteToDto(m models.ReverseProxyHostHttpRouteCreateRequest) data_models.ReverseProxyHostHttpRoute {
	result := data_models.ReverseProxyHostHttpRoute{
		Path:            m.Path,
		Pattern:         m.Pattern,
		TargetVmId:      m.TargetVmId,
		TargetPort:      m.TargetPort,
		TargetHost:      m.TargetHost,
		Schema:          m.Schema,
		RequestHeaders:  m.RequestHeaders,
		ResponseHeaders: m.ResponseHeaders,
	}

	return result
}

func ConfigReverseProxyHostHttpRouteToDto(m config_models.ReverseProxyConfigServerHttpRoute) data_models.ReverseProxyHostHttpRoute {
	return data_models.ReverseProxyHostHttpRoute{
		ID:              helpers.GenerateId(),
		Path:            m.Path,
		TargetPort:      m.TargetPort,
		TargetHost:      m.TargetHost,
		Schema:          m.Scheme,
		Pattern:         m.Pattern,
		RequestHeaders:  m.RequestHeaders,
		ResponseHeaders: m.ResponseHeaders,
	}
}

func ConfigReverseProxyHostHttpRoutesToDto(m []config_models.ReverseProxyConfigServerHttpRoute) []data_models.ReverseProxyHostHttpRoute {
	var routes []data_models.ReverseProxyHostHttpRoute
	for _, route := range m {
		routes = append(routes, ConfigReverseProxyHostHttpRouteToDto(route))
	}

	return routes
}
