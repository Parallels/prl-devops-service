package orchestrator

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostToken(ctx basecontext.ApiContext, hostId string) (string, string, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return "", "", err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return "", "", err
	}
	if host == nil {
		return "", "", errors.NewWithCodef(404, "Host %s not found", hostId)
	}
	if !host.Enabled {
		return "", "", errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}
	if host.State != "healthy" {
		return "", "", errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	httpClient := s.getApiClient(*host)
	hostUrl := fmt.Sprintf("%s/api/v1/auth/token", host.GetHost())
	auth, err := httpClient.Authorize(ctx, hostUrl)
	if err != nil {
		return "", "", err
	}

	if auth == nil {
		return "", "", errors.NewWithCodef(400, "Error getting token for host %s", host.Host)
	}

	if auth.BearerToken != "" {
		ctx.LogDebugf("[Api Client] Setting Authorization header to Bearer %s", helpers.ObfuscateString(auth.BearerToken))
		return "Authorization", "Bearer " + auth.BearerToken, nil
	} else if auth.ApiKey != "" {
		ctx.LogDebugf("[Api Client] Setting Authorization header to X-Api-Key %s", helpers.ObfuscateString(auth.ApiKey))
		return "X-Api-Key", auth.ApiKey, nil
	}

	return "", "", errors.NewWithCodef(400, "Error getting token for host %s", host.Host)
}
