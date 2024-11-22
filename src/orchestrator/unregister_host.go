package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) UnregisterHost(ctx basecontext.ApiContext, hostId string) error {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return err
	}

	err = dbService.DeleteOrchestratorHost(ctx, hostId)
	if err != nil {
		return err
	}

	s.Refresh()
	return nil
}
