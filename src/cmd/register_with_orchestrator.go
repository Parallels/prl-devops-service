package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/startup"
	"github.com/cjlapao/common-go/helper"
)

// minimalStartup initialises only what register-with-orchestrator needs:
// security primitives + service provider (DB + PD).  It deliberately avoids
// startup.Start so that background goroutines (EventEmitter, job manager,
// orchestrator service, etc.) are never spun up.
func minimalStartup(ctx basecontext.ApiContext) {
	startup.Init(ctx)
	serviceprovider.InitServices(ctx)
}

func processRegisterWithOrchestrator(ctx basecontext.ApiContext, command string) {
	if runtime.GOOS != "darwin" {
		ctx.LogErrorf("register-with-orchestrator is only supported on macOS systems.")
		os.Exit(1)
	}

	minimalStartup(ctx)

	// --- Parse flags ---
	orchestratorURL := ""
	enrollmentToken := ""
	hostName := ""
	tagsRaw := ""
	pdVersion := "latest"
	agentPort := ""

	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--"+constants.ORCHESTRATOR_URL_FLAG+"=") {
			orchestratorURL = strings.TrimPrefix(arg, "--"+constants.ORCHESTRATOR_URL_FLAG+"=")
		}
		if strings.HasPrefix(arg, "--"+constants.ORCHESTRATOR_TOKEN_FLAG+"=") {
			enrollmentToken = strings.TrimPrefix(arg, "--"+constants.ORCHESTRATOR_TOKEN_FLAG+"=")
		}
		if strings.HasPrefix(arg, "--"+constants.HOST_NAME_FLAG+"=") {
			hostName = strings.TrimPrefix(arg, "--"+constants.HOST_NAME_FLAG+"=")
		}
		if strings.HasPrefix(arg, "--"+constants.TAGS_FLAG+"=") {
			tagsRaw = strings.TrimPrefix(arg, "--"+constants.TAGS_FLAG+"=")
		}
		if strings.HasPrefix(arg, "--"+constants.PD_VERSION_FLAG+"=") {
			pdVersion = strings.TrimPrefix(arg, "--"+constants.PD_VERSION_FLAG+"=")
		}
		if strings.HasPrefix(arg, "--"+constants.API_PORT_FLAG+"=") {
			agentPort = strings.TrimPrefix(arg, "--"+constants.API_PORT_FLAG+"=")
		}
	}

	if orchestratorURL == "" || hostName == "" {
		ctx.LogErrorf("--orchestrator-url and --host-name are required.")
		os.Exit(1)
	}
	if enrollmentToken == "" {
		ctx.LogErrorf("--token is required.")
		os.Exit(1)
	}

	// --- Parallels Desktop check / install ---
	pdProvider := serviceprovider.Get()
	if pdProvider == nil || pdProvider.ParallelsDesktopService == nil {
		ctx.LogErrorf("Parallels Desktop service is not initialized")
		os.Exit(1)
	}
	pdService := pdProvider.ParallelsDesktopService

	if !pdService.Installed() {
		ctx.LogInfof("Parallels Desktop not found. Fetching latest version from Parallels livecheck...")
		version := pdVersion
		if version == "" || version == "latest" {
			latest, err := pdService.GetLatestVersion()
			if err != nil {
				ctx.LogErrorf("Could not determine latest Parallels Desktop version: %v", err)
				os.Exit(1)
			}
			version = latest
		}
		ctx.LogInfof("Installing Parallels Desktop %s...", version)
		if err := pdService.InstallFromDmg("", version, map[string]string{}); err != nil {
			ctx.LogErrorf("Failed to install Parallels Desktop: %v", err)
			os.Exit(1)
		}
	} else {
		ctx.LogInfof("Parallels Desktop is already installed. Version: %s", pdService.Version())
	}

	// --- Validate enrollment token against the orchestrator ---
	validateURL, err := url.Parse(orchestratorURL)
	if err != nil {
		ctx.LogErrorf("Invalid orchestrator URL: %v", err)
		os.Exit(1)
	}
	validateURL.Path = fmt.Sprintf("/api/v1/orchestrator/enrollment-token/%s/validate", enrollmentToken)

	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Get(validateURL.String())
	if err != nil {
		ctx.LogErrorf("Could not reach orchestrator at %s: %v", orchestratorURL, err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ctx.LogErrorf("Enrollment token validation failed (HTTP %d). Token may be expired or already used.", resp.StatusCode)
		os.Exit(1)
	}

	// --- Resolve the self URL this agent will advertise ---
	selfURL := resolveSelfBaseURL(ctx, agentPort)
	ctx.LogInfof("Agent will advertise URL: %s", selfURL)

	// --- Create a permanent local API key for the orchestrator to call back ---
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("Failed to initialise database: %v", err)
		os.Exit(1)
	}

	apiKeyReq := models.ApiKeyRequest{
		Name:   hostName,
		Key:    helper.RandomString(32),
		Secret: helper.RandomString(40),
	}
	dtoApiKey := mappers.ApiKeyRequestToDto(apiKeyReq)
	if _, err := dbService.CreateApiKey(ctx, dtoApiKey); err != nil {
		ctx.LogErrorf("Failed to create local API key: %v", err)
		os.Exit(1)
	}

	// --- Parse tags ---
	var tags []string
	for _, t := range strings.Split(tagsRaw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	// --- Build the registration request ---
	apiKey := apiKeyReq.Key + ":" + apiKeyReq.Secret
	regReq := models.OrchestratorHostRequest{
		Host:        selfURL,
		Description: hostName,
		Tags:        tags,
		Authentication: &models.OrchestratorAuthentication{
			ApiKey: apiKey,
		},
	}
	if err := regReq.Validate(); err != nil {
		ctx.LogErrorf("Invalid orchestrator host request: %v", err)
		os.Exit(1)
	}

	// --- POST to the orchestrator ---
	regURL, _ := url.Parse(orchestratorURL)
	regURL.Path = "/api/v1/orchestrator/hosts"

	body, _ := json.Marshal(regReq)
	req, err := http.NewRequest(http.MethodPost, regURL.String(), bytes.NewBuffer(body))
	if err != nil {
		ctx.LogErrorf("Failed to build HTTP request: %v", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(constants.ENROLLMENT_TOKEN_HEADER, enrollmentToken)

	regResp, err := httpClient.Do(req)
	if err != nil {
		ctx.LogErrorf("Failed to reach orchestrator: %v", err)
		os.Exit(1)
	}
	defer regResp.Body.Close()

	if regResp.StatusCode < 200 || regResp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(regResp.Body)
		ctx.LogErrorf("Orchestrator returned HTTP %d when registering host: %s", regResp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var hostResp models.OrchestratorHostResponse
	if err := json.NewDecoder(regResp.Body).Decode(&hostResp); err != nil {
		ctx.LogErrorf("Failed to parse registration response: %v", err)
		os.Exit(1)
	}

	ctx.LogInfof("Agent registered successfully")
	// Print the host ID to stdout in a machine-readable form so callers can parse it.
	fmt.Printf("HOST_ID=%s\n", hostResp.ID)

	os.Exit(0)
}

// resolveSelfBaseURL returns the URL this agent should advertise to the orchestrator.
// Priority: BASE_URL env var → portOverride flag → API_PORT env var → DEFAULT_API_PORT.
func resolveSelfBaseURL(ctx basecontext.ApiContext, portOverride string) string {
	cfg := config.Get()
	if base := cfg.GetKey(constants.BASE_URL_ENV_VAR); base != "" {
		return strings.TrimRight(base, "/")
	}

	schema := "http"
	if cfg.GetBoolKey(constants.TLS_ENABLED_ENV_VAR) {
		schema = "https"
	}
	port := portOverride
	if port == "" {
		port = cfg.GetKey(constants.API_PORT_ENV_VAR)
	}
	if port == "" {
		port = constants.DEFAULT_API_PORT
	}

	ip := detectLocalOutboundIP()
	return fmt.Sprintf("%s://%s:%s", schema, ip, port)
}

// detectLocalOutboundIP returns the first non-loopback, non-link-local IPv4 address.
func detectLocalOutboundIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "127.0.0.1"
}
