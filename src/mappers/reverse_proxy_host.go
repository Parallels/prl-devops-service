package mappers

import (
	config_models "github.com/Parallels/prl-devops-service/config"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoReverseProxyHostToApi(m data_models.ReverseProxyHost) models.ReverseProxyHost {
	r := models.ReverseProxyHost{
		ID:   m.ID,
		Name: m.Name,
		Host: m.Host,
		Port: m.Port,
	}

	if m.Tls != nil {
		e := DtoReverseProxyHostTlsToApi(*m.Tls)
		r.Tls = &e
	}

	if m.Cors != nil {
		e := DtoReverseProxyHostCorsToApi(*m.Cors)
		r.Cors = &e
	}

	if m.HttpRoutes != nil {
		for _, route := range m.HttpRoutes {
			e := DtoReverseProxyHostHttpRouteToApi(*route)
			r.HttpRoutes = append(r.HttpRoutes, &e)
		}
	}

	if m.TcpRoute != nil {
		e := DtoReverseProxyHostTcpRouteToApi(*m.TcpRoute)
		r.TcpRoute = &e
	}

	return r
}

func DtoReverseProxyHostsToApi(m []data_models.ReverseProxyHost) []models.ReverseProxyHost {
	var hosts []models.ReverseProxyHost
	for _, host := range m {
		hosts = append(hosts, DtoReverseProxyHostToApi(host))
	}

	return hosts
}

func ApiReverseProxyHostToDto(m models.ReverseProxyHost) data_models.ReverseProxyHost {
	r := data_models.ReverseProxyHost{
		ID:   m.ID,
		Name: m.Name,
		Host: m.Host,
		Port: m.Port,
	}

	if m.Tls != nil {
		e := ApiReverseProxyHostTlsToDto(*m.Tls)
		r.Tls = &e
	}

	if m.Cors != nil {
		e := ApiReverseProxyHostCorsToDto(*m.Cors)
		r.Cors = &e
	}

	if m.HttpRoutes != nil {
		for _, route := range m.HttpRoutes {
			e := ApiReverseProxyHostHttpRouteToDto(*route)
			r.HttpRoutes = append(r.HttpRoutes, &e)
		}
	}

	if m.TcpRoute != nil {
		e := ApiReverseProxyHostTcpRouteToDto(*m.TcpRoute)
		r.TcpRoute = &e
	}

	return r
}

func ApiReverseProxyHostsToDto(m []models.ReverseProxyHost) []data_models.ReverseProxyHost {
	var hosts []data_models.ReverseProxyHost
	for _, host := range m {
		hosts = append(hosts, ApiReverseProxyHostToDto(host))
	}

	return hosts
}

func ApiCreateRequestReverseProxyHostToDto(m models.ReverseProxyHostCreateRequest) data_models.ReverseProxyHost {
	r := data_models.ReverseProxyHost{
		Name: m.Name,
		Host: m.Host,
		Port: m.Port,
	}

	if m.Tls != nil {
		e := ApiReverseProxyHostTlsToDto(*m.Tls)
		r.Tls = &e
	}

	if m.Cors != nil {
		e := ApiReverseProxyHostCorsToDto(*m.Cors)
		r.Cors = &e
	}

	if m.HttpRoutes != nil {
		for _, route := range m.HttpRoutes {
			e := ApiReverseProxyHostHttpRouteToDto(*route)
			r.HttpRoutes = append(r.HttpRoutes, &e)
		}
	}

	if m.TcpRoute != nil {
		e := ApiReverseProxyHostTcpRouteToDto(*m.TcpRoute)
		r.TcpRoute = &e
	}

	return r
}

func ApiUpdateRequestReverseProxyHostToDto(m models.ReverseProxyHostUpdateRequest) data_models.ReverseProxyHost {
	r := data_models.ReverseProxyHost{
		Name: m.Name,
		Host: m.Host,
		Port: m.Port,
	}

	if m.Tls != nil {
		e := ApiReverseProxyHostTlsToDto(*m.Tls)
		r.Tls = &e
	}

	if m.Cors != nil {
		e := ApiReverseProxyHostCorsToDto(*m.Cors)
		r.Cors = &e
	}

	return r
}

func ConfigReverseProxyHostToDto(m config_models.ReverseProxyConfigHost) data_models.ReverseProxyHost {
	r := data_models.ReverseProxyHost{
		ID:   helpers.GenerateId(),
		Name: m.Name,
		Host: m.Host,
		Port: m.Port,
	}

	if m.HttpRoutes != nil {
		for _, route := range m.HttpRoutes {
			e := ConfigReverseProxyHostHttpRouteToDto(*route)
			r.HttpRoutes = append(r.HttpRoutes, &e)
		}
	}

	if m.TcpRoute != nil {
		e := ConfigReverseProxyHostTcpRouteToDto(*m.TcpRoute)
		r.TcpRoute = &e
	}

	return r
}
