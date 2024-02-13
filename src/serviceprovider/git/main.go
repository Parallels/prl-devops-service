package git

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/serviceprovider/interfaces"
	"github.com/Parallels/pd-api-service/serviceprovider/system"

	"github.com/cjlapao/common-go/commands"
)

var globalGitService *GitService

type GitService struct {
	ctx          basecontext.ApiContext
	executable   string
	installed    bool
	dependencies []interfaces.Service
}

func Get(ctx basecontext.ApiContext) *GitService {
	if globalGitService != nil {
		return globalGitService
	}

	return New(ctx)
}

func New(ctx basecontext.ApiContext) *GitService {
	globalGitService = &GitService{
		ctx: ctx,
	}
	if globalGitService.FindPath() == "" {
		ctx.LogWarnf("Running without support for git")
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
	s.ctx.LogInfof("Getting Git executable")
	out, err := commands.ExecuteWithNoOutput("which", "git")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		s.ctx.LogWarnf("Git executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.executable = path
		s.ctx.LogInfof("Git found at: %s", s.executable)
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/git"); err == nil {
			s.executable = "/opt/homebrew/bin/git"
		} else {
			s.ctx.LogWarnf("Git executable not found")
		}
		s.ctx.LogInfof("Git found at: %s", s.executable)
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
		s.ctx.LogInfof("%s already installed", s.Name())
		return nil
	}
	// Installing service dependency
	if s.dependencies != nil {
		for _, dependency := range s.dependencies {
			if dependency == nil {
				return errors.New("Dependency is nil")
			}
			s.ctx.LogInfof("Installing dependency %s for %s", dependency.Name(), s.Name())
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

	s.ctx.LogInfof("Installing %s with command: %v", s.Name(), cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}
	s.installed = true
	return nil
}

func (s *GitService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		s.ctx.LogInfof("Uninstalling %s", s.Name())

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
				s.ctx.LogInfof("Uninstalling dependency %s for %s", dependency.Name(), s.Name())
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

func (s *GitService) Clone(ctx basecontext.ApiContext, repoURL string, owner string, localPath string) (string, error) {
	var path string
	if owner == "" || owner == "root" {
		path = filepath.Join("/tmp", localPath)
	} else {
		home, err := system.Get().GetUserHome(ctx, owner)
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, localPath)
	}

	var cmd helpers.Command
	if owner == "" {
		cmd = helpers.Command{
			Command:          s.executable,
			WorkingDirectory: path,
			Args:             make([]string, 0),
		}
	} else {
		cmd = helpers.Command{
			Command:          "sudo",
			WorkingDirectory: path,
			Args:             []string{"-u", owner, s.executable},
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := helpers.CreateDirIfNotExist(path)
		if err != nil {
			return "", err
		}
		_, err = helpers.ExecuteWithNoOutput(helpers.Command{
			Command: "chown",
			Args:    []string{"-R", owner, path},
		})
		if err != nil {
			return "", err
		}

		cmd.Args = append(cmd.Args, "clone", repoURL, path)

		ctx.LogInfof(cmd.String())
		_, err = helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			buildError := errors.NewWithCodef(400, "failed to pull repository %v, error: %v", path, err.Error())
			return "", buildError
		}
	} else {
		cmd.Args = append(cmd.Args, "pull")

		ctx.LogInfof(cmd.String())
		if err != nil {
			buildError := errors.NewWithCodef(400, "failed to pull repository %v, error: %v", path, err.Error())

			return "", buildError
		}
	}

	ctx.LogInfof("Repository %s cloned to %s", repoURL, path)
	return path, nil
}
