package orchestrator

import (
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
)

func (s *OrchestratorService) DeleteHostVirtualMachine(host *models.OrchestratorHost, vmId string) error {
	httpClient := s.getApiClient(*host)
	path := "/v1/machines/" + vmId
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return err
	}

	apiResponse, err := httpClient.Get(url.String(), nil)
	if err != nil {
		return err
	}

	if apiResponse.StatusCode != 202 {
		return errors.NewWithCodef(400, "Error deleting virtual machine for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return nil
}
