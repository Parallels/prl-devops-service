package orchestrator

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/ssh"
)

// shellSingleQuote wraps s in single quotes and escapes any embedded single
// quotes so the result is safe to pass as a shell argument.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// detectOutboundIP returns the first non-loopback, non-link-local IPv4 address
// found on a local network interface, falling back to "127.0.0.1".
func detectOutboundIP() string {
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

// resolveOrchestratorBaseURL returns the orchestrator's publicly reachable URL.
// It reads BASE_URL from config first; if not set it falls back to the first
// non-loopback IPv4 interface address combined with the configured API port.
func resolveOrchestratorBaseURL(ctx basecontext.ApiContext) string {
	cfg := config.Get()
	if base := cfg.GetKey(constants.BASE_URL_ENV_VAR); base != "" {
		return strings.TrimRight(base, "/")
	}

	// Auto-detect
	schema := "http"
	if cfg.GetBoolKey(constants.TLS_ENABLED_ENV_VAR) {
		schema = "https"
	}
	port := cfg.GetKey(constants.API_PORT_ENV_VAR)
	if port == "" {
		port = constants.DEFAULT_API_PORT
	}

	ip := detectOutboundIP()

	return fmt.Sprintf("%s://%s:%s", schema, ip, port)
}

// DeployAndRegisterAgent installs the devops agent on a remote host via SSH and
// registers it with this orchestrator instance.  It is called by both the
// synchronous and asynchronous deploy handlers.
func (s *OrchestratorService) DeployAndRegisterAgent(ctx basecontext.ApiContext, req apimodels.DeployOrchestratorHostRequest, onOutput func(string)) (*apimodels.DeployOrchestratorHostResponse, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, fmt.Errorf("database unavailable: %w", err)
	}

	// 1. Generate a single-use enrollment token bound to the intended host name.
	ttl := req.EnrollmentTokenTTL
	if ttl <= 0 {
		ttl = constants.DEFAULT_ENROLLMENT_TOKEN_TTL_MINUTES
	}
	enrollToken, err := dbService.CreateEnrollmentToken(ctx, req.HostName, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to create enrollment token: %w", err)
	}

	// 2. Determine the URL the agent will use to reach this orchestrator.
	orchURL := resolveOrchestratorBaseURL(ctx)

	// 3. Build the remote command sequence.
	//    Step A – install the service.
	installArgs := []string{}
	if req.RootPassword != "" {
		installArgs = append(installArgs, "--root-password", req.RootPassword)
	}
	if req.EnabledModules != "" {
		installArgs = append(installArgs, "--modules", req.EnabledModules)
	}
	if req.PdVersion != "" {
		installArgs = append(installArgs, "--pd-version", req.PdVersion)
	}
	if req.AgentVersion != "" {
		installArgs = append(installArgs, "--version", req.AgentVersion)
	} else if req.PreRelease {
		installArgs = append(installArgs, "--pre-release")
	}
	installFlagsStr := strings.Join(installArgs, " ")

	//    Step B – register the agent with this orchestrator.
	//    Use the absolute path so that non-interactive SSH sessions (which often
	//    omit /usr/local/bin from PATH) can still find the binary.
	tagsStr := strings.Join(req.Tags, ",")
	// Resolve the port the agent will listen on: explicit field → default.
	agentPort := req.AgentPort
	if agentPort == "" {
		agentPort = constants.DEFAULT_API_PORT
	}

	// Use the SSH host as the advertised agent URL — it's the address we already
	// know is reachable from the outside.  Setting BASE_URL in the environment
	// takes top priority in resolveSelfBaseURL on the remote side.
	agentBaseURL := fmt.Sprintf("http://%s:%s", req.SshHost, agentPort)
	registerCmd := fmt.Sprintf(
		"BASE_URL=%s /usr/local/bin/prldevops register-with-orchestrator --orchestrator-url=%s --orchestrator-token=%s --host-name=%s --port=%s",
		agentBaseURL, orchURL, enrollToken.Token, req.HostName, agentPort,
	)
	if tagsStr != "" {
		registerCmd += " --tags=" + tagsStr
	}

	installCmd := fmt.Sprintf(
		`curl -fsSL https://raw.githubusercontent.com/Parallels/prl-devops-service/main/scripts/install.sh | bash -s -- %s`,
		installFlagsStr,
	)

	// Wait for the service to be reachable before registering.
	healthCmd := fmt.Sprintf(
		`for i in $(seq 1 30); do curl -sf http://localhost:%s/api/health/probe && break || sleep 2; done`,
		agentPort,
	)

	fullCmd := fmt.Sprintf("%s && %s && %s", installCmd, healthCmd, registerCmd)

	// 4. Execute over SSH.
	sshPort, _ := strconv.Atoi(req.SshPort)
	if sshPort == 0 {
		sshPort = 22
	}

	// Resolve the sudo password: explicit field takes precedence, then fall
	// back to the SSH login password (they're usually the same on macOS admin accounts).
	sudoPassword := req.SudoPassword
	if sudoPassword == "" {
		sudoPassword = req.SshPassword
	}

	// Wrap the full command to run as root via sudo -S so that all internal
	// sudo calls in the install script become no-ops and nothing blocks on a
	// TTY password prompt.  sudo -S reads the password from stdin, which we
	// pre-seed; no pseudo-terminal is required.
	execCmd := fullCmd
	if sudoPassword != "" {
		execCmd = fmt.Sprintf("echo %s | sudo -S bash -c %s",
			shellSingleQuote(sudoPassword), shellSingleQuote(fullCmd))
	}

	// Wrap the caller's callback so every line is also logged at info level.
	lineHandler := func(line string) {
		ctx.LogInfof("[deploy %s] %s", req.HostName, line)
		if onOutput != nil {
			onOutput(line)
		}
	}

	sshSvc := ssh.New(ctx)
	output, err := sshSvc.ExecuteWithCallback(ctx, req.SshHost, sshPort, req.SshUser, req.SshPassword, req.SshKey, execCmd, req.SshInsecureHostKey, lineHandler)
	if err != nil {
		return nil, fmt.Errorf("SSH execution failed: %w\noutput: %s", err, output)
	}

	// 5. Extract the host ID printed by register-with-orchestrator.
	hostID := ""
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HOST_ID=") {
			hostID = strings.TrimPrefix(line, "HOST_ID=")
		}
	}

	// 6. Wait up to 30 s for the host to appear in the orchestrator's DB.
	hostURL := ""
	if hostID != "" {
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			h, err := dbService.GetOrchestratorHost(ctx, hostID)
			if err == nil && h != nil {
				hostURL = h.Host
				break
			}
			time.Sleep(2 * time.Second)
		}
	}

	return &apimodels.DeployOrchestratorHostResponse{
		HostID:  hostID,
		Host:    hostURL,
		Message: "Agent deployed and registered successfully",
	}, nil
}
