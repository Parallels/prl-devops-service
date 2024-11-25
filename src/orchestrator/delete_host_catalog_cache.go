package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) DeleteHostCatalogCacheItem(ctx basecontext.ApiContext, hostId string, catalogId string, versionId string) error {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return err
	}
	if host == nil {
		return errors.NewWithCodef(404, "Host %s not found", hostId)
	}
	if !host.Enabled {
		return errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}
	if host.State != "healthy" {
		return errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	err = s.CallDeleteHostCatalogCacheItem(host, catalogId, versionId)
	if err != nil {
		return err
	}

	s.Refresh()
	return nil
}

func (s *OrchestratorService) CallDeleteHostCatalogCacheItem(host *models.OrchestratorHost, catalogId string, versionId string) error {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(5 * time.Minute) // sometimes deleting files can take a bit, waiting for 5 minutes
	path := "/v1/catalog/cache"
	if catalogId != "" {
		path = path + "/" + catalogId
	}
	if versionId != "" {
		path = path + "/" + versionId
	}

	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	apiResponse, err := httpClient.Delete(url.String(), nil)
	if err != nil {
		return err
	}

	if apiResponse.StatusCode != 202 {
		msg := "Error deleting catalog cache item"
		if catalogId != "" {
			msg = msg + " and catalogId " + catalogId
		}
		if versionId != "" {
			msg = msg + " and versionId " + versionId
		}
		var apiError *errors.SystemError
		if apiResponse.ApiError != nil {
			stackError := apiResponse.ApiError.ToError()
			apiError = errors.NewWithCodeAndNestedErrorf(*stackError, 400, "%v: %v", msg, stackError)
			apiError.Path = url.String()
		} else {
			apiError = errors.NewWithCodef(400, "%v: %v", msg, apiResponse.StatusCode)
			apiError.Path = url.String()
		}

		return apiError
	}

	s.Refresh()
	return nil
}
