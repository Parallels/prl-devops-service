package mappers

import (
	"net/url"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoOrchestratorHostToApiResponse(dto data_models.OrchestratorHost) models.OrchestratorHostResponse {
	result := models.OrchestratorHostResponse{
		ID:                       dto.ID,
		Enabled:                  dto.Enabled,
		Host:                     dto.GetHost(),
		Architecture:             dto.Architecture,
		CpuModel:                 dto.CpuModel,
		OsVersion:                dto.OsVersion,
		OsName:                   dto.OsName,
		ExternalIpAddress:        dto.ExternalIpAddress,
		DevOpsVersion:            dto.DevOpsVersion,
		Description:              dto.Description,
		ParallelsDesktopVersion:  dto.ParallelsDesktopVersion,
		ParallelsDesktopLicensed: dto.ParallelsDesktopLicensed,
		IsReverseProxyEnabled:    dto.IsReverseProxyEnabled,
		Tags:                     dto.Tags,
		RequiredClaims:           dto.RequiredClaims,
		RequiredRoles:            dto.RequiredRoles,
		State:                    dto.State,
	}

	if dto.Resources != nil {
		result.Resources = DtoOrchestratorResourceItemToApi(dto.Resources.Total)
	}

	if dto.ReverseProxy != nil {
		result.ReverseProxy = &models.HostReverseProxy{
			Host: dto.ReverseProxy.Host,
			Port: dto.ReverseProxy.Port,
		}
	}
	if len(dto.ReverseProxyHosts) > 0 {
		result.ReverseProxy.Hosts = make([]models.ReverseProxyHost, 0)
		for _, host := range dto.ReverseProxyHosts {
			result.ReverseProxy.Hosts = append(result.ReverseProxy.Hosts, DtoReverseProxyHostToApi(*host))
		}
	}
	return result
}

func ApiOrchestratorRequestToDto(request models.OrchestratorHostRequest) data_models.OrchestratorHost {
	result := data_models.OrchestratorHost{
		Host:           request.Host,
		Description:    request.Description,
		Tags:           request.Tags,
		RequiredClaims: request.RequiredClaims,
		RequiredRoles:  request.RequiredRoles,
	}

	if request.Authentication != nil {
		auth := ApiOrchestratorAuthenticationToDto(*request.Authentication)
		result.Authentication = &auth
	}

	hostUrl, err := url.Parse(request.Host)
	if err == nil {
		result.Host = hostUrl.Hostname()
		result.Port = hostUrl.Port()
		result.Schema = hostUrl.Scheme
		result.PathPrefix = hostUrl.Path
	}

	return result
}

func DtoOrchestratorAuthenticationToApi(dto data_models.OrchestratorHostAuthentication) models.OrchestratorAuthentication {
	result := models.OrchestratorAuthentication{
		Username: dto.Username,
		Password: dto.Password,
		ApiKey:   dto.ApiKey,
	}

	return result
}

func ApiOrchestratorAuthenticationToDto(request models.OrchestratorAuthentication) data_models.OrchestratorHostAuthentication {
	result := data_models.OrchestratorHostAuthentication{
		Username: request.Username,
		Password: request.Password,
		ApiKey:   request.ApiKey,
	}

	return result
}

func DtoOrchestratorResourceItemToApi(dto data_models.HostResourceItem) models.HostResourceItem {
	result := models.HostResourceItem{
		PhysicalCpuCount: dto.PhysicalCpuCount,
		LogicalCpuCount:  dto.LogicalCpuCount,
		MemorySize:       dto.MemorySize,
		DiskSize:         dto.DiskSize,
		FreeDiskSize:     dto.FreeDiskSize,
	}

	return result
}

func ApiOrchestratorResourceItemToDto(request models.HostResourceItem) data_models.HostResourceItem {
	result := data_models.HostResourceItem{
		PhysicalCpuCount: request.PhysicalCpuCount,
		LogicalCpuCount:  request.LogicalCpuCount,
		MemorySize:       request.MemorySize,
		DiskSize:         request.DiskSize,
		FreeDiskSize:     request.FreeDiskSize,
	}

	return result
}

func DtoOrchestratorReverseProxyToApi(dto data_models.HostReverseProxy) models.HostReverseProxy {
	result := models.HostReverseProxy{
		Enabled: dto.Enabled,
		Host:    dto.Host,
		Port:    dto.Port,
	}

	if len(dto.Hosts) > 0 {
		result.Hosts = make([]models.ReverseProxyHost, 0)
		for _, host := range dto.Hosts {
			result.Hosts = append(result.Hosts, DtoReverseProxyHostToApi(host))
		}
	}

	return result
}

func ApiOrchestratorReverseProxyToDto(request models.HostReverseProxy) data_models.HostReverseProxy {
	result := data_models.HostReverseProxy{
		Enabled: request.Enabled,
		Host:    request.Host,
		Port:    request.Port,
	}

	if len(request.Hosts) > 0 {
		result.Hosts = make([]data_models.ReverseProxyHost, 0)
		for _, host := range request.Hosts {
			result.Hosts = append(result.Hosts, ApiReverseProxyHostToDto(host))
		}
	}

	return result
}
