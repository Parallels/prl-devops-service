package mappers

import (
	"net/url"

	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/models"
)

func DtoOrchestratorHostToApiResponse(dto data_models.OrchestratorHost) models.OrchestratorHostResponse {
	result := models.OrchestratorHostResponse{
		ID:             dto.ID,
		Host:           dto.GetHost(),
		Architecture:   dto.Architecture,
		CpuModel:       dto.CpuModel,
		Description:    dto.Description,
		Tags:           dto.Tags,
		RequiredClaims: dto.RequiredClaims,
		RequiredRoles:  dto.RequiredRoles,
		State:          dto.State,
	}

	if dto.Resources != nil {
		result.Resources = DtoOrchestratorResourceItemToApi(dto.Resources.Total)
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
