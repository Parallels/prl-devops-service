package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	rp_models "github.com/Parallels/prl-devops-service/reverse_proxy/models"
)

func DtoReverseProxyToApi(m data_models.ReverseProxy) models.ReverseProxy {
	r := models.ReverseProxy{
		Host: m.Host,
		Port: m.Port,
	}

	return r
}

func ApiReverseProxyToDto(m models.ReverseProxy) data_models.ReverseProxy {
	r := data_models.ReverseProxy{
		Host: m.Host,
		Port: m.Port,
	}

	return r
}

func ConfigReverseProxyToDto(m rp_models.ReverseProxyConfig) data_models.ReverseProxy {
	r := data_models.ReverseProxy{
		Enabled: m.Enabled,
		Host:    m.Host,
		Port:    m.Port,
	}

	return r
}
