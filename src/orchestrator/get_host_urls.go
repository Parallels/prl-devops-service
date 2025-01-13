package orchestrator

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) GetHostBaseUrl(ctx basecontext.ApiContext, hostId string) (string, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return "", err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return "", err
	}
	if host == nil {
		return "", errors.NewWithCodef(404, "Host %s not found", hostId)
	}
	if !host.Enabled {
		return "", errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}
	if host.State != "healthy" {
		return "", errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	baseUrl := host.GetHost()
	if baseUrl == "" {
		return "", errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	return baseUrl, nil
}

func (s *OrchestratorService) GetHostWebsocketBaseUrl(ctx basecontext.ApiContext, hostId string) (string, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return "", err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return "", err
	}
	if host == nil {
		return "", errors.NewWithCodef(404, "Host %s not found", hostId)
	}
	if !host.Enabled {
		return "", errors.NewWithCodef(400, "Host %s is disabled", hostId)
	}
	if host.State != "healthy" {
		return "", errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	baseUrl := host.GetHost()
	if baseUrl == "" {
		return "", errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	baseUrl = strings.Replace(baseUrl, "http://", "ws://", 1)
	baseUrl = strings.Replace(baseUrl, "https://", "wss://", 1)
	return baseUrl, nil
}
