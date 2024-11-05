package mappers

import (
	config_models "github.com/Parallels/prl-devops-service/config"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
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

func ConfigReverseProxyToDto(m config_models.ReverseProxyConfig) data_models.ReverseProxy {
	r := data_models.ReverseProxy{
		Host: m.Host,
		Port: m.Port,
	}

	return r
}
