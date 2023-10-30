package git

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/errors"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/service_provider/interfaces"
	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
)

var globalGitService *GitService
var logger = common.Logger

type GitService struct {
	executable   string
	installed    bool
	dependencies []interfaces.Service
}

func Get() *GitService {
	if globalGitService != nil {
		return globalGitService
	}

	return New()
}

func New() *GitService {
	globalGitService = &GitService{}
	if globalGitService.FindPath() == "" {
		logger.Warn("Running without support for git")
		return nil
	} else {
		globalGitService.installed = true
	}

	globalGitService.SetDependencies([]interfaces.Service{})
	return globalGitService
}

func (s *GitService) Name() string {
	return "git"
}

func (s *GitService) FindPath() string {
	logger.Info("Getting Git executable")
	out, err := commands.ExecuteWithNoOutput("which", "git")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		logger.Warn("Git executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.executable = path
		logger.Info("Git found at: %s", s.executable)
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/git"); err == nil {
			s.executable = "/opt/homebrew/bin/git"
		} else {
			logger.Warn("Git executable not found")
		}
		logger.Info("Git found at: %s", s.executable)
	}

	return s.executable
}

func (s *GitService) Version() string {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"--version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(cmd)
	if err != nil {
		return "unknown"
	}

	vParts := strings.Split(stdout, " ")
	if len(vParts) > 2 {
		return strings.TrimSpace(vParts[2])
	} else {
		return stdout
	}
}

func (s *GitService) Install(asUser, version string, flags map[string]string) error {
	if s.installed {
		logger.Info("%s already installed", s.Name())
		return nil
	}
	// Installing service dependency
	if s.dependencies != nil {
		for _, dependency := range s.dependencies {
			if dependency == nil {
				return errors.New("Dependency is nil")
			}
			logger.Info("Installing dependency %s for %s", dependency.Name(), s.Name())
			if err := dependency.Install(asUser, "latest", flags); err != nil {
				return err
			}
		}
	}

	var cmd helpers.Command
	if asUser == "" {
		cmd = helpers.Command{
			Command: "brew",
		}
	} else {
		cmd = helpers.Command{
			Command: "sudo",
			Args:    []string{"-u", asUser, "brew"},
		}
	}
	if version == "" || version == "latest" {
		cmd.Args = append(cmd.Args, "install", "packer")
	} else {
		cmd.Args = append(cmd.Args, "install", "packer@"+version)
	}

	logger.Info("Installing %s with command: %v", s.Name(), cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}
	s.installed = true
	return nil
}

func (s *GitService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		logger.Info("Uninstalling %s", s.Name())

		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
				Args:    []string{"uninstall", "git"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew", "uninstall", "git"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}

	if uninstallDependencies {
		// Uninstall service dependency
		if s.dependencies != nil {
			for _, dependency := range s.dependencies {
				if dependency == nil {
					continue
				}
				logger.Info("Uninstalling dependency %s for %s", dependency.Name(), s.Name())
				if err := dependency.Uninstall(asUser, uninstallDependencies); err != nil {
					return err
				}
			}
		}
	}

	s.installed = false
	return nil
}

func (s *GitService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *GitService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *GitService) Installed() bool {
	return s.installed && s.executable != ""
}

func (s *GitService) Clone(repoURL string, localPath string) (string, error) {
	path := fmt.Sprintf("/tmp/%s", localPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := helpers.CreateDirIfNotExist(path)
		if err != nil {
			return "", err
		}

		_, err = commands.ExecuteWithNoOutput(s.executable, "clone", repoURL, path)
		if err != nil {
			buildError := errors.NewWithCodef(500, "failed to pull repository %v, error: %v", path, err.Error())
			return "", buildError
		}
	} else {
		_, err := helpers.ExecuteWithNoOutput(helpers.Command{
			Command:          s.executable,
			Args:             []string{"pull"},
			WorkingDirectory: path,
		})
		if err != nil {
			buildError := errors.NewWithCodef(500, "failed to pull repository %v, error: %v", path, err.Error())

			return "", buildError
		}
	}

	return path, nil
}
