package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostCatalogCache(ctx basecontext.ApiContext, hostId string) (*api_models.VirtualMachineCatalogManifestList, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", hostId)
	}
	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}
	if host.State != "healthy" {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	result, err := s.CallGetHostCatalogCache(host)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *OrchestratorService) CallGetHostCatalogCache(host *models.OrchestratorHost) (*api_models.VirtualMachineCatalogManifestList, error) {
	httpClient := s.getApiClient(*host)
	path := "/v1/catalog/cache"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response api_models.VirtualMachineCatalogManifestList
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting catalog cache for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return &response, nil
}
