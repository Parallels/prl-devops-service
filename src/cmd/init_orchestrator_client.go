package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/startup"
	"github.com/cjlapao/common-go/helper"
)

func processInitOrchestratorClient(ctx basecontext.ApiContext, command string) {
	if runtime.GOOS != "darwin" {
		ctx.LogErrorf("Init orchestrator client is only supported on macOS systems.")
		os.Exit(1)
	}

	startup.Init(ctx)
	startup.Start(ctx)

	orchestratorUrl := ""
	orchestratorToken := ""
	hostName := ""

	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--"+constants.ORCHESTRATOR_URL_FLAG+"=") {
			orchestratorUrl = strings.TrimPrefix(arg, "--"+constants.ORCHESTRATOR_URL_FLAG+"=")
		}
		if strings.HasPrefix(arg, "--"+constants.ORCHESTRATOR_TOKEN_FLAG+"=") {
			orchestratorToken = strings.TrimPrefix(arg, "--"+constants.ORCHESTRATOR_TOKEN_FLAG+"=")
		}
		if strings.HasPrefix(arg, "--"+constants.HOST_NAME_FLAG+"=") {
			hostName = strings.TrimPrefix(arg, "--"+constants.HOST_NAME_FLAG+"=")
		}
	}

	if orchestratorUrl == "" || hostName == "" {
		ctx.LogErrorf("Host name and orchestrator url must be provided.")
		os.Exit(1)
	}

	if orchestratorToken == "" {
		ctx.LogErrorf("Orchestrator token must be provided.")
		os.Exit(1)
	}

	pdProvider := serviceprovider.Get()
	if pdProvider == nil || pdProvider.ParallelsDesktopService == nil {
		ctx.LogErrorf("Parallels Desktop service is not initialized")
		os.Exit(1)
	}
	pdService := pdProvider.ParallelsDesktopService

	if !pdService.Installed() {
		ctx.LogInfof("Parallels Desktop not found. Attempting to install...")
		err := pdService.InstallFromDmg("", "latest", map[string]string{})
		if err != nil {
			ctx.LogErrorf(err.Error())
			os.Exit(1)
		}
	} else {
		ctx.LogInfof("Parallels Desktop is already installed. Version: %s", pdService.Version())
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("Failed to initialize database service: %v", err)
		os.Exit(1)
	}

	// Generate special API key locally
	keyName := hostName
	apiKeyReq := models.ApiKeyRequest{
		Name:   keyName,
		Key:    helper.RandomString(32),
		Secret: helper.RandomString(40),
	}

	dtoApiKey := mappers.ApiKeyRequestToDto(apiKeyReq)
	_, err = dbService.CreateApiKey(ctx, dtoApiKey)
	if err != nil {
		ctx.LogErrorf("Failed to generate local API key: %v", err)
		os.Exit(1)
	}

	apiKey := apiKeyReq.Key + ":" + apiKeyReq.Secret

	orchestratorHostReq := models.OrchestratorHostRequest{
		Host:        fmt.Sprintf("http://localhost"), // Typically self referring here for the host url if it dials back
		Description: hostName,
		Authentication: &models.OrchestratorAuthentication{
			ApiKey: apiKey,
		},
	}

	err = orchestratorHostReq.Validate()
	if err != nil {
		ctx.LogErrorf("Invalid orchestrator host request: %v", err)
		os.Exit(1)
	}

	registrationUrl, err := url.Parse(orchestratorUrl)
	if err != nil {
		ctx.LogErrorf("Invalid orchestrator URL: %v", err)
		os.Exit(1)
	}
	registrationUrl.Path = "/api/v1/orchestrator/hosts"

	reqBody, _ := json.Marshal(orchestratorHostReq)
	req, err := http.NewRequest("POST", registrationUrl.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		ctx.LogErrorf("Failed to create HTTP request: %v", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+orchestratorToken)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		ctx.LogErrorf("Failed to register host with orchestrator: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ctx.LogErrorf("Failed to register host. Orchestrator returned status: %d", resp.StatusCode)
		os.Exit(1)
	}

	ctx.LogInfof("Orchestrator client initialized successfully")
	ctx.LogInfof("Generated API Key: " + apiKey)
}
