package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoReverseProxyHostTlsToApi(m data_models.ReverseProxyHostTls) models.ReverseProxyHostTls {
	r := models.ReverseProxyHostTls{
		Enabled: m.Enabled,
		Cert:    m.Cert,
		Key:     m.Key,
	}

	return r
}

func ApiReverseProxyHostTlsToDto(m models.ReverseProxyHostTls) data_models.ReverseProxyHostTls {
	r := data_models.ReverseProxyHostTls{
		Enabled: m.Enabled,
		Cert:    m.Cert,
		Key:     m.Key,
	}

	return r
}
