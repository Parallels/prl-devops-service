package orchestrator

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) getHostDiskSpace(ctx basecontext.ApiContext, host data_models.OrchestratorHost, username string) (models.DiskSpaceAvailable, error) {
	httpClient := s.getApiClient(host)
	httpClient.WithTimeout(20 * time.Second)
	path := "/v1/config/diskspace"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return models.DiskSpaceAvailable{}, err
	}
	var request models.DiskSpaceAvailableRequest
	request.UserName = username
	request.FolderPath = ""

	var response models.DiskSpaceAvailable
	apiResponse, err := httpClient.Post(url.String(), request, &response)
	if err != nil {
		return models.DiskSpaceAvailable{}, err
	}

	if apiResponse.StatusCode != 200 {
		return models.DiskSpaceAvailable{}, errors.NewWithCodef(400, "Error getting hardware info for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return response, nil
}
