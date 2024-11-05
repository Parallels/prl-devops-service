package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoReverseProxyHostCorsToApi(m data_models.ReverseProxyHostCors) models.ReverseProxyHostCors {
	r := models.ReverseProxyHostCors{
		Enabled:        m.Enabled,
		AllowedOrigins: m.AllowedOrigins,
		AllowedMethods: m.AllowedMethods,
		AllowedHeaders: m.AllowedHeaders,
	}

	return r
}

func ApiReverseProxyHostCorsToDto(m models.ReverseProxyHostCors) data_models.ReverseProxyHostCors {
	r := data_models.ReverseProxyHostCors{
		Enabled:        m.Enabled,
		AllowedOrigins: m.AllowedOrigins,
		AllowedMethods: m.AllowedMethods,
		AllowedHeaders: m.AllowedHeaders,
	}

	return r
}
