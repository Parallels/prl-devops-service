package ssh

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/serviceprovider/interfaces"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// lineWriter is an io.Writer that accumulates all bytes into buf and calls
// onLine for each newline-terminated line as it arrives.
type lineWriter struct {
	buf    *bytes.Buffer
	onLine func(string)
	tail   strings.Builder // partial line not yet terminated
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for _, b := range p {
		if b == '\n' {
			line := w.tail.String()
			w.tail.Reset()
			if w.onLine != nil && line != "" {
				w.onLine(line)
			}
		} else {
			w.tail.WriteByte(b)
		}
	}
	return len(p), nil
}

// flush emits any remaining partial line (no trailing newline).
func (w *lineWriter) flush() {
	if w.onLine != nil && w.tail.Len() > 0 {
		w.onLine(w.tail.String())
		w.tail.Reset()
	}
}

type SshService struct {
	dependencies []interfaces.Service
}

func New(ctx basecontext.ApiContext) *SshService {
	return &SshService{
		dependencies: make([]interfaces.Service, 0),
	}
}

func (s *SshService) Name() string {
	return "SSH Service"
}

func (s *SshService) FindPath() string {
	return "ssh"
}

func (s *SshService) Version() string {
	return "1.0.0"
}

func (s *SshService) Install(asUser, version string, flags map[string]string) error {
	return nil
}

func (s *SshService) Uninstall(asUser string, uninstallDependencies bool) error {
	return nil
}

func (s *SshService) Installed() bool {
	return true
}

func (s *SshService) Dependencies() []interfaces.Service {
	return s.dependencies
}

func (s *SshService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *SshService) Execute(ctx basecontext.ApiContext, host string, port int, user, password, key, command string, enableInsecureKey bool) (string, error) {
	return s.execute(ctx, host, port, user, password, key, command, enableInsecureKey, false, "", nil)
}

// ExecuteWithCallback runs a command over SSH without a PTY and calls onLine
// for each line of output as it arrives.  Pass nil to skip streaming.
func (s *SshService) ExecuteWithCallback(ctx basecontext.ApiContext, host string, port int, user, password, key, command string, enableInsecureKey bool, onLine func(string)) (string, error) {
	return s.execute(ctx, host, port, user, password, key, command, enableInsecureKey, false, "", onLine)
}

// ExecuteWithPty runs a command over SSH with a pseudo-terminal allocated so
// that sudo prompts are satisfied automatically.  sudoPassword is written to
// the session's stdin before the command runs; pass an empty string when the
// remote user has passwordless sudo or when key-auth is used with NOPASSWD.
// onLine is called for each line of output as it arrives; pass nil to skip.
func (s *SshService) ExecuteWithPty(ctx basecontext.ApiContext, host string, port int, user, password, key, command string, enableInsecureKey bool, sudoPassword string, onLine func(string)) (string, error) {
	return s.execute(ctx, host, port, user, password, key, command, enableInsecureKey, true, sudoPassword, onLine)
}

func (s *SshService) execute(ctx basecontext.ApiContext, host string, port int, user, password, key, command string, enableInsecureKey bool, requestPty bool, sudoPassword string, onLine func(string)) (string, error) {
	cfg := config.Get()
	sshInsecureKey := cfg.GetBoolKey(constants.ENABLE_INSECURE_KEY_SSH_ENV_VAR)
	if enableInsecureKey {
		sshInsecureKey = true
	}
	var hostKeyCallback ssh.HostKeyCallback
	var hostKeyCallbackErr error
	if sshInsecureKey {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		hostKeyCallback, hostKeyCallbackErr = knownhosts.New("~/.ssh/known_hosts")
		if hostKeyCallbackErr != nil {
			return "", fmt.Errorf("failed to load known_hosts: %v", hostKeyCallbackErr)
		}
	}
	sshCfg := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	}

	if key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			return "", fmt.Errorf("unable to parse private key: %v", err)
		}
		sshCfg.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		sshCfg.Auth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	}

	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, sshCfg)
	if err != nil {
		return "", fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	if requestPty {
		modes := ssh.TerminalModes{
			ssh.ECHO:          0,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		if err := session.RequestPty("xterm", 40, 200, modes); err != nil {
			return "", fmt.Errorf("failed to request pty: %v", err)
		}
		// Pre-seed stdin with the sudo password so that sudo prompts are
		// answered automatically without blocking the session.
		if sudoPassword != "" {
			session.Stdin = bytes.NewBufferString(sudoPassword + "\n")
		}
	}

	var buf bytes.Buffer
	lw := &lineWriter{buf: &buf, onLine: onLine}
	session.Stdout = lw
	session.Stderr = lw

	if err := session.Run(command); err != nil {
		lw.flush()
		return buf.String(), fmt.Errorf("failed to run command: %v", err)
	}

	lw.flush()
	return buf.String(), nil
}
