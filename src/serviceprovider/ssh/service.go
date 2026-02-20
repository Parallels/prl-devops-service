package ssh

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/serviceprovider/interfaces"
	"golang.org/x/crypto/ssh"
)

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

func (s *SshService) Execute(ctx basecontext.ApiContext, host string, port int, user, password, key, command string) (string, error) {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	if key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			return "", fmt.Errorf("unable to parse private key: %v", err)
		}
		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		config.Auth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	}

	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return "", fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b // Capturing stderr in the same buffer for now

	if err := session.Run(command); err != nil {
		return b.String(), fmt.Errorf("failed to run command: %v, output: %s", err, b.String())
	}

	return b.String(), nil
}
