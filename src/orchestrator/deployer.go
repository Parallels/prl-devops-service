package orchestrator

import (
	"fmt"
	"net"
	"regexp"
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

// ansiEscape matches ANSI/VT100 escape sequences (colors, cursor moves, etc.)
var ansiEscape = regexp.MustCompile(`\x1b(\[[0-9;]*[A-Za-z]|[^[]|]|\][^\x07]*\x07)`)

// stripAnsiCodes removes all ANSI escape sequences from s.
func stripAnsiCodes(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// cleanOutput strips ANSI codes, normalises \r\n → \n, trims each line, and
// drops blank lines so the result is human-readable plain text.
func cleanOutput(s string) string {
	s = stripAnsiCodes(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

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
// onProgress is an optional callback invoked at key stages with a completion
// percentage (0–100) and a short status message.
func (s *OrchestratorService) DeployAndRegisterAgent(ctx basecontext.ApiContext, req apimodels.DeployOrchestratorHostRequest, onOutput func(string), onProgress func(int, string)) (*apimodels.DeployOrchestratorHostResponse, error) {
	progress := func(pct int, msg string) {
		if onProgress != nil {
			onProgress(pct, msg)
		}
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, fmt.Errorf("database unavailable: %w", err)
	}

	// 0. Guard against duplicate deployments: reject if a host with the same
	//    description (HostName) or the same SSH host address already exists.
	existingHosts, err := dbService.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to check existing hosts: %w", err)
	}
	for _, h := range existingHosts {
		if strings.EqualFold(h.Description, req.HostName) {
			return nil, fmt.Errorf("a host with the name %q already exists (id: %s)", req.HostName, h.ID)
		}
		if strings.EqualFold(h.Host, req.SshHost) {
			return nil, fmt.Errorf("a host with the address %q already exists (id: %s)", req.SshHost, h.ID)
		}
	}
	progress(5, "Configuration validated")

	// 1. Generate a single-use enrollment token bound to the intended host name.
	ttl := req.EnrollmentTokenTTL
	if ttl <= 0 {
		ttl = constants.DEFAULT_ENROLLMENT_TOKEN_TTL_MINUTES
	}
	enrollToken, err := dbService.CreateEnrollmentToken(ctx, req.HostName, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to create enrollment token: %w", err)
	}
	progress(10, "Enrollment token generated")

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
	// The loop explicitly exits 1 on the final attempt so that the && chain
	// short-circuits and the job is marked failed instead of silently continuing
	// to the register step against a dead agent.
	healthCmd := fmt.Sprintf(
		`for i in $(seq 1 30); do curl -sf http://localhost:%s/api/health/probe && break; if [ "$i" -eq 30 ]; then echo "Agent did not become reachable on port %s after 30 attempts" >&2; exit 1; fi; sleep 2; done`,
		agentPort, agentPort,
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
	// ANSI color codes and trailing carriage returns are stripped before
	// logging or forwarding so that job messages are plain text regardless
	// of the remote terminal's color output.
	lineHandler := func(line string) {
		clean := strings.TrimRight(stripAnsiCodes(line), "\r")
		if clean == "" {
			return
		}
		ctx.LogInfof("[deploy %s] %s", req.HostName, clean)
		if onOutput != nil {
			onOutput(clean)
		}
	}

	progress(15, "Connecting to host via SSH")
	sshSvc := ssh.New(ctx)
	output, err := sshSvc.ExecuteWithCallback(ctx, req.SshHost, sshPort, req.SshUser, req.SshPassword, req.SshKey, execCmd, req.SshInsecureHostKey, lineHandler)
	if err != nil {
		cleaned := cleanOutput(output)
		return nil, fmt.Errorf("SSH execution failed: %w\n\nOutput:\n%s", err, cleaned)
	}
	progress(80, "Agent installed, verifying startup")

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
		const maxAttempts = 15
		for attempt := 0; attempt < maxAttempts; attempt++ {
			h, err := dbService.GetOrchestratorHost(ctx, hostID)
			if err == nil && h != nil {
				hostURL = h.Host
				progress(95, "Host registered successfully")
				break
			}
			pct := 80 + (attempt+1)*15/maxAttempts
			progress(pct, "Waiting for host registration")
			time.Sleep(2 * time.Second)
		}
	}

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := apimodels.NewEventMessage(constants.EventTypeOrchestrator, "HOST_DEPLOYED", apimodels.HostDeployedEvent{
			HostID:  hostID,
			Host:    hostURL,
			Message: "Agent deployed and registered successfully",
		})
		if err := emitter.Broadcast(msg); err != nil {
			ctx.LogInfof("[Orchestrator] Failed to broadcast HOST_DEPLOYED for host %s: %v", hostID, err)
		}
	}

	return &apimodels.DeployOrchestratorHostResponse{
		HostID:  hostID,
		Host:    hostURL,
		Message: "Agent deployed and registered successfully",
	}, nil
}
